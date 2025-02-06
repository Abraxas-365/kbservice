package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Abraxas-365/kbservice/chathistory"
	"github.com/Abraxas-365/kbservice/llm"
	"github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) (*PostgresRepository, error) {
	if db == nil {
		return nil, errors.New("database connection is required")
	}
	return &PostgresRepository{db: db}, nil
}

// Required database schema
const schema = `
CREATE TABLE IF NOT EXISTS conversations (
    id TEXT PRIMARY KEY,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    conversation_id TEXT REFERENCES conversations(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    name TEXT,
    function_call JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    metadata JSONB,
    CONSTRAINT fk_conversation
        FOREIGN KEY(conversation_id)
        REFERENCES conversations(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_role ON messages(role);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_conversations_created_at ON conversations(created_at);
`

func (r *PostgresRepository) InitSchema(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, schema)
	return err
}

func (r *PostgresRepository) CreateConversation(ctx context.Context, conv chathistory.Conversation) error {
	metadata, err := json.Marshal(conv.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO conversations (id, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = r.db.ExecContext(ctx, query, conv.ID, metadata, conv.CreatedAt, conv.UpdatedAt)
	return err
}

func (r *PostgresRepository) AddMessage(ctx context.Context, conversationID string, message llm.Message) error {
	functionCall, err := json.Marshal(message.FuncCall)
	if err != nil {
		return fmt.Errorf("failed to marshal function call: %w", err)
	}

	metadata, err := json.Marshal(message.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO messages (conversation_id, role, content, name, function_call, created_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = r.db.ExecContext(ctx, query,
		conversationID,
		message.Role,
		message.Content,
		message.Name,
		functionCall,
		time.Now(),
		metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
	}

	// Update conversation updated_at timestamp
	updateQuery := `UPDATE conversations SET updated_at = NOW() WHERE id = $1`
	_, err = r.db.ExecContext(ctx, updateQuery, conversationID)
	return err
}

func (r *PostgresRepository) GetMessages(ctx context.Context, conversationID string, limit int) ([]llm.Message, error) {
	query := `
		SELECT role, content, name, function_call, created_at, metadata
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []llm.Message
	for rows.Next() {
		var msg llm.Message
		var functionCallJSON, metadataJSON []byte

		err := rows.Scan(
			&msg.Role,
			&msg.Content,
			&msg.Name,
			&functionCallJSON,
			&metadataJSON,
		)
		if err != nil {
			return nil, err
		}

		if len(functionCallJSON) > 0 {
			if err := json.Unmarshal(functionCallJSON, &msg.FuncCall); err != nil {
				return nil, err
			}
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &msg.Metadata); err != nil {
				return nil, err
			}
		}

		messages = append(messages, msg)
	}

	// Reverse the order to get chronological order
	for i := 0; i < len(messages)/2; i++ {
		j := len(messages) - i - 1
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *PostgresRepository) GetMessagesByFilter(ctx context.Context, conversationID string, filter chathistory.Filter, limit int) ([]llm.Message, error) {
	conditions := []string{"conversation_id = $1"}
	params := []interface{}{conversationID}
	paramCount := 2

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", paramCount))
		params = append(params, filter.StartTime)
		paramCount++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", paramCount))
		params = append(params, filter.EndTime)
		paramCount++
	}

	if len(filter.Roles) > 0 {
		conditions = append(conditions, fmt.Sprintf("role = ANY($%d)", paramCount))
		params = append(params, pq.Array(filter.Roles))
		paramCount++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("content ILIKE $%d", paramCount))
		params = append(params, "%"+filter.Search+"%")
		paramCount++
	}

	query := fmt.Sprintf(`
		SELECT role, content, name, function_call, created_at, metadata
		FROM messages
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d
	`, strings.Join(conditions, " AND "), paramCount)

	params = append(params, limit)
	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []llm.Message
	for rows.Next() {
		var msg llm.Message
		var functionCallJSON, metadataJSON []byte

		err := rows.Scan(
			&msg.Role,
			&msg.Content,
			&msg.Name,
			&functionCallJSON,
			&metadataJSON,
		)
		if err != nil {
			return nil, err
		}

		if len(functionCallJSON) > 0 {
			if err := json.Unmarshal(functionCallJSON, &msg.FuncCall); err != nil {
				return nil, err
			}
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &msg.Metadata); err != nil {
				return nil, err
			}
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *PostgresRepository) DeleteMessages(ctx context.Context, conversationID string, filter chathistory.Filter) error {
	conditions := []string{"conversation_id = $1"}
	params := []interface{}{conversationID}
	paramCount := 2

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", paramCount))
		params = append(params, filter.StartTime)
		paramCount++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", paramCount))
		params = append(params, filter.EndTime)
		paramCount++
	}

	if len(filter.Roles) > 0 {
		conditions = append(conditions, fmt.Sprintf("role = ANY($%d)", paramCount))
		params = append(params, pq.Array(filter.Roles))
		paramCount++
	}

	query := fmt.Sprintf(`
		DELETE FROM messages
		WHERE %s
	`, strings.Join(conditions, " AND "))

	_, err := r.db.ExecContext(ctx, query, params...)
	return err
}

func (r *PostgresRepository) ClearHistory(ctx context.Context, conversationID string) error {
	query := `DELETE FROM messages WHERE conversation_id = $1`
	_, err := r.db.ExecContext(ctx, query, conversationID)
	return err
}

func (r *PostgresRepository) DeleteConversation(ctx context.Context, conversationID string) error {
	query := `DELETE FROM conversations WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, conversationID)
	return err
}

func (r *PostgresRepository) GetConversation(ctx context.Context, conversationID string) (*chathistory.Conversation, error) {
	query := `
		SELECT id, metadata, created_at, updated_at
		FROM conversations
		WHERE id = $1
	`
	var conv chathistory.Conversation
	var metadataJSON []byte
	err := r.db.QueryRowContext(ctx, query, conversationID).Scan(
		&conv.ID,
		&metadataJSON,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &conv.Metadata); err != nil {
			return nil, err
		}
	}

	return &conv, nil
}

func (r *PostgresRepository) ListConversations(ctx context.Context, filter chathistory.Filter, limit, offset int) ([]chathistory.Conversation, error) {
	conditions := []string{"1=1"}
	params := []interface{}{}
	paramCount := 1

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", paramCount))
		params = append(params, filter.StartTime)
		paramCount++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", paramCount))
		params = append(params, filter.EndTime)
		paramCount++
	}

	query := fmt.Sprintf(`
		SELECT id, metadata, created_at, updated_at
		FROM conversations
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(conditions, " AND "), paramCount, paramCount+1)

	params = append(params, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []chathistory.Conversation
	for rows.Next() {
		var conv chathistory.Conversation
		var metadataJSON []byte
		err := rows.Scan(
			&conv.ID,
			&metadataJSON,
			&conv.CreatedAt,
			&conv.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &conv.Metadata); err != nil {
				return nil, err
			}
		}

		conversations = append(conversations, conv)
	}

	return conversations, nil
}

func (r *PostgresRepository) UpdateConversationMetadata(ctx context.Context, conversationID string, metadata map[string]any) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE conversations
		SET metadata = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err = r.db.ExecContext(ctx, query, metadataJSON, conversationID)
	return err
}

func (r *PostgresRepository) GetMessageCount(ctx context.Context, conversationID string, filter chathistory.Filter) (int, error) {
	conditions := []string{"conversation_id = $1"}
	params := []interface{}{conversationID}
	paramCount := 2

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", paramCount))
		params = append(params, filter.StartTime)
		paramCount++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", paramCount))
		params = append(params, filter.EndTime)
		paramCount++
	}

	if len(filter.Roles) > 0 {
		conditions = append(conditions, fmt.Sprintf("role = ANY($%d)", paramCount))
		params = append(params, pq.Array(filter.Roles))
		paramCount++
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM messages
		WHERE %s
	`, strings.Join(conditions, " AND "))

	var count int
	err := r.db.QueryRowContext(ctx, query, params...).Scan(&count)
	return count, err
}
