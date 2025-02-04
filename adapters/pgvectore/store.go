package pgvectore

import (
	"context"
	"fmt"
	"strings"

	"github.com/Abraxas-365/kbservice/vectorstore"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

// Distance represents the distance calculation method
type Distance string

const (
	Cosine       Distance = "cosine"
	Euclidean    Distance = "euclidean"
	InnerProduct Distance = "inner_product"
)

// IsValid checks if the distance metric is valid
func (d Distance) IsValid() bool {
	switch d {
	case Cosine, Euclidean, InnerProduct:
		return true
	default:
		return false
	}
}

type PGVectorStore struct {
	pool      *pgxpool.Pool
	tableName string
	dimension int
	distance  Distance
}

type Options struct {
	TableName string
	Dimension int
	Distance  Distance
}

// getOperatorAndFunction returns the appropriate operator and index operator class based on distance metric
func (p *PGVectorStore) getOperatorAndFunction() (string, string) {
	switch p.distance {
	case Euclidean:
		return "<->", "vector_l2_ops"
	case InnerProduct:
		return "<#>", "vector_ip_ops"
	default: // Cosine
		return "<=>", "vector_cosine_ops"
	}
}

func NewPGVectorStore(ctx context.Context, connString string, opts Options) (*PGVectorStore, error) {
	if opts.Distance == "" {
		opts.Distance = Cosine
	}

	if !opts.Distance.IsValid() {
		return nil, fmt.Errorf("invalid distance metric: %s", opts.Distance)
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error creating connection pool: %w", err)
	}

	store := &PGVectorStore{
		pool:      pool,
		tableName: opts.TableName,
		dimension: opts.Dimension,
		distance:  opts.Distance,
	}

	return store, nil
}

// InitDB initializes the database schema
func (p *PGVectorStore) InitDB(ctx context.Context, forceRecreate bool) error {
	// Enable pgvector extension
	_, err := p.pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		return fmt.Errorf("error creating vector extension: %w", err)
	}

	// Drop table if forceRecreate is true
	if forceRecreate {
		_, err = p.pool.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", p.tableName))
		if err != nil {
			return fmt.Errorf("error dropping table: %w", err)
		}
	}

	// Create table if not exists
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			metadata JSONB,
			embedding vector(%d),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`, p.tableName, p.dimension)

	_, err = p.pool.Exec(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating table: %w", err)
	}

	// Get the appropriate operator class for the index
	_, opClass := p.getOperatorAndFunction()

	// Create index for vector similarity search
	indexSQL := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %s_embedding_idx 
		ON %s 
		USING ivfflat (embedding %s)
		WITH (lists = 100)
	`, p.tableName, p.tableName, opClass)

	_, err = p.pool.Exec(ctx, indexSQL)
	if err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}

	return nil
}

func (p *PGVectorStore) AddDocuments(ctx context.Context, docs []vectorstore.Document, vectors [][]float32) error {
	batch := &pgx.Batch{}

	insertSQL := fmt.Sprintf(`
		INSERT INTO %s (content, metadata, embedding)
		VALUES ($1, $2, $3::vector)
	`, p.tableName)

	for i, doc := range docs {
		// Convert the vector to a PostgreSQL array format
		vectorStr := formatVectorForPG(vectors[i])
		batch.Queue(insertSQL, doc.PageContent, doc.Metadata, vectorStr)
	}

	results := p.pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(docs); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("error inserting document %d: %w", i, err)
		}
	}

	return nil
}

func (p *PGVectorStore) SimilaritySearch(ctx context.Context, vector []float32, limit int, filter vectorstore.Filter) ([]vectorstore.Document, error) {
	operator, _ := p.getOperatorAndFunction()

	// Format vector for PostgreSQL
	vectorStr := formatVectorForPG(vector)

	// Build the query with filters
	whereClause := ""
	args := []interface{}{vectorStr, limit}
	if len(filter) > 0 {
		conditions := make([]string, 0)
		for key, value := range filter {
			args = append(args, value)
			conditions = append(conditions, fmt.Sprintf("metadata->>'%s' = $%d", key, len(args)))
		}
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Adjust score calculation based on distance metric
	scoreExpr := ""
	switch p.distance {
	case Cosine:
		scoreExpr = fmt.Sprintf("1 - (embedding %s $1::vector)", operator)
	case InnerProduct:
		scoreExpr = fmt.Sprintf("(embedding %s $1::vector) * -1", operator)
	case Euclidean:
		scoreExpr = fmt.Sprintf("1 / (1 + (embedding %s $1::vector))", operator)
	}

	query := fmt.Sprintf(`
		SELECT 
			content,
			metadata,
			%s as similarity
		FROM %s
		%s
		ORDER BY embedding %s $1::vector
		LIMIT $2
	`, scoreExpr, p.tableName, whereClause, operator)

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing similarity search: %w", err)
	}
	defer rows.Close()

	var docs []vectorstore.Document
	for rows.Next() {
		var doc vectorstore.Document
		err := rows.Scan(&doc.PageContent, &doc.Metadata, &doc.Score)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		docs = append(docs, doc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return docs, nil
}

// formatVectorForPG converts a float32 slice to a PostgreSQL vector format
func formatVectorForPG(vector []float32) string {
	var b strings.Builder
	b.WriteString("[")
	for i, v := range vector {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(fmt.Sprintf("%.9f", float64(v))) // Use more precision
	}
	b.WriteString("]")
	return b.String()
}

func (p *PGVectorStore) Delete(ctx context.Context, filter vectorstore.Filter) error {
	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	i := 1

	for key, value := range filter {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf("metadata->>'%s' = $%d", key, i))
		i++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf("DELETE FROM %s %s", p.tableName, whereClause)

	_, err := p.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error deleting documents: %w", err)
	}

	return nil
}

// Close closes the database connection pool
func (p *PGVectorStore) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}
