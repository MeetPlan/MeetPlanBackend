package sql

type GradingTerm struct {
	ID                  string
	TeacherID           string `db:"teacher_id"`
	GradingID           string `db:"grading_id"`
	Date                string
	Hour                int
	Name                string
	Description         string
	Term                int // n-ti rok
	GradeAutoselectType int `db:"grade_autoselect_type"`

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (db *sqlImpl) GetGradingTerm(id string) (gradingTerm GradingTerm, err error) {
	err = db.db.Get(&gradingTerm, "SELECT * FROM grading_terms WHERE id=$1", id)
	return gradingTerm, err
}

func (db *sqlImpl) GetGradingTermsForGrading(gradingId string) (gradingTerms []GradingTerm, err error) {
	err = db.db.Select(&gradingTerms, "SELECT * FROM grading_terms WHERE grading_id=$1 ORDER BY date ASC", gradingId)
	return gradingTerms, err
}

func (db *sqlImpl) InsertGradingTerm(gradingTerm GradingTerm) error {
	i := `
	INSERT INTO grading_terms
	    (teacher_id, grading_id, date, hour, name, description, term, grade_autoselect_type) VALUES
	    (:teacher_id, :grading_id, :date, :hour, :name, :description, :term, :grade_autoselect_type)
	`
	_, err := db.db.NamedExec(
		i,
		gradingTerm)
	return err
}

func (db *sqlImpl) UpdateGradingTerm(gradingTerm GradingTerm) error {
	_, err := db.db.NamedExec(
		"UPDATE grading_terms SET teacher_id=:teacher_id, date=:date, hour=:hour, name=:name, description=:description, term=:term, grade_autoselect_type=:grade_autoselect_type WHERE id=:id",
		gradingTerm)
	return err
}

func (db *sqlImpl) DeleteGradingTermsByGrading(ID string) error {
	_, err := db.db.Exec("DELETE FROM grading_terms WHERE grading_id=$1", ID)
	return err
}

func (db *sqlImpl) DeleteGradingTerm(ID string) error {
	_, err := db.db.Exec("DELETE FROM grading_terms WHERE id=$1", ID)
	return err
}
