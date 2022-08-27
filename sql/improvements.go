package sql

type Improvement struct {
	ID        int
	StudentID int `db:"student_id"`
	MeetingID int `db:"meeting_id"`
	TeacherID int `db:"teacher_id"`
	Message   string

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (db *sqlImpl) GetLastImprovementID() int {
	var id int
	err := db.db.Get(&id, "SELECT id FROM improvements WHERE id = (SELECT MAX(id) FROM improvements)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) GetImprovement(id int) (improvement Improvement, err error) {
	err = db.db.Get(&improvement, "SELECT * FROM improvements WHERE id=$1", id)
	return improvement, err
}

func (db *sqlImpl) GetImprovementsForStudent(studentId int) (improvements []Improvement, err error) {
	err = db.db.Select(&improvements, "SELECT * FROM improvements WHERE student_id=$1 ORDER BY id ASC", studentId)
	if improvements == nil {
		improvements = make([]Improvement, 0)
	}
	return improvements, err
}

func (db *sqlImpl) InsertImprovement(improvement Improvement) error {
	_, err := db.db.NamedExec("INSERT INTO improvements (id, student_id, meeting_id, message, teacher_id) VALUES (:id, :student_id, :meeting_id, :message, :teacher_id)", improvement)
	return err
}

func (db *sqlImpl) UpdateImprovement(homework Homework) error {
	_, err := db.db.NamedExec(
		"UPDATE improvements SET message=:message WHERE id=:id",
		homework)
	return err
}

func (db *sqlImpl) DeleteImprovement(ID int) error {
	_, err := db.db.Exec("DELETE FROM improvements WHERE id=$1", ID)
	return err
}
