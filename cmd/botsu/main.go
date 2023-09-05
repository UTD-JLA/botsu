package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/bot"
	"github.com/UTD-JLA/botsu/internal/bot/commands"
	"github.com/UTD-JLA/botsu/internal/data/anime"
	"github.com/UTD-JLA/botsu/internal/guilds"
	"github.com/UTD-JLA/botsu/internal/users"
	"github.com/bwmarrin/discordgo"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"

	"github.com/jackc/pgx/v5/pgxpool"
)

var configPath = flag.String("config", "config.toml", "Path to config file")
var migrationSource = flag.String("migrations", "", "Path to migrations")

func main() {
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config := NewConfig()

	err := config.Load(*configPath)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Reading anime database file")

	dataChan := make(chan []*anime.AniDBEntry, 1)
	aodbChan := make(chan *anime.AnimeOfflineDatabase, 1)

	go func() {
		data, err := anime.ReadAniDBDump(config.AniDBDumpPath)

		if err != nil {
			panic(err)
		}

		dataChan <- data
	}()

	go func() {
		aodb, err := anime.ReadAODBFile(config.AoDBPath)

		if err != nil {
			panic(err)
		}

		aodbChan <- aodb
	}()

	mappings := anime.CreateAIDMappingFromAODB(<-aodbChan)
	joined := anime.JoinAniDBAndAODB(mappings, <-dataChan)

	searcher := anime.NewAnimeSearcher(joined)

	// check if index exists
	if _, err = os.Stat("anime-index.bluge"); err != nil {
		log.Println("Creating index")
		err = searcher.CreateIndex("anime-index.bluge")

		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Loading index")
	_, err = searcher.LoadIndex("anime-index.bluge")

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connecting to database")

	if *migrationSource != "" {
		log.Println("Running migrations")

		m, err := migrate.New(
			*migrationSource,
			config.Database.ConnectionString()+"?sslmode=disable")

		if err != nil {
			log.Fatal(err)
		}

		err = m.Up()

		if err != nil {
			log.Fatal(err)
		}
	}

	pool, err := pgxpool.New(context.Background(), config.Database.ConnectionString())

	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()

	activityRepo := activities.NewActivityRepository(pool)
	userRepo := users.NewUserRepository(pool)
	guildRepo := guilds.NewGuildRepository(pool)

	bot := bot.NewBot(guildRepo)

	bot.AddCommand(commands.LogCommandData, commands.NewLogCommand(activityRepo, userRepo, guildRepo, searcher))
	bot.AddCommand(commands.ConfigCommandData, commands.NewConfigCommand(userRepo))
	bot.AddCommand(commands.HistoryCommandData, commands.NewHistoryCommand(activityRepo))
	bot.AddCommand(commands.LeaderboardCommandData, commands.NewLeaderboardCommand(activityRepo, userRepo, guildRepo))
	bot.AddCommand(commands.UndoCommandData, commands.NewUndoCommand(activityRepo))
	bot.AddCommand(commands.ChartCommandData, commands.NewChartCommand(activityRepo, userRepo, guildRepo))
	bot.AddCommand(commands.GuildConfigCommandData, commands.NewGuildConfigCommand(guildRepo))
	bot.AddCommand(commands.ExportCommandData, commands.NewExportCommand(activityRepo))

	log.Println("Logging in")

	intents := discordgo.IntentsNone

	if config.UseMembersIntent {
		intents = discordgo.IntentsGuildMembers
	}

	err = bot.Login(config.Token, intents)

	if err != nil {
		log.Fatal(err)
	}

	defer bot.Close()

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running. Press CTRL-C to exit")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-c
}
