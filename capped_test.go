package capped

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIndexer(t *testing.T) {
	assert.Panics(t, func() { NewIndexer(-1) })
	assert.Panics(t, func() { NewIndexer(0) })
	idx := NewIndexer(42)
	assert.Equal(t, 0, idx.Len())

	assert.Equal(t, -1, idx.ReadIndex())
	assert.Equal(t, -1, idx.ReadIndex())
	assert.Equal(t, -1, idx.ReadIndex())
	assert.Equal(t, 0, idx.Len())

	assert.Equal(t, 0, idx.WriteIndex())
	assert.Equal(t, 1, idx.Len())
	assert.Equal(t, 1, idx.WriteIndex())
	assert.Equal(t, 2, idx.Len())

	assert.Equal(t, 0, idx.ReadIndex())
	assert.Equal(t, 1, idx.Len())

	assert.Equal(t, 2, idx.WriteIndex())
	assert.Equal(t, 2, idx.Len())
	assert.Equal(t, 1, idx.ReadIndex())
	assert.Equal(t, 1, idx.Len())
	assert.Equal(t, 2, idx.ReadIndex())
	assert.Equal(t, 0, idx.Len())
	assert.Equal(t, -1, idx.ReadIndex())
	assert.Equal(t, 0, idx.Len())
	assert.Equal(t, -1, idx.ReadIndex())
	assert.Equal(t, 0, idx.Len())
}

func TestWriteAndReadRespectLenght(t *testing.T) {
	idx := NewIndexer(42)
	idx.WriteIndex()
	idx.WriteIndex()
	idx.WriteIndex()

	expectedWriteIdx := 0
	for i := 3; i < 86; i++ {
		expectedWriteIdx = i % idx.size
		assert.Equal(t, expectedWriteIdx, idx.WriteIndex())
	}

	expectedReadIdx := mod(expectedWriteIdx+1, idx.size)
	for i := 42; i > -42; i-- {
		if i <= 0 {
			assert.Equal(t, 0, idx.Len())
			assert.Equal(t, -1, idx.ReadIndex())
		} else {
			assert.Equal(t, i, idx.Len())
			assert.Equal(t, expectedReadIdx, idx.ReadIndex())
			expectedReadIdx = mod(expectedReadIdx+1, idx.size)
		}
	}
	assert.Equal(t, mod(expectedWriteIdx+1, idx.size), idx.WriteIndex())
	assert.Equal(t, 1, idx.Len())
}

func TestReadIndexFollow(t *testing.T) {
	collectionSize := 13
	idx := NewIndexer(collectionSize)
	for i := 0; i < 42; i++ {
		t.Logf("Case indexFollow#%d i=%d idx=%#v", i+1, i, idx)
		expectedIdx := i % collectionSize
		assert.Equal(t, expectedIdx, idx.WriteIndex(), "unexpected write index")
		assert.Equal(t, 1, idx.Len(), "unexpected lenght")
		assert.Equal(t, expectedIdx, idx.ReadIndex(), "unexpected read index")
		assert.Equal(t, 0, idx.Len(), "unexpected lenght")
	}
}

func TestReadIndexPushed(t *testing.T) {
	collectionSize := 15
	idx := NewIndexer(collectionSize)
	for i := 0; i < 42; i++ {
		expectedWriteIdx := i % collectionSize
		assert.Equal(t, expectedWriteIdx, idx.WriteIndex(), "unexpected write index")
		t.Logf("Case indexPushed#%d expectedWriteIdx=%d idx=%#v", i+1, expectedWriteIdx, idx)
		if i < 15 {
			assert.Equal(t, -1, idx.readIndex, "unexpected prev read index")
			assert.Equal(t, i+1, idx.Len(), "unexpected lenght")
		} else {
			assert.Equal(t, mod(idx.writeIndex+1, idx.size), idx.readIndex, "unexpected read index")
			assert.Equal(t, 15, idx.Len(), "unexpected lenght")
		}
	}
}

func TestReadIndexPushedAfterRead(t *testing.T) {
	collectionSize := 11
	idx := NewIndexer(collectionSize)
	for i := 0; i < 42; i++ {
		expectedWriteIdx := i % collectionSize
		assert.Equal(t, expectedWriteIdx, idx.WriteIndex(), "unexpected write index")
		t.Logf("Case indexPushedAfterRead#%d expectedWriteIdx=%d idx=%#v", i+1, expectedWriteIdx, idx)
		if i == 0 {
			assert.Equal(t, 1, idx.Len(), "unexpected lenght")
			assert.Equal(t, 0, idx.ReadIndex(), "unexpected prev read index")
			assert.Equal(t, 0, idx.Len(), "unexpected lenght")
		} else if i < 11 {
			assert.Equal(t, 0, idx.readIndex, "unexpected read index")
			assert.Equal(t, i, idx.Len(), "unexpected lenght")
		} else {
			assert.Equal(t, mod(idx.writeIndex+1, idx.size), idx.readIndex, "unexpected read index")
			assert.Equal(t, 11, idx.Len(), "unexpected lenght")
		}
	}
}

func TestIndexPushedRead(t *testing.T) {
	collectionSize := 11
	idx := NewIndexer(collectionSize)
	for i := 0; i < 42; i++ {
		expectedWriteIdx := i % collectionSize
		assert.Equal(t, expectedWriteIdx, idx.WriteIndex(), "unexpected write index")

		t.Logf("Case indexPushedAfterRead#%d expectedWriteIdx=%d idx=%#v", i+1, expectedWriteIdx, idx)
		if i == 0 {
			assert.Equal(t, 1, idx.Len(), "unexpected lenght")
			assert.Equal(t, 0, idx.ReadIndex(), "unexpected prev read index")
			assert.Equal(t, 0, idx.Len(), "unexpected lenght")
		} else if i < 11 {
			assert.Equal(t, 0, idx.readIndex, "unexpected read index")
			assert.Equal(t, i, idx.Len(), "unexpected lenght")
		} else {
			assert.Equal(t, 11, idx.Len(), "unexpected lenght")
			if i%3 == 0 {
				_ = idx.ReadIndex()
				assert.Equal(t, 10, idx.Len(), "unexpected lenght")
			}
		}
	}
}
