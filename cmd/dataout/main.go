package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"

	"github.com/UTD-JLA/botsu/internal/data/anime"
)

var (
	anidbDumpFile = flag.String("anidb-dump-file", "anime-titles.dat", "Path to the AniDB dump file")
	aodbFile      = flag.String("aodb-file", "anime-offline-database.json", "Path to the Anime Offline Database file")
	output        = flag.String("output", "output.json", "Path to the output file")
	pretty        = flag.Bool("pretty", false, "Pretty print the output")
)

func main() {
	flag.Parse()

	dataChan := make(chan []*anime.AniDBEntry, 1)
	aodbChan := make(chan *anime.AnimeOfflineDatabase, 1)

	go func() {
		data, err := anime.ReadAniDBDump(*anidbDumpFile)

		if err != nil {
			panic(err)
		}

		dataChan <- data
	}()

	go func() {
		aodb, err := anime.ReadAODBFile(*aodbFile)

		if err != nil {
			panic(err)
		}

		aodbChan <- aodb
	}()

	mappings := anime.CreateAIDMappingFromAODB(<-aodbChan)
	joined := anime.JoinAniDBAndAODB(mappings, <-dataChan)

	outFile, err := os.Create(*output)

	if err != nil {
		panic(err)
	}

	defer outFile.Close()

	encoder := json.NewEncoder(outFile)

	if *pretty {
		encoder.SetIndent("", "  ")
	}

	err = encoder.Encode(joined)

	if err != nil {
		panic(err)
	}

	// TESTING
	searcher := anime.NewAnimeSearcher(joined)
	_, err = searcher.LoadIndex("test")

	if err != nil {
		panic(err)
	}

	results, err := searcher.Search(context.Background(), "love live", 25)

	if err != nil {
		panic(err)
	}

	for _, result := range results {
		displayTitle := result.Anime.PrimaryTitle

		switch result.Field {
		case "japaneseOfficialTitle":
			displayTitle = result.Anime.JapaneseOfficialTitle
		case "xJatOfficialTitle":
			displayTitle = result.Anime.XJatOfficialTitle
		case "englishOfficialTitle":
			displayTitle = result.Anime.EnglishOfficialTitle
		}

		println(displayTitle)
		println(result.Anime.ID, result.Field, result.Score)
	}
}
