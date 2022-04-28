package sql

import "encoding/json"

func remove(s []int, i int) []int {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

type Class struct {
	ID             int
	Name           string
	Teacher        int
	Students       string
	ClassYear      string `db:"class_year"`
	SOK            int
	EOK            int
	LastSchoolDate int `db:"last_school_date"`
}

func (db *sqlImpl) GetClass(id int) (class Class, err error) {
	err = db.db.Get(&class, "SELECT * FROM classes WHERE id=$1", id)
	return class, err
}

func (db *sqlImpl) InsertClass(class Class) (err error) {
	_, err = db.db.NamedExec(
		"INSERT INTO classes (id, teacher, name, class_year, sok, eok, last_school_date) VALUES (:id, :teacher, :name, :class_year, :sok, :eok, :last_school_date)",
		class)
	return err
}

func (db *sqlImpl) UpdateClass(class Class) error {
	_, err := db.db.NamedExec(
		"UPDATE classes SET teacher=:teacher, students=:students, name=:name, class_year=:class_year, sok=:sok, eok=:eok, last_school_date=:last_school_date WHERE id=:id",
		class)
	return err
}

func (db *sqlImpl) GetLastClassID() (id int) {
	err := db.db.Get(&id, "SELECT id FROM classes WHERE id = (SELECT MAX(id) FROM classes)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) GetClasses() (classes []Class, err error) {
	err = db.db.Select(&classes, "SELECT * FROM classes ORDER BY id ASC")
	return classes, err
}

func (db *sqlImpl) DeleteClass(ID int) error {
	_, err := db.db.Exec("DELETE FROM classes WHERE id=$1", ID)
	return err
}

func (db *sqlImpl) DeleteTeacherClasses(teacherId int) error {
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
				if subjects[n].InheritsClass && subjects[n].ClassID == classes[i].ID {
					db.DeleteMeetingsForSubject(subjects[n].ID)
					db.DeleteSubject(subjects[n])
				}
			}
			db.DeleteClass(classes[i].ID)
		}
	}
	return nil
}

func (db *sqlImpl) DeleteUserClasses(userId int) {
	classes, _ := db.GetClasses()
	for i := 0; i < len(classes); i++ {
		class := classes[i]
		var users []int
		json.Unmarshal([]byte(class.Students), &users)
		for n := 0; n < len(users); n++ {
			if users[n] == userId {
				users = remove(users, n)
			}
		}
		marshal, _ := json.Marshal(users)
		class.Students = string(marshal)
		db.UpdateClass(class)
	}
}
