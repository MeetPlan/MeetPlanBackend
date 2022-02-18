package httphandlers

import (
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (server *httpImpl) NewHomework(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "student" {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		return
	}
	subjectId, err := strconv.Atoi(mux.Vars(r)["subject_id"])
	if err != nil {
		return
	}
	if jwt["role"] == "teacher" {
		subject, err := server.db.GetSubject(subjectId)
		if err != nil {
			return
		}
		if subject.TeacherID != userId {
			WriteForbiddenJWT(w)
			return
		}
	}
	homework := sql.Homework{
		ID:          server.db.GetLastHomeworkID(),
		TeacherID:   userId,
		SubjectID:   subjectId,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		// TODO: We do not support this currently
		ToDate:   "",
		FromDate: "",
	}
	err = server.db.InsertHomework(homework)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusCreated)
}

func (server *httpImpl) GetAllHomeworksForSpecificSubject(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "student" {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		return
	}
	subjectId, err := strconv.Atoi(mux.Vars(r)["subject_id"])
	if err != nil {
		return
	}
	if jwt["role"] == "teacher" {
		subject, err := server.db.GetSubject(subjectId)
		if err != nil {
			return
		}
		if subject.TeacherID != userId {
			WriteForbiddenJWT(w)
			return
		}
	}
	homework, err := server.db.GetHomeworkForSubject(subjectId)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Data: homework, Success: true}, http.StatusOK)
}
