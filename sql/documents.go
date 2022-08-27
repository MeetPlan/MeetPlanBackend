package sql

type Document struct {
	ID           string
	ExportedBy   int  `db:"exported_by"`
	DocumentType int  `db:"document_type"`
	Timestamp    int  // Unix timestamp
	IsSigned     bool `db:"is_signed"`

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (db *sqlImpl) GetDocument(id string) (document Document, err error) {
	err = db.db.Get(&document, "SELECT * FROM documents WHERE id=$1", id)
	return document, err
}

func (db *sqlImpl) GetAllDocuments() (documents []Document, err error) {
	err = db.db.Select(&documents, "SELECT * FROM documents ORDER BY timestamp DESC")
	return documents, err
}

func (db *sqlImpl) InsertDocument(document Document) error {
	_, err := db.db.NamedExec(
		"INSERT INTO documents (id, exported_by, document_type, timestamp, is_signed) VALUES (:id, :exported_by, :document_type, :timestamp, :is_signed)",
		document)
	return err
}

func (db *sqlImpl) DeleteDocument(id string) {
	db.db.Exec("DELETE FROM documents WHERE id=$1", id)
}
