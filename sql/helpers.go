package sql

import "encoding/json"

func (db *sqlImpl) GetStudentsFromSubject(subject *Subject) []string {
	if subject == nil {
		return make([]string, 0)
	}
	var students []string
	if subject.InheritsClass {
		class, err := db.GetClass(*subject.ClassID)
		if err != nil {
			return make([]string, 0)
		}
		err = json.Unmarshal([]byte(class.Students), &students)
		if err != nil {
			return make([]string, 0)
		}
	} else {
		err := json.Unmarshal([]byte(subject.Students), &students)
		if err != nil {
			return make([]string, 0)
		}
	}
	return students
}
