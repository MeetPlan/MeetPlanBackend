package sql

type Subject struct {
	ID            int
	TeacherID     int `db:"teacher_id"`
	Name          string
	InheritsClass bool `db:"inherits_class"`
	ClassID       int  `db:"class_id"`
	Students      string
}

func (db *sqlImpl) GetLastSubjectID() int {
	var id int
	err := db.db.Get(&id, "SELECT id FROM subject WHERE id = (SELECT MAX(id) FROM subject)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) GetSubject(id int) (subject Subject, err error) {
	err = db.db.Get(&subject, "SELECT * FROM subject WHERE id=$1", id)
	return subject, err
}

func (db *sqlImpl) GetAllSubjectsForTeacher(id int) (subject []Subject, err error) {
	err = db.db.Select(&subject, "SELECT * FROM subject WHERE teacher_id=$1", id)
	return subject, err
}

func (db *sqlImpl) GetAllSubjects() (subject []Subject, err error) {
	err = db.db.Select(&subject, "SELECT * FROM subject")
	return subject, err
}

func (db *sqlImpl) GetAllSubjectsForUser(id int) (subject []Subject, err error) {
	err = db.db.Select(&subject, "SELECT * FROM subject WHERE user_id=$1", id)
	return subject, err
}

func (db *sqlImpl) InsertSubject(subject Subject) error {
	_, err := db.db.NamedExec(
		"INSERT INTO subject (id, teacher_id, name, inherits_class, class_id, students) VALUES (:id, :teacher_id, :name, :inherits_class, :class_id, :students)",
		subject)
	return err
}

func (db *sqlImpl) UpdateSubject(subject Subject) error {
	_, err := db.db.NamedExec(
		"UPDATE subject SET teacher_id=:teacher_id, name=:name, inherits_class=:inherits_class, class_id=:class_id, students=:students WHERE id=:id",
		subject)
	return err
}

func (db *sqlImpl) DeleteSubject(subject Subject) error {
	_, err := db.db.NamedExec(
		"DELETE FROM subject WHERE id=:id",
		subject)
	return err
}
