package httphandlers

import (
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (server *httpImpl) ManageTeacherAbsences(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		meetingId, err := strconv.Atoi(mux.Vars(r)["meeting_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		absences, err := server.proton.ManageAbsences(meetingId)
		if err != nil {
			WriteJSON(w, Response{Data: "Proton failed to optimize timetable", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		WriteJSON(w, Response{Data: absences, Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}
