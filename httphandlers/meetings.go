package httphandlers

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
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
	SubjectName string
	Subject     Subject
}

type TimetableDate struct {
	Meetings [][]Meeting `json:"meetings"`
	Date     string      `json:"date"`
}

type Absence struct {
	sql.Absence
	TeacherName string
	UserName    string
	MeetingName string
}

func (server *httpImpl) GetTimetable(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}

	var users []string
	myMeetings := false

	if r.URL.Query().Get("classId") != "" {
		classId := r.URL.Query().Get("classId")
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
		subjectId := r.URL.Query().Get("subjectId")
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
			class, err := server.db.GetClass(*subject.ClassID)
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
		if user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == PARENT || user.Role == SCHOOL_PSYCHOLOGIST {
			users = make([]string, 0)
			// TODO: This doesn't seem right
			users = append(users, user.ID)
			myMeetings = true
		} else {
			WriteForbiddenJWT(w)
			return
		}
	} else if r.URL.Query().Get("teacherId") != "" {
		if user.Role == ADMIN || user.Role == PRINCIPAL_ASSISTANT || user.Role == PRINCIPAL {
			teacherId := r.URL.Query().Get("teacherId")
			if err != nil {
				WriteBadRequest(w)
				return
			}
			users = make([]string, 0)
			users = append(users, teacherId)
			myMeetings = true
		} else {
			WriteForbiddenJWT(w)
			return
		}
	} else {
		// my user
		users = make([]string, 0)
		users = append(users, user.ID)
		myMeetings = true
	}
	if user.Role == STUDENT {
		var isIn = false
		for n := 0; n < len(users); n++ {
			if users[n] == user.ID {
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
		if i > 1000 {
			WriteJSON(w, Response{Data: "Exiting due to exceeded maximum depth", Success: false}, http.StatusInternalServerError)
			return
		}
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

	if len(users) == 0 {
		return
	}
	currentUser, err := server.db.GetUser(users[0])
	if err != nil {
		return
	}

	var meetingsJson = make([]TimetableDate, 0)
	for i := 0; i < len(dates); i++ {
		date := dates[i]
		meetings, err := server.db.GetMeetingsOnSpecificDate(date,
			user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == TEACHER || user.Role == SCHOOL_PSYCHOLOGIST,
		)
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
			var u []string
			if subject.InheritsClass {
				class, err := server.db.GetClass(*subject.ClassID)
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
			currentUser2, err := server.db.GetUser(user.ID)
			if err != nil {
				return
			}
			var studentsParent []string
			err = json.Unmarshal([]byte(currentUser2.Users), &studentsParent)
			if err != nil {
				return
			}

			// Check if at least one user belongs to class
			for x := 0; x < len(u); x++ {
				if (myMeetings && (user.Role == TEACHER || user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == SCHOOL_PSYCHOLOGIST)) || helpers.Contains(users, u[x]) {
					if user.Role == PARENT {
						if !helpers.Contains(studentsParent, u[x]) {
							continue
						}
					}
					cont = true
					break
				}
			}

			if cont {
				if (currentUser.Role == TEACHER || currentUser.Role == SCHOOL_PSYCHOLOGIST) && myMeetings {
					if meeting.TeacherID == currentUser.ID {
						m = append(m, meeting)
					}
				} else {
					if user.Role == STUDENT {
						if helpers.Contains(u, user.ID) {
							m = append(m, meeting)
						}
					} else if user.Role == ADMIN ||
						user.Role == PRINCIPAL ||
						user.Role == PRINCIPAL_ASSISTANT ||
						(!myMeetings && user.Role == TEACHER) ||
						(myMeetings && user.Role == TEACHER && meeting.TeacherID == user.ID) ||
						(!myMeetings && user.Role == SCHOOL_PSYCHOLOGIST) ||
						(myMeetings && user.Role == SCHOOL_PSYCHOLOGIST && meeting.TeacherID == user.ID) ||
						user.Role == PARENT {
						m = append(m, meeting)
					}
				}
			}
		}
		dateMeetingsJson := make([][]Meeting, 0)
		for n := 0; n < 9; n++ {
			hour := make([]Meeting, 0)
			for c := 0; c < len(m); c++ {
				meeting := m[c]
				teacher, err := server.db.GetUser(meeting.TeacherID)
				if err != nil {
					WriteJSON(w, Response{Data: "failed while retrieving the teacher from the database", Error: err.Error(), Success: false}, http.StatusNotFound)
					return
				}
				subject, err := server.db.GetSubject(meeting.SubjectID)
				if err != nil {
					WriteJSON(w, Response{Data: "failed while retrieving the subject from the database", Error: err.Error(), Success: false}, http.StatusNotFound)
					return
				}
				subjectTeacher, err := server.db.GetUser(subject.TeacherID)
				if err != nil {
					WriteJSON(w, Response{Data: "failed while retrieving the teacher from the database", Error: err.Error(), Success: false}, http.StatusNotFound)
					return
				}
				if meeting.Hour == n {
					hour = append(hour, Meeting{
						Meeting:     meeting,
						TeacherName: teacher.Name,
						SubjectName: subject.Name,
						Subject: Subject{
							Subject:         subject,
							TeacherName:     subjectTeacher.Name,
							User:            nil,
							RealizationDone: 0,
							TeacherID:       subject.TeacherID,
						},
					})
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
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == TEACHER || user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == SCHOOL_PSYCHOLOGIST) {
		WriteForbiddenJWT(w)
		return
	}
	dates := make([]string, 0)

	date := r.FormValue("date")
	dates = append(dates, date)

	hour, err := strconv.Atoi(r.FormValue("hour"))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	subjectId := r.FormValue("subjectId")
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

	isCorrectionTest, err := strconv.ParseBool(r.FormValue("is_correction_test"))
	if err != nil {
		WriteBadRequest(w)
		return
	}

	isTestString := r.FormValue("is_test")
	var isTest = false
	if isTestString == "true" {
		isTest = true
	}

	if r.FormValue("last_date") != "" {
		repeatCycle, err := strconv.Atoi(r.FormValue("repeat_cycle"))
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed at converting repeat_cycle to int", Success: false}, http.StatusBadRequest)
			return
		}
		lastDate, err := time.Parse("02-01-2006", r.FormValue("last_date"))
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed at converting last_date to Time", Success: false}, http.StatusBadRequest)
			return
		}
		date, err := time.Parse("02-01-2006", date)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed at converting date to Time", Success: false}, http.StatusBadRequest)
			return
		}
		for {
			m := 24 * 7 * repeatCycle
			date = date.Add(time.Hour * time.Duration(m))
			if date.After(lastDate) {
				break
			}
			dates = append(dates, date.Format("02-01-2006"))
		}
	}

	for i := 0; i < len(dates); i++ {
		date := dates[i]

		meeting := sql.Meeting{
			MeetingName:         name,
			TeacherID:           user.ID,
			SubjectID:           subjectId,
			Hour:                hour,
			Date:                date,
			IsMandatory:         isMandatory,
			URL:                 url,
			Details:             details,
			Location:            r.FormValue("location"),
			IsGrading:           isGrading,
			IsWrittenAssessment: isWrittenAssessment,
			IsTest:              isTest,
			IsCorrectionTest:    isCorrectionTest,
			IsSubstitution:      false,
			IsBeta:              false,
		}

		err = server.db.InsertMeeting(meeting)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}

func (server *httpImpl) PatchMeeting(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == TEACHER || user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == SCHOOL_PSYCHOLOGIST) {
		WriteForbiddenJWT(w)
		return
	}
	id := mux.Vars(r)["id"]
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

	isCorrectionTest, err := strconv.ParseBool(r.FormValue("is_correction_test"))
	if err != nil {
		WriteBadRequest(w)
		return
	}

	originalmeeting, err := server.db.GetMeeting(id)

	subject, err := server.db.GetSubject(originalmeeting.SubjectID)
	if err != nil {
		return
	}

	teacherId := subject.TeacherID
	isSubstitutionString := r.FormValue("is_substitution")
	var isSubstitution = false
	if (user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) && isSubstitutionString == "true" {
		isSubstitution = true
		teacherId = r.FormValue("teacherId")
		if err != nil {
			WriteBadRequest(w)
			return
		}
	}

	if !(subject.TeacherID == teacherId || originalmeeting.TeacherID == teacherId) && (user.Role == TEACHER || user.Role == SCHOOL_PSYCHOLOGIST) {
		WriteForbiddenJWT(w)
		return
	}

	meeting := sql.Meeting{
		ID:                  id,
		MeetingName:         name,
		TeacherID:           teacherId,
		SubjectID:           subject.ID,
		Hour:                hour,
		Date:                date,
		IsMandatory:         isMandatory,
		URL:                 url,
		Details:             details,
		IsGrading:           isGrading,
		IsWrittenAssessment: isWrittenAssessment,
		IsTest:              isTest,
		IsSubstitution:      isSubstitution,
		IsCorrectionTest:    isCorrectionTest,
		IsBeta:              originalmeeting.IsBeta,
		Location:            r.FormValue("location"),
	}

	err = server.db.UpdateMeeting(meeting)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}

func (server *httpImpl) DeleteMeeting(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == TEACHER || user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == SCHOOL_PSYCHOLOGIST) {
		WriteForbiddenJWT(w)
		return
	}
	id := mux.Vars(r)["id"]
	if err != nil {
		WriteBadRequest(w)
		return
	}

	originalmeeting, err := server.db.GetMeeting(id)
	if originalmeeting.TeacherID != user.ID && (user.Role == TEACHER || user.Role == SCHOOL_PSYCHOLOGIST) {
		WriteForbiddenJWT(w)
		return
	}

	err = server.db.DeleteMeeting(id)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}

func (server *httpImpl) GetMeeting(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	meetingId := mux.Vars(r)["meeting_id"]
	if err != nil {
		WriteBadRequest(w)
		return
	}
	meeting, err := server.db.GetMeeting(meetingId)
	if err != nil {
		return
	}
	if user.Role == STUDENT || user.Role == PARENT {
		subjects, err := server.db.GetAllSubjects()
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		for i := 0; i < len(subjects); i++ {
			subject := subjects[i]
			if subject.ID == meeting.SubjectID {
				var users []string
				if subject.InheritsClass {
					class, err := server.db.GetClass(*subject.ClassID)
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
				if user.Role == STUDENT {
					var isIn = false
					for n := 0; n < len(users); n++ {
						if users[n] == user.ID {
							isIn = true
							break
						}
					}
					if !isIn {
						WriteForbiddenJWT(w)
						return
					}
				} else if user.Role == PARENT {
					var students []string

					err := json.Unmarshal([]byte(user.Users), &students)
					if err != nil {
						return
					}

					var ok = false

					for i := 0; i < len(users); i++ {
						if helpers.Contains(students, users[i]) {
							ok = true
							break
						}
					}

					if !ok {
						WriteForbiddenJWT(w)
						return
					}
				} else {
					WriteForbiddenJWT(w)
					return
				}
			}
		}
	} else if !(user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == ADMIN || user.Role == SCHOOL_PSYCHOLOGIST || user.Role == TEACHER) {
		WriteForbiddenJWT(w)
		return
	}
	teacher, err := server.db.GetUser(meeting.TeacherID)
	if err != nil {
		return
	}
	subject, err := server.db.GetSubject(meeting.SubjectID)
	if err != nil {
		return
	}
	meetingsList, err := server.db.GetMeetingsForSubjectWithIDLower(meeting.CreatedAt, subject.ID)
	if err != nil {
		return
	}
	m := Meeting{meeting, teacher.Name, subject.Name, Subject{
		Subject:         subject,
		TeacherName:     teacher.Name,
		User:            nil,
		RealizationDone: float32(len(meetingsList)),
		TeacherID:       subject.TeacherID,
	}}
	WriteJSON(w, Response{Data: m, Success: true}, http.StatusOK)
}

func (server *httpImpl) GetAbsencesTeacher(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == TEACHER || user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == SCHOOL_PSYCHOLOGIST) {
		WriteForbiddenJWT(w)
		return
	}
	meetingId := mux.Vars(r)["meeting_id"]
	if err != nil {
		WriteBadRequest(w)
		return
	}
	meeting, err := server.db.GetMeeting(meetingId)
	if err != nil {
		return
	}
	subject, err := server.db.GetSubject(meeting.SubjectID)
	if err != nil {
		return
	}
	if (user.Role == TEACHER || user.Role == SCHOOL_PSYCHOLOGIST) && !(subject.TeacherID == user.ID || meeting.TeacherID == user.ID) {
		WriteForbiddenJWT(w)
		return
	}
	var users []string
	if subject.InheritsClass {
		class, err := server.db.GetClass(*subject.ClassID)
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
		currentUser, err := server.db.GetUser(userId)
		if err != nil {
			return
		}
		absence, err := server.db.GetAbsenceForUserMeeting(meetingId, userId)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				absence := sql.Absence{

					UserID:      userId,
					TeacherID:   user.ID,
					MeetingID:   meetingId,
					AbsenceType: "UNMANAGED",
				}
				err := server.db.InsertAbsence(absence)
				if err != nil {
					return
				}
				absences = append(absences, Absence{
					Absence:     absence,
					TeacherName: user.Name,
					UserName:    currentUser.Name,
				})
			} else {
				return
			}
		} else {
			absences = append(absences, Absence{
				Absence:     absence,
				TeacherName: user.Name,
				UserName:    currentUser.Name,
			})
		}
	}
	WriteJSON(w, Response{Success: true, Data: absences}, http.StatusOK)
}

func (server *httpImpl) PatchAbsence(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == TEACHER || user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == SCHOOL_PSYCHOLOGIST) {
		WriteForbiddenJWT(w)
		return
	}
	absenceId := mux.Vars(r)["absence_id"]
	if err != nil {
		WriteBadRequest(w)
		return
	}
	absence, err := server.db.GetAbsence(absenceId)
	if err != nil {
		return
	}
	meeting, err := server.db.GetMeeting(absence.MeetingID)
	if err != nil {
		return
	}
	subject, err := server.db.GetSubject(meeting.SubjectID)
	if err != nil {
		return
	}
	if (user.Role == TEACHER || user.Role == SCHOOL_PSYCHOLOGIST) && !(subject.TeacherID == user.ID || meeting.TeacherID == user.ID) {
		WriteForbiddenJWT(w)
		return
	}
	absence.TeacherID = user.ID
	absence.AbsenceType = r.FormValue("absence_type")
	err = server.db.UpdateAbsence(absence)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true, Data: "OK"}, http.StatusOK)
}

func (server *httpImpl) GetUsersForMeeting(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == TEACHER || user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == SCHOOL_PSYCHOLOGIST) {
		WriteForbiddenJWT(w)
		return
	}
	meetingId := mux.Vars(r)["meeting_id"]
	if err != nil {
		WriteBadRequest(w)
		return
	}
	meeting, err := server.db.GetMeeting(meetingId)
	if err != nil {
		return
	}
	subject, err := server.db.GetSubject(meeting.SubjectID)
	if err != nil {
		return
	}
	if (user.Role == TEACHER || user.Role == SCHOOL_PSYCHOLOGIST) && !(subject.TeacherID == user.ID || meeting.TeacherID == user.ID) {
		WriteForbiddenJWT(w)
		return
	}
	var students []string
	if subject.InheritsClass {
		class, err := server.db.GetClass(*subject.ClassID)
		if err != nil {
			return
		}
		err = json.Unmarshal([]byte(class.Students), &students)
		if err != nil {
			return
		}
	} else {
		err = json.Unmarshal([]byte(subject.Students), &students)
		if err != nil {
			return
		}
	}
	users := make([]UserJSON, 0)
	for i := 0; i < len(students); i++ {
		student := students[i]
		studentUser, err := server.db.GetUser(student)
		if err != nil {
			return
		}
		users = append(users, UserJSON{
			Name: studentUser.Name,
			ID:   studentUser.ID,
		})
	}
	WriteJSON(w, Response{Data: users, Success: true}, http.StatusOK)
}

func (server *httpImpl) MigrateBetaMeetings(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	err = server.db.MigrateBetaMeetingsToNonBeta()
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while migrating beta meetings to non-beta meetings", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}

func (server *httpImpl) DeleteBetaMeetings(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	err = server.db.DeleteBetaMeetings()
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while deleting beta meetings", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}
