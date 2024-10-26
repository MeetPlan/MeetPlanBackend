package httphandlers

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/MeetPlan/MeetPlanBackend/proton"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"math/rand"
	"net/http"
	"strconv"
)

func (server *httpImpl) ManageTeacherAbsences(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	meetingId := mux.Vars(r)["meeting_id"]
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
}

func (server *httpImpl) PostProcessTimetable(classes []sql.Class, stableTimetable []proton.ProtonMeeting, cancelPostProcessingBeforeDone bool) ([]proton.ProtonMeeting, error) {
	// Dogajajo se pripetljaji. Ni vsako polnjenje lukenj popolno, zato gremo "zlikati" ta urnik večkrat.
	for i := 0; i < proton.PROTON_REPEAT_POST_PROCESSING; i++ {
		server.logger.Debugw("izvajam post-procesiranje", "nivo", i)

		// Post-procesiranje urnika za vsak razred posebej.
		for i := 0; i < len(classes); i++ {
			class := classes[i]

			server.logger.Debugw("izvajam post-procesiranje", "class", class)

			var err error
			stableTimetable, err = server.proton.TimetablePostProcessing(stableTimetable, class, cancelPostProcessingBeforeDone)
			if err != nil {
				return nil, err
			}
		}
	}
	return stableTimetable, nil
}

func (server *httpImpl) NewProtonRule(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
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
		// - subjects - Array/[] of numbers/int corresponding to ID(s) of subject(s) in the database
		var subjects []string
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
}

func (server *httpImpl) GetProtonRules(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	protonConfig := server.proton.GetProtonConfig()
	for i := 0; i < len(protonConfig.Rules); i++ {
		if protonConfig.Rules[i].ID == "" {
			UUID, err := uuid.NewUUID()
			if err != nil {
				continue
			}
			protonConfig.Rules[i].ID = UUID.String()
		}
	}
	server.proton.SaveConfig(protonConfig)

	WriteJSON(w, Response{Data: server.proton.GetProtonConfig(), Success: false}, http.StatusOK)
}

func GenerateRandomHourForBeforeAfterSubjects() int {
	return rand.Intn(proton.PROTON_MAX_AFTER_CLASS_HOUR-proton.PROTON_MIN_AFTER_CLASS_HOUR) + proton.PROTON_MIN_AFTER_CLASS_HOUR
}

func GenerateBeforeAfterHour(stackedSubjects []string, subject sql.Subject) int {
	var hour int

	k := rand.Intn(2)
	// naključna izbira med preduro in pouro
	if k == 0 || helpers.Contains(stackedSubjects, subject.ID) {
		// Tako ali tako bomo "zafilali" te luknje
		hour = GenerateRandomHourForBeforeAfterSubjects()
	} else {
		hour = 0
	}

	return hour
}

func (server *httpImpl) AssembleTimetable(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	subjects, err := server.db.GetAllSubjects()
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to retrieve subjects", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	classes, err := server.db.GetClasses()
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to retrieve classes", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	// Before & After class subjects will be treated differently
	beforeAfterSubjects := server.proton.GetSubjectsBeforeOrAfterClass()
	//beforeAfterSubjects := make([]string, 0)
	stackedSubjects := server.proton.GetSubjectsWithStackedHours()

	k := float32(0)
	for i := 0; i < len(subjects); i++ {
		k += subjects[i].SelectedHours
	}

	subjectGroups := server.proton.GetSubjectGroups()

	stableTimetable := make([]proton.ProtonMeeting, 0)

	depth := 0

	failRate := 0
	failResetCount := 0

	for {
		if failResetCount >= proton.PROTON_ALLOWED_FAIL_RESET_RATE {
			WriteJSON(w, Response{Data: "Fail reset rate was exceeded. Aborted.", Success: false}, http.StatusInternalServerError)
			return
		}

		if failRate >= proton.PROTON_ALLOWED_FAIL_RATE {
			var b [8]byte
			_, err = crypto_rand.Read(b[:])
			if err != nil {
				WriteJSON(w, Response{Data: "cannot seed math/rand package with cryptographically secure random number generator", Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))

			server.logger.Debug("fail rate was exceeded. now resetting stableTimetable.")
			depth = 0
			failRate = 0

			stableTimetable = make([]proton.ProtonMeeting, 0)

			//server.logger.Debug(stableTimetable)

			failResetCount++
		}

		if depth >= proton.PROTON_ALLOWED_WHILE_DEPTH {
			WriteJSON(w, Response{Error: "Failed to make a timetable", Data: stableTimetable, Success: false}, http.StatusInternalServerError)
			return
		}

		subjectNum := rand.Intn(len(subjects))
		subject := subjects[subjectNum]

		// Tako dobimo boljšo naključnost
		date := rand.Intn(5)

		var hour int
		// Je izbirni predmet, ampak ni predura
		if helpers.Contains(beforeAfterSubjects, subject.ID) {
			hour = GenerateBeforeAfterHour(stackedSubjects, subject)
		} else {
			hour = rand.Intn(proton.PROTON_MAX_NORMAL_HOUR-proton.PROTON_MIN_NORMAL_HOUR) + proton.PROTON_MIN_NORMAL_HOUR
		}

		t := float32(0)

		// imamo dva tedna, posledično moramo vse deliti z 2
		if float32(len(stableTimetable)/2) >= k {
			break
		}

		for i := 0; i < len(stableTimetable); i++ {
			m := stableTimetable[i]
			if m.SubjectID == subject.ID {
				t++
			}
		}

		if t/2 >= subject.SelectedHours {
			continue
		}

		timetable := make([]proton.ProtonMeeting, 0)
		timetable = append(timetable, stableTimetable...)

		var subjectGroup = make([]string, 0)

		for i := 0; i < len(subjectGroups); i++ {
			group := subjectGroups[i]

			// Check if this group contains OUR SPECIFIED SUBJECT
			var ok = false

			for x := 0; x < len(group.Objects); x++ {
				obj := group.Objects[x]
				if obj.Type == "subject" && obj.ObjectID == subject.ID {
					ok = true
					break
				}
			}

			if ok {
				for n := 0; n < len(group.Objects); n++ {
					object := group.Objects[n]
					if object.Type == "subject" && !helpers.Contains(subjectGroup, object.ObjectID) {
						subjectGroup = append(subjectGroup, object.ObjectID)
					}
				}
			}
		}

		if len(subjectGroup) == 0 {
			subjectGroup = append(subjectGroup, subject.ID)
		}

		// S tem bomo preverili, če so vsi predmeti v skupini predmetov kompatibilni med seboj, tj. imajo isto število ur na teden.
		// V nasprotnem primeru ne moremo ustvariti urnika in javimo "fatal" napako.
		currentSubjectSelectedHours := float32(0)

		var hasSetHour = false

		for i := 0; i < len(subjectGroup); i++ {
			subjectId := subjectGroup[i]
			var currentSubject sql.Subject
			if subject.ID == subjectId {
				currentSubject = subject
			} else {
				currentSubject, err = server.db.GetSubject(subjectId)
				if err != nil {
					server.logger.Error(fmt.Sprintf("failed to retrieve subject %s from the database. skipping.", fmt.Sprint(subjectId)))
					continue
				}
			}

			if currentSubjectSelectedHours == 0 {
				currentSubjectSelectedHours = currentSubject.SelectedHours
			} else if currentSubjectSelectedHours != currentSubject.SelectedHours {
				WriteJSON(w, Response{Data: fmt.Sprintf("Nekompatibilna sestava Proton konfiguracije. Predmet %s je nekompatibilen v številu ur z ostalimi v skupini. Ne morem ustvariti urnika.", fmt.Sprint(subjectId)), Success: false}, http.StatusConflict)
				return
			}
			UUID, err2 := uuid.NewUUID()
			if err2 != nil {
				return
			}

			var generateOnlyOneHour = false

			// Dej, naj mi kdo pove, če je kaka boljša opcija za preverjanje polur.
			if currentSubjectSelectedHours-float32(int(currentSubjectSelectedHours)) == 0.5 {
				// preverimo, če je že vpisane pol ure v naslednjemu tednu
				var hours = 0
				for n := 0; n < len(stableTimetable); n++ {
					meeting := stableTimetable[n]
					if meeting.SubjectID == currentSubject.ID {
						hours++
					}
				}
				if float32(hours/2) == currentSubjectSelectedHours-0.5 {
					if !hasSetHour {
						// V tem primeru nam manjka samo te pol ure, posledično bomo samo dodali tole uro na 2. teden na naključno uro (katero ustvarimo z generatorjem naključnih števil za predure in poure)
						hour = GenerateBeforeAfterHour(stackedSubjects, subject)

						// Preprečimo, da bi se generirali dve različni uri za predmete v isti skupini srečanj
						hasSetHour = true
					}

					// Generiramo samo eno uro v 2. tednu (Week 1)
					generateOnlyOneHour = true
				}
			}

			var classId = make([]string, 0)
			if currentSubject.InheritsClass {
				classId = append(classId, *currentSubject.ClassID)
			} else {
				var students []string
				err := json.Unmarshal([]byte(currentSubject.Students), &students)
				if err != nil {
					return
				}
				for i := 0; i < len(classes); i++ {
					var classStudents []string
					err := json.Unmarshal([]byte(classes[i].Students), &classStudents)
					if err != nil {
						return
					}
					for n := 0; n < len(students); n++ {
						if helpers.Contains(classStudents, students[n]) && !helpers.Contains(classId, classes[i].ID) {
							classId = append(classId, classes[i].ID)
						}
					}
				}
			}

			m := proton.ProtonMeeting{
				ID:           UUID.String(),
				TeacherID:    currentSubject.TeacherID,
				SubjectID:    currentSubject.ID,
				Hour:         hour,
				DayOfTheWeek: date,
				SubjectName:  currentSubject.Name,
				Week:         1,
				ClassID:      classId,
				IsHalfHour:   generateOnlyOneHour,
			}
			timetable = append(timetable, m)

			if generateOnlyOneHour {
				continue
			}

			UUID, err2 = uuid.NewUUID()
			if err2 != nil {
				return
			}

			m = proton.ProtonMeeting{
				ID:           UUID.String(),
				TeacherID:    currentSubject.TeacherID,
				SubjectID:    currentSubject.ID,
				Hour:         hour,
				DayOfTheWeek: date,
				SubjectName:  currentSubject.Name,
				Week:         0,
				ClassID:      classId,
				IsHalfHour:   false,
			}

			timetable = append(timetable, m)

			if server.proton.SubjectHasDoubleHours(subjectId) {
				UUID, err2 := uuid.NewUUID()
				if err2 != nil {
					return
				}

				hour++

				m := proton.ProtonMeeting{
					ID:           UUID.String(),
					TeacherID:    currentSubject.TeacherID,
					SubjectID:    currentSubject.ID,
					Hour:         hour,
					DayOfTheWeek: date,
					SubjectName:  currentSubject.Name,
					Week:         1,
					ClassID:      classId,
					IsHalfHour:   false,
				}
				timetable = append(timetable, m)

				m = proton.ProtonMeeting{
					ID:           UUID.String(),
					TeacherID:    currentSubject.TeacherID,
					SubjectID:    currentSubject.ID,
					Hour:         hour,
					DayOfTheWeek: date,
					SubjectName:  currentSubject.Name,
					Week:         0,
					ClassID:      classId,
					IsHalfHour:   false,
				}
				timetable = append(timetable, m)
			}
		}

		//server.logger.Debug(timetable, stableTimetable)
		ok, err := server.proton.CheckIfProtonConfigIsOk(timetable)
		if ok {
			//server.logger.Debugw("successfully added new meetings", "timetable", timetable)
			stableTimetable = make([]proton.ProtonMeeting, 0)
			stableTimetable = append(stableTimetable, timetable...)

			failRate = 0
		} else {
			server.logger.Debugw("fail while trying to make a timetable using proton", "error", err.Error(), "ttlen", len(stableTimetable), "k", k)

			failRate++
		}

		depth++
	}

	server.logger.Info("done generating timetable. now passing post-processing to the proton package.")

	stableTimetable, err = server.PostProcessTimetable(classes, stableTimetable, false)
	if err != nil {
		WriteJSON(w, Response{Data: "Fail while post-processing the timetable", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: stableTimetable, Success: true}, http.StatusOK)
}

func (server *httpImpl) ManualPostProcessRepeat(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
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

func (server *httpImpl) AcceptAssembledTimetable(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
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

	meetings, err := server.proton.AssembleMeetingsFromProtonMeetings(protonMeetings, server.config)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while assembling meetings from proton meetings", Success: false, Error: err.Error()}, http.StatusInternalServerError)
		return
	}

	for i := 0; i < len(meetings); i++ {
		meeting := meetings[i]
		err := server.db.InsertMeeting(meeting)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed while inserting new meeting", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
	}

	WriteJSON(w, Response{Data: meetings, Error: "OK", Success: true}, http.StatusCreated)
}

func (server *httpImpl) DeleteProtonRule(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	server.proton.DeleteRule(r.FormValue("ruleId"))
	WriteJSON(w, Response{Data: server.proton.GetProtonConfig(), Success: true}, http.StatusOK)
}
