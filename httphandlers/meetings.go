package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Meeting struct {
	sql.Meeting
	TeacherName string
}

type TimetableDate struct {
	Meetings [][]sql.Meeting `json:"meetings"`
	Date     string          `json:"date"`
}

type Absence struct {
	sql.Absence
	TeacherName string
	UserName    string
	MeetingName string
}

func containsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (server *httpImpl) GetTimetable(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	uid, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}

	var users []int
	myMeetings := false

	if r.URL.Query().Get("classId") != "" {
		classId, err := strconv.Atoi(r.URL.Query().Get("classId"))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		class, err := server.db.GetClass(classId)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal([]byte(class.Students), &users)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
	} else if r.URL.Query().Get("subjectId") != "" {
		subjectId, err := strconv.Atoi(r.URL.Query().Get("subjectId"))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		subject, err := server.db.GetSubject(subjectId)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		if subject.InheritsClass {
			class, err := server.db.GetClass(subject.ClassID)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			err = json.Unmarshal([]byte(class.Students), &users)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
		} else {
			err = json.Unmarshal([]byte(subject.Students), &users)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
		}
	} else if r.URL.Query().Get("studentId") != "" {
		if jwt["role"] == "student" {
			WriteForbiddenJWT(w)
			return
		}
		users = make([]int, 0)
		users = append(users, uid)
		myMeetings = true
	} else {
		// my user
		users = make([]int, 0)
		users = append(users, uid)
		myMeetings = true
	}
	if jwt["role"] == "student" {
		var isIn = false
		for n := 0; n < len(users); n++ {
			if users[n] == uid {
				isIn = true
				break
			}
		}
		if !isIn {
			WriteForbiddenJWT(w)
			return
		}
	}
	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")
	var dates = make([]string, 0)
	// Do some fancy logic to get dates
	i := 0
	for {
		ndate, err := time.Parse("02-01-2006", startDate)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		ndate = ndate.AddDate(0, 0, i)
		d2 := strings.Split(ndate.Format("02-01-2006"), " ")[0]
		dates = append(dates, d2)
		if d2 == endDate {
			break
		}
		i++
	}
	server.logger.Debug(dates)
	var meetingsJson = make([]TimetableDate, 0)
	for i := 0; i < len(dates); i++ {
		date := dates[i]
		meetings, err := server.db.GetMeetingsOnSpecificDate(date)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		var m = make([]sql.Meeting, 0)
		for n := 0; n < len(meetings); n++ {
			meeting := meetings[n]
			subject, err := server.db.GetSubject(meeting.SubjectID)
			if err != nil {
				return
			}
			var u []int
			if subject.InheritsClass {
				class, err := server.db.GetClass(subject.ClassID)
				if err != nil {
					return
				}
				err = json.Unmarshal([]byte(class.Students), &u)
				if err != nil {
					return
				}
			} else {
				err = json.Unmarshal([]byte(subject.Students), &u)
				if err != nil {
					return
				}
			}
			var cont = false
			currentUser, err := server.db.GetUser(uid)
			if err != nil {
				return
			}
			var studentsParent []int
			err = json.Unmarshal([]byte(currentUser.Users), &studentsParent)
			if err != nil {
				return
			}

			server.logger.Debug(studentsParent, u, users)

			// Check if at least one user belongs to class
			for x := 0; x < len(u); x++ {
				if (myMeetings && (jwt["role"] == "teacher" || jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant")) || contains(users, u[x]) {
					if jwt["role"] == "parent" {
						if !contains(studentsParent, u[x]) {
							continue
						}
					}
					cont = true
					break
				}
			}
			if cont {
				if jwt["role"] == "student" {
					if contains(u, uid) {
						m = append(m, meeting)
					}
				} else if jwt["role"] == "admin" || jwt["role"] == "principal" || (jwt["role"] == "teacher" && meeting.TeacherID == uid) || jwt["role"] == "parent" {
					m = append(m, meeting)
				}
			}
		}
		dateMeetingsJson := make([][]sql.Meeting, 0)
		for n := 0; n < 9; n++ {
			hour := make([]sql.Meeting, 0)
			for c := 0; c < len(m); c++ {
				meeting := m[c]
				if meeting.Hour == n {
					hour = append(hour, meeting)
				}
			}
			dateMeetingsJson = append(dateMeetingsJson, hour)
		}
		ttdate := TimetableDate{Date: date, Meetings: dateMeetingsJson}
		meetingsJson = append(meetingsJson, ttdate)
	}
	WriteJSON(w, Response{Data: meetingsJson, Success: true}, http.StatusOK)
}

func (server *httpImpl) NewMeeting(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "teacher" || jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		date := r.FormValue("date")
		hour, err := strconv.Atoi(r.FormValue("hour"))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		subjectId, err := strconv.Atoi(r.FormValue("subjectId"))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		teacherId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		name := r.FormValue("name")

		isMandatoryString := r.FormValue("is_mandatory")
		var isMandatory = true
		if isMandatoryString == "false" {
			isMandatory = false
		}

		url := r.FormValue("url")
		details := r.FormValue("details")

		isGradingString := r.FormValue("is_grading")
		var isGrading = false
		if isGradingString == "true" {
			isGrading = true
		}

		isWrittenAssessmentString := r.FormValue("is_written_assessment")
		var isWrittenAssessment = false
		if isWrittenAssessmentString == "true" {
			isWrittenAssessment = true
		}

		isTestString := r.FormValue("is_test")
		var isTest = false
		if isTestString == "true" {
			isTest = true
		}

		meeting := sql.Meeting{
			ID:                  server.db.GetLastMeetingID(),
			MeetingName:         name,
			TeacherID:           teacherId,
			SubjectID:           subjectId,
			Hour:                hour,
			Date:                date,
			IsMandatory:         isMandatory,
			URL:                 url,
			Details:             details,
			IsGrading:           isGrading,
			IsWrittenAssessment: isWrittenAssessment,
			IsTest:              isTest,
			IsSubstitution:      false,
		}

		err = server.db.InsertMeeting(meeting)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}

func (server *httpImpl) PatchMeeting(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "teacher" || jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		date := r.FormValue("date")
		hour, err := strconv.Atoi(r.FormValue("hour"))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		subjectId, err := strconv.Atoi(r.FormValue("subjectId"))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		teacherId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		name := r.FormValue("name")

		isMandatoryString := r.FormValue("is_mandatory")
		var isMandatory = true
		if isMandatoryString == "false" {
			isMandatory = false
		}

		url := r.FormValue("url")
		details := r.FormValue("details")

		isGradingString := r.FormValue("is_grading")
		var isGrading = false
		if isGradingString == "true" {
			isGrading = true
		}

		isWrittenAssessmentString := r.FormValue("is_written_assessment")
		var isWrittenAssessment = false
		if isWrittenAssessmentString == "true" {
			isWrittenAssessment = true
		}

		isSubstitutionString := r.FormValue("is_substitution")
		var isSubstitution = false
		if (jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant") && isSubstitutionString == "true" {
			isSubstitution = true
			teacherId, err = strconv.Atoi(r.FormValue("teacherId"))
			if err != nil {
				WriteBadRequest(w)
				return
			}
		}

		isTestString := r.FormValue("is_test")
		var isTest = false
		if isTestString == "true" {
			isTest = true
		}

		originalmeeting, err := server.db.GetMeeting(id)
		if originalmeeting.TeacherID != teacherId && jwt["role"] == "teacher" {
			WriteForbiddenJWT(w)
			return
		}

		meeting := sql.Meeting{
			ID:                  id,
			MeetingName:         name,
			TeacherID:           teacherId,
			SubjectID:           subjectId,
			Hour:                hour,
			Date:                date,
			IsMandatory:         isMandatory,
			URL:                 url,
			Details:             details,
			IsGrading:           isGrading,
			IsWrittenAssessment: isWrittenAssessment,
			IsTest:              isTest,
			IsSubstitution:      isSubstitution,
		}

		err = server.db.UpdateMeeting(meeting)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}

func (server *httpImpl) DeleteMeeting(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "teacher" || jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}

		teacherId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteBadRequest(w)
			return
		}

		originalmeeting, err := server.db.GetMeeting(id)
		if originalmeeting.TeacherID != teacherId && jwt["role"] == "teacher" {
			WriteForbiddenJWT(w)
			return
		}

		err = server.db.DeleteMeeting(id)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}

func (server *httpImpl) GetMeeting(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	meetingId, err := strconv.Atoi(mux.Vars(r)["meeting_id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	meeting, err := server.db.GetMeeting(meetingId)
	if err != nil {
		return
	}
	if jwt["role"] == "student" {
		uid, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteForbiddenJWT(w)
			return
		}
		subjects, err := server.db.GetAllSubjects()
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		for i := 0; i < len(subjects); i++ {
			subject := subjects[i]
			if subject.ID == meeting.SubjectID {
				var users []int
				if subject.InheritsClass {
					class, err := server.db.GetClass(subject.ClassID)
					if err != nil {
						WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
						return
					}
					err = json.Unmarshal([]byte(class.Students), &users)
					if err != nil {
						WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
						return
					}
				} else {
					err := json.Unmarshal([]byte(subject.Students), &users)
					if err != nil {
						WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
						return
					}
				}
				var isIn = false
				for n := 0; n < len(users); n++ {
					if users[n] == uid {
						isIn = true
						break
					}
				}
				if !isIn {
					WriteForbiddenJWT(w)
					return
				}
			}
		}
	}
	teacher, err := server.db.GetUser(meeting.TeacherID)
	if err != nil {
		return
	}
	m := Meeting{meeting, teacher.Name}
	WriteJSON(w, Response{Data: m, Success: true}, http.StatusOK)
}

func (server *httpImpl) GetAbsencesTeacher(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "teacher" || jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		meetingId, err := strconv.Atoi(mux.Vars(r)["meeting_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		meeting, err := server.db.GetMeeting(meetingId)
		if err != nil {
			return
		}
		teacherId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			return
		}
		if jwt["role"] == "teacher" && meeting.TeacherID != teacherId {
			WriteForbiddenJWT(w)
			return
		}
		teacher, err := server.db.GetUser(teacherId)
		if err != nil {
			return
		}
		subject, err := server.db.GetSubject(meeting.SubjectID)
		if err != nil {
			return
		}
		var users []int
		if subject.InheritsClass {
			class, err := server.db.GetClass(subject.ClassID)
			if err != nil {
				return
			}
			err = json.Unmarshal([]byte(class.Students), &users)
			if err != nil {
				return
			}
		} else {
			err = json.Unmarshal([]byte(subject.Students), &users)
			if err != nil {
				return
			}
		}
		var absences = make([]Absence, 0)
		for i := 0; i < len(users); i++ {
			userId := users[i]
			user, err := server.db.GetUser(userId)
			if err != nil {
				return
			}
			absence, err := server.db.GetAbsenceForUserMeeting(meetingId, userId)
			if err != nil {
				if err.Error() == "sql: no rows in result set" {
					absence := sql.Absence{
						ID:          server.db.GetLastAbsenceID(),
						UserID:      userId,
						TeacherID:   teacherId,
						MeetingID:   meetingId,
						AbsenceType: "UNMANAGED",
					}
					err := server.db.InsertAbsence(absence)
					if err != nil {
						return
					}
					absences = append(absences, Absence{
						Absence:     absence,
						TeacherName: teacher.Name,
						UserName:    user.Name,
					})
				} else {
					return
				}
			} else {
				absences = append(absences, Absence{
					Absence:     absence,
					TeacherName: teacher.Name,
					UserName:    user.Name,
				})
			}
		}
		WriteJSON(w, Response{Success: true, Data: absences}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}

func (server *httpImpl) PatchAbsence(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "teacher" || jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		absenceId, err := strconv.Atoi(mux.Vars(r)["absence_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		absence, err := server.db.GetAbsence(absenceId)
		if err != nil {
			return
		}
		teacherId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			return
		}
		if jwt["role"] == "teacher" && absence.TeacherID != teacherId {
			WriteForbiddenJWT(w)
			return
		}
		absence.TeacherID = teacherId
		absence.AbsenceType = r.FormValue("absence_type")
		err = server.db.UpdateAbsence(absence)
		if err != nil {
			return
		}
		WriteJSON(w, Response{Success: true, Data: "OK"}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}
