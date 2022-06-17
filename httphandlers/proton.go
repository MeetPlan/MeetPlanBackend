package httphandlers

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/proton"
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

func (server *httpImpl) NewProtonRule(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		ruleId, err := strconv.Atoi(r.FormValue("protonRuleId"))
		if err != nil {
			WriteJSON(w, Response{Data: "Failed at converting protonRuleId to integer", Error: err.Error(), Success: false}, http.StatusBadRequest)
			return
		}
		var protonRule = proton.ProtonRule{
			Objects:  make([]proton.ProtonObject, 0),
			RuleName: "Proton pravilo",
			RuleType: ruleId,
		}
		if ruleId == 0 {
			// Full teacher's days on the school - Polni dnevi učitelja na šoli
			//
			// Required arguments:
			// - days - Array/[] of numbers/int corresponding to days in the week (Monday = 0, Sunday = 6)
			// - teacherId - number/int corresponding ID of the teacher in the database
			var days []int
			err := json.Unmarshal([]byte(r.FormValue("days")), &days)
			if err != nil {
				WriteBadRequest(w)
				return
			}

			teacherId, err := strconv.Atoi(r.FormValue("teacherId"))
			if err != nil {
				WriteBadRequest(w)
				return
			}

			for i := 0; i < len(days); i++ {
				day := days[i]
				protonRule.Objects = append(protonRule.Objects, proton.ProtonObject{
					ObjectID: day,
					Type:     "day",
				})
			}
			protonRule.Objects = append(protonRule.Objects, proton.ProtonObject{
				ObjectID: teacherId,
				Type:     "teacher",
			})
		} else if ruleId == 1 {
			// Teacher's hours on the school - Ure učitelja na šoli
			//
			// Required arguments:
			// - days - Array/[] of numbers/int corresponding to days in the week (Monday = 0, Sunday = 6)
			// - hours - Array/[] of numbers/int corresponding to hours of the school day (0th hour = 0, 6th hour = 6)
			// - teacherId - number/int corresponding ID of the teacher in the database
			var days []int
			err := json.Unmarshal([]byte(r.FormValue("days")), &days)
			if err != nil {
				WriteBadRequest(w)
				return
			}

			var hours []int
			err = json.Unmarshal([]byte(r.FormValue("hours")), &hours)
			if err != nil {
				WriteBadRequest(w)
				return
			}

			teacherId, err := strconv.Atoi(r.FormValue("teacherId"))
			if err != nil {
				WriteBadRequest(w)
				return
			}

			for i := 0; i < len(days); i++ {
				day := days[i]
				protonRule.Objects = append(protonRule.Objects, proton.ProtonObject{
					ObjectID: day,
					Type:     "day",
				})
			}
			for i := 0; i < len(hours); i++ {
				hour := hours[i]
				protonRule.Objects = append(protonRule.Objects, proton.ProtonObject{
					ObjectID: hour,
					Type:     "hour",
				})
			}
			protonRule.Objects = append(protonRule.Objects, proton.ProtonObject{
				ObjectID: teacherId,
				Type:     "teacher",
			})
		} else if ruleId == 2 || ruleId == 3 || ruleId == 4 {
			// Subject groups - Skupine predmetov - Rule ID 2
			// Subjects before or after class - Predmeti pred ali po pouku - Rule ID 3
			// Subjects with stacked hours - Predmeti z blok urami - Rule ID 4
			//
			// Required arguments:
			// - subject - Array/[] of numbers/int corresponding to ID(s) of subject(s) in the database
			var subjects []int
			err := json.Unmarshal([]byte(r.FormValue("subjects")), &subjects)
			if err != nil {
				WriteBadRequest(w)
				return
			}

			for i := 0; i < len(subjects); i++ {
				subject := subjects[i]
				protonRule.Objects = append(protonRule.Objects, proton.ProtonObject{
					ObjectID: subject,
					Type:     "subject",
				})
			}
		}
		err = server.proton.NewProtonRule(protonRule)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to add a new rule", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		marshal, err := json.Marshal(protonRule)
		WriteJSON(w, Response{Data: "Successfully added a new rule", Error: string(marshal), Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}

func (server *httpImpl) GetProtonRules(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		WriteJSON(w, Response{Data: server.proton.GetProtonConfig(), Success: false}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}
