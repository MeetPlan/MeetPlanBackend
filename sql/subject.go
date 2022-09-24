package sql

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
)

type Subject struct {
	ID            string
	TeacherID     string `db:"teacher_id"`
	Name          string
	InheritsClass bool    `db:"inherits_class"`
	ClassID       *string `db:"class_id"`
	Students      string
	LongName      string `db:"long_name"`
	Realization   float32
	SelectedHours float32 `db:"selected_hours"`
	Color         string
	Location      string `db:"location"`
	IsGraded      bool   `db:"is_graded"`

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (db *sqlImpl) GetSubject(id string) (subject Subject, err error) {
	err = db.db.Get(&subject, "SELECT * FROM subject WHERE id=$1", id)
	return subject, err
}

func (db *sqlImpl) GetSubjectsWithSpecificLongName(longName string) (subject []Subject, err error) {
	err = db.db.Select(&subject, "SELECT * FROM subject WHERE long_name=$1", longName)
	return subject, err
}

func (db *sqlImpl) GetAllSubjectsForTeacher(id string) (subject []Subject, err error) {
	err = db.db.Select(&subject, "SELECT * FROM subject WHERE teacher_id=$1 ORDER BY id ASC", id)
	return subject, err
}

func (db *sqlImpl) GetAllSubjects() (subject []Subject, err error) {
	err = db.db.Select(&subject, "SELECT * FROM subject ORDER BY id ASC")
	return subject, err
}

func (db *sqlImpl) GetAllSubjectsForUser(id string) (subjects []Subject, err error) {
	subjectsAll, err := db.GetAllSubjects()
	if err != nil {
		return make([]Subject, 0), err
	}
	subjects = make([]Subject, 0)
	for i := 0; i < len(subjectsAll); i++ {
		subject := subjectsAll[i]
		var users []string
		if subject.InheritsClass {
			class, err := db.GetClass(*subject.ClassID)
			if err != nil {
				return make([]Subject, 0), err
			}
			err = json.Unmarshal([]byte(class.Students), &users)
			if err != nil {
				return make([]Subject, 0), err
			}
		} else {
			err := json.Unmarshal([]byte(subject.Students), &users)
			if err != nil {
				return make([]Subject, 0), err
			}
		}
		if helpers.Contains(users, id) {
			subjects = append(subjects, subject)
		}
	}
	return subjects, nil
}

func (db *sqlImpl) InsertSubject(subject Subject) error {
	_, err := db.db.NamedExec(
		"INSERT INTO subject (teacher_id, name, inherits_class, class_id, students, long_name, realization, selected_hours, color, location, is_graded) VALUES (:teacher_id, :name, :inherits_class, :class_id, :students, :long_name, :realization, :selected_hours, :color, :location, :is_graded)",
		subject)
	return err
}

func (db *sqlImpl) UpdateSubject(subject Subject) error {
	_, err := db.db.NamedExec(
		"UPDATE subject SET teacher_id=:teacher_id, name=:name, inherits_class=:inherits_class, class_id=:class_id, students=:students, long_name=:long_name, realization=:realization, selected_hours=:selected_hours, color=:color, location=:location, is_graded=:is_graded WHERE id=:id",
		subject)
	return err
}

func (db *sqlImpl) DeleteSubject(subject Subject) error {
	_, err := db.db.NamedExec(
		"DELETE FROM subject WHERE id=:id",
		subject)
	return err
}

func (db *sqlImpl) DeleteStudentSubject(userId string) {
	subjects, _ := db.GetAllSubjects()
	for i := 0; i < len(subjects); i++ {
		subject := subjects[i]
		var users []string
		json.Unmarshal([]byte(subject.Students), &users)
		if subject.InheritsClass {
			for n := 0; n < len(users); n++ {
				if users[n] == userId {
					users = helpers.Remove(users, n)
				}
			}
			marshal, _ := json.Marshal(users)
			subject.Students = string(marshal)
			db.UpdateSubject(subject)
		}
	}
}
