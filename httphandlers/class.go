package httphandlers

import (
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"net/http"
	"strconv"
)

func (server *httpImpl) NewClass(w http.ResponseWriter, r *http.Request) {
	className := r.FormValue("name")
	teacherIdStr := fmt.Sprint(r.FormValue("teacher_id"))
	server.logger.Debug(teacherIdStr)

	teacherId, err := strconv.Atoi(teacherIdStr)
	if err != nil {
		WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
		return
	}

	class := sql.Class{ID: server.db.GetLastClassID(), Name: className, Teacher: teacherId}
	server.logger.Debug(class)
	err = server.db.InsertClass(class)
	if err != nil {
		WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Success: true, Data: class.ID}, http.StatusOK)
}

func (server *httpImpl) GetClasses(w http.ResponseWriter, r *http.Request) {
	classes, err := server.db.GetClasses()
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true, Data: classes}, http.StatusOK)
}
