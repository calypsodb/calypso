package calypso

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	art "github.com/plar/go-adaptive-radix-tree"

	internal "github.com/calypsodb/calypso/octopus"
	"github.com/calypsodb/calypso/octopus/data"
	"github.com/calypsodb/calypso/octopus/metadata"
)

func (b *Bitcask) maybeRotate() error {
	size := b.curr.Size()
	if size < int64(b.config.MaxDatafileSize) {
		return nil
	}

	err := b.curr.Close()
	if err != nil {
		return err
	}

	id := b.curr.FileID()

	df, err := data.NewDatafile(
		b.path, id, true,
		b.config.MaxKeySize,
		b.config.MaxValueSize,
		b.config.FileFileModeBeforeUmask,
	)
	if err != nil {
		return err
	}

	b.datafiles[id] = df

	id = b.curr.FileID() + 1
	curr, err := data.NewDatafile(
		b.path, id, false,
		b.config.MaxKeySize,
		b.config.MaxValueSize,
		b.config.FileFileModeBeforeUmask,
	)
	if err != nil {
		return err
	}
	b.curr = curr
	err = b.saveIndexes()
	if err != nil {
		return err
	}

	return nil
}

// put inserts a new (key, value). Both key and value are valid inputs.
func (b *Bitcask) put(key, value []byte) (int64, int64, error) {
	if err := b.maybeRotate(); err != nil {
		return -1, 0, fmt.Errorf("error rotating active datafile: %w", err)
	}

	return b.curr.Write(internal.NewEntry(key, value, nil))
}

// putWithExpiry inserts a new (key, value, expiry).
// Both key and value are valid inputs.
func (b *Bitcask) putWithExpiry(key, value []byte, expiry time.Time) (int64, int64, error) {
	if err := b.maybeRotate(); err != nil {
		return -1, 0, fmt.Errorf("error rotating active datafile: %w", err)
	}

	return b.curr.Write(internal.NewEntry(key, value, &expiry))
}

// closeCurrentFile closes current datafile and makes it read only.
func (b *Bitcask) closeCurrentFile() error {
	if err := b.curr.Close(); err != nil {
		return err
	}

	id := b.curr.FileID()
	df, err := data.NewDatafile(
		b.path, id, true,
		b.config.MaxKeySize,
		b.config.MaxValueSize,
		b.config.FileFileModeBeforeUmask,
	)
	if err != nil {
		return err
	}

	b.datafiles[id] = df
	return nil
}

// openNewWritableFile opens new datafile for writing data
func (b *Bitcask) openNewWritableFile() error {
	id := b.curr.FileID() + 1
	curr, err := data.NewDatafile(
		b.path, id, false,
		b.config.MaxKeySize,
		b.config.MaxValueSize,
		b.config.FileFileModeBeforeUmask,
	)
	if err != nil {
		return err
	}
	b.curr = curr
	return nil
}

// saveIndex saves index and ttl_index currently in RAM to disk
func (b *Bitcask) saveIndexes() error {
	tempIdx := "temp_index"
	if err := b.indexer.Save(b.trie, filepath.Join(b.path, tempIdx)); err != nil {
		return err
	}
	if err := os.Rename(filepath.Join(b.path, tempIdx), filepath.Join(b.path, "index")); err != nil {
		return err
	}
	if err := b.ttlIndexer.Save(b.ttlIndex, filepath.Join(b.path, tempIdx)); err != nil {
		return err
	}
	return os.Rename(filepath.Join(b.path, tempIdx), filepath.Join(b.path, ttlIndexFile))
}

// saveMetadata saves metadata into disk
func (b *Bitcask) saveMetadata() error {
	return b.metadata.Save(filepath.Join(b.path, "meta.json"), b.config.DirFileModeBeforeUmask)
}

func loadDatafiles(path string, maxKeySize uint32, maxValueSize uint64, fileModeBeforeUmask os.FileMode) (datafiles map[int]data.Datafile, lastID int, err error) {
	fns, err := internal.GetDatafiles(path)
	if err != nil {
		return nil, 0, err
	}

	ids, err := internal.ParseIds(fns)
	if err != nil {
		return nil, 0, err
	}

	datafiles = make(map[int]data.Datafile, len(ids))
	for _, id := range ids {
		datafiles[id], err = data.NewDatafile(
			path, id, true,
			maxKeySize,
			maxValueSize,
			fileModeBeforeUmask,
		)
		if err != nil {
			return
		}

	}
	if len(ids) > 0 {
		lastID = ids[len(ids)-1]
	}
	return
}

func getSortedDatafiles(datafiles map[int]data.Datafile) []data.Datafile {
	out := make([]data.Datafile, len(datafiles))
	idx := 0
	for _, df := range datafiles {
		out[idx] = df
		idx++
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].FileID() < out[j].FileID()
	})
	return out
}

// loadIndexes loads index from disk to memory. If index is not available or partially available (last bitcask process crashed)
// then it iterates over last datafile and construct index
// we construct ttl_index here also along with normal index
func loadIndexes(b *Bitcask, datafiles map[int]data.Datafile, lastID int) (art.Tree, art.Tree, error) {
	t, found, err := b.indexer.Load(filepath.Join(b.path, "index"), b.config.MaxKeySize)
	if err != nil {
		return nil, nil, err
	}
	ttlIndex, _, err := b.ttlIndexer.Load(filepath.Join(b.path, ttlIndexFile), b.config.MaxKeySize)
	if err != nil {
		return nil, nil, err
	}
	if found && b.metadata.IndexUpToDate {
		return t, ttlIndex, nil
	}
	if found {
		if err := loadIndexFromDatafile(t, ttlIndex, datafiles[lastID]); err != nil {
			return nil, ttlIndex, err
		}
		return t, ttlIndex, nil
	}
	sortedDatafiles := getSortedDatafiles(datafiles)
	for _, df := range sortedDatafiles {
		if err := loadIndexFromDatafile(t, ttlIndex, df); err != nil {
			return nil, ttlIndex, err
		}
	}
	return t, ttlIndex, nil
}

func loadIndexFromDatafile(t art.Tree, ttlIndex art.Tree, df data.Datafile) error {
	var offset int64
	for {
		e, n, err := df.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		// Tombstone value  (deleted key)
		if len(e.Value) == 0 {
			t.Delete(e.Key)
			offset += n
			continue
		}
		item := internal.Item{FileID: df.FileID(), Offset: offset, Size: n}
		t.Insert(e.Key, item)
		if e.Expiry != nil {
			ttlIndex.Insert(e.Key, *e.Expiry)
		}
		offset += n
	}
	return nil
}

func loadMetadata(path string) (*metadata.MetaData, error) {
	if !internal.Exists(filepath.Join(path, "meta.json")) {
		meta := new(metadata.MetaData)
		return meta, nil
	}
	return metadata.Load(filepath.Join(path, "meta.json"))
}
