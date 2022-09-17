package sql

type Document struct {
	ID           string
	ExportedBy   string `db:"exported_by"`
	DocumentType int    `db:"document_type"`
	Timestamp    string `db:"created_at"`
	IsSigned     bool   `db:"is_signed"`
	UpdatedAt    string `db:"updated_at"`
}

func (db *sqlImpl) GetDocument(id string) (document Document, err error) {
	err = db.db.Get(&document, "SELECT * FROM documents WHERE id=$1", id)
	return document, err
}

func (db *sqlImpl) GetAllDocuments() (documents []Document, err error) {
	err = db.db.Select(&documents, "SELECT * FROM documents ORDER BY created_at DESC")
	return documents, err
}

func (db *sqlImpl) InsertDocument(document Document) error {
	_, err := db.db.NamedExec(
		"INSERT INTO documents (id, exported_by, document_type, is_signed) VALUES (:id, :exported_by, :document_type, :is_signed)",
		document)
	return err
}

func (db *sqlImpl) DeleteDocument(id string) {
	db.db.Exec("DELETE FROM documents WHERE id=$1", id)
}
