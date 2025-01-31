package memory

// type longTermMemoryImpl struct {
// 	db *sql.DB
// }

// func NewLongTermMemory(db *sql.DB) LongTermMemory {
// 	return &longTermMemoryImpl{db: db}
// }

// func (l *longTermMemoryImpl) Store(ctx context.Context, entry Entry) error {
// 	query := `
//         INSERT INTO memories (id, type, content, metadata, timestamp, importance, tags)
//         VALUES ($1, $2, $3, $4, $5, $6, $7)
//     `
// 	_, err := l.db.ExecContext(ctx, query, entry.ID, entry.Type, entry.Content,
// 		entry.Metadata, entry.Timestamp, entry.Importance, entry.Tags)
// 	return err
// }
