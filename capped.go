package capped

// Indexer it is an index manager for any given collection with access
// via indexes. It allows to emulate the behavior of capped collections
type Indexer struct {
	readIndex  int
	writeIndex int
	size       int
	overwrite  bool
}

func mod(a, b int) int {
	m := a % b
	if a < 0 {
		m += b
	}
	return m
}

// NewIndexer creates an index manager for a collection
// to ensure the behavior of a capped collection
func NewIndexer(collectionSize int) *Indexer {
	if collectionSize < 2 {
		panic("incorrect collection size < 2")
	}
	return &Indexer{
		readIndex:  -1,
		writeIndex: -1,
		size:       collectionSize,
		overwrite:  false,
	}
}

// WriteIndex return the next index where to write to the collection
func (i *Indexer) WriteIndex() (int, bool) {
	prevIndex := i.writeIndex
	i.writeIndex = mod(prevIndex+1, i.size)

	if i.readIndex == -1 {
		if i.writeIndex == 0 && prevIndex != -1 {
			i.overwrite = true
			i.readIndex = 1
		}
	} else if i.writeIndex == i.readIndex {
		i.overwrite = true
		i.readIndex = mod(i.writeIndex+1, i.size)
	}
	return i.writeIndex, i.overwrite
}

// ReadIndex return the next index where to read from the collection
func (i *Indexer) ReadIndex() int {
	if i.writeIndex == -1 {
		return -1
	}
	nextIndex := mod(i.readIndex+1, i.size)
	if nextIndex == mod(i.writeIndex+1, i.size) {
		return -1
	}
	if i.overwrite {
		i.overwrite = false
		return i.readIndex
	}
	i.readIndex = nextIndex
	return i.readIndex
}

// Len return the number of items written
// but not read from the collection
func (i *Indexer) Len() int {
	if i.writeIndex == i.readIndex {
		return 0
	}
	if i.readIndex == -1 {
		return mod(i.writeIndex+1, i.size+1)
	}
	size := mod(i.writeIndex-i.readIndex, i.size)
	if mod(i.writeIndex+1, i.size) == i.readIndex {
		if i.overwrite {
			size++
		}
	}
	return size
}
