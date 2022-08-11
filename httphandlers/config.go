package httphandlers

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"net/http"
	"strconv"
)

type ParentConfig struct {
	ParentViewGrades   bool `json:"parent_view_grades"`
	ParentViewAbsences bool `json:"parent_view_absences"`
	ParentViewHomework bool `json:"parent_view_homework"`
	ParentViewGradings bool `json:"parent_view_gradings"`
}

func (server *httpImpl) GetConfig(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	WriteJSON(w, Response{Data: server.config, Success: true}, http.StatusOK)
}

func (server *httpImpl) UpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	schoolPostCode, err := strconv.Atoi(r.FormValue("school_post_code"))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	parentViewGrades, err := strconv.ParseBool(r.FormValue("parent_view_grades"))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	parentViewAbsences, err := strconv.ParseBool(r.FormValue("parent_view_absences"))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	parentViewHomework, err := strconv.ParseBool(r.FormValue("parent_view_homework"))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	parentViewGradings, err := strconv.ParseBool(r.FormValue("parent_view_gradings"))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	blockRegistrations, err := strconv.ParseBool(r.FormValue("block_registrations"))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	blockMeals, err := strconv.ParseBool(r.FormValue("block_meals"))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	// admins, pls no shady business when patching dates, otherwise, system will not work anymore
	err = json.Unmarshal([]byte(r.FormValue("school_free_days")), &server.config.SchoolFreeDays)
	if err != nil {
		WriteBadRequest(w)
		return
	}
	server.config.SchoolPostCode = schoolPostCode
	server.config.SchoolCountry = r.FormValue("school_country")
	server.config.SchoolAddress = r.FormValue("school_address")
	server.config.SchoolCity = r.FormValue("school_city")
	server.config.SchoolName = r.FormValue("school_name")
	server.config.ParentViewGrades = parentViewGrades
	server.config.ParentViewAbsences = parentViewAbsences
	server.config.ParentViewHomework = parentViewHomework
	server.config.ParentViewGradings = parentViewGradings
	server.config.BlockRegistrations = blockRegistrations
	server.config.BlockMeals = blockMeals
	err = sql.SaveConfig(server.config)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to save config", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)

}

func (server *httpImpl) ParentConfig(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if user.Role != PARENT {
		WriteForbiddenJWT(w)
		return
	}
	WriteJSON(w, Response{Data: ParentConfig{
		ParentViewGrades:   server.config.ParentViewGrades,
		ParentViewAbsences: server.config.ParentViewAbsences,
		ParentViewHomework: server.config.ParentViewHomework,
		ParentViewGradings: server.config.ParentViewGradings,
	}, Success: true}, http.StatusOK)
}
