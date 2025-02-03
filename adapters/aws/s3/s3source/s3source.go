package s3source

import (
	"context"
	"io"
	"path/filepath"

	"github.com/Abraxas-365/kbservice/datasource"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Source struct {
	client *s3.Client
	bucket string
	prefix string
}

func NewS3Source(client *s3.Client, bucket, prefix string) *S3Source {
	return &S3Source{
		client: client,
		bucket: bucket,
		prefix: prefix,
	}
}

func (s *S3Source) Load(ctx context.Context, opts ...datasource.Option) ([]datasource.Document, error) {
	options := &datasource.LoadOptions{}
	for _, opt := range opts {
		opt(options)
	}

	input := &s3.ListObjectsV2Input{
		Bucket: &s.bucket,
		Prefix: &s.prefix,
	}

	var documents []datasource.Document
	paginator := s3.NewListObjectsV2Paginator(s.client, input)

	for paginator.HasMorePages() {
		if options.MaxItems > 0 && len(documents) >= options.MaxItems {
			break
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &datasource.DataSourceError{
				Source:  "s3",
				Op:      "Load",
				Err:     err,
				Code:    datasource.ErrCodeInternal,
				Message: "failed to list objects",
			}
		}

		for _, obj := range page.Contents {
			if options.MaxItems > 0 && len(documents) >= options.MaxItems {
				break
			}

			if !options.Recursive && filepath.Dir(*obj.Key) != s.prefix {
				continue
			}

			metadata := map[string]interface{}{
				"key":           *obj.Key,
				"last_modified": *obj.LastModified,
				"size":          obj.Size,
				"etag":          *obj.ETag,
			}

			if options.Filter != nil && !options.Filter(metadata) {
				continue
			}

			content, err := s.getObjectContent(ctx, *obj.Key)
			if err != nil {
				return nil, err
			}

			doc := datasource.Document{
				Content:  content,
				Metadata: metadata,
				Source:   "s3://" + s.bucket + "/" + *obj.Key,
			}

			documents = append(documents, doc)
		}
	}

	return documents, nil
}

func (s *S3Source) getObjectContent(ctx context.Context, key string) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return "", &datasource.DataSourceError{
			Source:  "s3",
			Op:      "getObjectContent",
			Err:     err,
			Code:    datasource.ErrCodeInternal,
			Message: "failed to get object content",
		}
	}
	defer result.Body.Close()

	content, err := io.ReadAll(result.Body)
	if err != nil {
		return "", &datasource.DataSourceError{
			Source:  "s3",
			Op:      "getObjectContent",
			Err:     err,
			Code:    datasource.ErrCodeInternal,
			Message: "failed to read object content",
		}
	}

	return string(content), nil
}
