package document

import (
	"strings"
	"testing"
)

func TestNewTiktokenSplitter(t *testing.T) {
	tests := []struct {
		name           string
		tokensPerChunk int
		chunkOverlap   int
		model          string
		wantErr        bool
		errMessage     string
	}{
		{
			name:           "Valid parameters",
			tokensPerChunk: 100,
			chunkOverlap:   20,
			model:          "text-embedding-3-small",
			wantErr:        false,
		},
		{
			name:           "Zero tokens per chunk",
			tokensPerChunk: 0,
			chunkOverlap:   20,
			model:          "text-embedding-3-small",
			wantErr:        true,
			errMessage:     "tokensPerChunk must be positive",
		},
		{
			name:           "Negative chunk overlap",
			tokensPerChunk: 100,
			chunkOverlap:   -1,
			model:          "text-embedding-3-small",
			wantErr:        true,
			errMessage:     "chunkOverlap must be non-negative",
		},
		{
			name:           "Overlap larger than chunk size",
			tokensPerChunk: 100,
			chunkOverlap:   150,
			model:          "text-embedding-3-small",
			wantErr:        true,
			errMessage:     "chunkOverlap must be less than tokensPerChunk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			splitter, err := NewTiktokenSplitter(tt.tokensPerChunk, tt.chunkOverlap, tt.model)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewTiktokenSplitter() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.errMessage) {
					t.Errorf("NewTiktokenSplitter() error = %v, want error containing %v", err, tt.errMessage)
				}
				return
			}
			if err != nil {
				t.Errorf("NewTiktokenSplitter() unexpected error = %v", err)
				return
			}
			if splitter == nil {
				t.Error("NewTiktokenSplitter() returned nil splitter")
			}
		})
	}
}

func TestTiktokenSplitter_SplitText(t *testing.T) {
	// Create a long text for testing
	longText := strings.Repeat("This is a test sentence. ", 100)
	shortText := "This is a short test sentence."

	tests := []struct {
		name           string
		text           string
		tokensPerChunk int
		chunkOverlap   int
		model          string
		wantChunks     int
		wantErr        bool
	}{
		{
			name:           "Empty text",
			text:           "",
			tokensPerChunk: 100,
			chunkOverlap:   20,
			model:          "text-embedding-3-small",
			wantChunks:     0,
			wantErr:        false,
		},
		{
			name:           "Short text within chunk size",
			text:           shortText,
			tokensPerChunk: 100,
			chunkOverlap:   20,
			model:          "text-embedding-3-small",
			wantChunks:     1,
			wantErr:        false,
		},
		{
			name:           "Long text with multiple chunks",
			text:           longText,
			tokensPerChunk: 50,
			chunkOverlap:   10,
			model:          "text-embedding-3-small",
			wantChunks:     10, // Approximate, actual number depends on tokenization
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			splitter, err := NewTiktokenSplitter(tt.tokensPerChunk, tt.chunkOverlap, tt.model)
			if err != nil {
				t.Fatalf("Failed to create splitter: %v", err)
			}

			chunks, err := splitter.SplitText(tt.text)
			if tt.wantErr {
				if err == nil {
					t.Error("SplitText() error = nil, wantErr true")
				}
				return
			}
			if err != nil {
				t.Errorf("SplitText() unexpected error = %v", err)
				return
			}

			if tt.wantChunks > 0 && len(chunks) == 0 {
				t.Error("SplitText() returned no chunks when chunks were expected")
			}

			// Verify chunk contents
			if len(chunks) > 0 {
				// Check that no chunk is empty
				for i, chunk := range chunks {
					if chunk == "" {
						t.Errorf("Chunk %d is empty", i)
					}
				}

				// Check that chunks can be rejoined (approximately)
				if tt.text == shortText {
					joined := strings.Join(chunks, "")
					if !strings.Contains(joined, shortText) {
						t.Error("Joined chunks do not contain original text")
					}
				}
			}
		})
	}
}

func TestTiktokenSplitter_SplitDocuments(t *testing.T) {
	docs := []Document{
		{
			PageContent: "This is document 1.",
			Metadata: map[string]interface{}{
				"source": "test1",
			},
		},
		{
			PageContent: strings.Repeat("This is document 2 with longer content. ", 50),
			Metadata: map[string]interface{}{
				"source": "test2",
			},
		},
	}

	tests := []struct {
		name           string
		docs           []Document
		tokensPerChunk int
		chunkOverlap   int
		model          string
		wantErr        bool
	}{
		{
			name:           "Split multiple documents",
			docs:           docs,
			tokensPerChunk: 50,
			chunkOverlap:   10,
			model:          "text-embedding-3-small",
			wantErr:        false,
		},
		{
			name:           "Empty document list",
			docs:           []Document{},
			tokensPerChunk: 50,
			chunkOverlap:   10,
			model:          "text-embedding-3-small",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			splitter, err := NewTiktokenSplitter(tt.tokensPerChunk, tt.chunkOverlap, tt.model)
			if err != nil {
				t.Fatalf("Failed to create splitter: %v", err)
			}

			splitDocs, err := splitter.SplitDocuments(tt.docs)
			if tt.wantErr {
				if err == nil {
					t.Error("SplitDocuments() error = nil, wantErr true")
				}
				return
			}
			if err != nil {
				t.Errorf("SplitDocuments() unexpected error = %v", err)
				return
			}

			// Verify that metadata is preserved
			for _, doc := range splitDocs {
				if doc.Metadata == nil {
					t.Error("Split document has nil metadata")
					continue
				}
				if doc.PageContent == "" {
					t.Error("Split document has empty content")
				}
			}

			// Verify that longer documents produce more chunks
			if len(tt.docs) > 0 {
				_ = tt.docs[1].PageContent
				longDocChunks := 0
				for _, doc := range splitDocs {
					if doc.Metadata["source"] == "test2" {
						longDocChunks++
					}
				}
				if longDocChunks <= 1 {
					t.Errorf("Expected multiple chunks for long document, got %d", longDocChunks)
				}
			}
		})
	}
}

func TestGetEncodingForModel(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected string
	}{
		{
			name:     "GPT-4",
			model:    "gpt-4",
			expected: "cl100k_base",
		},
		{
			name:     "GPT-3.5 Turbo",
			model:    "gpt-3.5-turbo",
			expected: "cl100k_base",
		},
		{
			name:     "Text Embedding 3 Small",
			model:    "text-embedding-3-small",
			expected: "cl100k_base",
		},
		{
			name:     "Davinci",
			model:    "text-davinci-002",
			expected: "p50k_base",
		},
		{
			name:     "Unknown model",
			model:    "unknown-model",
			expected: "cl100k_base",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getEncodingForModel(tt.model)
			if result != tt.expected {
				t.Errorf("getEncodingForModel() = %v, want %v", result, tt.expected)
			}
		})
	}
}
