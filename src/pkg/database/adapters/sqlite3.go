package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db     *sql.DB
	dbPath string
}

func NewSQLiteStore(dbPath string) *SQLiteStore {
	return &SQLiteStore{
		dbPath: dbPath,
	}
}

// Connect establishes connection to the SQLite database
func (s *SQLiteStore) Connect(ctx context.Context) error {
	logger.GetLogger().Infof("Connecting to SQLite database..., %s", s.dbPath)

	// Ensure the directory exists
	dir := filepath.Dir(s.dbPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := sql.Open("sqlite3", s.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	s.db = db
	return nil
}

// CreateTable creates a new table with the given schema
func (s *SQLiteStore) CreateTable(ctx context.Context, tableName string, schema string) error {
	if s.db == nil {
		return fmt.Errorf("database connection not established")
	}

	// Clean and validate table name to prevent SQL injection
	tableName = sanitizeIdentifier(tableName)
	if tableName == "" {
		return fmt.Errorf("invalid table name")
	}

	// Create table query
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, schema)

	// Execute the query
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	return nil
}

// Insert adds new data to the specified table
func (s *SQLiteStore) Insert(ctx context.Context, tableName string, data map[string]interface{}) error {
	if s.db == nil {
		return fmt.Errorf("database connection not established")
	}

	// Clean table name
	tableName = sanitizeIdentifier(tableName)
	if tableName == "" {
		return fmt.Errorf("invalid table name")
	}

	// Prepare column names and values
	var columns []string
	var placeholders []string
	var values []interface{}

	for col, val := range data {
		col = sanitizeIdentifier(col)
		if col != "" {
			columns = append(columns, col)
			placeholders = append(placeholders, "?")
			values = append(values, val)
		}
	}

	if len(columns) == 0 {
		return fmt.Errorf("no valid columns provided for insert")
	}

	// Build the query
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	// Execute the query
	_, err := s.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to insert data into %s: %w", tableName, err)
	}

	return nil
}

// Update modifies existing data in the specified table
func (s *SQLiteStore) Update(ctx context.Context, tableName string, id string, data map[string]interface{}) error {
	if s.db == nil {
		return fmt.Errorf("database connection not established")
	}

	// Clean table name and validate id
	tableName = sanitizeIdentifier(tableName)
	if tableName == "" {
		return fmt.Errorf("invalid table name")
	}
	if id == "" {
		return fmt.Errorf("invalid id")
	}

	// Prepare SET clause
	var setStatements []string
	var values []interface{}

	for col, val := range data {
		col = sanitizeIdentifier(col)
		if col != "" {
			setStatements = append(setStatements, fmt.Sprintf("%s = ?", col))
			values = append(values, val)
		}
	}

	if len(setStatements) == 0 {
		return fmt.Errorf("no valid columns provided for update")
	}

	// Add id to values
	values = append(values, id)

	// Build the query
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = ?",
		tableName,
		strings.Join(setStatements, ", "),
	)

	// Execute the query
	result, err := s.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to update data in %s: %w", tableName, err)
	}

	// Check if any row was affected
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("no record found with id: %s", id)
	}

	return nil
}

// Delete removes data from the specified table
func (s *SQLiteStore) Delete(ctx context.Context, tableName string, id string) error {
	if s.db == nil {
		return fmt.Errorf("database connection not established")
	}

	// Clean table name and validate id
	tableName = sanitizeIdentifier(tableName)
	if tableName == "" {
		return fmt.Errorf("invalid table name")
	}
	if id == "" {
		return fmt.Errorf("invalid id")
	}

	// Build and execute query
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete data from %s: %w", tableName, err)
	}

	// Check if any row was affected
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("no record found with id: %s", id)
	}

	return nil
}

// Get retrieves data from the specified table
func (s *SQLiteStore) Get(ctx context.Context, tableName string, id string) (map[string]interface{}, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}

	// Clean table name and validate id
	tableName = sanitizeIdentifier(tableName)
	if tableName == "" {
		return nil, fmt.Errorf("invalid table name")
	}
	if id == "" {
		return nil, fmt.Errorf("invalid id")
	}

	// Build and execute query
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", tableName)

	row := s.db.QueryRowContext(ctx, query, id)

	// Scan the result
	var createdAt time.Time
	var content []byte
	if err := row.Scan(&id, &createdAt, &content); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get data from %s: %w", tableName, err)
	}

	return map[string]interface{}{
		"id":         id,
		"created_at": createdAt,
		"content":    content,
	}, nil
}

// GetAll retrieves all records from the specified table
func (s *SQLiteStore) GetAll(ctx context.Context, tableName string) ([]map[string]interface{}, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}

	// Clean table name
	tableName = sanitizeIdentifier(tableName)
	if tableName == "" {
		return nil, fmt.Errorf("invalid table name")
	}

	// Build and execute query
	query := fmt.Sprintf("SELECT * FROM %s ORDER BY created_at DESC", tableName)

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table %s: %w", tableName, err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get column names: %w", err)
	}

	// Prepare result slice
	var results []map[string]interface{}

	// Prepare value holders for scanning
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	// Iterate through rows
	for rows.Next() {
		// Scan the row into value holders
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for this row's data
		entry := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Handle null values
			if val == nil {
				entry[col] = nil
				continue
			}

			entry[col] = val
		}

		results = append(results, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
		s.db = nil
	}
	return nil
}

// Additional helper methods that could be useful

func (s *SQLiteStore) Query(ctx context.Context, tableName string, query string, args ...interface{}) (*sql.Rows, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}
	return s.db.QueryContext(ctx, query, args...)
}

func (s *SQLiteStore) QueryRow(ctx context.Context, tableName string, query string, args ...interface{}) *sql.Row {
	return s.db.QueryRowContext(ctx, query, args...)
}

func (s *SQLiteStore) Begin(ctx context.Context) (*sql.Tx, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}
	return s.db.BeginTx(ctx, nil)
}

func sanitizeIdentifier(identifier string) string {
	// Remove any characters that aren't alphanumeric or underscores
	cleaned := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' {
			return r
		}
		return -1
	}, identifier)

	// Ensure it doesn't start with a number
	if len(cleaned) > 0 && cleaned[0] >= '0' && cleaned[0] <= '9' {
		return ""
	}

	return cleaned
}
