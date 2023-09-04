package anime

import (
	"encoding/json"
	"os"
	"strings"
)

type AnimeOfflineDatabase struct {
	License struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"license"`
	Repository string       `json:"repository"`
	LastUpdate string       `json:"lastUpdate"`
	Data       []*AODBAnime `json:"data"`
}

type AODBAnime struct {
	Sources     []string `json:"sources"`
	Title       string   `json:"title"`
	Type        string   `json:"type"`
	Episodes    int      `json:"episodes"`
	Status      string   `json:"status"`
	AnimeSeason struct {
		Season string `json:"season"`
		Year   *int   `json:"year"`
	} `json:"animeSeason"`
	Picture   string   `json:"picture"`
	Thumbnail string   `json:"thumbnail"`
	Synonyms  []string `json:"synonyms"`
	Relations []string `json:"relations"`
	Tags      []string `json:"tags"`
}

func ReadAODBFile(path string) (aodb *AnimeOfflineDatabase, err error) {
	var db AnimeOfflineDatabase

	file, err := os.Open(path)

	if err != nil {
		return
	}

	defer file.Close()

	err = json.NewDecoder(file).Decode(&db)

	if err != nil {
		return
	}

	aodb = &db
	return
}

func CreateAIDMappingFromAODB(aodb *AnimeOfflineDatabase) (aidMap map[string]*AODBAnime) {
	aidMap = make(map[string]*AODBAnime, len(aodb.Data))

	for _, anime := range aodb.Data {
		sources := anime.Sources

		for _, source := range sources {
			if strings.HasPrefix(source, "https://anidb.net/anime/") {
				aid := strings.TrimPrefix(source, "https://anidb.net/anime/")
				aidMap[aid] = anime
				break
			}
		}
	}

	return
}
