package sql

type Meeting struct {
	ID          int    `db:"id"`
	MeetingName string `db:"meeting_name"`
	TeacherID   int    `db:"teacher_id"`
	ClassID     int    `db:"class_id"`
	Hour        int    `db:"hour"`
	Date        string `db:"date"`
	IsMandatory bool   `db:"is_mandatory"`
	URL         string `db:"url"`
	Details     string `db:"details"`
	// Ocenjevanje
	IsGrading           bool `db:"is_grading"`
	IsWrittenAssessment bool `db:"is_written_assessment"`
	// Preverjanje znanja
	IsTest bool `db:"is_test"`
}

func (db *sqlImpl) GetMeeting(id int) (meeting Meeting, err error) {
	err = db.db.Get(&meeting, "SELECT * FROM meetings WHERE id=$1", id)
	return meeting, err
}

func (db *sqlImpl) GetMeetingsOnSpecificTime(date string, hour int) (meetings []Meeting, err error) {
	err = db.db.Select(&meetings, "SELECT * FROM meetings WHERE date=$1 AND hour=$2", date, hour)
	return meetings, err
}

func (db *sqlImpl) GetMeetingsOnSpecificDate(date string) (meetings []Meeting, err error) {
	err = db.db.Select(&meetings, "SELECT * FROM meetings WHERE date=$1", date)
	return meetings, err
}

func (db *sqlImpl) GetMeetingsOnSpecificDateAndClass(date string, classId int) (meetings []Meeting, err error) {
	err = db.db.Select(&meetings, "SELECT * FROM meetings WHERE date=$1 AND class_id=$2", date, classId)
	return meetings, err
}

func (db *sqlImpl) InsertMeeting(meeting Meeting) (err error) {
	i := `
	INSERT INTO meetings (id, meeting_name, teacher_id, class_id, hour, date, is_mandatory, url, details, is_grading, is_written_assessment, is_test)
		VALUES (:id, :meeting_name, :teacher_id, :class_id, :hour, :date, :is_mandatory, :url, :details, :is_grading, :is_written_assessment, :is_test)
	`
	_, err = db.db.NamedExec(
		i,
		meeting)
	return err
}

func (db *sqlImpl) UpdateMeeting(meeting Meeting) error {
	i := `
	UPDATE meetings SET meeting_name=:meeting_name, teacher_id=:teacher_id,
	                    class_id=:class_id, hour=:hour, date=:date,
	                    is_mandatory=:is_mandatory, url=:url, details=:details,
	                    is_grading=:is_grading, is_written_assessment=:is_written_assessment,
	                    is_test=:is_test WHERE id=:id
	`
	_, err := db.db.NamedExec(
		i,
		meeting)
	return err
}

func (db *sqlImpl) GetLastMeetingID() (id int) {
	err := db.db.Get(&id, "SELECT id FROM meetings WHERE id = (SELECT MAX(id) FROM meetings)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) GetMeetings() (meetings []Meeting, err error) {
	err = db.db.Select(&meetings, "SELECT * FROM meetings")
	return meetings, err
}
func (db *sqlImpl) DeleteMeeting(ID int) error {
	_, err := db.db.Exec("DELETE FROM meetings WHERE id=$1", ID)
	return err
}
