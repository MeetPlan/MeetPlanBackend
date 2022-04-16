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
	Init()

	UpdateTestingResult(testing Testing) error
	InsertTestingResult(testing Testing) error
	GetTestingResults(date string, classId int) ([]TestingJSON, error)
	GetAllTestingsForUser(id int) (testing []Testing, err error)
	GetTestingResult(date string, id int) (Testing, error)
	GetTestingResultByID(id int) (Testing, error)
	GetLastTestingID() int

	GetUser(id int) (message User, err error)
	InsertUser(user User) (err error)
	GetLastUserID() (id int)
	GetUserByEmail(email string) (message User, err error)
	CheckIfAdminIsCreated() bool
	GetAllUsers() (users []User, err error)
	UpdateUser(user User) error
	DeleteUser(ID int) error
	GetTeachers() ([]User, error)

	GetClass(id int) (Class, error)
	InsertClass(class Class) (err error)
	GetLastClassID() (id int)
	UpdateClass(class Class) error
	GetClasses() ([]Class, error)
	DeleteClass(ID int) error

	GetMeeting(id int) (meeting Meeting, err error)
	GetMeetingsOnSpecificTime(date string, hour int) (meetings []Meeting, err error)
	GetMeetingsOnSpecificDateAndClass(date string, classId int) (meetings []Meeting, err error)
	InsertMeeting(meeting Meeting) (err error)
	UpdateMeeting(meeting Meeting) error
	GetLastMeetingID() (id int)
	GetMeetings() (meetings []Meeting, err error)
	DeleteMeeting(ID int) error
	GetMeetingsOnSpecificDate(date string) (meetings []Meeting, err error)

	GetLastAbsenceID() int
	GetAbsence(id int) (absence Absence, err error)
	GetAllAbsences(id int) (absences []Absence, err error)
	InsertAbsence(absence Absence) error
	UpdateAbsence(absence Absence) error
	GetAbsenceForUserMeeting(meeting_id int, user_id int) (absence Absence, err error)
	GetAbsencesForUser(user_id int) (absence []Absence, err error)

	GetLastSubjectID() int
	GetSubject(id int) (subject Subject, err error)
	GetAllSubjectsForTeacher(id int) (subject []Subject, err error)
	GetAllSubjectsForUser(id int) (subject []Subject, err error)
	InsertSubject(subject Subject) error
	UpdateSubject(subject Subject) error
	GetAllSubjects() (subject []Subject, err error)
	GetStudents() (message []User, err error)
	DeleteSubject(subject Subject) error

	GetLastGradeID() int
	GetGrade(id int) (grade Grade, err error)
	GetGradesForUser(userId int) (grades []Grade, err error)
	GetGradesForUserInSubject(userId int, subjectId int) (grades []Grade, err error)
	CheckIfFinal(userId int, subjectId int) (grade Grade, err error)
	InsertGrade(grade Grade) error
	UpdateGrade(grade Grade) error
	DeleteGrade(ID int) error

	GetLastHomeworkID() int
	GetHomework(id int) (homework Homework, err error)
	GetHomeworkForSubject(id int) (homework []Homework, err error)
	InsertHomework(homework Homework) error
	UpdateHomework(homework Homework) error
	DeleteHomework(ID int) error

	GetLastStudentHomeworkID() int
	GetStudentHomework(id int) (homework StudentHomework, err error)
	GetStudentHomeworkForUser(homeworkId int, userId int) (homework StudentHomework, err error)
	GetStudentsHomeworkByHomeworkID(id int, meetingId int) (homework []StudentHomeworkJSON, err error)
	GetStudentsHomework(id int) (homework []StudentHomework, err error)
	InsertStudentHomework(homework StudentHomework) error
	UpdateStudentHomework(homework StudentHomework) error
	DeleteStudentHomework(ID int) error

	GetCommunication(id int) (communication Communication, err error)
	InsertCommunication(communication Communication) (err error)
	UpdateCommunication(communication Communication) error
	GetLastCommunicationID() (id int)
	GetCommunications() (communication []Communication, err error)
	DeleteCommunication(ID int) error

	GetMessage(id int) (message Message, err error)
	GetCommunicationMessages(communicationId int) (messages []Message, err error)
	GetAllUnreadMessages(userId int) (messages []Message, err error)
	InsertMessage(message Message) (err error)
	UpdateMessage(message Message) error
	GetLastMessageID() (id int)
	GetAllMessages() (messages []Message, err error)
	DeleteMessage(ID int) error

	GetMeal(id int) (meal Meal, err error)
	InsertMeal(meal Meal) (err error)
	UpdateMeal(meal Meal) error
	GetLastMealID() (id int)
	GetMeals() (meals []Meal, err error)
	DeleteMeal(ID int) error

	GetNotification(id int) (notification NotificationSQL, err error)
	GetAllNotifications() (notifications []NotificationSQL, err error)
	InsertNotification(notification NotificationSQL) (err error)
	UpdateNotification(notification NotificationSQL) error
	GetLastNotificationID() (id int)
	DeleteNotification(ID int) error
}

func NewSQL(driver string, drivername string, logger *zap.SugaredLogger) (SQL, error) {
	db, err := sqlx.Connect(driver, drivername)
	return &sqlImpl{
		db:     db,
		logger: logger,
	}, err
}
