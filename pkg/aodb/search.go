package aodb

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/blevesearch/bleve/v2"
)

var db *AnimeOfflineDatabase
var index bleve.Index

func ReadDatabaseFile(path string) (err error) {
	file, err := os.Open(path)

	if err != nil {
		return err
	}

	defer file.Close()

	err = json.NewDecoder(file).Decode(&db)

	if err != nil {
		return err
	}

	return nil
}

func CreateIndex() (err error) {
	// only index the title and synonyms
	documentMapping := bleve.NewDocumentMapping()
	documentMapping.AddFieldMappingsAt("title", bleve.NewTextFieldMapping())
	documentMapping.AddFieldMappingsAt("synonyms", bleve.NewTextFieldMapping())

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("anime", documentMapping)

	mapping := bleve.NewIndexMapping()
	mapping.AddDocumentMapping("anime", documentMapping)

	if _, err = os.Stat("anime-offline-database.bleve"); err == nil {
		index, err = bleve.Open("anime-offline-database.bleve")
		return
	}

	index, err = bleve.New("anime-offline-database.bleve", mapping)

	if err != nil {
		return
	}

	fmt.Printf("Indexing database... (%d entries)\n", len(db.Data))

	batch := index.NewBatch()

	for i, anime := range db.Data {
		batch.Index(strconv.Itoa(i), &anime)
	}

	err = index.Batch(batch)

	return

}

type SearchResult struct {
	Anime *Anime
	Score float64
	Index int
}

func GetEntry(index int) (anime *Anime, err error) {
	if index < 0 || index >= len(db.Data) {
		return &Anime{}, fmt.Errorf("index out of bounds")
	}

	return &db.Data[index], nil
}

func Search(query string) (results []SearchResult, err error) {
	q := bleve.NewMatchQuery(query)
	search := bleve.NewSearchRequest(q)

	search.SortBy([]string{"-_score"})
	search.Size = 25

	searchResults, err := index.Search(search)

	if err != nil {
		return nil, err
	}

	for _, result := range searchResults.Hits {
		id, _ := strconv.Atoi(result.ID)
		anime := &db.Data[id]
		results = append(results, SearchResult{
			Anime: anime,
			Score: result.Score,
			Index: id,
		})
	}

	return results, nil
}
