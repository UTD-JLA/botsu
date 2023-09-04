package anime

type Anime struct {
	ID                    string   `json:"id"`
	PrimaryTitle          string   `json:"primaryTitle"`
	XJatOfficialTitle     string   `json:"xJatOfficialTitle"`
	XJatSynonyms          []string `json:"xJatSynonyms"`
	JapaneseOfficialTitle string   `json:"japaneseOfficialTitle"`
	JapaneseSynonyms      []string `json:"japaneseSynonyms"`
	EnglishOfficialTitle  string   `json:"englishOfficialTitle"`
	EnglishSynonyms       []string `json:"englishSynonyms"`
	Sources               []string `json:"sources"`
	Type                  string   `json:"type"`
	Episodes              int      `json:"episodes"`
	Status                string   `json:"status"`
	AnimeSeason           struct {
		Season string `json:"season"`
		Year   *int   `json:"year"`
	} `json:"animeSeason"`
	Picture   string   `json:"picture"`
	Thumbnail string   `json:"thumbnail"`
	Relations []string `json:"relations"`
	Tags      []string `json:"tags"`
}

func JoinAniDBAndAODB(mappings map[string]*AODBAnime, titleEntries []*AniDBEntry) (joined []*Anime) {
	joined = make([]*Anime, 0, len(mappings))

	for _, entry := range titleEntries {
		aid := entry.AID

		if anime, ok := mappings[aid]; ok {
			joined = append(joined, &Anime{
				ID:                    aid,
				PrimaryTitle:          entry.PrimaryTitle,
				XJatOfficialTitle:     entry.XJatOfficialTitle,
				XJatSynonyms:          entry.XJatSynonyms,
				JapaneseOfficialTitle: entry.JapaneseOfficialTitle,
				JapaneseSynonyms:      entry.JapaneseSynonyms,
				EnglishOfficialTitle:  entry.EnglishOfficialTitle,
				EnglishSynonyms:       entry.EnglishSynonyms,
				Sources:               anime.Sources,
				Type:                  anime.Type,
				Episodes:              anime.Episodes,
				Status:                anime.Status,
				AnimeSeason:           anime.AnimeSeason,
				Picture:               anime.Picture,
				Thumbnail:             anime.Thumbnail,
				Relations:             anime.Relations,
				Tags:                  anime.Tags,
			})
		}
	}

	return
}
