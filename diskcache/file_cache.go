package diskcache

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/afero"
	"golang.org/x/sync/singleflight"
)

// numStripes bounds lock memory: keys hash onto a fixed set of RWMutexes
// instead of growing a per-key map forever.
const numStripes = 64

type FileCache struct {
	fs afero.Fs

	stripes [numStripes]sync.RWMutex
	// sf coalesces concurrent Store calls for the same key so a thumbnail
	// stampede generates/writes once instead of N times.
	sf singleflight.Group
}

func New(fs afero.Fs, root string) *FileCache {
	return &FileCache{
		fs: afero.NewBasePathFs(fs, root),
	}
}

func (f *FileCache) shardFor(key string) *sync.RWMutex {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return &f.stripes[h.Sum32()%numStripes]
}

func (f *FileCache) Store(ctx context.Context, key string, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err, _ := f.sf.Do("store:"+key, func() (interface{}, error) {
		return nil, f.storeAtomic(ctx, key, value)
	})
	return err
}

// storeAtomic writes to a temp file in the same directory and renames it
// over the destination, so concurrent Load calls never observe a partial
// write. Must be called with singleflight coalescing; takes the shard's
// write lock.
func (f *FileCache) storeAtomic(ctx context.Context, key string, value []byte) error {
	shard := f.shardFor(key)
	shard.Lock()
	defer shard.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	fileName := f.getFileName(key)
	if err := f.fs.MkdirAll(filepath.Dir(fileName), 0700); err != nil {
		return err
	}

	tmpName := fileName + ".tmp"
	if err := afero.WriteFile(f.fs, tmpName, value, 0600); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		_ = f.fs.Remove(tmpName)
		return err
	}

	if err := f.fs.Rename(tmpName, fileName); err != nil {
		_ = f.fs.Remove(tmpName)
		return err
	}

	return nil
}

func (f *FileCache) Load(ctx context.Context, key string) (value []byte, exist bool, err error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}

	// RLock covers open+read+close so a concurrent Store rename cannot
	// interleave with the read; rename atomicity keeps the FD valid.
	shard := f.shardFor(key)
	shard.RLock()
	defer shard.RUnlock()

	if err := ctx.Err(); err != nil {
		return nil, false, err
	}

	r, ok, err := f.open(key)
	if err != nil || !ok {
		return nil, ok, err
	}
	defer r.Close()

	value, err = io.ReadAll(r)
	if err != nil {
		return nil, false, err
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	return value, true, nil
}

func (f *FileCache) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	shard := f.shardFor(key)
	shard.Lock()
	defer shard.Unlock()

	fileName := f.getFileName(key)
	if err := f.fs.Remove(fileName); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (f *FileCache) open(key string) (afero.File, bool, error) {
	fileName := f.getFileName(key)
	file, err := f.fs.Open(fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return file, true, nil
}

func (f *FileCache) getFileName(key string) string {
	hasher := sha1.New()
	_, _ = hasher.Write([]byte(key))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return fmt.Sprintf("%s/%s/%s", hash[:1], hash[1:3], hash)
}
