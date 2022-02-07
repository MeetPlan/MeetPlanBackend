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

func (server *httpImpl) GetTimetable(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	classId, err := strconv.Atoi(r.URL.Query().Get("classId"))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	if jwt["role"] == "student" {
		uid, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteForbiddenJWT(w)
			return
		}
		classes, err := server.db.GetClasses()
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		for i := 0; i < len(classes); i++ {
			class := classes[i]
			if class.ID == classId {
				var users []int
				err := json.Unmarshal([]byte(class.Students), &users)
				if err != nil {
					WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
					return
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
		dateMeetingsJson := make([][]sql.Meeting, 0)
		for n := 0; n < 9; n++ {
			hour := make([]sql.Meeting, 0)
			for c := 0; c < len(meetings); c++ {
				meeting := meetings[c]
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
	if jwt["role"] == "teacher" || jwt["role"] == "admin" {
		date := r.FormValue("date")
		hour, err := strconv.Atoi(r.FormValue("hour"))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		classId, err := strconv.Atoi(r.FormValue("classId"))
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
			ClassID:             classId,
			Hour:                hour,
			Date:                date,
			IsMandatory:         isMandatory,
			URL:                 url,
			Details:             details,
			IsGrading:           isGrading,
			IsWrittenAssessment: isWrittenAssessment,
			IsTest:              isTest,
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
	if jwt["role"] == "teacher" || jwt["role"] == "admin" {
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
		classId, err := strconv.Atoi(r.FormValue("classId"))
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

		originalmeeting, err := server.db.GetMeeting(id)
		if originalmeeting.TeacherID != teacherId && jwt["role"] == "teacher" {
			WriteForbiddenJWT(w)
			return
		}

		meeting := sql.Meeting{
			ID:                  id,
			MeetingName:         name,
			TeacherID:           teacherId,
			ClassID:             classId,
			Hour:                hour,
			Date:                date,
			IsMandatory:         isMandatory,
			URL:                 url,
			Details:             details,
			IsGrading:           isGrading,
			IsWrittenAssessment: isWrittenAssessment,
			IsTest:              isTest,
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
	if jwt["role"] == "teacher" || jwt["role"] == "admin" {
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
		classes, err := server.db.GetClasses()
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		for i := 0; i < len(classes); i++ {
			class := classes[i]
			if class.ID == meeting.ClassID {
				var users []int
				err := json.Unmarshal([]byte(class.Students), &users)
				if err != nil {
					WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
					return
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
