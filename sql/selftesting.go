package sql

type Testing struct {
	ID        int
	UserID    int `db:"user_id"`
	Date      string
	TeacherID int `db:"teacher_id"`
	ClassID   int `db:"class_id"`
	Result    string
}

func (db *sqlImpl) GetTestingResults(date string, classId int) ([]Testing, error) {
	var message []Testing

	err := db.db.Select(&message, "SELECT * FROM testing WHERE date=$1 AND class_id=$2", date, classId)
	return message, err
}

func (db *sqlImpl) GetLastTestingID() int {
	var id int
	err := db.db.Get(&id, "SELECT id FROM testing WHERE id = (SELECT MAX(id) FROM testing)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
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
