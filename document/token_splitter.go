package document

import (
	"fmt"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

type TiktokenSplitter struct {
	TokensPerChunk int
	ChunkOverlap   int
	Model          string
	encoding       *tiktoken.Tiktoken
}

// getEncodingForModel returns the appropriate encoding name for a given model
func getEncodingForModel(model string) string {
	// GPT-4 Preview models
	if strings.HasPrefix(model, "gpt-4o") {
		return "o200k_base"
	}

	// GPT-4 and GPT-3.5 models
	if strings.HasPrefix(model, "gpt-4") ||
		strings.HasPrefix(model, "gpt-3.5-turbo") ||
		model == "text-embedding-ada-002" ||
		model == "text-embedding-3-small" ||
		model == "text-embedding-3-large" {
		return "cl100k_base"
	}

	// Codex and certain Davinci models
	if strings.HasPrefix(model, "code-") ||
		model == "text-davinci-002" ||
		model == "text-davinci-003" {
		return "p50k_base"
	}

	// GPT-3 models
	if strings.HasPrefix(model, "text-davinci-001") ||
		strings.HasPrefix(model, "text-curie-001") ||
		strings.HasPrefix(model, "text-babbage-001") ||
		strings.HasPrefix(model, "text-ada-001") ||
		model == "davinci" ||
		model == "curie" ||
		model == "babbage" ||
		model == "ada" ||
		strings.HasPrefix(model, "text-similarity-") ||
		strings.HasPrefix(model, "text-search-") ||
		strings.HasPrefix(model, "code-search-") {
		return "r50k_base"
	}

	// Default to cl100k_base if model is unknown
	return "cl100k_base"
}

func NewTiktokenSplitter(tokensPerChunk int, chunkOverlap int, model string) (*TiktokenSplitter, error) {
	// Validate parameters
	if tokensPerChunk <= 0 {
		return nil, &SplitterError{
			Op:      "new_tiktoken_splitter",
			Message: "tokensPerChunk must be positive",
			Err:     fmt.Errorf("invalid tokensPerChunk: %d", tokensPerChunk),
		}
	}

	if chunkOverlap < 0 {
		return nil, &SplitterError{
			Op:      "new_tiktoken_splitter",
			Message: "chunkOverlap must be non-negative",
			Err:     fmt.Errorf("invalid chunkOverlap: %d", chunkOverlap),
		}
	}

	if chunkOverlap >= tokensPerChunk {
		return nil, &SplitterError{
			Op:      "new_tiktoken_splitter",
			Message: "chunkOverlap must be less than tokensPerChunk",
			Err:     fmt.Errorf("overlap %d >= chunk size %d", chunkOverlap, tokensPerChunk),
		}
	}

	encodingName := getEncodingForModel(model)
	encoding, err := tiktoken.GetEncoding(encodingName)
	if err != nil {
		return nil, &SplitterError{
			Op:      "new_tiktoken_splitter",
			Message: fmt.Sprintf("failed to get %s encoding for model %s", encodingName, model),
			Err:     err,
		}
	}

	return &TiktokenSplitter{
		TokensPerChunk: tokensPerChunk,
		ChunkOverlap:   chunkOverlap,
		Model:          model,
		encoding:       encoding,
	}, nil
}

func (ts *TiktokenSplitter) SplitText(text string) ([]string, error) {
	if text == "" {
		return nil, nil
	}

	// Get tokens for the text
	tokens := ts.encoding.Encode(text, nil, nil)
	if len(tokens) == 0 {
		return nil, nil
	}

	var chunks []string
	start := 0

	// Add safety check for infinite loop
	maxIterations := len(tokens) // Maximum possible chunks
	iteration := 0

	for start < len(tokens) {
		// Safety check
		iteration++
		if iteration > maxIterations {
			return nil, &SplitterError{
				Op:      "split_text",
				Message: "infinite loop detected in token splitting",
				Err:     fmt.Errorf("exceeded maximum iterations of %d", maxIterations),
			}
		}

		// Calculate end position
		end := start + ts.TokensPerChunk
		if end > len(tokens) {
			end = len(tokens)
		}

		// Create chunk
		chunkTokens := tokens[start:end]
		chunk := ts.encoding.Decode(chunkTokens)
		chunks = append(chunks, chunk)

		// Calculate next start position and ensure forward progress
		start = end - ts.ChunkOverlap
		if end == len(tokens) || start >= end {
			break
		}
	}

	return chunks, nil
}

func (ts *TiktokenSplitter) SplitDocuments(docs []Document) ([]Document, error) {
	var result []Document

	for _, doc := range docs {
		chunks, err := ts.SplitText(doc.PageContent)
		if err != nil {
			return nil, &SplitterError{
				Op:      "split_documents",
				Message: "failed to split document text",
				Err:     err,
			}
		}

		for _, chunk := range chunks {
			// Create a new document for each chunk with the same metadata
			newDoc := Document{
				PageContent: chunk,
				Metadata:    doc.Metadata,
			}
			result = append(result, newDoc)
		}
	}

	return result, nil
}
