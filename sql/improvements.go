package sql

type Improvement struct {
	ID        string
	StudentID string `db:"student_id"`
	MeetingID string `db:"meeting_id"`
	TeacherID string `db:"teacher_id"`
	Message   string

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (db *sqlImpl) GetImprovement(id string) (improvement Improvement, err error) {
	err = db.db.Get(&improvement, "SELECT * FROM improvements WHERE id=$1", id)
	return improvement, err
}

func (db *sqlImpl) GetImprovementsForStudent(studentId string) (improvements []Improvement, err error) {
	err = db.db.Select(&improvements, "SELECT * FROM improvements WHERE student_id=$1 ORDER BY id ASC", studentId)
	if improvements == nil {
		improvements = make([]Improvement, 0)
	}
	return improvements, err
}

func (db *sqlImpl) InsertImprovement(improvement Improvement) error {
	_, err := db.db.NamedExec("INSERT INTO improvements (student_id, meeting_id, message, teacher_id) VALUES (:student_id, :meeting_id, :message, :teacher_id)", improvement)
	return err
}

func (db *sqlImpl) UpdateImprovement(homework Homework) error {
	_, err := db.db.NamedExec(
		"UPDATE improvements SET message=:message WHERE id=:id",
		homework)
	return err
}

func (db *sqlImpl) DeleteImprovement(ID string) error {
	_, err := db.db.Exec("DELETE FROM improvements WHERE id=$1", ID)
	return err
}
