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
}

func NewSQL(driver string, drivername string, logger *zap.SugaredLogger) (SQL, error) {
	db, err := sqlx.Connect(driver, drivername)
	return &sqlImpl{
		db:     db,
		logger: logger,
	}, err
}
