package sql

import (
	"encoding/json"
)

type Testing struct {
	ID        string
	UserID    string `db:"user_id"`
	Date      string
	TeacherID string `db:"teacher_id"`
	ClassID   string `db:"class_id"`
	Result    string

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type TestingJSON struct {
	ID          string
	UserID      string `db:"user_id"`
	Date        string
	TeacherID   string `db:"teacher_id"`
	TeacherName string
	ClassID     string `db:"class_id"`
	ValidUntil  string
	Result      string
	IsDone      bool
	UserName    string
}

func (db *sqlImpl) GetTestingResults(date string, classId string) ([]TestingJSON, error) {
	var testing = make([]TestingJSON, 0)

	class, err := db.GetClass(classId)
	if err != nil {
		db.logger.Debug(err)
		return nil, err
	}
	var students []string
	err = json.Unmarshal([]byte(class.Students), &students)
	if err != nil {
		db.logger.Debug(err)
		return nil, err
	}

	for i := 0; i < len(students); i++ {
		student := students[i]
		var test Testing
		var tjson TestingJSON
		user, err := db.GetUser(student)
		if err != nil {
			db.logger.Debug(err)
			return nil, err
		}
		err = db.db.Get(&test, "SELECT * FROM testing WHERE date=$1 AND class_id=$2 AND user_id=$3", date, classId, student)
		if err != nil || test.Result == "" || test.Result == "SE NE TESTIRA" {
			db.logger.Debug(err)
			if err != nil && err.Error() == "sql: no rows in result set" {
				tjson = TestingJSON{IsDone: false, UserID: student, ClassID: classId, Date: date, UserName: user.Name}
			} else if test.Result == "" || test.Result == "SE NE TESTIRA" {
				tjson = TestingJSON{IsDone: false, UserID: student, ClassID: classId, Date: date, UserName: user.Name, Result: test.Result}
			} else {
				db.logger.Debug(err)
				return nil, err
			}
		} else {
			tjson = TestingJSON{IsDone: true, ClassID: test.ClassID, Date: test.Date, Result: test.Result, TeacherID: test.TeacherID, ID: test.ID, UserID: test.UserID, UserName: user.Name}
		}
		testing = append(testing, tjson)
	}
	return testing, nil
}

func (db *sqlImpl) GetTestingResult(date string, id string) (Testing, error) {
	var message Testing

	err := db.db.Get(&message, "SELECT * FROM testing WHERE user_id=$1 AND date=$2", id, date)
	return message, err
}

func (db *sqlImpl) GetAllTestingsForUser(id string) (testing []Testing, err error) {
	err = db.db.Select(&testing, "SELECT * FROM testing WHERE user_id=$1 ORDER BY id ASC", id)
	return testing, err
}

func (db *sqlImpl) GetTestingResultByID(id string) (Testing, error) {
	var message Testing

	err := db.db.Get(&message, "SELECT * FROM testing WHERE id=$1", id)
	return message, err
}

func (db *sqlImpl) InsertTestingResult(testing Testing) error {
	_, err := db.db.NamedExec(
		"INSERT INTO testing (user_id, date, teacher_id, class_id, result) VALUES (:user_id, :date, :teacher_id, :class_id, :result)",
		testing)
	return err
}

func (db *sqlImpl) UpdateTestingResult(testing Testing) error {
	_, err := db.db.NamedExec(
		"UPDATE testing SET user_id=:user_id, date=:date, teacher_id=:teacher_id, class_id=:class_id, result=:result WHERE id=:id",
		testing)
	return err
}

func (db *sqlImpl) DeleteTeacherSelfTesting(teacherId string) error {
	_, err := db.db.Exec("DELETE FROM testing WHERE teacher_id=$1", teacherId)
	return err
}

func (db *sqlImpl) DeleteUserSelfTesting(userId string) error {
	_, err := db.db.Exec("DELETE FROM testing WHERE user_id=$1", userId)
	return err
}
