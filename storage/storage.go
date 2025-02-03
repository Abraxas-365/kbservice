package storage

import (
	"context"
	"io"
	"time"
)

// ObjectInfo represents metadata about a stored object
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
	ContentType  string
	Metadata     map[string]string
}

// DataStore represents a generic interface for object storage operations
type DataStore interface {
	Put(ctx context.Context, key string, data io.Reader, options ...PutOption) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]ObjectInfo, error)
	Exists(ctx context.Context, key string) (bool, error)

	GetPresignedPutURL(ctx context.Context, key string, expires time.Duration, options ...PresignedPutOption) (PresignedURL, error)
	GetPresignedGetURL(ctx context.Context, key string, expires time.Duration) (PresignedURL, error)
}

// PutOption allows customizing Put operations
type PutOption func(*PutOptions)

// PutOptions contains configuration for Put operations
type PutOptions struct {
	ContentType        string
	Metadata           map[string]string
	CacheControl       string
	ContentEncoding    string
	ContentDisposition string
}

// WithContentType sets the content type for the object
func WithContentType(contentType string) PutOption {
	return func(o *PutOptions) {
		o.ContentType = contentType
	}
}

// WithMetadata sets additional metadata for the object
func WithMetadata(metadata map[string]string) PutOption {
	return func(o *PutOptions) {
		o.Metadata = metadata
	}
}

// WithCacheControl sets the Cache-Control header for the object
func WithCacheControl(cacheControl string) PutOption {
	return func(o *PutOptions) {
		o.CacheControl = cacheControl
	}
}

// WithContentEncoding sets the Content-Encoding header for the object
func WithContentEncoding(contentEncoding string) PutOption {
	return func(o *PutOptions) {
		o.ContentEncoding = contentEncoding
	}
}

// WithContentDisposition sets the Content-Disposition header for the object
func WithContentDisposition(contentDisposition string) PutOption {
	return func(o *PutOptions) {
		o.ContentDisposition = contentDisposition
	}
}

// PresignedURL represents a presigned URL with its associated metadata
type PresignedURL struct {
	URL     string
	Method  string
	Headers map[string]string
}

// PresignedPutOption allows customizing presigned Put URL generation
type PresignedPutOption func(*PresignedPutOptions)

// PresignedPutOptions contains configuration for presigned Put URLs
type PresignedPutOptions struct {
	ContentType        string
	ContentLength      *int64
	Metadata           map[string]string
	AllowedExtensions  []string
	CacheControl       string
	ContentEncoding    string
	ContentDisposition string
}

// WithPresignedContentType sets the content type for the presigned URL
func WithPresignedContentType(contentType string) PresignedPutOption {
	return func(o *PresignedPutOptions) {
		o.ContentType = contentType
	}
}

// WithPresignedContentLength sets the expected content length
func WithPresignedContentLength(length int64) PresignedPutOption {
	return func(o *PresignedPutOptions) {
		o.ContentLength = &length
	}
}

// WithPresignedMetadata sets additional metadata for the object
func WithPresignedMetadata(metadata map[string]string) PresignedPutOption {
	return func(o *PresignedPutOptions) {
		o.Metadata = metadata
	}
}

// WithPresignedAllowedExtensions sets the allowed file extensions for upload
func WithPresignedAllowedExtensions(extensions []string) PresignedPutOption {
	return func(o *PresignedPutOptions) {
		o.AllowedExtensions = extensions
	}
}

// WithPresignedCacheControl sets the Cache-Control header for the presigned URL
func WithPresignedCacheControl(cacheControl string) PresignedPutOption {
	return func(o *PresignedPutOptions) {
		o.CacheControl = cacheControl
	}
}

// WithPresignedContentEncoding sets the Content-Encoding header for the presigned URL
func WithPresignedContentEncoding(contentEncoding string) PresignedPutOption {
	return func(o *PresignedPutOptions) {
		o.ContentEncoding = contentEncoding
	}
}

// WithPresignedContentDisposition sets the Content-Disposition header for the presigned URL
func WithPresignedContentDisposition(contentDisposition string) PresignedPutOption {
	return func(o *PresignedPutOptions) {
		o.ContentDisposition = contentDisposition
	}
}
