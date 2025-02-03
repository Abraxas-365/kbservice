package websource

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/Abraxas-365/kbservice/datasource"
)

type WebSource struct {
	urls    []string
	client  *http.Client
	timeout time.Duration
}

func NewWebSource(urls []string, timeout time.Duration) *WebSource {
	return &WebSource{
		urls:    urls,
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (w *WebSource) Load(ctx context.Context, opts ...datasource.Option) ([]datasource.Document, error) {
	options := &datasource.LoadOptions{}
	for _, opt := range opts {
		opt(options)
	}

	var documents []datasource.Document

	for _, url := range w.urls {
		if options.MaxItems > 0 && len(documents) >= options.MaxItems {
			break
		}

		metadata := map[string]interface{}{
			"url": url,
		}

		if options.Filter != nil && !options.Filter(metadata) {
			continue
		}

		content, err := w.fetchURL(ctx, url)
		if err != nil {
			return nil, err
		}

		doc := datasource.Document{
			Content:  content,
			Metadata: metadata,
			Source:   url,
		}

		documents = append(documents, doc)
	}

	return documents, nil
}

func (w *WebSource) fetchURL(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", &datasource.DataSourceError{
			Source:  "web",
			Op:      "fetchURL",
			Err:     err,
			Code:    datasource.ErrCodeInvalidSource,
			Message: "invalid URL",
		}
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return "", &datasource.DataSourceError{
			Source:  "web",
			Op:      "fetchURL",
			Err:     err,
			Code:    datasource.ErrCodeInternal,
			Message: "failed to fetch URL",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", &datasource.DataSourceError{
			Source:  "web",
			Op:      "fetchURL",
			Code:    datasource.ErrCodeNotFound,
			Message: "failed to fetch URL: " + resp.Status,
		}
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &datasource.DataSourceError{
			Source:  "web",
			Op:      "fetchURL",
			Err:     err,
			Code:    datasource.ErrCodeInternal,
			Message: "failed to read response body",
		}
	}

	return string(content), nil
}
