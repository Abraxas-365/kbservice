package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Abraxas-365/kbservice/adapters/openai"
	"github.com/Abraxas-365/kbservice/adapters/pgvectore"
	"github.com/Abraxas-365/kbservice/adapters/web/websource"
	"github.com/Abraxas-365/kbservice/document"
	"github.com/Abraxas-365/kbservice/kb"
)

func main() {
	ctx := context.Background()

	// Initialize components
	embedder := openai.NewOpenAIEmbedder(os.Getenv("OPENAI_API_KEY"))

	store, err := pgvectore.NewPGVectorStore(ctx,
		os.Getenv("DATABASE_URL"),
		pgvectore.Options{
			TableName: "documents",
			Dimension: 1536,
			Distance:  pgvectore.Cosine,
		},
	)
	if err != nil {
		log.Fatalf("Failed to create vector store: %v", err)
	}

	// Create character splitter
	splitter := document.NewCharacterSplitter(
		120, // chunk size
		50,  // chunk overlap
		" ", // separator
	)

	// Create knowledge base
	knowledgeBase, err := kb.New(
		embedder,
		store,
		splitter,
	)
	if err != nil {
		log.Fatalf("Failed to create knowledge base: %v", err)
	}
	defer knowledgeBase.Close()

	// Initialize the store
	if err := knowledgeBase.InitStore(ctx, true); err != nil {
		log.Fatalf("Failed to init store: %v", err)
	}

	// Example 1: Loading from single URL
	fmt.Println("=== Loading Single URL ===")
	singleSource := websource.NewWebSource(
		[]string{"https://www.iana.org/help/example-domains"},
		10*time.Second,
	)

	// Use Sync to process the documents
	if err := knowledgeBase.Sync(ctx, singleSource); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Single URL processed successfully")

	// Example 2: Similarity search
	fmt.Println("\n=== Similarity Search ===")
	query := "what is described in RFC 2606"
	results, err := knowledgeBase.SimilaritySearch(ctx, query, 2, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d relevant documents for query: %s\n", len(results), query)
	for i, doc := range results {
		fmt.Printf("%d. Score: %.4f\n   Content: %s\n", i+1, doc.Score, doc.PageContent)
	}
}
