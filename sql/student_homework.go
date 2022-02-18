package sql

type StudentHomework struct {
	ID         int
	UserID     int
	HomeworkID int `db:"homework_id"`
	Status     string
}

func (db *sqlImpl) GetLastStudentHomeworkID() int {
	var id int
	err := db.db.Get(&id, "SELECT id FROM student_homework WHERE id = (SELECT MAX(id) FROM student_homework)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) GetStudentHomework(id int) (homework StudentHomework, err error) {
	err = db.db.Get(&homework, "SELECT * FROM student_homework WHERE id=$1", id)
	return homework, err
}

func (db *sqlImpl) GetStudentsHomework(id int) (homework []StudentHomework, err error) {
	err = db.db.Select(&homework, "SELECT * FROM student_homework WHERE user_id=$1", id)
	if homework == nil {
		homework = make([]StudentHomework, 0)
	}
	return homework, err
}

func (db *sqlImpl) InsertStudentHomework(homework StudentHomework) error {
	_, err := db.db.NamedExec(
		"INSERT INTO student_homework (id, user_id, homework_id, status) VALUES (:id, :user_id, :homework_id, :status)",
		homework)
	return err
}

func (db *sqlImpl) UpdateStudentHomework(homework StudentHomework) error {
	_, err := db.db.NamedExec(
		"UPDATE student_homework SET status=:status WHERE id=:id",
		homework)
	return err
}

func (db *sqlImpl) DeleteStudentHomework(ID int) error {
	_, err := db.db.Exec("DELETE FROM student_homework WHERE id=$1", ID)
	return err
}
