// examples/basic/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Abraxas-365/kbservice/adapters/openai"
	"github.com/Abraxas-365/kbservice/adapters/pgvectore"
	"github.com/Abraxas-365/kbservice/document"
	"github.com/Abraxas-365/kbservice/vectorstore"
)

func main() {
	ctx := context.Background()

	// Initialize components
	embedder := openai.NewOpenAIEmbedder(os.Getenv("OPENAI_API_KEY"))

	pgStore, err := pgvectore.NewPGVectorStore(ctx,
		os.Getenv("DATABASE_URL"),
		pgvectore.Options{
			TableName: "documents",
			Dimension: 1536,
			Distance:  pgvectore.Cosine,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize database (recreate table if needed)
	err = pgStore.InitDB(ctx, true)
	if err != nil {
		log.Fatal(err)
	}

	// Create vector store
	store := vectorstore.New(
		pgStore,
		embedder,
		vectorstore.WithScoreThreshold(0.7),
	)

	// Sample documents to index
	docs := []document.Document{
		{
			PageContent: "The quick brown fox jumps over the lazy dog",
			Metadata: map[string]interface{}{
				"type": "sentence",
				"tag":  "animals",
			},
		},
		{
			PageContent: "OpenAI is a leading artificial intelligence research laboratory",
			Metadata: map[string]interface{}{
				"type": "fact",
				"tag":  "technology",
			},
		},
		{
			PageContent: "Python is a popular programming language known for its simplicity",
			Metadata: map[string]interface{}{
				"type": "fact",
				"tag":  "programming",
			},
		},
		{
			PageContent: "Machine learning is a subset of artificial intelligence",
			Metadata: map[string]interface{}{
				"type": "fact",
				"tag":  "technology",
			},
		},
		{
			PageContent: "The Earth revolves around the Sun in approximately 365 days",
			Metadata: map[string]interface{}{
				"type": "fact",
				"tag":  "science",
			},
		},
	}

	// Add documents to the vector store
	fmt.Println("Indexing documents...")
	err = store.AddDocuments(ctx, docs)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Documents indexed successfully")

	// Perform different searches
	searches := []string{
		"Tell me about artificial intelligence",
		"What do you know about animals?",
		"Tell me about programming",
		"Facts about space",
	}

	for _, query := range searches {
		fmt.Printf("\nSearching for: %s\n", query)
		fmt.Println("----------------------------------------")

		results, err := store.SimilaritySearch(ctx, query, 3, nil)
		if err != nil {
			log.Fatal(err)
		}

		for _, doc := range results {
			fmt.Printf("Score: %.4f\n", doc.Score)
			fmt.Printf("Content: %s\n", doc.PageContent)
			fmt.Printf("Metadata: %v\n", doc.Metadata)
			fmt.Println("----------------------------------------")
		}
	}
}
