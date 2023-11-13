package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"
	"time"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/bot"
	"github.com/UTD-JLA/botsu/internal/bot/commands"
	"github.com/UTD-JLA/botsu/internal/data"
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

var (
	configPath      = flag.String("config", "config.toml", "Path to config file")
	migrationSource = flag.String("migrations", "", "Path to migrations")
	enableProfiling = flag.Bool("profiling", false, "Enable profiling")
)

const (
	staleAODBThreshold  = 7 * 24 * time.Hour
	staleAniDBThreshold = 7 * 24 * time.Hour
	staleVNDBThreshold  = 7 * 24 * time.Hour
)

type dataSource struct {
	path           string
	staleThreshold time.Duration
	downloadFunc   func(string) error
}

func ensureDataSourceExists(logger *slog.Logger, source dataSource) (err error) {
	logger.Info("Checking data source")

	stat, err := os.Stat(source.path)

	if os.IsNotExist(err) {
		logger.Info("Downloading source")

		dir := path.Dir(source.path)

		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			logger.Error("Unable to create directory", slog.String("err", err.Error()))
			return
		}

		err = source.downloadFunc(source.path)

		if err != nil {
			logger.Error("Unable to download data", slog.String("err", err.Error()))
			return
		}
	} else if err != nil {
		logger.Error("Unable to stat file", slog.String("err", err.Error()))
		return
	} else if time.Since(stat.ModTime()) > source.staleThreshold {
		logger.Warn("Data is stale, consider updating it", slog.Duration("age", time.Since(stat.ModTime())))
	}

	return
}

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

	discordgo.Logger = func(msgL, _caller int, format string, a ...interface{}) {
		msg := fmt.Sprintf("[DGO] "+format, a...)

		switch msgL {
		case discordgo.LogError:
			logger.Error(msg)
		case discordgo.LogWarning:
			logger.Warn(msg)
		case discordgo.LogInformational:
			logger.Info(msg)
		case discordgo.LogDebug:
			logger.Debug(msg)
		}
	}

	var sources = map[string]dataSource{
		"aodb": {
			path:           config.AoDBPath,
			downloadFunc:   anime.DownloadAnimeOfflineDatabase,
			staleThreshold: staleAODBThreshold,
		},
		"anidb": {
			path:           config.AniDBDumpPath,
			downloadFunc:   anime.DownloadAniDBDump,
			staleThreshold: staleAniDBThreshold,
		},
		"vndb": {
			path:           config.VNDBDumpPath,
			downloadFunc:   vn.DownloadVNDBDump,
			staleThreshold: staleVNDBThreshold,
		},
	}

	errs := make(chan error, len(sources))

	for name, source := range sources {
		sourceLogger := logger.WithGroup(name).With(slog.String("path", source.path))

		go func(src dataSource) {
			errs <- ensureDataSourceExists(sourceLogger, src)
		}(source)
	}

	errsSlice := make([]error, 0, len(sources))

	for range sources {
		if err := <-errs; err != nil {
			errsSlice = append(errsSlice, err)
		}
	}

	if err = errors.Join(errsSlice...); err != nil {
		logger.Error("Encountered error(s) while ensuring data sources exist, exiting")
		os.Exit(1)
	}

	animeStore := data.NewDocumentStore[anime.Anime](
		context.Background(),
		data.NewDefaultConfig("data/.index/anime.bluge").WithSearchFields(anime.SearchFields...),
	)

	// check if index exists
	if _, err = os.Stat("data/.index/anime.bluge"); os.IsNotExist(err) {
		dataChan := make(chan []*anime.AniDBEntry, 1)
		aodbChan := make(chan *anime.AnimeOfflineDatabase, 1)

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

		for _, entry := range joined {
			if err = animeStore.Store(entry); err != nil {
				logger.Error("Unable to store anime", slog.String("err", err.Error()))
				os.Exit(1)
			}
		}
	}

	animeStore.Flush()

	vnStore := data.NewDocumentStore[vn.VisualNovel](
		context.Background(),
		data.NewDefaultConfig("data/.index/vn.bluge").WithSearchFields(vn.SearchFields...),
	)

	if _, err = os.Stat("data/.index/vn.bluge"); os.IsNotExist(err) {
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

		for _, entry := range vns {
			if err = vnStore.Store(entry); err != nil {
				logger.Error("Unable to store vn", slog.String("err", err.Error()))
				os.Exit(1)
			}
		}
	}

	vnStore.Flush()

	// Force GC before continuing with the rest of the setup
	runtime.GC()

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
	bot.SetNoPanic(config.NoPanic)

	bot.AddCommand(commands.LogCommandData, commands.NewLogCommand(activityRepo, userRepo, guildRepo, animeStore, vnStore))
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
		logger.Warn("Profiling server enabled")

		go func() {
			err := http.ListenAndServe("localhost:6060", nil)
			logger.Warn("Profiling server exited", slog.String("err", err.Error()))
		}()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-c
}
