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
	dt := time.Now()
	date := dt.Format("02-01-2006")
	results, err := server.db.GetTestingResults(date)
	if err != nil {
		WriteJSON(w, Response{Success: false, Error: err}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Success: true, Data: results}, http.StatusOK)
}

func (server *httpImpl) PatchSelfTesting(w http.ResponseWriter, r *http.Request) {
	student_id, err := strconv.Atoi(mux.Vars(r)["student_id"])
	if err != nil {
		return
	}
	dt := time.Now()
	date := dt.Format("02-01-2006")

	results, err := server.db.GetTestingResult(date, student_id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			t := sql.Testing{
				Date: date,
				ID: 
			}
		}
		WriteJSON(w, Response{Success: false, Error: err}, http.StatusInternalServerError)
		return
	}
	if results
	results.Result = r.FormValue("result")

	WriteJSON(w, Response{Success: true, Data: results}, http.StatusOK)
}
