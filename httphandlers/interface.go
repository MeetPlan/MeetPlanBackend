package httphandlers

import (
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"go.uber.org/zap"
	"net/http"
)

type httpImpl struct {
	logger *zap.SugaredLogger
	db     sql.SQL
}

type HTTP interface {
	Login(w http.ResponseWriter, r *http.Request)
	NewUser(w http.ResponseWriter, r *http.Request)

	GetSelfTestingTeacher(w http.ResponseWriter, r *http.Request)
	PatchSelfTesting(w http.ResponseWriter, r *http.Request)
	GetPDFSelfTestingReportStudent(w http.ResponseWriter, r *http.Request)

	NewClass(w http.ResponseWriter, r *http.Request)
	GetClasses(w http.ResponseWriter, r *http.Request)
}

func NewHTTPInterface(logger *zap.SugaredLogger, db sql.SQL) HTTP {
	return &httpImpl{
		logger: logger,
		db:     db,
	}
}
