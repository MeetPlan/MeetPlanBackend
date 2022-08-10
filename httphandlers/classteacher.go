package httphandlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (server *httpImpl) ExcuseAbsence(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		return
	}

	if user.Role != TEACHER {
		WriteForbiddenJWT(w)
		return
	}
	studentId, err := strconv.Atoi(mux.Vars(r)["student_id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	absenceId, err := strconv.Atoi(mux.Vars(r)["absence_id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	classes, err := server.db.GetClasses()
	if err != nil {
		return
	}
	var valid = false
	for i := 0; i < len(classes); i++ {
		class := classes[i]
		var users []int
		err := json.Unmarshal([]byte(class.Students), &users)
		if err != nil {
			return
		}
		for j := 0; j < len(users); j++ {
			if users[j] == studentId && class.Teacher == user.ID {
				valid = true
				break
			}
		}
		if valid {
			break
		}
	}
	if !valid {
		WriteForbiddenJWT(w)
		return
	}
	absence, err := server.db.GetAbsence(absenceId)
	if err != nil {
		return
	}
	absence.IsExcused = true
	err = server.db.UpdateAbsence(absence)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}
