package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
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
	_, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		return
	}
	classes, err := server.db.GetClasses()
	if err != nil {
		return
	}
	if classes == nil {
		classes = make([]sql.Class, 0)
	}
	WriteJSON(w, Response{Success: true, Data: classes}, http.StatusOK)
}

func (server *httpImpl) AssignUserToClass(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		return
	}
	if jwt["role"] != "admin" {
		WriteForbiddenJWT(w)
		return
	}

	classId, err := strconv.Atoi(mux.Vars(r)["class_id"])
	if err != nil {
		return
	}
	userId, err := strconv.Atoi(mux.Vars(r)["user_id"])
	if err != nil {
		return
	}
	class, err := server.db.GetClass(classId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	var m []int
	err = json.Unmarshal([]byte(class.Students), &m)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	m = append(m, userId)

	s, err := json.Marshal(m)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	class.Students = string(s)

	err = server.db.UpdateClass(class)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
}
