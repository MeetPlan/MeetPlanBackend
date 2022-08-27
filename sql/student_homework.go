package sql

import "encoding/json"

type StudentHomework struct {
	ID         int
	UserID     int `db:"user_id"`
	HomeworkID int `db:"homework_id"`
	Status     string

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type StudentHomeworkJSON struct {
	StudentHomework
	Name        string
	TeacherName string
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

func (db *sqlImpl) GetStudentHomeworkForUser(homeworkId int, userId int) (homework StudentHomework, err error) {
	err = db.db.Get(&homework, "SELECT * FROM student_homework WHERE homework_id=$1 AND user_id=$2", homeworkId, userId)
	return homework, err
}

func (db *sqlImpl) GetStudentsHomework(id int) (homework []StudentHomework, err error) {
	err = db.db.Select(&homework, "SELECT * FROM student_homework WHERE user_id=$1 ORDER BY id ASC", id)
	if homework == nil {
		homework = make([]StudentHomework, 0)
	}
	return homework, err
}

func (db *sqlImpl) GetStudentsHomeworkByHomeworkID(id int, meeetingId int) (homework []StudentHomeworkJSON, err error) {
	baseHomework, err := db.GetHomework(id)
	if err != nil {
		return make([]StudentHomeworkJSON, 0), err
	}
	subject, err := db.GetSubject(baseHomework.SubjectID)
	if err != nil {
		return make([]StudentHomeworkJSON, 0), err
	}
	var students []int
	if subject.InheritsClass {
		class, err := db.GetClass(subject.ClassID)
		if err != nil {
			return make([]StudentHomeworkJSON, 0), err
		}
		err = json.Unmarshal([]byte(class.Students), &students)
		if err != nil {
			return make([]StudentHomeworkJSON, 0), err
		}
	} else {
		err := json.Unmarshal([]byte(subject.Students), &students)
		if err != nil {
			return make([]StudentHomeworkJSON, 0), err
		}
	}

	teacher, err := db.GetUser(baseHomework.TeacherID)
	if err != nil {
		return make([]StudentHomeworkJSON, 0), err
	}

	homework = make([]StudentHomeworkJSON, 0)
	for i := 0; i < len(students); i++ {
		student, err := db.GetUser(students[i])
		if err != nil {
			return make([]StudentHomeworkJSON, 0), err
		}
		homeworkUser, err := db.GetStudentHomeworkForUser(baseHomework.ID, students[i])
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				var status = " "
				absence, err := db.GetAbsenceForUserMeeting(meeetingId, students[i])
				if err != nil {
					if err.Error() != "sql: no rows in result set" {
						return make([]StudentHomeworkJSON, 0), err
					} else {
						status = ""
					}
				}
				if status != "" {
					if absence.AbsenceType == "ABSENT" {
						status = "ABSENT"
					}
				}
				studentHomework := StudentHomework{ID: db.GetLastStudentHomeworkID(), UserID: students[i], HomeworkID: id, Status: status}
				homework = append(homework, StudentHomeworkJSON{
					StudentHomework: studentHomework,
					Name:            student.Name,
					TeacherName:     teacher.Name,
				})
				err = db.InsertStudentHomework(studentHomework)
				if err != nil {
					return make([]StudentHomeworkJSON, 0), err
				}
			} else {
				return make([]StudentHomeworkJSON, 0), err
			}
		} else {
			homework = append(homework, StudentHomeworkJSON{
				StudentHomework: homeworkUser,
				Name:            student.Name,
				TeacherName:     teacher.Name,
			})
		}
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

func (db *sqlImpl) DeleteStudentHomeworkByHomeworkID(ID int) error {
	_, err := db.db.Exec("DELETE FROM student_homework WHERE homework_id=$1", ID)
	return err
}

func (db *sqlImpl) DeleteStudentHomeworkByStudentID(ID int) error {
	_, err := db.db.Exec("DELETE FROM student_homework WHERE user_id=$1", ID)
	return err
}
