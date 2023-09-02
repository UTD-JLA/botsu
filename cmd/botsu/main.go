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
	"github.com/UTD-JLA/botsu/internal/commands"
	"github.com/UTD-JLA/botsu/internal/users"
	"github.com/UTD-JLA/botsu/pkg/aodb"

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

	err = aodb.ReadDatabaseFile(config.AodbPath)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating index")

	err = aodb.CreateIndex()

	if err != nil {
		log.Fatal(err)
	}

	bot := bot.NewBot()

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

	bot.AddCommand(commands.LogCommandData, commands.NewLogCommand(activityRepo, userRepo))
	bot.AddCommand(commands.ConfigCommandData, commands.NewConfigCommand(userRepo))
	bot.AddCommand(commands.HistoryCommandData, commands.NewHistoryCommand(activityRepo))
	bot.AddCommand(commands.LeaderboardCommandData, commands.NewLeaderboardCommand(activityRepo, userRepo))
	bot.AddCommand(commands.UndoCommandData, commands.NewUndoCommand(activityRepo))

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
