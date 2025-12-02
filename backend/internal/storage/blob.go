// Package storage provides a portable blob storage abstraction using GoCloud CDK.
// It supports local file storage (file://) for development and S3 (s3://) for production.
package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob" // Register file:// scheme
	_ "gocloud.dev/blob/s3blob"   // Register s3:// scheme
)

// BlobStorage provides a portable interface for blob storage operations.
type BlobStorage struct {
	bucket *blob.Bucket
	prefix string
	url    string
}

// NewBlobStorage creates a new blob storage client from a URL.
// Supported URL schemes:
//   - file:///path/to/dir - Local filesystem storage
//   - s3://bucket-name?region=us-east-1 - AWS S3 storage
func NewBlobStorage(ctx context.Context, storageURL, prefix string) (*BlobStorage, error) {
	if storageURL == "" {
		storageURL = "file://./exports"
	}

	// Parse the URL to validate it
	parsedURL, err := url.Parse(storageURL)
	if err != nil {
		return nil, fmt.Errorf("invalid storage URL: %w", err)
	}

	// For file:// URLs, ensure the directory exists
	if parsedURL.Scheme == "file" {
		path := parsedURL.Path
		if path == "" {
			path = parsedURL.Opaque
		}
		// fileblob requires the path to exist
		if err := ensureDir(path); err != nil {
			return nil, fmt.Errorf("failed to create storage directory: %w", err)
		}
	}

	bucket, err := blob.OpenBucket(ctx, storageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket: %w", err)
	}

	log.Info().
		Str("url", storageURL).
		Str("prefix", prefix).
		Msg("Blob storage initialized")

	return &BlobStorage{
		bucket: bucket,
		prefix: prefix,
		url:    storageURL,
	}, nil
}

// Close closes the blob storage connection.
func (b *BlobStorage) Close() error {
	if b.bucket != nil {
		return b.bucket.Close()
	}
	return nil
}

// Write writes data to a blob at the specified key.
func (b *BlobStorage) Write(ctx context.Context, key string, data []byte, contentType string) error {
	fullKey := b.fullKey(key)

	opts := &blob.WriterOptions{}
	if contentType != "" {
		opts.ContentType = contentType
	}

	writer, err := b.bucket.NewWriter(ctx, fullKey, opts)
	if err != nil {
		return fmt.Errorf("failed to create writer for %s: %w", fullKey, err)
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return fmt.Errorf("failed to write to %s: %w", fullKey, err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer for %s: %w", fullKey, err)
	}

	log.Debug().
		Str("key", fullKey).
		Int("size", len(data)).
		Str("content_type", contentType).
		Msg("Blob written")

	return nil
}

// WriteStream writes data from a reader to a blob at the specified key.
func (b *BlobStorage) WriteStream(ctx context.Context, key string, reader io.Reader, contentType string) error {
	fullKey := b.fullKey(key)

	opts := &blob.WriterOptions{}
	if contentType != "" {
		opts.ContentType = contentType
	}

	writer, err := b.bucket.NewWriter(ctx, fullKey, opts)
	if err != nil {
		return fmt.Errorf("failed to create writer for %s: %w", fullKey, err)
	}

	if _, err := io.Copy(writer, reader); err != nil {
		writer.Close()
		return fmt.Errorf("failed to write stream to %s: %w", fullKey, err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer for %s: %w", fullKey, err)
	}

	log.Debug().
		Str("key", fullKey).
		Str("content_type", contentType).
		Msg("Blob stream written")

	return nil
}

// Read reads data from a blob at the specified key.
func (b *BlobStorage) Read(ctx context.Context, key string) ([]byte, error) {
	fullKey := b.fullKey(key)

	data, err := b.bucket.ReadAll(ctx, fullKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", fullKey, err)
	}

	log.Debug().
		Str("key", fullKey).
		Int("size", len(data)).
		Msg("Blob read")

	return data, nil
}

// ReadStream returns a reader for the blob at the specified key.
func (b *BlobStorage) ReadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	fullKey := b.fullKey(key)

	reader, err := b.bucket.NewReader(ctx, fullKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open reader for %s: %w", fullKey, err)
	}

	return reader, nil
}

// Delete deletes a blob at the specified key.
func (b *BlobStorage) Delete(ctx context.Context, key string) error {
	fullKey := b.fullKey(key)

	if err := b.bucket.Delete(ctx, fullKey); err != nil {
		return fmt.Errorf("failed to delete %s: %w", fullKey, err)
	}

	log.Debug().
		Str("key", fullKey).
		Msg("Blob deleted")

	return nil
}

// Exists checks if a blob exists at the specified key.
func (b *BlobStorage) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := b.fullKey(key)

	exists, err := b.bucket.Exists(ctx, fullKey)
	if err != nil {
		return false, fmt.Errorf("failed to check existence of %s: %w", fullKey, err)
	}

	return exists, nil
}

// List lists all blobs with the given prefix.
func (b *BlobStorage) List(ctx context.Context, prefix string) ([]string, error) {
	fullPrefix := b.fullKey(prefix)

	var keys []string
	iter := b.bucket.List(&blob.ListOptions{Prefix: fullPrefix})

	for {
		obj, err := iter.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list blobs: %w", err)
		}
		// Remove the storage prefix from the key
		key := strings.TrimPrefix(obj.Key, b.prefix)
		key = strings.TrimPrefix(key, "/")
		keys = append(keys, key)
	}

	return keys, nil
}

// GetURL returns the full URL for a blob key.
// For S3, this returns the s3:// URL. For file://, this returns the local path.
func (b *BlobStorage) GetURL(key string) string {
	fullKey := b.fullKey(key)

	parsedURL, err := url.Parse(b.url)
	if err != nil {
		return fullKey
	}

	switch parsedURL.Scheme {
	case "file":
		path := parsedURL.Path
		if path == "" {
			path = parsedURL.Opaque
		}
		return filepath.Join(path, fullKey)
	case "s3":
		return fmt.Sprintf("s3://%s/%s", parsedURL.Host, fullKey)
	default:
		return fullKey
	}
}

// fullKey returns the full key including the prefix.
func (b *BlobStorage) fullKey(key string) string {
	if b.prefix == "" {
		return key
	}
	return filepath.Join(b.prefix, key)
}

// ContentTypeForExtension returns the appropriate content type for a file extension.
func ContentTypeForExtension(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".svg":
		return "image/svg+xml"
	case ".csv":
		return "text/csv"
	case ".json":
		return "application/json"
	case ".html":
		return "text/html"
	default:
		return "application/octet-stream"
	}
}

