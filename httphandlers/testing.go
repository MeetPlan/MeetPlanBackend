package httphandlers

import (
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	Error   interface{} `json:"error"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

func (server *httpImpl) GetSelfTestingTeacher(w http.ResponseWriter, r *http.Request) {
	classId, err := strconv.Atoi(mux.Vars(r)["class_id"])
	if err != nil {
		return
	}
	dt := time.Now()
	date := dt.Format("02-01-2006")
	results, err := server.db.GetTestingResults(date, classId)
	if err != nil {
		WriteJSON(w, Response{Success: false, Error: err}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Success: true, Data: results}, http.StatusOK)
}

func (server *httpImpl) PatchSelfTesting(w http.ResponseWriter, r *http.Request) {
	studentId, err := strconv.Atoi(mux.Vars(r)["student_id"])
	if err != nil {
		return
	}
	dt := time.Now()
	date := dt.Format("02-01-2006")

	results, err := server.db.GetTestingResult(date, studentId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			results = sql.Testing{
				Date:      date,
				ID:        server.db.GetLastTestingID(),
				UserID:    studentId,
				TeacherID: 0,
				ClassID:   0,
				Result:    r.FormValue("result"),
			}
			err := server.db.InsertTestingResult(results)
			if err != nil {
				WriteJSON(w, Response{Success: false, Error: err}, http.StatusInternalServerError)
				return
			}
		} else {
			WriteJSON(w, Response{Success: false, Error: err}, http.StatusInternalServerError)
			return
		}
	}
	results.Result = r.FormValue("result")
	err = server.db.UpdateTestingResult(results)
	if err != nil {
		WriteJSON(w, Response{Success: false, Error: err}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Success: true, Data: results}, http.StatusOK)
}
