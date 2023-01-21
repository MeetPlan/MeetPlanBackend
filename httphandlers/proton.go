package httphandlers

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/proton"
	"github.com/MeetPlan/MeetPlanBackend/proton/genetic"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"net/http"
	"strconv"
)

// TODO: Implement
func (server *httpImpl) ManageTeacherAbsences(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	//meetingId := mux.Vars(r)["meeting_id"]
	if err != nil {
		WriteBadRequest(w)
		return
	}
	//absences, err := server.proton.ManageAbsences(meetingId)
	if err != nil {
		WriteJSON(w, Response{Data: "Proton failed to optimize timetable", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: []proton.TierGradingList{}, Success: true}, http.StatusOK)
}

func (server *httpImpl) PostProcessTimetable(classes []sql.Class, stableTimetable []proton.ProtonMeeting, cancelPostProcessingBeforeDone bool) ([]proton.ProtonMeeting, error) {
	// Dogajajo se pripetljaji. Ni vsako polnjenje lukenj popolno, zato gremo "zlikati" ta urnik večkrat.
	for i := 0; i < proton.PROTON_REPEAT_POST_PROCESSING; i++ {
		server.logger.Debugw("izvajam post-procesiranje", "nivo", i)

		// Post-procesiranje urnika za vsak razred posebej.
		for i := 0; i < len(classes); i++ {
			class := classes[i]

			server.logger.Debugw("izvajam post-procesiranje", "class", class)

			//var err error
			//stableTimetable, err = server.proton.TimetablePostProcessing(stableTimetable, class, cancelPostProcessingBeforeDone)
			//if err != nil {
			//	return nil, err
			//}
		}
	}
	return stableTimetable, nil
}

func (server *httpImpl) NewProtonRule(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	ruleId, err := strconv.Atoi(r.FormValue("protonRuleId"))
	if err != nil {
		WriteJSON(w, Response{Data: "Failed at converting protonRuleId to integer", Error: err.Error(), Success: false}, http.StatusBadRequest)
		return
	}
	var protonRule = genetic.ProtonRule{
		Objects:  make([]genetic.ProtonObject, 0),
		RuleName: "Proton pravilo",
		RuleType: ruleId,
	}
	if ruleId == 0 {
		// Full teacher's days on the school - Polni dnevi učitelja na šoli
		//
		// Required arguments:
		// - days - Array/[] of numbers/int corresponding to days in the week (Monday = 0, Sunday = 6)
		// - teacherId - number/int corresponding ID of the teacher in the database
		var days []string
		err := json.Unmarshal([]byte(r.FormValue("days")), &days)
		if err != nil {
			WriteBadRequest(w)
			return
		}

		teacherId := r.FormValue("teacherId")
		if err != nil {
			WriteBadRequest(w)
			return
		}

		for i := 0; i < len(days); i++ {
			day := days[i]
			protonRule.Objects = append(protonRule.Objects, genetic.ProtonObject{
				ObjectID: day,
				Type:     "day",
			})
		}
		protonRule.Objects = append(protonRule.Objects, genetic.ProtonObject{
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
		var days []string
		err := json.Unmarshal([]byte(r.FormValue("days")), &days)
		if err != nil {
			WriteBadRequest(w)
			return
		}

		var hours []string
		err = json.Unmarshal([]byte(r.FormValue("hours")), &hours)
		if err != nil {
			WriteBadRequest(w)
			return
		}

		teacherId := r.FormValue("teacherId")
		if err != nil {
			WriteBadRequest(w)
			return
		}

		for i := 0; i < len(days); i++ {
			day := days[i]
			protonRule.Objects = append(protonRule.Objects, genetic.ProtonObject{
				ObjectID: day,
				Type:     "day",
			})
		}
		for i := 0; i < len(hours); i++ {
			hour := hours[i]
			protonRule.Objects = append(protonRule.Objects, genetic.ProtonObject{
				ObjectID: hour,
				Type:     "hour",
			})
		}
		protonRule.Objects = append(protonRule.Objects, genetic.ProtonObject{
			ObjectID: teacherId,
			Type:     "teacher",
		})
	} else if ruleId == 2 || ruleId == 3 || ruleId == 4 {
		// Subject groups - Skupine predmetov - Rule ID 2
		// Subjects before or after class - Predmeti pred ali po pouku - Rule ID 3
		// Subjects with stacked hours - Predmeti z blok urami - Rule ID 4
		//
		// Required arguments:
		// - subjects - Array/[] of numbers/int corresponding to ID(s) of subject(s) in the database
		var subjects []string
		err := json.Unmarshal([]byte(r.FormValue("subjects")), &subjects)
		if err != nil {
			WriteBadRequest(w)
			return
		}

		for i := 0; i < len(subjects); i++ {
			subject := subjects[i]
			protonRule.Objects = append(protonRule.Objects, genetic.ProtonObject{
				ObjectID: subject,
				Type:     "subject",
			})
		}
	}
	//err = server.proton.NewProtonRule(protonRule)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to add a new rule", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	marshal, err := json.Marshal(protonRule)
	WriteJSON(w, Response{Data: "Successfully added a new rule", Error: string(marshal), Success: true}, http.StatusOK)
}

// TODO: Implement
func (server *httpImpl) GetProtonRules(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	//protonConfig := server.proton.GetProtonConfig()
	//for i := 0; i < len(protonConfig.Rules); i++ {
	//	if protonConfig.Rules[i].ID == "" {
	//		UUID, err := uuid.NewUUID()
	//		if err != nil {
	//			continue
	//		}
	//		protonConfig.Rules[i].ID = UUID.String()
	//	}
	//}
	//server.proton.SaveConfig(protonConfig)

	WriteJSON(w, Response{Data: genetic.ProtonConfig{
		Version: "1.0",
		Rules:   []genetic.ProtonRule{},
	}, Success: true}, http.StatusOK)
}

//func GenerateRandomHourForBeforeAfterSubjects() int {
//	return rand.Intn(proton.PROTON_MAX_AFTER_CLASS_HOUR-proton.PROTON_MIN_AFTER_CLASS_HOUR) + proton.PROTON_MIN_AFTER_CLASS_HOUR
//}

//func GenerateBeforeAfterHour(stackedSubjects []string, subject sql.Subject) int {
//	var hour int
//
//	k := rand.Intn(2)
//	// naključna izbira med preduro in pouro
//	if k == 0 || helpers.Contains(stackedSubjects, subject.ID) {
//		// Tako ali tako bomo "zafilali" te luknje
//		hour = GenerateRandomHourForBeforeAfterSubjects()
//	} else {
//		hour = 0
//	}
//
//	return hour
//}

func (server *httpImpl) AssembleTimetable(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	timetable, err := server.proton.AssembleTimetable()
	if err != nil {
		WriteJSON(w, Response{Data: "Error while generating timetable", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: timetable, Success: true}, http.StatusOK)
}

func (server *httpImpl) ManualPostProcessRepeat(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}

	var stableTimetable []proton.ProtonMeeting
	err = json.Unmarshal([]byte(r.FormValue("timetable")), &stableTimetable)
	if err != nil {
		WriteJSON(w, Response{Data: stableTimetable, Error: err.Error(), Success: false}, http.StatusBadRequest)
		return
	}

	classes, err := server.db.GetClasses()
	if err != nil {
		WriteJSON(w, Response{Data: stableTimetable, Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	cancelPostProcessingBeforeDone, err := strconv.ParseBool(r.FormValue("cancelPostProcessingBeforeDone"))
	if err != nil {
		WriteJSON(w, Response{Data: stableTimetable, Success: false, Error: err.Error()}, http.StatusBadRequest)
		return
	}

	stableTimetable, err = server.PostProcessTimetable(classes, stableTimetable, cancelPostProcessingBeforeDone)
	if err != nil {
		WriteJSON(w, Response{Data: "Fail while post-processing the timetable", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: stableTimetable, Success: true}, http.StatusOK)
}

// TODO: Implement
func (server *httpImpl) AcceptAssembledTimetable(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	timetableString := r.FormValue("timetable")
	var protonMeetings []proton.ProtonMeeting
	err = json.Unmarshal([]byte(timetableString), &protonMeetings)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while unmarshalling proton meetings", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	//meetings, err := server.proton.AssembleMeetingsFromProtonMeetings(protonMeetings, server.config)
	//if err != nil {
	//	WriteJSON(w, Response{Data: "Failed while assembling meetings from proton meetings", Success: false, Error: err.Error()}, http.StatusInternalServerError)
	//	return
	//}
	//
	//for i := 0; i < len(meetings); i++ {
	//	meeting := meetings[i]
	//	err := server.db.InsertMeeting(meeting)
	//	if err != nil {
	//		WriteJSON(w, Response{Data: "Failed while inserting new meeting", Error: err.Error(), Success: false}, http.StatusInternalServerError)
	//		return
	//	}
	//}
	//
	//WriteJSON(w, Response{Data: meetings, Error: "OK", Success: true}, http.StatusCreated)
}

// TODO: Implement
func (server *httpImpl) DeleteProtonRule(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	//server.proton.DeleteRule(r.FormValue("ruleId"))
	//WriteJSON(w, Response{Data: server.proton.GetProtonConfig(), Success: true}, http.StatusOK)
}
