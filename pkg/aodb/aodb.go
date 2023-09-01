package aodb

type AnimeOfflineDatabase struct {
	License struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"license"`
	Repository string  `json:"repository"`
	LastUpdate string  `json:"lastUpdate"`
	Data       []Anime `json:"data"`
}

type Anime struct {
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

func (a *Anime) BleveType() string {
	return "anime"
}
