package s3

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/Abraxas-365/kbservice/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Store struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}

func NewS3Store(client *s3.Client, bucket string) *S3Store {
	return &S3Store{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bucket:        bucket,
	}
}

func (s *S3Store) Put(ctx context.Context, key string, data io.Reader, options ...storage.PutOption) error {
	opts := &storage.PutOptions{}
	for _, opt := range options {
		opt(opts)
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   data,
	}

	if opts.ContentType != "" {
		input.ContentType = aws.String(opts.ContentType)
	}

	if opts.CacheControl != "" {
		input.CacheControl = aws.String(opts.CacheControl)
	}

	if opts.ContentEncoding != "" {
		input.ContentEncoding = aws.String(opts.ContentEncoding)
	}

	if opts.ContentDisposition != "" {
		input.ContentDisposition = aws.String(opts.ContentDisposition)
	}

	if opts.Metadata != nil {
		input.Metadata = opts.Metadata
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return storage.NewStorageError("Put", key, err, storage.ErrCodeInternal, "failed to put object")
	}

	return nil
}

func (s *S3Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		var nsk *types.NoSuchKey
		if err != nil && errors.As(err, &nsk) {
			return nil, storage.NewStorageError("Get", key, err, storage.ErrCodeNotFound, "object not found")
		}
		return nil, storage.NewStorageError("Get", key, err, storage.ErrCodeInternal, "failed to get object")
	}

	return result.Body, nil
}

func (s *S3Store) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return storage.NewStorageError("Delete", key, err, storage.ErrCodeInternal, "failed to delete object")
	}

	return nil
}

func (s *S3Store) List(ctx context.Context, prefix string) ([]storage.ObjectInfo, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	}

	var objects []storage.ObjectInfo

	paginator := s3.NewListObjectsV2Paginator(s.client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, storage.NewStorageError("List", prefix, err, storage.ErrCodeInternal, "failed to list objects")
		}

		for _, obj := range page.Contents {
			object := storage.ObjectInfo{
				Key:          *obj.Key,
				Size:         *obj.Size,
				LastModified: *obj.LastModified,
				ETag:         *obj.ETag,
			}
			objects = append(objects, object)
		}
	}

	return objects, nil
}

func (s *S3Store) Exists(ctx context.Context, key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.HeadObject(ctx, input)
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return false, nil
		}
		return false, storage.NewStorageError("Exists", key, err, storage.ErrCodeInternal, "failed to check object existence")
	}

	return true, nil
}

func (s *S3Store) GetPresignedPutURL(ctx context.Context, key string, expires time.Duration, options ...storage.PresignedPutOption) (storage.PresignedURL, error) {
	opts := &storage.PresignedPutOptions{}
	for _, opt := range options {
		opt(opts)
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	if opts.ContentType != "" {
		input.ContentType = aws.String(opts.ContentType)
	}

	if opts.ContentLength != nil {
		input.ContentLength = opts.ContentLength
	}

	if opts.CacheControl != "" {
		input.CacheControl = aws.String(opts.CacheControl)
	}

	if opts.ContentEncoding != "" {
		input.ContentEncoding = aws.String(opts.ContentEncoding)
	}

	if opts.ContentDisposition != "" {
		input.ContentDisposition = aws.String(opts.ContentDisposition)
	}

	if opts.Metadata != nil {
		input.Metadata = opts.Metadata
	}

	presignedReq, err := s.presignClient.PresignPutObject(ctx, input,
		s3.WithPresignExpires(expires))
	if err != nil {
		return storage.PresignedURL{}, storage.NewStorageError("GetPresignedPutURL", key, err, storage.ErrCodeInternal, "failed to generate presigned URL")
	}

	// Convert http.Header to map[string]string
	headers := make(map[string]string)
	for k, v := range presignedReq.SignedHeader {
		if len(v) > 0 {
			headers[k] = v[0] // Take the first value if multiple values exist
		}
	}

	return storage.PresignedURL{
		URL:     presignedReq.URL,
		Method:  string(presignedReq.Method),
		Headers: headers,
	}, nil
}

func (s *S3Store) GetPresignedGetURL(ctx context.Context, key string, expires time.Duration) (storage.PresignedURL, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	presignedReq, err := s.presignClient.PresignGetObject(ctx, input,
		s3.WithPresignExpires(expires))
	if err != nil {
		return storage.PresignedURL{}, storage.NewStorageError("GetPresignedGetURL", key, err, storage.ErrCodeInternal, "failed to generate presigned URL")
	}

	// Convert http.Header to map[string]string
	headers := make(map[string]string)
	for k, v := range presignedReq.SignedHeader {
		if len(v) > 0 {
			headers[k] = v[0] // Take the first value if multiple values exist
		}
	}

	return storage.PresignedURL{
		URL:     presignedReq.URL,
		Method:  string(presignedReq.Method),
		Headers: headers,
	}, nil
}
