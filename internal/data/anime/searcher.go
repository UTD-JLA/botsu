package anime

import (
	"context"
	"errors"
	"fmt"

	"github.com/blugelabs/bluge"
)

var ErrReaderNotInitialized = errors.New("reader not initialized")

var SearchFieldEnglishTitle = "englishOfficialTitle"
var SearchFieldJapaneseTitle = "japaneseOfficialTitle"
var SearchFieldXJatTitle = "xJatOfficialTitle"
var SearchFieldPrimaryTitle = "primaryTitle"

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
	animeMap map[string]*Anime
	reader   *bluge.Reader
}

func NewAnimeSearcher(anime []*Anime) *AnimeSearcher {
	animeMap := make(map[string]*Anime, len(anime))

	for _, entry := range anime {
		animeMap[entry.ID] = entry
	}

	return &AnimeSearcher{
		animeMap: animeMap,
	}
}

func (s *AnimeSearcher) GetAnime(id string) (*Anime, error) {
	if anime, ok := s.animeMap[id]; ok {
		return anime, nil
	}

	return nil, fmt.Errorf("anime with id %s not found", id)
}

func (s *AnimeSearcher) Search(ctx context.Context, queryStr string, limit int) ([]*Match, error) {
	if s.reader == nil {
		return nil, ErrReaderNotInitialized
	}

	fmt.Println("Searching for:", queryStr)

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
			err = next.VisitStoredFields(func(field string, value []byte) bool {
				if field == "_id" {
					matches.addMatch(s.animeMap[string(value)], next.Score, query.Field())
				}
				return true
			})
			if err != nil {
				return nil, err
			}
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

func (s *AnimeSearcher) CreateIndex(path string) error {
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

	for _, entry := range s.animeMap {
		doc := bluge.NewDocument(entry.ID).
			AddField(bluge.NewKeywordField("id", entry.ID)).
			AddField(bluge.NewTextField(SearchFieldPrimaryTitle, entry.PrimaryTitle)).
			AddField(bluge.NewTextField(SearchFieldXJatTitle, entry.XJatOfficialTitle)).
			AddField(bluge.NewTextField(SearchFieldJapaneseTitle, entry.JapaneseOfficialTitle)).
			AddField(bluge.NewTextField(SearchFieldEnglishTitle, entry.EnglishOfficialTitle))
		// AddField(bluge.NewCompositeFieldIncluding("xJatSynonyms", entry.XJatSynonyms)).
		// AddField(bluge.NewCompositeFieldIncluding("japaneseSynonyms", entry.JapaneseSynonyms)).
		// AddField(bluge.NewCompositeFieldIncluding("englishSynonyms", entry.EnglishSynonyms))

		batch.Update(doc.ID(), doc)
	}

	return index.Batch(batch)
}
