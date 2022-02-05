package sql

type Class struct {
	ID       int
	Name     string
	Teacher  int
	Students string
}

func (db *sqlImpl) GetClass(id int) (class Class, err error) {
	err = db.db.Get(&class, "SELECT * FROM classes WHERE id=$1", id)
	return class, err
}

func (db *sqlImpl) InsertClass(class Class) (err error) {
	_, err = db.db.NamedExec(
		"INSERT INTO classes (id, teacher, name) VALUES (:id, :teacher, :name)",
		class)
	return err
}

func (db *sqlImpl) UpdateClass(class Class) error {
	_, err := db.db.NamedExec(
		"UPDATE classes SET teacher=:teacher, students=:students, name=:name WHERE id=:id",
		class)
	return err
}

func (db *sqlImpl) GetLastClassID() (id int) {
	err := db.db.Get(&id, "SELECT id FROM classes WHERE id = (SELECT MAX(id) FROM classes)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) GetClasses() (classes []Class, err error) {
	err = db.db.Select(&classes, "SELECT * FROM classes")
	return classes, err
}
func (db *sqlImpl) DeleteClass(ID int) error {
	_, err := db.db.Exec("DELETE FROM classes WHERE id=$1", ID)
	return err
}
