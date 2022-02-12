package sql

type Grade struct {
	ID          int
	UserID      int `db:"user_id"`
	TeacherID   int `db:"teacher_id"`
	SubjectID   int `db:"subject_id"`
	Grade       int
	Date        string
	IsWritten   bool `db:"is_written"`
	Period      int
	Description string
}

func (db *sqlImpl) GetLastGradeID() int {
	var id int
	err := db.db.Get(&id, "SELECT id FROM grades WHERE id = (SELECT MAX(id) FROM grades)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) GetGrade(id int) (grade Grade, err error) {
	err = db.db.Get(&grade, "SELECT * FROM grades WHERE id=$1", id)
	return grade, err
}

func (db *sqlImpl) GetGradesForUser(userId int) (grades []Grade, err error) {
	err = db.db.Select(&grades, "SELECT * FROM grades WHERE user_id=$1", userId)
	return grades, err
}

func (db *sqlImpl) InsertGrade(grade Grade) error {
	i := `
	INSERT INTO grades
	    (id, user_id, teacher_id, subject_id, date, is_written, grade, period, description) VALUES
	    (:id, :user_id, :teacher_id, :subject_id, :date, :is_written, :grade, :period, :description)
	`
	_, err := db.db.NamedExec(
		i,
		grade)
	return err
}

func (db *sqlImpl) UpdateGrade(grade Grade) error {
	_, err := db.db.NamedExec(
		"UPDATE grades SET user_id=:user_id, teacher_id=:teacher_id, subject_id=:subject_id, date=:date, is_written=:is_written, grade=:grade, period=:period, description=:description WHERE id=:id",
		grade)
	return err
}

func (db *sqlImpl) DeleteGrade(ID int) error {
	_, err := db.db.Exec("DELETE FROM grades WHERE id=$1", ID)
	return err
}
