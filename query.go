package calypso

import (
	"bytes"

	"github.com/abcum/lcp"
	art "github.com/plar/go-adaptive-radix-tree"
)

// Sift iterates over all keys in the database calling the function `f` for
// each key. If the KV pair is expired or the function returns true, that key is
// deleted from the database.
// If the function returns an error on any key, no further keys are processed, no
// keys are deleted, and the first error is returned.
func (b *Bitcask) Sift(f func(key []byte) (bool, error)) (err error) {
	keysToDelete := art.New()

	b.mu.RLock()
	b.trie.ForEach(func(node art.Node) bool {
		if b.isExpired(node.Key()) {
			keysToDelete.Insert(node.Key(), true)
			return true
		}
		var shouldDelete bool
		if shouldDelete, err = f(node.Key()); err != nil {
			return false
		} else if shouldDelete {
			keysToDelete.Insert(node.Key(), true)
		}
		return true
	})
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()
	keysToDelete.ForEach(func(node art.Node) (cont bool) {
		b.delete(node.Key())
		return true
	})
	return
}

// Scan performs a prefix scan of keys matching the given prefix and calling
// the function `f` with the keys found. If the function returns an error
// no further keys are processed and the first error is returned.
func (b *Bitcask) Scan(prefix []byte, f func(key []byte) error) (err error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	b.trie.ForEachPrefix(prefix, func(node art.Node) bool {
		// Skip the root node
		if len(node.Key()) == 0 {
			return true
		}

		if err = f(node.Key()); err != nil {
			return false
		}
		return true
	})
	return
}

// SiftScan iterates over all keys in the database beginning with the given
// prefix, calling the function `f` for each key. If the KV pair is expired or
// the function returns true, that key is deleted from the database.
//
//	If the function returns an error on any key, no further keys are processed,
//
// no keys are deleted, and the first error is returned.
func (b *Bitcask) SiftScan(prefix []byte, f func(key []byte) (bool, error)) (err error) {
	keysToDelete := art.New()

	b.mu.RLock()
	b.trie.ForEachPrefix(prefix, func(node art.Node) bool {
		// Skip the root node
		if len(node.Key()) == 0 {
			return true
		}
		if b.isExpired(node.Key()) {
			keysToDelete.Insert(node.Key(), true)
			return true
		}
		var shouldDelete bool
		if shouldDelete, err = f(node.Key()); err != nil {
			return false
		} else if shouldDelete {
			keysToDelete.Insert(node.Key(), true)
		}
		return true
	})
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()
	keysToDelete.ForEach(func(node art.Node) (cont bool) {
		b.delete(node.Key())
		return true
	})
	return
}

// Range performs a range scan of keys matching a range of keys between the
// start key and end key and calling the function `f` with the keys found.
// If the function returns an error no further keys are processed and the
// first error returned.
func (b *Bitcask) Range(start, end []byte, f func(key []byte) error) (err error) {
	if bytes.Compare(start, end) == 1 {
		return ErrInvalidRange
	}

	commonPrefix := lcp.LCP(start, end)
	if commonPrefix == nil {
		return ErrInvalidRange
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	b.trie.ForEachPrefix(commonPrefix, func(node art.Node) bool {
		if bytes.Compare(node.Key(), start) >= 0 && bytes.Compare(node.Key(), end) <= 0 {
			if err = f(node.Key()); err != nil {
				return false
			}
			return true
		} else if bytes.Compare(node.Key(), start) >= 0 && bytes.Compare(node.Key(), end) > 0 {
			return false
		}
		return true
	})
	return
}

// SiftRange performs a range scan of keys matching a range of keys between the
// start key and end key and calling the function `f` with the keys found.
// If the KV pair is expired or the function returns true, that key is deleted
// from the database.
// If the function returns an error on any key, no further keys are processed, no
// keys are deleted, and the first error is returned.
func (b *Bitcask) SiftRange(start, end []byte, f func(key []byte) (bool, error)) (err error) {
	if bytes.Compare(start, end) == 1 {
		return ErrInvalidRange
	}

	commonPrefix := lcp.LCP(start, end)
	if commonPrefix == nil {
		return ErrInvalidRange
	}

	keysToDelete := art.New()

	b.mu.RLock()
	b.trie.ForEachPrefix(commonPrefix, func(node art.Node) bool {
		if bytes.Compare(node.Key(), start) >= 0 && bytes.Compare(node.Key(), end) <= 0 {
			if b.isExpired(node.Key()) {
				keysToDelete.Insert(node.Key(), true)
				return true
			}
			var shouldDelete bool
			if shouldDelete, err = f(node.Key()); err != nil {
				return false
			} else if shouldDelete {
				keysToDelete.Insert(node.Key(), true)
			}
			return true
		} else if bytes.Compare(node.Key(), start) >= 0 && bytes.Compare(node.Key(), end) > 0 {
			return false
		}
		return true
	})
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	keysToDelete.ForEach(func(node art.Node) (cont bool) {
		b.delete(node.Key())
		return true
	})

	return
}
