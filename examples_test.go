package capped

import (
	"io"
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

func (e *ExampleBytes) push(b byte) {
	idx := e.index.WriteIndex()
	e.store[idx] = b
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

func (e *ExampleStrings) push(s string) {
	idx := e.index.WriteIndex()
	e.store[idx] = s
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
