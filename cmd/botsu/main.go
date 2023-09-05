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

	"github.com/jackc/pgx/v5/pgxpool"
)

var configPath = flag.String("config", "config.toml", "Path to config file")

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

	pool, err := pgxpool.New(context.Background(), config.Database.ConnectionString())

	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()

	if err != nil {
		log.Fatal(err)
	}

	activityRepo := activities.NewActivityRepository(pool)
	userRepo := users.NewUserRepository(pool)
	guildRepo := guilds.NewGuildRepository(pool)

	bot := bot.NewBot()

	bot.AddCommand(commands.LogCommandData, commands.NewLogCommand(activityRepo, userRepo, guildRepo, searcher))
	bot.AddCommand(commands.ConfigCommandData, commands.NewConfigCommand(userRepo))
	bot.AddCommand(commands.HistoryCommandData, commands.NewHistoryCommand(activityRepo))
	bot.AddCommand(commands.LeaderboardCommandData, commands.NewLeaderboardCommand(activityRepo, userRepo, guildRepo))
	bot.AddCommand(commands.UndoCommandData, commands.NewUndoCommand(activityRepo))
	bot.AddCommand(commands.ChartCommandData, commands.NewChartCommand(activityRepo, userRepo, guildRepo))
	bot.AddCommand(commands.GuildConfigCommandData, commands.NewGuildConfigCommand(guildRepo))

	log.Println("Logging in")

	err = bot.Login(config.Token)

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
