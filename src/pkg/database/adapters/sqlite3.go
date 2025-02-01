package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	fmt.Println("Connecting to SQLite database...", s.dbPath)

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
