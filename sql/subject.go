package sql

import "encoding/json"

type Subject struct {
	ID            int
	TeacherID     int `db:"teacher_id"`
	Name          string
	InheritsClass bool `db:"inherits_class"`
	ClassID       int  `db:"class_id"`
	Students      string
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (db *sqlImpl) GetLastSubjectID() int {
	var id int
	err := db.db.Get(&id, "SELECT id FROM subject WHERE id = (SELECT MAX(id) FROM subject)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) GetSubject(id int) (subject Subject, err error) {
	err = db.db.Get(&subject, "SELECT * FROM subject WHERE id=$1", id)
	return subject, err
}

func (db *sqlImpl) GetAllSubjectsForTeacher(id int) (subject []Subject, err error) {
	err = db.db.Select(&subject, "SELECT * FROM subject WHERE teacher_id=$1", id)
	return subject, err
}

func (db *sqlImpl) GetAllSubjects() (subject []Subject, err error) {
	err = db.db.Select(&subject, "SELECT * FROM subject")
	return subject, err
}

func (db *sqlImpl) GetAllSubjectsForUser(id int) (subjects []Subject, err error) {
	subjectsAll, err := db.GetAllSubjects()
	if err != nil {
		return make([]Subject, 0), err
	}
	subjects = make([]Subject, 0)
	for i := 0; i < len(subjectsAll); i++ {
		subject := subjectsAll[i]
		var users []int
		if subject.InheritsClass {
			class, err := db.GetClass(subject.ClassID)
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
		if contains(users, id) {
			subjects = append(subjects, subject)
		}
	}
	return subjects, nil
}

func (db *sqlImpl) InsertSubject(subject Subject) error {
	_, err := db.db.NamedExec(
		"INSERT INTO subject (id, teacher_id, name, inherits_class, class_id, students) VALUES (:id, :teacher_id, :name, :inherits_class, :class_id, :students)",
		subject)
	return err
}

func (db *sqlImpl) UpdateSubject(subject Subject) error {
	_, err := db.db.NamedExec(
		"UPDATE subject SET teacher_id=:teacher_id, name=:name, inherits_class=:inherits_class, class_id=:class_id, students=:students WHERE id=:id",
		subject)
	return err
}

func (db *sqlImpl) DeleteSubject(subject Subject) error {
	_, err := db.db.NamedExec(
		"DELETE FROM subject WHERE id=:id",
		subject)
	return err
}
