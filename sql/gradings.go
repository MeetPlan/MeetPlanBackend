package sql

type Grading struct {
	ID          string
	SubjectID   string `db:"subject_id"`
	TeacherID   string `db:"teacher_id"`
	Name        string
	Description string
	GradingType int    `db:"grading_type"`
	SchoolYear  string `db:"school_year"`
	Period      int

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (db *sqlImpl) GetGrading(id string) (grading Grading, err error) {
	err = db.db.Get(&grading, "SELECT * FROM gradings WHERE id=$1", id)
	return grading, err
}

func (db *sqlImpl) GetGradingsForSubject(subjectId string) (gradings []Grading, err error) {
	err = db.db.Select(&gradings, "SELECT * FROM gradings WHERE subject_id=$1 ORDER BY period ASC, created_at ASC", subjectId)
	return gradings, err
}

func (db *sqlImpl) InsertGrading(grading Grading) error {
	i := `
	INSERT INTO gradings
	    (subject_id, teacher_id, name, description, grading_type, school_year, period) VALUES
	    (:subject_id, :teacher_id, :name, :description, :grading_type, :school_year, :period)
	`
	_, err := db.db.NamedExec(
		i,
		grading)
	return err
}

func (db *sqlImpl) UpdateGrading(grading Grading) error {
	_, err := db.db.NamedExec(
		"UPDATE gradings SET teacher_id=:teacher_id, name=:name, description=:description, grading_type=:grading_type, school_year=:school_year, period=:period WHERE id=:id",
		grading)
	return err
}

func (db *sqlImpl) DeleteGrading(ID string) error {
	_, err := db.db.Exec("DELETE FROM gradings WHERE id=$1", ID)
	return err
}
