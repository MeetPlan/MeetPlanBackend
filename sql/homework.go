package sql

type Homework struct {
	ID          string
	TeacherID   string `db:"teacher_id"`
	SubjectID   string `db:"subject_id"`
	Name        string
	Description string
	ToDate      string `db:"to_date"`
	FromDate    string `db:"from_date"`

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (db *sqlImpl) GetHomework(id string) (homework Homework, err error) {
	err = db.db.Get(&homework, "SELECT * FROM homework WHERE id=$1", id)
	return homework, err
}

func (db *sqlImpl) GetHomeworkForSubject(id string) (homework []Homework, err error) {
	err = db.db.Select(&homework, "SELECT * FROM homework WHERE subject_id=$1 ORDER BY id ASC", id)
	if homework == nil {
		homework = make([]Homework, 0)
	}
	return homework, err
}

func (db *sqlImpl) GetHomeworkForTeacher(teacherId string) (homework []Homework, err error) {
	err = db.db.Select(&homework, "SELECT * FROM homework WHERE teacher_id=$1 ORDER BY id ASC", teacherId)
	if homework == nil {
		homework = make([]Homework, 0)
	}
	return homework, err
}

func (db *sqlImpl) InsertHomework(homework Homework) error {
	i := `
	INSERT INTO homework
	    (teacher_id, subject_id, name, description, from_date, to_date) VALUES
	    (:teacher_id, :subject_id, :name, :description, :from_date, :to_date)
	`
	_, err := db.db.NamedExec(
		i,
		homework)
	return err
}

func (db *sqlImpl) UpdateHomework(homework Homework) error {
	_, err := db.db.NamedExec(
		"UPDATE homework SET from_date=:from_date, to_date=:to_date, name=:name, description=:description WHERE id=:id",
		homework)
	return err
}

func (db *sqlImpl) DeleteHomework(ID string) error {
	_, err := db.db.Exec("DELETE FROM homework WHERE id=$1", ID)
	db.DeleteStudentHomeworkByHomeworkID(ID)
	return err
}

func (db *sqlImpl) DeleteAllTeacherHomeworks(ID string) {
	homework, _ := db.GetHomeworkForTeacher(ID)
	for i := 0; i < len(homework); i++ {
		db.DeleteHomework(homework[i].ID)
	}
}
