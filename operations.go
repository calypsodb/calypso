package calypso

import (
	"hash/crc32"
	"time"

	art "github.com/plar/go-adaptive-radix-tree"

	internal "github.com/calypsodb/calypso/octopus"
	"github.com/calypsodb/calypso/octopus/data"
	"github.com/calypsodb/calypso/octopus/data/codec"
)

// Get fetches value for a key
func (b *Bitcask) Get(key []byte) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	e, err := b.get(key)
	if err != nil {
		return nil, err
	}
	return e.Value, nil
}

// Has returns true if the key exists in the database, false otherwise.
func (b *Bitcask) Has(key []byte) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, found := b.trie.Search(key)
	if found {
		return !b.isExpired(key)
	}
	return found
}

// Put stores the key and value in the database.
func (b *Bitcask) Put(key, value []byte) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}
	if b.config.MaxKeySize > 0 && uint32(len(key)) > b.config.MaxKeySize {
		return ErrKeyTooLarge
	}
	if b.config.MaxValueSize > 0 && uint64(len(value)) > b.config.MaxValueSize {
		return ErrValueTooLarge
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	offset, n, err := b.put(key, value)
	if err != nil {
		return err
	}

	if b.config.Sync {
		if err := b.curr.Sync(); err != nil {
			return err
		}
	}

	// in case of successful `put`, IndexUpToDate will be always be false
	b.metadata.IndexUpToDate = false

	if oldItem, found := b.trie.Search(key); found {
		b.metadata.ReclaimableSpace += oldItem.(internal.Item).Size
	}

	item := internal.Item{FileID: b.curr.FileID(), Offset: offset, Size: n}
	b.trie.Insert(key, item)

	return nil
}

// PutWithTTL stores the key and value in the database with the given TTL
func (b *Bitcask) PutWithTTL(key, value []byte, ttl time.Duration) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}
	if b.config.MaxKeySize > 0 && uint32(len(key)) > b.config.MaxKeySize {
		return ErrKeyTooLarge
	}
	if b.config.MaxValueSize > 0 && uint64(len(value)) > b.config.MaxValueSize {
		return ErrValueTooLarge
	}

	expiry := time.Now().Add(ttl)

	b.mu.Lock()
	defer b.mu.Unlock()
	offset, n, err := b.putWithExpiry(key, value, expiry)
	if err != nil {
		return err
	}

	if b.config.Sync {
		if err := b.curr.Sync(); err != nil {
			return err
		}
	}

	// in case of successful `put`, IndexUpToDate will be always be false
	b.metadata.IndexUpToDate = false

	if oldItem, found := b.trie.Search(key); found {
		b.metadata.ReclaimableSpace += oldItem.(internal.Item).Size
	}

	item := internal.Item{FileID: b.curr.FileID(), Offset: offset, Size: n}
	b.trie.Insert(key, item)
	b.ttlIndex.Insert(key, expiry)

	return nil
}

// Delete deletes the named key.
func (b *Bitcask) Delete(key []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delete(key)
}

// delete deletes the named key. If the key doesn't exist or an I/O error
// occurs the error is returned.
func (b *Bitcask) delete(key []byte) error {
	_, _, err := b.put(key, []byte{})
	if err != nil {
		return err
	}
	if item, found := b.trie.Search(key); found {
		b.metadata.ReclaimableSpace += item.(internal.Item).Size + codec.MetaInfoSize + int64(len(key))
	}
	b.trie.Delete(key)
	b.ttlIndex.Delete(key)

	return nil
}

// DeleteAll deletes all the keys. If an I/O error occurs the error is returned.
func (b *Bitcask) DeleteAll() (err error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	b.trie.ForEach(func(node art.Node) bool {
		_, _, err = b.put(node.Key(), []byte{})
		if err != nil {
			return false
		}
		item, _ := b.trie.Search(node.Key())
		b.metadata.ReclaimableSpace += item.(internal.Item).Size + codec.MetaInfoSize + int64(len(node.Key()))
		return true
	})
	b.trie = art.New()
	b.ttlIndex = art.New()

	return
}

// Len returns the total number of keys in the database
func (b *Bitcask) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.trie.Size()
}

// Keys returns all keys in the database as a channel of keys
func (b *Bitcask) Keys() chan []byte {
	ch := make(chan []byte)
	go func() {
		b.mu.RLock()
		defer b.mu.RUnlock()

		for it := b.trie.Iterator(); it.HasNext(); {
			node, _ := it.Next()
			if b.isExpired(node.Key()) {
				continue
			}
			ch <- node.Key()
		}
		close(ch)
	}()

	return ch
}

// RunGC deletes all expired keys
func (b *Bitcask) RunGC() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.runGC()
}

// runGC deletes all keys that are expired
// caller function should take care of the locking when calling this method
func (b *Bitcask) runGC() (err error) {
	keysToDelete := art.New()

	b.ttlIndex.ForEach(func(node art.Node) (cont bool) {
		if !b.isExpired(node.Key()) {
			// later, return false here when the ttlIndex is sorted
			return true
		}
		keysToDelete.Insert(node.Key(), true)
		return true
	})

	keysToDelete.ForEach(func(node art.Node) (cont bool) {
		b.delete(node.Key())
		return true
	})

	return nil
}

// Fold iterates over all keys in the database calling the function `f` for
// each key. If the function returns an error, no further keys are processed
// and the error is returned.
func (b *Bitcask) Fold(f func(key []byte) error) (err error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	b.trie.ForEach(func(node art.Node) bool {
		if err = f(node.Key()); err != nil {
			return false
		}
		return true
	})

	return
}

// get retrieves the value of the given key
func (b *Bitcask) get(key []byte) (internal.Entry, error) {
	var df data.Datafile

	value, found := b.trie.Search(key)
	if !found {
		return internal.Entry{}, ErrKeyNotFound
	}
	if b.isExpired(key) {
		return internal.Entry{}, ErrKeyExpired
	}

	item := value.(internal.Item)

	if item.FileID == b.curr.FileID() {
		df = b.curr
	} else {
		df = b.datafiles[item.FileID]
	}

	e, err := df.ReadAt(item.Offset, item.Size)
	if err != nil {
		return internal.Entry{}, err
	}

	checksum := crc32.ChecksumIEEE(e.Value)
	if checksum != e.Checksum {
		return internal.Entry{}, ErrChecksumFailed
	}

	return e, nil
}

// isExpired returns true if a key has expired
// it returns false if key does not exist in ttl index
func (b *Bitcask) isExpired(key []byte) bool {
	expiry, found := b.ttlIndex.Search(key)
	if !found {
		return false
	}
	return expiry.(time.Time).Before(time.Now().UTC())
}
