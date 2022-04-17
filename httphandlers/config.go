package httphandlers

import (
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"net/http"
	"strconv"
)

type ParentConfig struct {
	ParentViewGrades   bool `json:"parent_view_grades"`
	ParentViewAbsences bool `json:"parent_view_absences"`
	ParentViewHomework bool `json:"parent_view_homework"`
}

func (server *httpImpl) GetConfig(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		WriteJSON(w, Response{Data: server.config, Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) UpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
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
		server.config.SchoolPostCode = schoolPostCode
		server.config.SchoolCountry = r.FormValue("school_country")
		server.config.SchoolAddress = r.FormValue("school_address")
		server.config.SchoolCity = r.FormValue("school_city")
		server.config.SchoolName = r.FormValue("school_name")
		server.config.ParentViewGrades = parentViewGrades
		server.config.ParentViewAbsences = parentViewAbsences
		server.config.ParentViewHomework = parentViewHomework
		err = sql.SaveConfig(server.config)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to save config", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) ParentConfig(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] != "parent" {
		WriteForbiddenJWT(w)
		return
	}
	WriteJSON(w, Response{Data: ParentConfig{
		ParentViewGrades:   server.config.ParentViewGrades,
		ParentViewAbsences: server.config.ParentViewAbsences,
		ParentViewHomework: server.config.ParentViewHomework,
	}, Success: true}, http.StatusOK)
}
