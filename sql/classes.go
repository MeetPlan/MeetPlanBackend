package sql

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
)

type Class struct {
	ID             string
	Name           string
	Teacher        string
	Students       string
	ClassYear      string `db:"class_year"`
	SOK            int
	EOK            int
	LastSchoolDate int `db:"last_school_date"`

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (db *sqlImpl) GetClass(id string) (class Class, err error) {
	err = db.db.Get(&class, "SELECT * FROM classes WHERE id=$1", id)
	return class, err
}

func (db *sqlImpl) InsertClass(class Class) (err error) {
	_, err = db.db.NamedExec(
		"INSERT INTO classes (teacher, name, class_year, sok, eok, last_school_date) VALUES (:teacher, :name, :class_year, :sok, :eok, :last_school_date)",
		class)
	return err
}

func (db *sqlImpl) UpdateClass(class Class) error {
	_, err := db.db.NamedExec(
		"UPDATE classes SET teacher=:teacher, students=:students, name=:name, class_year=:class_year, sok=:sok, eok=:eok, last_school_date=:last_school_date WHERE id=:id",
		class)
	return err
}

func (db *sqlImpl) GetClasses() (classes []Class, err error) {
	err = db.db.Select(&classes, "SELECT * FROM classes ORDER BY id ASC")
	return classes, err
}

func (db *sqlImpl) DeleteClass(ID string) error {
	_, err := db.db.Exec("DELETE FROM classes WHERE id=$1", ID)
	return err
}

func (db *sqlImpl) DeleteTeacherClasses(teacherId string) error {
	classes, err := db.GetClasses()
	if err != nil {
		return err
	}
	for i := 0; i < len(classes); i++ {
		if classes[i].Teacher == teacherId {
			subjects, err := db.GetAllSubjects()
			if err != nil {
				return err
			}
			for n := 0; n < len(subjects); n++ {
				if subjects[n].InheritsClass && *subjects[n].ClassID == classes[i].ID {
					db.DeleteMeetingsForSubject(subjects[n].ID)
					db.DeleteSubject(subjects[n])
				}
			}
			db.DeleteClass(classes[i].ID)
		}
	}
	return nil
}

func (db *sqlImpl) DeleteUserClasses(userId string) {
	classes, _ := db.GetClasses()
	for i := 0; i < len(classes); i++ {
		class := classes[i]
		var users []string
		json.Unmarshal([]byte(class.Students), &users)
		for n := 0; n < len(users); n++ {
			if users[n] == userId {
				users = helpers.Remove(users, n)
			}
		}
		marshal, _ := json.Marshal(users)
		class.Students = string(marshal)
		db.UpdateClass(class)
	}
}
