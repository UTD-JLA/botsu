package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/bot"
	"github.com/UTD-JLA/botsu/internal/bot/commands"
	"github.com/UTD-JLA/botsu/internal/data/anime"
	"github.com/UTD-JLA/botsu/internal/data/vn"
	"github.com/UTD-JLA/botsu/internal/guilds"
	"github.com/UTD-JLA/botsu/internal/users"
	"github.com/bwmarrin/discordgo"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lmittmann/tint"
)

var configPath = flag.String("config", "config.toml", "Path to config file")
var migrationSource = flag.String("migrations", "", "Path to migrations")
var enableProfiling = flag.Bool("profiling", false, "Enable profiling")

func main() {
	flag.Parse()
	config := NewConfig()

	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelWarn)

	logHandler := tint.NewHandler(os.Stdout, &tint.Options{
		Level: logLevel,
	})

	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	err := config.Load(*configPath)

	if err != nil && !os.IsNotExist(err) {
		logger.Error("Unable to load config", slog.String("err", err.Error()), slog.String("path", *configPath))
		os.Exit(1)
	}

	err = config.LoadEnv()

	if err != nil {
		logger.Error("Unable to load config from env", slog.String("err", err.Error()))
		os.Exit(1)
	}

	config.LoadDefaults()

	// Config is loaded, now we can set the log level
	logLevel.Set(config.LogLevel)
	logger.Info("Log level set", slog.String("level", config.LogLevel.String()))

	logger.Info("Reading anime database file", slog.String("path", config.AoDBPath))

	dataChan := make(chan []*anime.AniDBEntry, 1)
	aodbChan := make(chan *anime.AnimeOfflineDatabase, 1)

	_, err = os.Stat(config.AoDBPath)

	if os.IsNotExist(err) {
		logger.Info("Downloading anime offline database")

		dir := path.Dir(config.AoDBPath)

		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			logger.Error("Unable to create directory", slog.String("err", err.Error()), slog.String("path", dir))
			os.Exit(1)
		}

		err = anime.DownloadAnimeOfflineDatabase(config.AoDBPath)

		if err != nil {
			logger.Error("Unable to download anime offline database", slog.String("err", err.Error()))
			os.Exit(1)
		}
	} else if err != nil {
		logger.Error("Unable to stat anime offline database", slog.String("err", err.Error()))
		os.Exit(1)
	}

	logger.Info("Reading anidb dump file", slog.String("path", config.AniDBDumpPath))

	_, err = os.Stat(config.AniDBDumpPath)

	if os.IsNotExist(err) {
		logger.Info("Downloading anidb dump")

		dir := path.Dir(config.AniDBDumpPath)

		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			logger.Error("Unable to create directory", slog.String("err", err.Error()), slog.String("path", dir))
			os.Exit(1)
		}

		err = anime.DownloadAniDBDump(config.AniDBDumpPath)

		if err != nil {
			logger.Error("Unable to download anidb dump", slog.String("err", err.Error()))
		}
	} else if err != nil {
		logger.Error("Unable to stat anidb dump", slog.String("err", err.Error()))
		os.Exit(1)
	}

	logger.Info("Reading vndb dump file", slog.String("path", config.VNDBDumpPath))

	_, err = os.Stat(config.VNDBDumpPath)

	if os.IsNotExist(err) {
		dir := path.Dir(config.VNDBDumpPath)

		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			logger.Error("Unable to create directory", slog.String("err", err.Error()), slog.String("path", dir))
			os.Exit(1)
		}

		logger.Info("Downloading vndb dump")

		err = vn.DownloadVNDBDump(config.VNDBDumpPath)

		if err != nil {
			logger.Error("Unable to download vndb dump", slog.String("err", err.Error()))
			os.Exit(1)
		}
	} else if err != nil {
		logger.Error("Unable to stat vndb dump", slog.String("err", err.Error()))
		os.Exit(1)
	}

	searcher := anime.NewAnimeSearcher()

	// check if index exists
	if _, err = os.Stat("anime-index.bluge"); os.IsNotExist(err) {
		logger.Info("Creating anime index")

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

		err = searcher.CreateIndex("anime-index.bluge", joined)

		if err != nil {
			logger.Error("Unable to create index", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}

	logger.Info("Loading anime index")
	_, err = searcher.LoadIndex("anime-index.bluge")

	if err != nil {
		logger.Error("Unable to load index", slog.String("err", err.Error()))
		os.Exit(1)
	}

	vnSearcher := vn.NewVNSearcher()

	if _, err = os.Stat("vndb-index.bluge"); os.IsNotExist(err) {
		logger.Info("Creating vndb index")

		titles, err := vn.ReadVNDBTitlesFile(config.VNDBDumpPath + "/db/vn_titles")

		if err != nil {
			logger.Error("Unable to read vndb titles file", slog.String("err", err.Error()))
			os.Exit(1)
		}

		data, err := vn.ReadVNDBDataFile(config.VNDBDumpPath + "/db/vn")

		if err != nil {
			logger.Error("Unable to read vndb data file", slog.String("err", err.Error()))
			os.Exit(1)
		}

		vns := vn.JoinTitlesAndData(titles, data)

		err = vnSearcher.CreateIndex("vndb-index.bluge", vns)

		if err != nil {
			logger.Error("Unable to create index", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}

	_, err = vnSearcher.LoadIndex("vndb-index.bluge")

	if err != nil {
		logger.Error("Unable to load index", slog.String("err", err.Error()))
		os.Exit(1)
	}

	logger.Info("Connecting to database")

	migrationURL := config.Database.ConnectionURL()
	q := migrationURL.Query()
	q.Add("sslmode", "disable")
	migrationURL.RawQuery = q.Encode()

	if *migrationSource != "" {
		logger.Info("Running migrations", slog.String("source", *migrationSource))

		m, err := migrate.New(*migrationSource, migrationURL.String())

		if err != nil {
			logger.Error("Unable to create migration", slog.String("err", err.Error()))
			os.Exit(1)
		}

		err = m.Up()

		if err != nil {
			logger.Error("Unable to run migrations", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}

	pool, err := pgxpool.New(context.Background(), config.Database.ConnectionString())

	if err != nil {
		logger.Error("Unable to connect to database", slog.String("err", err.Error()))
		os.Exit(1)
	}

	defer pool.Close()

	activityRepo := activities.NewActivityRepository(pool)
	userRepo := users.NewUserRepository(pool)
	guildRepo := guilds.NewGuildRepository(pool)

	bot := bot.NewBot(logger.WithGroup("bot"), guildRepo)

	bot.AddCommand(commands.LogCommandData, commands.NewLogCommand(activityRepo, userRepo, guildRepo, searcher, vnSearcher))
	bot.AddCommand(commands.ConfigCommandData, commands.NewConfigCommand(userRepo, activityRepo))
	bot.AddCommand(commands.HistoryCommandData, commands.NewHistoryCommand(activityRepo))
	bot.AddCommand(commands.LeaderboardCommandData, commands.NewLeaderboardCommand(activityRepo, userRepo, guildRepo))
	bot.AddCommand(commands.UndoCommandData, commands.NewUndoCommand(activityRepo))
	bot.AddCommand(commands.ChartCommandData, commands.NewChartCommand(activityRepo, userRepo, guildRepo))
	bot.AddCommand(commands.GuildConfigCommandData, commands.NewGuildConfigCommand(guildRepo))
	bot.AddCommand(commands.ExportCommandData, commands.NewExportCommand(activityRepo))
	bot.AddCommand(commands.ImportCommandData, commands.NewImportCommand(activityRepo))

	logger.Info("Starting bot")

	intents := discordgo.IntentsNone

	if config.UseMembersIntent {
		intents = discordgo.IntentsGuildMembers
	}

	err = bot.Login(config.Token, intents)

	if err != nil {
		logger.Error("Unable to login", slog.String("err", err.Error()))
		os.Exit(1)
	}

	defer bot.Close()

	// Wait here until CTRL-C or other term signal is received.
	logger.Info("Setup completed, press CTRL-C to exit")

	if *enableProfiling {
		go func() {
			err := http.ListenAndServe("localhost:6060", nil)
			logger.Warn("Profiling server exited", slog.String("err", err.Error()))
		}()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-c
}
