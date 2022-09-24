package sql

type Meeting struct {
	ID             string `db:"id"`
	MeetingName    string `db:"meeting_name"`
	TeacherID      string `db:"teacher_id"`
	SubjectID      string `db:"subject_id"`
	Hour           int    `db:"hour"`
	Date           string `db:"date"`
	IsMandatory    bool   `db:"is_mandatory"`
	URL            string `db:"url"`
	Details        string `db:"details"`
	IsSubstitution bool   `db:"is_substitution"`
	Location       string `db:"location"`
	// Ocenjevanje
	IsGrading           bool `db:"is_grading"`
	IsWrittenAssessment bool `db:"is_written_assessment"`
	// Preverjanje znanja
	IsTest bool `db:"is_test"`
	// Beta srečanja so tista, ustvarjena z Proton layerjem. Proton bo prvo ustvaril nova beta, srečanja, katera so neodvisna od drugih in katerih učenci ne morejo videti.
	// Vsak učitelj bo lahko preveril svoje učne ure in svoj urnik s tem ustvarjenim urnikom.
	// Če učitelji ne bodo zadovoljni, se z enim klikom izbriše ta srečanja in se ustvari nov urnik z Proton layerjem, drugače pa se jih z enim klikom spremeni v normalna srečanja,
	// vidna tudi učencem
	IsBeta bool `db:"is_beta"`

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (db *sqlImpl) GetMeeting(id string) (meeting Meeting, err error) {
	err = db.db.Get(&meeting, "SELECT * FROM meetings WHERE id=$1", id)
	return meeting, err
}

func (db *sqlImpl) GetMeetingsOnSpecificTime(date string, hour int) (meetings []Meeting, err error) {
	err = db.db.Select(&meetings, "SELECT * FROM meetings WHERE date=$1 AND hour=$2 ORDER BY id ASC", date, hour)
	return meetings, err
}

func (db *sqlImpl) GetMeetingsOnSpecificDate(date string, includeBeta bool) (meetings []Meeting, err error) {
	if includeBeta {
		err = db.db.Select(&meetings, "SELECT * FROM meetings WHERE date=$1 ORDER BY id ASC", date)
		return meetings, err
	}

	err = db.db.Select(&meetings, "SELECT * FROM meetings WHERE date=$1 AND is_beta=false ORDER BY id ASC", date)
	return meetings, err
}

func (db *sqlImpl) GetMeetingsForTeacherOnSpecificDate(teacherId string, date string) (meetings []Meeting, err error) {
	err = db.db.Select(&meetings, "SELECT * FROM meetings WHERE date=$1 AND teacher_id=$2 ORDER BY id ASC", date, teacherId)
	return meetings, err
}

func (db *sqlImpl) GetMeetingsForSubject(subjectId string) (meetings []Meeting, err error) {
	err = db.db.Select(&meetings, "SELECT * FROM meetings WHERE subject_id=$1 ORDER BY id ASC", subjectId)
	return meetings, err
}

func (db *sqlImpl) InsertMeeting(meeting Meeting) (err error) {
	i := `
	INSERT INTO meetings (meeting_name, teacher_id, subject_id, hour, date, is_mandatory, url, details, is_grading, is_written_assessment, is_test, is_substitution, is_beta, location)
		VALUES (:meeting_name, :teacher_id, :subject_id, :hour, :date, :is_mandatory, :url, :details, :is_grading, :is_written_assessment, :is_test, :is_substitution, :is_beta, :location)
	`
	_, err = db.db.NamedExec(
		i,
		meeting)
	return err
}

func (db *sqlImpl) UpdateMeeting(meeting Meeting) error {
	i := `
	UPDATE meetings SET meeting_name=:meeting_name, teacher_id=:teacher_id,
	                    subject_id=:subject_id, hour=:hour, date=:date,
	                    is_mandatory=:is_mandatory, url=:url, details=:details,
	                    is_grading=:is_grading, is_written_assessment=:is_written_assessment,
	                    is_test=:is_test, is_substitution=:is_substitution, is_beta=:is_beta,
	                    location=:location WHERE id=:id
	`
	_, err := db.db.NamedExec(
		i,
		meeting)
	return err
}

func (db *sqlImpl) MigrateBetaMeetingsToNonBeta() error {
	_, err := db.db.Exec("UPDATE meetings SET is_beta=false WHERE is_beta=true")
	return err
}

func (db *sqlImpl) DeleteBetaMeetings() error {
	_, err := db.db.Exec("DELETE FROM meetings WHERE is_beta=true")
	return err
}

func (db *sqlImpl) GetMeetings() (meetings []Meeting, err error) {
	err = db.db.Select(&meetings, "SELECT * FROM meetings ORDER BY id ASC")
	return meetings, err
}

func (db *sqlImpl) GetMeetingsForSubjectWithIDLower(createdAt string, subjectId string) (meetings []Meeting, err error) {
	// Ok, ja, to je tut ful bad, ampak približno isto, kakor smo prej določevali z ID-ji.
	err = db.db.Select(&meetings, "SELECT * FROM meetings WHERE created_at<=$1 AND subject_id=$2", createdAt, subjectId)
	return meetings, err
}

func (db *sqlImpl) DeleteMeeting(ID string) error {
	_, err := db.db.Exec("DELETE FROM meetings WHERE id=$1", ID)
	return err
}

func (db *sqlImpl) DeleteMeetingsForTeacher(ID string) error {
	_, err := db.db.Exec("DELETE FROM meetings WHERE teacher_id=$1", ID)
	return err
}

func (db *sqlImpl) DeleteMeetingsForSubject(ID string) error {
	_, err := db.db.Exec("DELETE FROM meetings WHERE subject_id=$1", ID)
	return err
}
