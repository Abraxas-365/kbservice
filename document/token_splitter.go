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

	tokens := ts.encoding.Encode(text, nil, nil)
	if len(tokens) == 0 {
		return nil, nil
	}

	var chunks []string
	start := 0

	for start < len(tokens) {
		end := start + ts.TokensPerChunk
		if end > len(tokens) {
			end = len(tokens)
		}

		chunkTokens := tokens[start:end]
		chunk := ts.encoding.Decode(chunkTokens)
		chunks = append(chunks, chunk)

		start = end - ts.ChunkOverlap
		if start < 0 {
			start = 0
		}
		if start >= len(tokens) {
			break
		}
	}

	return chunks, nil
}
