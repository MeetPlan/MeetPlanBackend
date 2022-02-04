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

	GetClass(id int) (Class, error)
	InsertClass(class Class) (err error)
	GetLastClassID() (id int)
	UpdateClass(class Class) error
	GetClasses() ([]Class, error)
}

func NewSQL(driver string, drivername string, logger *zap.SugaredLogger) (SQL, error) {
	db, err := sqlx.Connect(driver, drivername)
	return &sqlImpl{
		db:     db,
		logger: logger,
	}, err
}
