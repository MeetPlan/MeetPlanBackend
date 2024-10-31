package sql

type Grade struct {
	ID          string
	UserID      string  `db:"user_id"`
	TeacherID   string  `db:"teacher_id"`
	TermID      *string `db:"term_id"`
	SubjectID   string  `db:"subject_id"`
	Grade       int
	Date        string
	IsWritten   bool `db:"is_written"`
	IsFinal     bool `db:"is_final"`
	Period      int
	Description string
	CanPatch    bool `db:"can_patch"`

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (db *sqlImpl) GetGrade(id string) (grade Grade, err error) {
	err = db.db.Get(&grade, "SELECT * FROM grades WHERE id=$1", id)
	return grade, err
}

func (db *sqlImpl) GetGradesForUser(userId string) (grades []Grade, err error) {
	err = db.db.Select(&grades, "SELECT * FROM grades WHERE user_id=$1 ORDER BY id ASC", userId)
	return grades, err
}

func (db *sqlImpl) GetGradesForTerm(termId string) (grades []Grade, err error) {
	err = db.db.Select(&grades, "SELECT * FROM grades WHERE term_id=$1 ORDER BY id ASC", termId)
	return grades, err
}

func (db *sqlImpl) GetGradeForTermAndUser(termId string, userId string) (grade Grade, err error) {
	err = db.db.Get(&grade, "SELECT * FROM grades WHERE term_id=$1 AND user_id=$2", termId, userId)
	return grade, err
}

func (db *sqlImpl) CheckIfFinal(userId string, subjectId string) (grade Grade, err error) {
	err = db.db.Get(&grade, "SELECT * FROM grades WHERE user_id=$1 AND subject_id=$2 AND is_final=true", userId, subjectId)
	return grade, err
}

func (db *sqlImpl) GetGradesForUserInSubject(userId string, subjectId string) (grades []Grade, err error) {
	err = db.db.Select(&grades, "SELECT * FROM grades WHERE user_id=$1 AND subject_id=$2 ORDER BY id ASC", userId, subjectId)
	return grades, err
}

func (db *sqlImpl) InsertGrade(grade Grade) error {
	i := `
	INSERT INTO grades
	    (user_id, teacher_id, term_id, subject_id, date, is_written, grade, period, description, is_final, can_patch) VALUES
	    (:user_id, :teacher_id, :term_id, :subject_id, :date, :is_written, :grade, :period, :description, :is_final, :can_patch)
	`
	_, err := db.db.NamedExec(
		i,
		grade)
	return err
}

func (db *sqlImpl) UpdateGrade(grade Grade) error {
	_, err := db.db.NamedExec(
		"UPDATE grades SET user_id=:user_id, teacher_id=:teacher_id, term_id=:term_id, subject_id=:subject_id, date=:date, is_written=:is_written, grade=:grade, period=:period, description=:description, can_patch=:can_patch WHERE id=:id",
		grade)
	return err
}

func (db *sqlImpl) DeleteGrade(ID string) error {
	_, err := db.db.Exec("DELETE FROM grades WHERE id=$1", ID)
	return err
}

func (db *sqlImpl) DeleteGradeByTermAndUser(termId string, userId string) error {
	_, err := db.db.Exec("DELETE FROM grades WHERE term_id=$1 AND user_id=$2", termId, userId)
	return err
}

func (db *sqlImpl) DeleteGradesByTeacherID(ID string) error {
	_, err := db.db.Exec("DELETE FROM grades WHERE teacher_id=$1", ID)
	return err
}

func (db *sqlImpl) DeleteGradesByTermID(ID string) error {
	_, err := db.db.Exec("DELETE FROM grades WHERE term_id=$1", ID)
	return err
}

func (db *sqlImpl) DeleteGradesByUserID(ID string) error {
	_, err := db.db.Exec("DELETE FROM grades WHERE user_id=$1", ID)
	return err
}
