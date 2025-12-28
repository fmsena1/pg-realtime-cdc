package db

import (
	"context"
	"fmt"
	"log"

	"github.com/fmsena1/pg-realtime-cdc/backend/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

func Init() error {
	config, err := pgxpool.ParseConfig("postgres://realtime:realtime@postgres:5432/realtime")
	if err != nil {
		return fmt.Errorf("unable to parse config: %w", err)
	}

	pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("✅ Database connected")
	return nil
}

func GetAllMessages() ([]models.Message, error) {
	rows, err := pool.Query(context.Background(), "SELECT id, content, created_at FROM messages ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, rows.Err()
}

func CreateMessage(content string) (*models.Message, error) {
	var m models.Message
	err := pool.QueryRow(
		context.Background(),
		"INSERT INTO messages (content) VALUES ($1) RETURNING id, content, created_at",
		content,
	).Scan(&m.ID, &m.Content, &m.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &m, nil
}

func UpdateMessage(id int, content string) (*models.Message, error) {
	var m models.Message
	err := pool.QueryRow(
		context.Background(),
		"UPDATE messages SET content = $1 WHERE id = $2 RETURNING id, content, created_at",
		content, id,
	).Scan(&m.ID, &m.Content, &m.CreatedAt)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("message not found")
		}
		return nil, err
	}

	return &m, nil
}

func DeleteMessage(id int) error {
	result, err := pool.Exec(
		context.Background(),
		"DELETE FROM messages WHERE id = $1",
		id,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}

func Close() {
	if pool != nil {
		pool.Close()
	}
}

