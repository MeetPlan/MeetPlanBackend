package sql

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type sqlImpl struct {
	db     *sqlx.DB
	logger *zap.SugaredLogger
}

func (db *sqlImpl) Init() {
	db.db.MustExec(schema)
}

type SQL interface {
	CheckToken(loginToken string) (User, error)
	GetRandomToken(currentUser User) (string, error)

	Init()
	Exec(query string) error

	UpdateTestingResult(testing Testing) error
	InsertTestingResult(testing Testing) error
	GetTestingResults(date string, classId string) ([]TestingJSON, error)
	GetAllTestingsForUser(id string) (testing []Testing, err error)
	GetTestingResult(date string, id string) (Testing, error)
	GetTestingResultByID(id string) (Testing, error)

	DeleteTeacherSelfTesting(teacherId string) error
	DeleteUserSelfTesting(userId string) error

	GetUser(id string) (user User, err error)
	GetUserByLoginToken(loginToken string) (user User, err error)
	InsertUser(user User) (err error)

	GetUserByEmail(email string) (user User, err error)
	CheckIfAdminIsCreated() bool
	GetAllUsers() (users []User, err error)
	UpdateUser(user User) error
	DeleteUser(ID string) error
	GetTeachers() ([]User, error)
	GetPrincipal() (principal User, err error)

	GetClass(id string) (Class, error)
	InsertClass(class Class) (err error)

	UpdateClass(class Class) error
	GetClasses() ([]Class, error)
	DeleteClass(ID string) error
	DeleteTeacherClasses(teacherId string) error
	DeleteUserClasses(userId string)

	GetMeeting(id string) (meeting Meeting, err error)
	GetMeetingsOnSpecificTime(date string, hour int) (meetings []Meeting, err error)
	GetMeetingsForSubject(subjectId string) (meetings []Meeting, err error)
	GetMeetingsForTeacherOnSpecificDate(teacherId string, date string) (meetings []Meeting, err error)
	InsertMeeting(meeting Meeting) (err error)
	UpdateMeeting(meeting Meeting) error

	GetMeetings() (meetings []Meeting, err error)
	GetMeetingsForSubjectWithIDLower(createdAt string, subjectId string) (meetings []Meeting, err error)
	DeleteMeeting(ID string) error
	GetMeetingsOnSpecificDate(date string, includeBeta bool) (meetings []Meeting, err error)
	DeleteMeetingsForTeacher(ID string) error
	DeleteMeetingsForSubject(ID string) error
	MigrateBetaMeetingsToNonBeta() error
	DeleteBetaMeetings() error

	GetAbsence(id string) (absence Absence, err error)
	GetAllAbsences(id string) (absences []Absence, err error)
	InsertAbsence(absence Absence) error
	UpdateAbsence(absence Absence) error
	GetAbsenceForUserMeeting(meeting_id string, user_id string) (absence Absence, err error)
	GetAbsencesForUser(user_id string) (absence []Absence, err error)
	DeleteAbsencesForTeacher(userId string)
	DeleteAbsencesForUser(userId string)

	GetSubject(id string) (subject Subject, err error)
	GetAllSubjectsForTeacher(id string) (subject []Subject, err error)
	GetAllSubjectsForUser(id string) (subject []Subject, err error)
	GetSubjectsWithSpecificLongName(longName string) (subject []Subject, err error)
	InsertSubject(subject Subject) error
	UpdateSubject(subject Subject) error
	GetAllSubjects() (subject []Subject, err error)
	GetStudents() (message []User, err error)
	DeleteSubject(subject Subject) error
	DeleteStudentSubject(userId string)

	GetGrade(id string) (grade Grade, err error)
	GetGradesForUser(userId string) (grades []Grade, err error)
	GetGradesForUserInSubject(userId string, subjectId string) (grades []Grade, err error)
	CheckIfFinal(userId string, subjectId string) (grade Grade, err error)
	InsertGrade(grade Grade) error
	UpdateGrade(grade Grade) error
	DeleteGrade(ID string) error
	DeleteGradesByTeacherID(ID string) error
	DeleteGradesByUserID(ID string) error

	GetHomework(id string) (homework Homework, err error)
	GetHomeworkForSubject(id string) (homework []Homework, err error)
	InsertHomework(homework Homework) error
	UpdateHomework(homework Homework) error
	DeleteHomework(ID string) error

	GetStudentHomework(id string) (homework StudentHomework, err error)
	GetStudentHomeworkForUser(homeworkId string, userId string) (homework StudentHomework, err error)
	DeleteStudentHomeworkByStudentID(ID string) error
	GetHomeworkForTeacher(teacherId string) (homework []Homework, err error)
	GetStudentsHomeworkByHomeworkID(id string, meetingId string) (homework []StudentHomeworkJSON, err error)
	GetStudentsHomework(id string) (homework []StudentHomework, err error)
	InsertStudentHomework(homework StudentHomework) error
	UpdateStudentHomework(homework StudentHomework) error
	DeleteStudentHomework(ID string) error
	DeleteStudentHomeworkByHomeworkID(ID string) error
	DeleteAllTeacherHomeworks(ID string)

	GetCommunication(id string) (communication Communication, err error)
	InsertCommunication(communication Communication) (err error)
	UpdateCommunication(communication Communication) error

	GetCommunications() (communication []Communication, err error)
	DeleteCommunication(ID string) error
	DeleteUserCommunications(userId string)

	GetMessage(id string) (message Message, err error)
	GetCommunicationMessages(communicationId string) (messages []Message, err error)
	GetAllUnreadMessages(userId string) (messages []Message, err error)
	InsertMessage(message Message) (err error)
	UpdateMessage(message Message) error

	GetAllMessages() (messages []Message, err error)
	DeleteMessage(ID string) error

	GetMeal(id string) (meal Meal, err error)
	InsertMeal(meal Meal) (err error)
	UpdateMeal(meal Meal) error

	GetMeals() (meals []Meal, err error)
	DeleteMeal(ID string) error

	GetNotification(id string) (notification NotificationSQL, err error)
	GetAllNotifications() (notifications []NotificationSQL, err error)
	InsertNotification(notification NotificationSQL) (err error)
	UpdateNotification(notification NotificationSQL) error

	DeleteNotification(ID string) error

	GetImprovement(id string) (improvement Improvement, err error)
	GetImprovementsForStudent(studentId string) (improvements []Improvement, err error)
	InsertImprovement(improvement Improvement) error
	UpdateImprovement(homework Homework) error
	DeleteImprovement(ID string) error

	GetDocument(id string) (document Document, err error)
	GetAllDocuments() (documents []Document, err error)
	InsertDocument(document Document) error
	DeleteDocument(id string)
}

func NewSQL(driver string, drivername string, logger *zap.SugaredLogger) (SQL, error) {
	db, err := sqlx.Connect(driver, drivername)
	return &sqlImpl{
		db:     db,
		logger: logger,
	}, err
}

func (db *sqlImpl) Exec(query string) error {
	_, err := db.db.Exec(query)
	return err
}
