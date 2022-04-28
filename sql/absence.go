package sql

type Absence struct {
	ID          int
	UserID      int    `db:"user_id"`
	TeacherID   int    `db:"teacher_id"`
	MeetingID   int    `db:"meeting_id"`
	AbsenceType string `db:"absence_type"`
	IsExcused   bool   `db:"is_excused"`
}

func (db *sqlImpl) GetLastAbsenceID() int {
	var id int
	err := db.db.Get(&id, "SELECT id FROM absence WHERE id = (SELECT MAX(id) FROM absence)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) GetAbsence(id int) (absence Absence, err error) {
	err = db.db.Get(&absence, "SELECT * FROM absence WHERE id=$1", id)
	return absence, err
}

func (db *sqlImpl) GetAbsenceForUserMeeting(meeting_id int, user_id int) (absence Absence, err error) {
	err = db.db.Get(&absence, "SELECT * FROM absence WHERE meeting_id=$1 AND user_id=$2", meeting_id, user_id)
	return absence, err
}

func (db *sqlImpl) GetAbsencesForUser(user_id int) (absence []Absence, err error) {
	err = db.db.Select(&absence, "SELECT * FROM absence WHERE user_id=$1 ORDER BY id ASC", user_id)
	return absence, err
}

func (db *sqlImpl) GetAllAbsences(id int) (absences []Absence, err error) {
	err = db.db.Select(&absences, "SELECT * FROM absence WHERE user_id=$1 ORDER BY id ASC", id)
	return absences, err
}

func (db *sqlImpl) InsertAbsence(absence Absence) error {
	_, err := db.db.NamedExec(
		"INSERT INTO absence (id, user_id, teacher_id, meeting_id, absence_type, is_excused) VALUES (:id, :user_id, :teacher_id, :meeting_id, :absence_type, :is_excused)",
		absence)
	return err
}

func (db *sqlImpl) UpdateAbsence(absence Absence) error {
	_, err := db.db.NamedExec(
		"UPDATE absence SET user_id=:user_id, teacher_id=:teacher_id, meeting_id=:meeting_id, absence_type=:absence_type, is_excused=:is_excused WHERE id=:id",
		absence)
	return err
}

func (db *sqlImpl) DeleteAbsencesForTeacher(userId int) {
	db.db.Exec("DELETE FROM absence WHERE teacher_id=$1", userId)
}

func (db *sqlImpl) DeleteAbsencesForUser(userId int) {
	db.db.Exec("DELETE FROM absence WHERE user_id=$1", userId)
}
