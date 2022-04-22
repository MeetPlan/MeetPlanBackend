package sql

import (
	"encoding/json"
)

type Testing struct {
	ID        int
	UserID    int `db:"user_id"`
	Date      string
	TeacherID int `db:"teacher_id"`
	ClassID   int `db:"class_id"`
	Result    string
}

type TestingJSON struct {
	ID          int
	UserID      int `db:"user_id"`
	Date        string
	TeacherID   int `db:"teacher_id"`
	TeacherName string
	ClassID     int `db:"class_id"`
	ValidUntil  string
	Result      string
	IsDone      bool
	UserName    string
}

func (db *sqlImpl) GetTestingResults(date string, classId int) ([]TestingJSON, error) {
	var testing = make([]TestingJSON, 0)

	class, err := db.GetClass(classId)
	if err != nil {
		db.logger.Debug(err)
		return nil, err
	}
	var students []int
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
				tjson = TestingJSON{IsDone: false, UserID: student, ClassID: classId, Date: date, TeacherID: -1, ID: -1, UserName: user.Name}
			} else if test.Result == "" || test.Result == "SE NE TESTIRA" {
				tjson = TestingJSON{IsDone: false, UserID: student, ClassID: classId, Date: date, TeacherID: -1, ID: -1, UserName: user.Name, Result: test.Result}
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

	err := db.db.Get(&message, "SELECT * FROM testing WHERE user_id=$1 AND date=$2", id, date)
	return message, err
}

func (db *sqlImpl) GetAllTestingsForUser(id int) (testing []Testing, err error) {
	err = db.db.Select(&testing, "SELECT * FROM testing WHERE user_id=$1", id)
	return testing, err
}

func (db *sqlImpl) GetTestingResultByID(id int) (Testing, error) {
	var message Testing

	err := db.db.Get(&message, "SELECT * FROM testing WHERE id=$1", id)
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

func (db *sqlImpl) DeleteTeacherSelfTesting(teacherId int) error {
	_, err := db.db.Exec("DELETE FROM testing WHERE teacher_id=$1", teacherId)
	return err
}

func (db *sqlImpl) DeleteUserSelfTesting(userId int) error {
	_, err := db.db.Exec("DELETE FROM testing WHERE user_id=$1", userId)
	return err
}
