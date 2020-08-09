package blobfs

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strings"

	"gocloud.dev/blob"
)

type BlobFS struct {
	storageURL string
}

func New(storageURL string) *BlobFS {
	return &BlobFS{storageURL: storageURL}
}

func NewInMemory() *BlobFS {
	return New("mem://")
}

func (fs *BlobFS) WriteFile(ctx context.Context, path string, data []byte) error {
	dir, filename := filepath.Split(path)
	bucket, err := fs.openBucket(ctx, dir)
	if err != nil {
		return err
	}
	defer bucket.Close()

	w, err := bucket.NewWriter(ctx, filename, nil)
	if err != nil {
		return err
	}
	_, writeErr := w.Write(data)
	// Always check the return value of Close when writing.
	closeErr := w.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return nil
}

func (fs *BlobFS) ReadFile(ctx context.Context, path string) ([]byte, error) {
	dir, filename := filepath.Split(path)
	bucket, err := fs.openBucket(ctx, dir)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()
	// Open the key "foo.txt" for reading with the default options.
	r, err := bucket.NewReader(ctx, filename, nil)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (fs *BlobFS) Exists(ctx context.Context, path string) (bool, error) {
	dir, filename := filepath.Split(path)
	bucket, err := fs.openBucket(ctx, dir)
	if err != nil {
		return false, err
	}
	defer bucket.Close()

	return bucket.Exists(context.TODO(), filename)
}

func (fs *BlobFS) SignedURL(ctx context.Context, path string) (string, error) {
	dir, filename := filepath.Split(path)
	bucket, err := fs.openBucket(ctx, dir)
	if err != nil {
		return "", err
	}
	defer bucket.Close()

	return bucket.SignedURL(ctx, filename, nil)
}

func (fs *BlobFS) openBucket(ctx context.Context, dir string) (*blob.Bucket, error) {
	bucket, err := blob.OpenBucket(ctx, fs.storageURL)
	if err != nil {
		return nil, err
	}
	return blob.PrefixedBucket(bucket, strings.Trim(dir, "/")+"/"), nil
}
