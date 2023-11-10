package anime

import (
	"encoding/json"
	"fmt"

	"github.com/blugelabs/bluge"
)

const (
	SearchFieldPrimaryTitle          = "primary_title"
	SearchFieldRomajiOfficialTitle   = "romaji_official_title"
	SearchFieldJapaneseOfficialTitle = "japanese_official_title"
	SearchFieldEnglishOfficialTitle  = "english_official_title"
)

var SearchFields = []string{
	SearchFieldPrimaryTitle,
	SearchFieldRomajiOfficialTitle,
	SearchFieldJapaneseOfficialTitle,
	SearchFieldEnglishOfficialTitle,
}

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
}

func (a Anime) Marshal() (*bluge.Document, error) {
	doc := bluge.NewDocument(a.ID)

	doc.AddField(bluge.NewTextField(SearchFieldPrimaryTitle, a.PrimaryTitle).StoreValue())
	doc.AddField(bluge.NewTextField(SearchFieldRomajiOfficialTitle, a.RomajiOfficialTitle).StoreValue())
	doc.AddField(bluge.NewTextField(SearchFieldJapaneseOfficialTitle, a.JapaneseOfficialTitle).StoreValue())
	doc.AddField(bluge.NewTextField(SearchFieldEnglishOfficialTitle, a.EnglishOfficialTitle).StoreValue())
	doc.AddField(bluge.NewStoredOnlyField("picture", []byte(a.Picture)))
	doc.AddField(bluge.NewStoredOnlyField("thumbnail", []byte(a.Thumbnail)))

	if sourcesBytes, err := json.Marshal(a.Sources); err == nil {
		doc.AddField(bluge.NewStoredOnlyField("sources", sourcesBytes))
	} else {
		return nil, fmt.Errorf("unable to marshal sources: %w", err)
	}

	if tagsBytes, err := json.Marshal(a.Tags); err == nil {
		doc.AddField(bluge.NewStoredOnlyField("tags", tagsBytes))
	} else {
		return nil, fmt.Errorf("unable to marshal tags: %w", err)
	}

	return doc, nil
}

func (a *Anime) Unmarshal(fields map[string][]byte) error {
	a.ID = string(fields["_id"])
	a.PrimaryTitle = string(fields[SearchFieldPrimaryTitle])
	a.RomajiOfficialTitle = string(fields[SearchFieldRomajiOfficialTitle])
	a.JapaneseOfficialTitle = string(fields[SearchFieldJapaneseOfficialTitle])
	a.EnglishOfficialTitle = string(fields[SearchFieldEnglishOfficialTitle])
	a.Picture = string(fields["picture"])
	a.Thumbnail = string(fields["thumbnail"])

	if err := json.Unmarshal(fields["sources"], &a.Sources); err != nil {
		return fmt.Errorf("unable to unmarshal sources: %w: %s", err, string(fields["sources"]))
	}

	if err := json.Unmarshal(fields["tags"], &a.Tags); err != nil {
		return fmt.Errorf("unable to unmarshal tags: %w: %s", err, string(fields["tags"]))
	}

	return nil
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
