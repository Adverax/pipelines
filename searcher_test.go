package pipelines

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sort"
	"sync"
	"testing"
	"time"
)

// Пример реализации параллельного агрегирующего поиска на нескольких поисковиках (SearchEngine).
// Данный пример затрагивает полный цикл:
// - легкое добавление новых поисковиков
// - декомпозиция задачи (Searcher, Aggregator, pipelines)
// - ограничение времени на выполнение
// - обработка ошибок
// - gracefully shutdown
type SearchEngine interface {
	Search(ctx context.Context, request *SearchRequest) *SearchResponse
}

type CustomAggregator interface {
	Aggregate(ctx context.Context, response *SearchResponse) error
	Summary() ([]string, error)
}

type AggregatorFactory interface {
	New() CustomAggregator
}

type SearchRequest struct {
	Query  string
	engine SearchEngine
}

type SearchResponse struct {
	Text string
	Err  error
}

type DummySearchEngine struct {
	name     string
	duration time.Duration
}

func (that *DummySearchEngine) Search(ctx context.Context, request *SearchRequest) *SearchResponse {
	select {
	case <-ctx.Done():
		return &SearchResponse{
			Err: ctx.Err(),
		}
	case <-time.After(that.duration):
	}

	return &SearchResponse{Text: fmt.Sprintf("founded by %s", that.name)}
}

type Searcher struct {
	engines []SearchEngine
	results AggregatorFactory
	timeout time.Duration
}

func NewSearcher(
	results AggregatorFactory,
	timeout time.Duration,
	engines ...SearchEngine,
) *Searcher {
	return &Searcher{
		results: results,
		timeout: timeout,
		engines: engines,
	}
}

func (that *Searcher) Search(ctx context.Context, query string) ([]string, error) {
	if that.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, that.timeout)
		defer cancel()
	}

	return that.search(ctx, query)
}

func (that *Searcher) search(ctx context.Context, query string) ([]string, error) {
	result := that.results.New()

	err := Consume(
		context.Background(), // non cancelable consume
		NewFanIn(
			ctx,
			NewOperations[*SearchRequest, *SearchResponse](
				ctx,
				OpFunc[*SearchRequest, *SearchResponse](
					func(ctx context.Context, request *SearchRequest) *SearchResponse {
						return request.engine.Search(ctx, request)
					},
				),
				NewFanOut(
					ctx,
					NewIterator(ctx, that.newRequestIterator(query)),
					len(that.engines),
				)...,
			)...,
		),
		result,
	)
	if err != nil && !errors.Is(err, context.Canceled) {
		return nil, err
	}

	return result.Summary()
}

func (that *Searcher) newRequestIterator(query string) func(yield func(*SearchRequest) bool) {
	return func(yield func(*SearchRequest) bool) {
		for _, engine := range that.engines {
			request := &SearchRequest{
				engine: engine,
				Query:  query,
			}
			if !yield(request) {
				return
			}
		}
	}
}

type SearchAggregatorFactory struct{}

func (that *SearchAggregatorFactory) New() CustomAggregator {
	return &SearchAggregator{}
}

type SearchAggregator struct {
	sync.Mutex
	values []string
	error  error
}

func (that *SearchAggregator) Aggregate(ctx context.Context, response *SearchResponse) error {
	that.Lock()
	defer that.Unlock()

	if response.Err != nil {
		if errors.Is(response.Err, context.DeadlineExceeded) {
			return nil
		}

		if that.error == nil {
			that.error = response.Err
		}

		return nil
	}

	that.values = append(that.values, response.Text)
	return nil
}

func (that *SearchAggregator) Summary() ([]string, error) {
	that.Lock()
	defer that.Unlock()

	sort.Strings(that.values)
	return that.values, that.error
}

func TestSearcher(t *testing.T) {
	searcher := NewSearcher(
		&SearchAggregatorFactory{},
		500*time.Millisecond,
		&DummySearchEngine{name: "engine1", duration: 50 * time.Duration(time.Millisecond)},
		&DummySearchEngine{name: "engine2", duration: 10 * time.Duration(time.Millisecond)},
		&DummySearchEngine{name: "engine3", duration: 1000 * time.Duration(time.Millisecond)},
		&DummySearchEngine{name: "engine4", duration: 70 * time.Duration(time.Millisecond)},
	)

	result, err := searcher.Search(context.Background(), "query")
	assert.NoError(t, err)
	assert.Equal(t, []string{"founded by engine1", "founded by engine2", "founded by engine4"}, result)
}
