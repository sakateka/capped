package capped

import (
	"io"
	"math/rand"
	"runtime"

	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ExampleBytes struct {
	store []byte
	index *Indexer
}

func NewExampleBytes(size int) ExampleBytes {
	return ExampleBytes{
		store: make([]byte, size),
		index: NewIndexer(size),
	}
}

func (e *ExampleBytes) pop() (byte, error) {
	idx := e.index.ReadIndex()
	if idx == -1 {
		return 0, io.EOF
	}
	return e.store[idx], nil
}

func (e *ExampleBytes) push(b byte) bool {
	idx, overwrite := e.index.WriteIndex()
	e.store[idx] = b
	return overwrite
}

func TestCappedBytes(t *testing.T) {
	cappedBytes := NewExampleBytes(5)

	data := "example data"
	for _, i := range data {
		cappedBytes.push(byte(i))
	}

	var tail []byte
	for b, err := cappedBytes.pop(); err == nil; b, err = cappedBytes.pop() {
		tail = append(tail, b)
	}
	t.Logf("Expected tail %q", data[len(data)-5:])
	assert.Equal(t, data[len(data)-5:], string(tail))
}

// ExampleStrings capped slice of string
type ExampleStrings struct {
	store []string
	index *Indexer
}

func NewExampleStrings(size int) ExampleStrings {
	return ExampleStrings{
		store: make([]string, size),
		index: NewIndexer(size),
	}
}

func (e *ExampleStrings) pop() (string, error) {
	idx := e.index.ReadIndex()
	if idx == -1 {
		return "", io.EOF
	}
	return e.store[idx], nil
}

func (e *ExampleStrings) push(s string) bool {
	idx, overwrite := e.index.WriteIndex()
	e.store[idx] = s
	return overwrite
}

func TestCappedStrings(t *testing.T) {
	strings := NewExampleStrings(3)

	strings.push("one")
	assert.Equal(t, []string{"one", "", ""}, strings.store)

	strings.push("two")
	assert.Equal(t, []string{"one", "two", ""}, strings.store)

	strings.push("three")
	assert.Equal(t, []string{"one", "two", "three"}, strings.store)

	strings.push("four")
	assert.Equal(t, []string{"four", "two", "three"}, strings.store)

	strings.push("five")
	assert.Equal(t, []string{"four", "five", "three"}, strings.store)

	three, err := strings.pop()
	assert.NoError(t, err)
	assert.Equal(t, three, "three")
}

// ExampleConcurrent thread safe usage
type ExampleConcurrent struct {
	sync.Mutex
	store []int
	index *Indexer
}

func NewExampleConcurrent(size int) ExampleConcurrent {
	return ExampleConcurrent{
		store: make([]int, size),
		index: NewIndexer(size),
	}
}

func (e *ExampleConcurrent) pop() (int, error) {
	e.Lock()
	defer e.Unlock()
	idx := e.index.ReadIndex()
	if idx == -1 {
		return 0, io.EOF
	}
	return e.store[idx], nil
}

func (e *ExampleConcurrent) push(s int) bool {
	e.Lock()
	defer e.Unlock()
	idx, overwrite := e.index.WriteIndex()
	e.store[idx] = s
	return overwrite
}

func TestExampleConcurrent(t *testing.T) {
	ints := NewExampleConcurrent(30)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); ints.push(1) }()
	wg.Add(1)
	go func() { defer wg.Done(); ints.push(2) }()
	wg.Add(1)
	go func() { defer wg.Done(); ints.push(3) }()

	wg.Wait()
	n, err := ints.pop()
	assert.NoError(t, err)
	assert.True(t, n > 0)

	// round 2
	wg = sync.WaitGroup{}
	results := []int{}
	expected := []int{}

	syncCh := make(chan struct{})
	for i := 0; i < len(ints.store); i++ {
		expected = append(expected, i) // 0, 1, 3, 4, 5, 6, ...
		wg.Add(1)
		go func(n int) {
			<-syncCh
			defer wg.Done()
			runtime.Gosched()
			t.Logf("thread #%d push int=%[1]d", n)
			ints.push(n)
		}(i)
	}
	close(syncCh)
	wg.Wait()
	assert.Equal(t, len(ints.store), ints.index.Len())

	// A far-fetched Babylonian mess
	// tries to create high parallelism in access to the collection
	wg = sync.WaitGroup{}
	syncCh = make(chan struct{})
	resCh := make(chan int, len(ints.store))
	for j := 1; j < 5; j++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-syncCh
			for {
				runtime.Gosched()
				i, err := ints.pop()
				if err == io.EOF {
					break
				}
				t.Logf("thread #%d pop int=%d", id, i)
				if rand.Int()%3 == 0 {
					t.Logf("thread #%d push int=%d", id, i)
					ints.push(i)
					continue
				}
				t.Logf("thread #%d send int=%d", id, i)
				resCh <- i
			}
		}(j)
	}

	go func() {
		wg.Wait()
		close(resCh)
	}()

	close(syncCh)
	for i := range resCh {
		results = append(results, i)
	}
	sort.Ints(results)
	assert.Equal(t, expected, results)
	assert.Zero(t, ints.index.Len())
}
