package data

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/blugelabs/bluge"
)

type Store interface {
	Marshal() (*bluge.Document, error)
}

type Read[T any] interface {
	*T
	Unmarshal(map[string][]byte) error
}

type DocumentStore[T any, PT Read[T]] struct {
	config         *storeConfig
	writer         *bluge.Writer
	reader         *bluge.Reader
	recordChan     chan Store
	parentCtx      context.Context
	writeCtx       context.Context
	writeCtxCancel context.CancelFunc
	mu             sync.RWMutex
	closeMu        sync.Mutex
	closedSignal   chan struct{}
}

func NewDocumentStore[T any, PT Read[T]](parentCtx context.Context, options ...Option) *DocumentStore[T, PT] {
	s := &DocumentStore[T, PT]{
		config:       &storeConfig{},
		recordChan:   make(chan Store),
		closedSignal: make(chan struct{}, 1),
		parentCtx:    parentCtx,
	}

	s.config.applyDefaults()
	s.config.applyOptions(options)

	return s
}

func (s *DocumentStore[T, PT]) handleWrites() {
	buff := bluge.NewBatch()
	buffLen := 0

	flush := func() (err error) {
		slog.Debug("flushing batch", slog.Int("len", buffLen))
		defer slog.Debug("batch flushed", slog.Int("len", buffLen))
		slog.Debug("acquiring lock")
		s.mu.Lock()
		slog.Debug("acquired lock")
		defer s.mu.Unlock()
		defer buff.Reset()

		if buffLen == 0 {
			return
		}

		if err = s.writer.Batch(buff); err != nil {
			slog.Error("unable to write batch", slog.String("err", err.Error()))
			return
		}

		buffLen = 0
		return
	}

	defer func() {
		slog.Debug("closing writer")
		defer slog.Debug("writer closed")
		if err := flush(); err != nil {
			panic(err)
		}
		if err := s.writer.Close(); err != nil {
			panic(err)
		}

		s.writer = nil

		// reset the reader to update the index
		if s.reader != nil {
			s.reader.Close()
		}

		var err error

		if s.reader, err = bluge.OpenReader(bluge.DefaultConfig(s.config.Path)); err != nil {
			panic(err)
		}

		s.closedSignal <- struct{}{}
	}()

	for {
		select {
		case <-s.writeCtx.Done():
			slog.Debug("flushing batch due to context")
			return
		case <-time.After(s.config.MaxBuffAge):
			slog.Debug("flushing batch due to age", slog.Int("len", buffLen))
			if err := flush(); err != nil {
				panic(err)
			}
		case rec := <-s.recordChan:
			doc, err := rec.Marshal()

			if err != nil {
				panic(err)
			}

			buff.Insert(doc)
			buffLen++

			if buffLen >= s.config.MaxBuffLen {
				slog.Debug("flushing batch due to len", slog.Int("len", buffLen))
				if err := flush(); err != nil {
					panic(err)
				}
			}
		}
	}
}

// Closes the writer and waits for it to finish flushing.
// This store may still be used for both reading and writing after this
// operation is complete. It is best to call this method when you are
// done writing to the store.
func (s *DocumentStore[T, PT]) Flush() {
	s.closeMu.Lock()
	if s.writer != nil {
		s.writeCtxCancel()
		<-s.closedSignal
	}
	s.closeMu.Unlock()
}

// Closes the reader and waits for it to finish flushing.
// This store should not be used after this operation is complete.
func (s *DocumentStore[T, PT]) Close() {
	s.Flush()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.reader != nil {
		s.reader.Close()
		s.reader = nil
	}

	close(s.recordChan)
	s.recordChan = nil
	s.closedSignal = nil
}

func (s *DocumentStore[T, PT]) Store(record Store) (err error) {
	s.mu.Lock()

	if s.writer == nil {
		s.writer, err = bluge.OpenWriter(bluge.DefaultConfig(s.config.Path))
		s.writeCtx, s.writeCtxCancel = context.WithCancel(s.parentCtx)

		if err != nil {
			return
		}

		go s.handleWrites()
	} else if err = s.writeCtx.Err(); err != nil {
		return
	}

	s.mu.Unlock()
	s.recordChan <- record
	return
}

func (s *DocumentStore[T, PT]) Get(ctx context.Context, id string) (record PT, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.reader == nil {
		s.reader, err = bluge.OpenReader(bluge.DefaultConfig(s.config.Path))

		if err != nil {
			return
		}
	}

	req := bluge.NewTermQuery(id).SetField("_id")
	dmi, err := s.reader.Search(ctx, bluge.NewTopNSearch(1, req))

	if err != nil {
		return
	}

	next, err := dmi.Next()

	if err != nil {
		return
	}

	if next == nil {
		err = fmt.Errorf("record store: record with id %s not found", id)
	}

	record = PT(new(T))
	fields := make(map[string][]byte)

	err = next.VisitStoredFields(func(field string, value []byte) bool {
		fields[field] = value
		return true
	})

	if err != nil {
		return
	}

	err = record.Unmarshal(fields)
	return
}

func (s *DocumentStore[T, PT]) Search(ctx context.Context, matchQuery string, limit int) (matches []Match[T], err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.reader == nil {
		s.reader, err = bluge.OpenReader(bluge.DefaultConfig(s.config.Path))

		if err != nil {
			return
		}
	}

	matchList := make(MatchList[T])

	for _, field := range s.config.SearchFields {
		query := bluge.NewMatchQuery(matchQuery).SetField(field)
		searchRequest := bluge.NewTopNSearch(limit, query)
		dmi, err := s.reader.Search(ctx, searchRequest)

		if err != nil {
			return nil, err
		}

		next, err := dmi.Next()

		for next != nil && err == nil {
			record := PT(new(T))
			fields := make(map[string][]byte)

			err = next.VisitStoredFields(func(field string, value []byte) bool {
				fields[field] = value
				return true
			})

			if err != nil {
				break
			}

			if err = record.Unmarshal(fields); err != nil {
				break
			}

			match := Match[T]{string(fields["_id"]), record, next.Score, field}
			matchList.Insert(match)
			next, err = dmi.Next()
		}

		if err != nil {
			return nil, err
		}
	}

	matches = matchList.Top(limit)

	return
}
