package anime

type Anime struct {
	ID                    string   `json:"id"`
	PrimaryTitle          string   `json:"primaryTitle"`
	RomajiOfficialTitle   string   `json:"romajiOfficialTitle"`
	JapaneseOfficialTitle string   `json:"japaneseOfficialTitle"`
	Sources               []string `json:"sources"`
	Picture               string   `json:"picture"`
	Thumbnail             string   `json:"thumbnail"`
	Tags                  []string `json:"tags"`
	EnglishOfficialTitle  string   `json:"englishOfficialTitle"`
	//XJatSynonyms          []string `json:"xJatSynonyms"`
	//JapaneseSynonyms      []string `json:"japaneseSynonyms"`
	//EnglishSynonyms       []string `json:"englishSynonyms"`
	//Type        string   `json:"type"`
	//Episodes    int      `json:"episodes"`
	//Status      string   `json:"status"`
	//AnimeSeason struct {
	//	Season string `json:"season"`
	//	Year   *int   `json:"year"`
	//} `json:"animeSeason"`
	//Relations []string `json:"relations"`
}

func JoinAniDBAndAODB(mappings map[string]*AODBAnime, titleEntries []*AniDBEntry) (joined []*Anime) {
	joined = make([]*Anime, 0, len(mappings))

	for _, entry := range titleEntries {
		aid := entry.AID

		if anime, ok := mappings[aid]; ok {
			joined = append(joined, &Anime{
				ID:                    aid,
				PrimaryTitle:          entry.PrimaryTitle,
				RomajiOfficialTitle:   entry.XJatOfficialTitle,
				JapaneseOfficialTitle: entry.JapaneseOfficialTitle,
				Sources:               anime.Sources,
				Picture:               anime.Picture,
				Thumbnail:             anime.Thumbnail,
				Tags:                  anime.Tags,
			})
		}
	}

	return
}
