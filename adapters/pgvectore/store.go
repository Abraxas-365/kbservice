package pgvectore

import (
	"context"
	"fmt"
	"strings"

	"github.com/Abraxas-365/kbservice/document"
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
		return nil, &vectorstore.VectorStoreError{
			Code:    vectorstore.ErrCodeInitFailed,
			Op:      "NewPGVectorStore",
			Store:   "pgvector",
			Message: fmt.Sprintf("invalid distance metric: %s", opts.Distance),
		}
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, &vectorstore.VectorStoreError{
			Code:    vectorstore.ErrCodeInitFailed,
			Op:      "NewPGVectorStore",
			Store:   "pgvector",
			Message: "error parsing connection string",
			Err:     err,
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, &vectorstore.VectorStoreError{
			Code:    vectorstore.ErrCodeInitFailed,
			Op:      "NewPGVectorStore",
			Store:   "pgvector",
			Message: "error creating connection pool",
			Err:     err,
		}
	}

	store := &PGVectorStore{
		pool:      pool,
		tableName: opts.TableName,
		dimension: opts.Dimension,
		distance:  opts.Distance,
	}

	return store, nil
}

func (p *PGVectorStore) InitDB(ctx context.Context, forceRecreate bool) error {
	// Check if table exists
	if !forceRecreate {
		var exists bool
		err := p.pool.QueryRow(ctx,
			"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)",
			p.tableName).Scan(&exists)
		if err == nil && exists {
			return vectorstore.NewDBExistsError("pgvector", nil)
		}
	}

	// Enable pgvector extension
	_, err := p.pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		return vectorstore.NewInitFailedError("pgvector", fmt.Errorf("failed to create vector extension: %w", err))
	}

	// Drop table if forceRecreate is true
	if forceRecreate {
		_, err = p.pool.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", p.tableName))
		if err != nil {
			return vectorstore.NewInitFailedError("pgvector", fmt.Errorf("failed to drop table: %w", err))
		}
	}

	// Create table
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
		return vectorstore.NewInitFailedError("pgvector", fmt.Errorf("failed to create table: %w", err))
	}

	// Create vector similarity index
	_, opClass := p.getOperatorAndFunction()
	vectorIndexSQL := fmt.Sprintf(`
        CREATE INDEX IF NOT EXISTS %s_embedding_idx 
        ON %s 
        USING ivfflat (embedding %s)
        WITH (lists = 100)
    `, p.tableName, p.tableName, opClass)

	_, err = p.pool.Exec(ctx, vectorIndexSQL)
	if err != nil {
		return vectorstore.NewInitFailedError("pgvector", fmt.Errorf("failed to create vector index: %w", err))
	}

	// Create index for source and last_modified lookups
	metadataIndexSQL := fmt.Sprintf(`
        CREATE INDEX IF NOT EXISTS %s_metadata_source_lastmod_idx 
        ON %s ((metadata->>'source'), (metadata->>'last_modified'))
    `, p.tableName, p.tableName)

	_, err = p.pool.Exec(ctx, metadataIndexSQL)
	if err != nil {
		return vectorstore.NewInitFailedError("pgvector", fmt.Errorf("failed to create metadata index: %w", err))
	}

	// Create index for general metadata filters
	filterIndexSQL := fmt.Sprintf(`
        CREATE INDEX IF NOT EXISTS %s_metadata_gin_idx 
        ON %s USING GIN (metadata)
    `, p.tableName, p.tableName)

	_, err = p.pool.Exec(ctx, filterIndexSQL)
	if err != nil {
		return vectorstore.NewInitFailedError("pgvector", fmt.Errorf("failed to create metadata GIN index: %w", err))
	}

	return nil
}

func (p *PGVectorStore) AddDocuments(ctx context.Context, docs []vectorstore.Document, vectors [][]float32) error {
	// Validate vector dimensions
	for _, vec := range vectors {
		if len(vec) != p.dimension {
			return vectorstore.NewInvalidDimensionsError("pgvector", p.dimension, len(vec))
		}
	}

	batch := &pgx.Batch{}
	insertSQL := fmt.Sprintf(`
        INSERT INTO %s (content, metadata, embedding)
        VALUES ($1, $2, $3::vector)
    `, p.tableName)

	for i, doc := range docs {
		vectorStr := formatVectorForPG(vectors[i])
		batch.Queue(insertSQL, doc.PageContent, doc.Metadata, vectorStr)
	}

	results := p.pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(docs); i++ {
		_, err := results.Exec()
		if err != nil {
			return vectorstore.NewAddFailedError("pgvector", fmt.Errorf("failed to insert document %d: %w", i, err))
		}
	}

	return nil
}

func (p *PGVectorStore) SimilaritySearch(ctx context.Context, vector []float32, limit int, filter vectorstore.Filter) ([]vectorstore.Document, error) {
	// Validate vector dimension
	if len(vector) != p.dimension {
		return nil, vectorstore.NewInvalidDimensionsError("pgvector", p.dimension, len(vector))
	}

	// Validate filter
	if err := p.validateFilter(filter); err != nil {
		return nil, vectorstore.NewInvalidFilterError("pgvector", err.Error())
	}

	operator, _ := p.getOperatorAndFunction()
	vectorStr := formatVectorForPG(vector)

	// Build query with filters
	whereClause, args := p.buildWhereClause(filter)
	args = append([]interface{}{vectorStr, limit}, args...)

	scoreExpr := p.buildScoreExpression(operator)
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
		return nil, vectorstore.NewSearchFailedError("pgvector", err)
	}
	defer rows.Close()

	var docs []vectorstore.Document
	for rows.Next() {
		var doc vectorstore.Document
		err := rows.Scan(&doc.PageContent, &doc.Metadata, &doc.Score)
		if err != nil {
			return nil, vectorstore.NewSearchFailedError("pgvector", fmt.Errorf("failed to scan row: %w", err))
		}
		docs = append(docs, doc)
	}

	if err = rows.Err(); err != nil {
		return nil, vectorstore.NewSearchFailedError("pgvector", err)
	}

	return docs, nil
}

func (p *PGVectorStore) buildDeleteWhereClause(filter vectorstore.Filter) (string, []interface{}) {
	if len(filter) == 0 {
		return "", nil
	}

	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	i := 1 // Start from 1 for delete operations

	for key, value := range filter {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf("metadata->>'%s' = $%d", key, i))
		i++
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (p *PGVectorStore) Delete(ctx context.Context, filter vectorstore.Filter) error {
	whereClause, args := p.buildDeleteWhereClause(filter)
	query := fmt.Sprintf("DELETE FROM %s %s", p.tableName, whereClause)

	_, err := p.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

// Helper methods

func (p *PGVectorStore) validateFilter(filter vectorstore.Filter) error {
	if filter == nil {
		return nil
	}

	for key, value := range filter {
		if key == "" {
			return fmt.Errorf("empty key in filter")
		}
		if value == nil {
			return fmt.Errorf("nil value for key %s", key)
		}
	}
	return nil
}

func (p *PGVectorStore) buildWhereClause(filter vectorstore.Filter) (string, []interface{}) {
	if len(filter) == 0 {
		return "", nil
	}

	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	i := 3 // Starting from 3 because $1 and $2 are used for vector and limit

	for key, value := range filter {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf("metadata->>'%s' = $%d", key, i))
		i++
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (p *PGVectorStore) buildScoreExpression(operator string) string {
	switch p.distance {
	case Cosine:
		return fmt.Sprintf("1 - (embedding %s $1::vector)", operator)
	case InnerProduct:
		return fmt.Sprintf("(embedding %s $1::vector) * -1", operator)
	case Euclidean:
		return fmt.Sprintf("1 / (1 + (embedding %s $1::vector))", operator)
	default:
		return fmt.Sprintf("1 - (embedding %s $1::vector)", operator)
	}
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

func (p *PGVectorStore) DocumentExists(ctx context.Context, docs []document.Document) ([]bool, error) {
	exists := make([]bool, len(docs))

	batch := &pgx.Batch{}
	selectSQL := fmt.Sprintf(`
        SELECT EXISTS (
            SELECT 1 FROM %s 
            WHERE metadata->>'source' = $1 
            AND metadata->>'last_modified' = $2::text
        )
    `, p.tableName)

	for _, doc := range docs {
		source, _ := doc.Metadata["source"].(string)
		lastMod, _ := doc.Metadata["last_modified"].(string)
		batch.Queue(selectSQL, source, lastMod)
	}

	results := p.pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := range docs {
		err := results.QueryRow().Scan(&exists[i])
		if err != nil {
			return nil, vectorstore.NewSearchFailedError("pgvector",
				fmt.Errorf("failed to check document existence: %w", err))
		}
	}

	return exists, nil
}
