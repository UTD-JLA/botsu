package anime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/blugelabs/bluge"
)

var ErrReaderNotInitialized = errors.New("reader not initialized")

const SearchFieldEnglishTitle = "englishOfficialTitle"
const SearchFieldJapaneseTitle = "japaneseOfficialTitle"
const SearchFieldXJatTitle = "xJatOfficialTitle"
const SearchFieldPrimaryTitle = "primaryTitle"

type Match struct {
	Anime *Anime
	Score float64
	Field string
}

type orderedMatches struct {
	matches []*Match
	maxLen  int
}

func newOrderedMatches(maxLen int) *orderedMatches {
	return &orderedMatches{
		matches: make([]*Match, 0, maxLen),
		maxLen:  maxLen,
	}
}

func (o *orderedMatches) addMatch(anime *Anime, score float64, field string) {
	if len(o.matches) < o.maxLen {
		o.matches = append(o.matches, &Match{
			Anime: anime,
			Score: score,
			Field: field,
		})
	} else {
		i := o.findInsertIndex(score)

		if i == o.maxLen {
			return
		}

		o.matches[i] = &Match{
			Anime: anime,
			Score: score,
			Field: field,
		}
	}
}

func (o *orderedMatches) findInsertIndex(score float64) int {
	i := 0
	maxLen := len(o.matches)

	for i < maxLen {
		if score > o.matches[i].Score {
			return i
		}
		i++
	}

	return maxLen
}

type AnimeSearcher struct {
	reader *bluge.Reader
}

func NewAnimeSearcher() *AnimeSearcher {
	return &AnimeSearcher{}
}

func (s *AnimeSearcher) GetAnime(ctx context.Context, id string) (*Anime, error) {
	if s.reader == nil {
		return nil, ErrReaderNotInitialized
	}

	searchRequest := bluge.NewTopNSearch(1, bluge.NewMatchQuery(id).SetField("id"))

	dmi, err := s.reader.Search(ctx, searchRequest)

	if err != nil {
		return nil, err
	}

	next, err := dmi.Next()

	if err != nil {
		return nil, err
	}

	if next != nil {
		anime := &Anime{}

		err = next.VisitStoredFields(func(field string, value []byte) bool {
			if field == "id" {
				anime.ID = string(value)
			} else if field == "primaryTitle" {
				anime.PrimaryTitle = string(value)
			} else if field == "xJatOfficialTitle" {
				anime.RomajiOfficialTitle = string(value)
			} else if field == "japaneseOfficialTitle" {
				anime.JapaneseOfficialTitle = string(value)
			} else if field == "englishOfficialTitle" {
				anime.EnglishOfficialTitle = string(value)
			} else if field == "picture" {
				anime.Picture = string(value)
			} else if field == "thumbnail" {
				anime.Thumbnail = string(value)
			} else if field == "sources" {
				err := json.Unmarshal(value, &anime.Sources)

				if err != nil {
					panic(err)
				}
			} else if field == "tags" {
				err := json.Unmarshal(value, &anime.Tags)

				if err != nil {
					panic(err)
				}
			}

			return true
		})

		if err != nil {
			return nil, err
		}

		return anime, nil
	}

	return nil, fmt.Errorf("anime with id %s not found", id)
}

func (s *AnimeSearcher) Search(ctx context.Context, queryStr string, limit int) ([]*Match, error) {
	if s.reader == nil {
		return nil, ErrReaderNotInitialized
	}

	// search all fields
	ptQuery := bluge.NewMatchQuery(queryStr).
		SetField(SearchFieldPrimaryTitle)

	xJatOfficialTitleQuery := bluge.NewMatchQuery(queryStr).
		SetField(SearchFieldXJatTitle)

	japaneseOfficialTitleQuery := bluge.NewMatchQuery(queryStr).
		SetField(SearchFieldJapaneseTitle)

	englishOfficialTitleQuery := bluge.NewMatchQuery(queryStr).
		SetField(SearchFieldEnglishTitle)

	queries := []*bluge.MatchQuery{
		ptQuery,
		xJatOfficialTitleQuery,
		japaneseOfficialTitleQuery,
		englishOfficialTitleQuery,
	}

	matches := newOrderedMatches(limit)

	for _, query := range queries {
		searchRequest := bluge.NewTopNSearch(limit, query)

		dmi, err := s.reader.Search(ctx, searchRequest)

		if err != nil {
			return nil, err
		}

		next, err := dmi.Next()
		for err == nil && next != nil {
			anime := &Anime{}

			err = next.VisitStoredFields(func(field string, value []byte) bool {
				if field == "id" {
					anime.ID = string(value)
				} else if field == "primaryTitle" {
					anime.PrimaryTitle = string(value)
				} else if field == "xJatOfficialTitle" {
					anime.RomajiOfficialTitle = string(value)
				} else if field == "japaneseOfficialTitle" {
					anime.JapaneseOfficialTitle = string(value)
				} else if field == "englishOfficialTitle" {
					anime.EnglishOfficialTitle = string(value)
				} else if field == "picture" {
					anime.Picture = string(value)
				} else if field == "thumbnail" {
					anime.Thumbnail = string(value)
				} else if field == "sources" {
					err := json.Unmarshal(value, &anime.Sources)

					if err != nil {
						panic(err)
					}
				} else if field == "tags" {
					err := json.Unmarshal(value, &anime.Tags)

					if err != nil {
						panic(err)
					}
				}

				return true
			})

			if err != nil {
				return nil, err
			}

			matches.addMatch(anime, next.Score, query.Field())
			next, err = dmi.Next()
		}
		if err != nil {
			return nil, err
		}
	}

	return matches.matches, nil
}

func (s *AnimeSearcher) LoadIndex(path string) (*bluge.Reader, error) {
	config := bluge.DefaultConfig(path)

	if s.reader != nil {
		return s.reader, nil
	}

	index, err := bluge.OpenReader(config)

	if err != nil {
		return nil, err
	}

	s.reader = index

	return index, nil
}

func (s *AnimeSearcher) CreateIndex(path string, anime []*Anime) error {
	config := bluge.DefaultConfig(path)

	index, err := bluge.OpenWriter(config)

	if err != nil {
		return err
	}

	defer func() {
		err := index.Close()

		if err != nil {
			panic(err)
		}
	}()

	batch := bluge.NewBatch()

	for _, entry := range anime {
		sourcesBytes, err := json.Marshal(entry.Sources)

		if err != nil {
			return err
		}

		tagsBytes, err := json.Marshal(entry.Tags)

		if err != nil {
			return err
		}

		doc := bluge.NewDocument(entry.ID).
			AddField(bluge.NewKeywordField("id", entry.ID).StoreValue()).
			AddField(bluge.NewTextField(SearchFieldPrimaryTitle, entry.PrimaryTitle).StoreValue()).
			AddField(bluge.NewTextField(SearchFieldXJatTitle, entry.RomajiOfficialTitle).StoreValue()).
			AddField(bluge.NewTextField(SearchFieldJapaneseTitle, entry.JapaneseOfficialTitle).StoreValue()).
			AddField(bluge.NewTextField(SearchFieldEnglishTitle, entry.EnglishOfficialTitle).StoreValue()).
			AddField(bluge.NewStoredOnlyField("picture", []byte(entry.Picture))).
			AddField(bluge.NewStoredOnlyField("thumbnail", []byte(entry.Thumbnail))).
			AddField(bluge.NewStoredOnlyField("sources", sourcesBytes)).
			AddField(bluge.NewStoredOnlyField("tags", tagsBytes))

		batch.Update(doc.ID(), doc)
	}

	return index.Batch(batch)
}
