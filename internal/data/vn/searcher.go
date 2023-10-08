package vn

import (
	"context"
	"errors"
	"fmt"

	"github.com/blugelabs/bluge"
)

var ErrReaderNotInitialized = errors.New("reader not initialized")

const SearchFieldEnglishTitle = "englishTitle"
const SearchFieldJapaneseTitle = "japaneseTitle"
const SearchFieldRomaji = "romajiTitle"

type Match struct {
	VN    *VisualNovel
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

func (o *orderedMatches) addMatch(vn *VisualNovel, score float64, field string) {
	if len(o.matches) < o.maxLen {
		o.matches = append(o.matches, &Match{
			VN:    vn,
			Score: score,
			Field: field,
		})
	} else {
		i := o.findInsertIndex(score)

		if i == o.maxLen {
			return
		}

		o.matches[i] = &Match{
			VN:    vn,
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

type VNSearcher struct {
	reader *bluge.Reader
}

func NewVNSearcher() *VNSearcher {
	return &VNSearcher{}
}

func (s *VNSearcher) GetVN(ctx context.Context, id string) (*VisualNovel, error) {
	if s.reader == nil {
		return nil, ErrReaderNotInitialized
	}

	// search by id
	idQuery := bluge.NewMatchQuery(id).
		SetField("id")

	results, err := s.reader.Search(ctx, bluge.NewTopNSearch(1, idQuery))

	if err != nil {
		return nil, err
	}

	next, err := results.Next()

	if err != nil {
		return nil, err
	}

	if next != nil {
		vn := &VisualNovel{}

		err = next.VisitStoredFields(func(field string, value []byte) bool {
			if field == "id" {
				vn.ID = string(value)
			} else if field == "englishTitle" {
				vn.EnglishTitle = string(value)
			} else if field == "japaneseTitle" {
				vn.JapaneseTitle = string(value)
			} else if field == "romajiTitle" {
				vn.RomajiTitle = string(value)
			} else if field == "image" {
				vn.Image = string(value)
			}

			return true
		})

		if err != nil {
			return nil, err
		}

		return vn, nil
	}

	return nil, fmt.Errorf("vn with id %s not found", id)
}

func (s *VNSearcher) Search(ctx context.Context, queryStr string, limit int) ([]*Match, error) {
	if s.reader == nil {
		return nil, ErrReaderNotInitialized
	}

	japaneseQuery := bluge.NewMatchQuery(queryStr).
		SetField(SearchFieldJapaneseTitle)

	englishQuery := bluge.NewMatchQuery(queryStr).
		SetField(SearchFieldEnglishTitle)

	romajiQuery := bluge.NewMatchQuery(queryStr).
		SetField(SearchFieldRomaji)

	queries := []*bluge.MatchQuery{
		japaneseQuery,
		englishQuery,
		romajiQuery,
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
			vn := &VisualNovel{}
			err = next.VisitStoredFields(func(field string, value []byte) bool {
				fmt.Println(field, string(value))

				if field == "id" {
					vn.ID = string(value)
				} else if field == "englishTitle" {
					vn.EnglishTitle = string(value)
				} else if field == "japaneseTitle" {
					vn.JapaneseTitle = string(value)
				} else if field == "romajiTitle" {
					vn.RomajiTitle = string(value)
				} else if field == "image" {
					vn.Image = string(value)
				}

				return true
			})
			if err != nil {
				return nil, err
			}

			matches.addMatch(vn, next.Score, query.Field())
			next, err = dmi.Next()
		}
		if err != nil {
			return nil, err
		}
	}

	return matches.matches, nil
}

func (s *VNSearcher) LoadIndex(path string) (*bluge.Reader, error) {
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

func (s *VNSearcher) CreateIndex(path string, visualNovels []*VisualNovel) error {
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

	for _, entry := range visualNovels {
		doc := bluge.NewDocument(entry.ID).
			AddField(bluge.NewKeywordField("id", entry.ID).StoreValue()).
			AddField(bluge.NewTextField("englishTitle", entry.EnglishTitle).StoreValue()).
			AddField(bluge.NewTextField("japaneseTitle", entry.JapaneseTitle).StoreValue()).
			AddField(bluge.NewTextField("romajiTitle", entry.RomajiTitle).StoreValue()).
			AddField(bluge.NewStoredOnlyField("image", []byte(entry.Image)))

		batch.Update(doc.ID(), doc)
	}

	return index.Batch(batch)
}
