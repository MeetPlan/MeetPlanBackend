package sql

type Testing struct {
	ID        int
	UserID    int `db:"user_id"`
	Date      string
	TeacherID int `db:"teacher_id"`
	ClassID   int `db:"class_id"`
	Result    string
}

func (db *sqlImpl) GetTestingResults(date string) ([]Testing, error) {
	var message []Testing

	err := db.db.Select(&message, "SELECT * FROM testing WHERE date=$1", date)
	return message, err
}

func (db *sqlImpl) GetTestingResult(date string, id int) (Testing, error) {
	var message Testing

	err := db.db.Select(&message, "SELECT * FROM testing WHERE user_id=$1 AND date=$2", id, date)
	return message, err
}

func (db *sqlImpl) InsertTestingResult(testing Testing) error {
	_, err := db.db.NamedExec(
		"INSERT INTO testing (id, user_id, date, teacher_id, class_id, result) VALUES (:id, :user_id, :date, :teacher_id, :class_id, :result)",
		testing)
	return err
}

func (db *sqlImpl) UpdateTestingResult(testing Testing) error {
	_, err := db.db.NamedExec(
		"UPDATE testing SET user_id=:user_id, date=:date, teacher_id=:teacher_id, class_id=:class_id, result=:result WHERE id=:id",
		testing)
	return err
}
