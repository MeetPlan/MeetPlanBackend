package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type Improvement struct {
	sql.Improvement
	TeacherName string
	MeetingName string
}

func (server *httpImpl) NewImprovement(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	teacherId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	studentId, err := strconv.Atoi(mux.Vars(r)["student_id"])
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	meetingId, err := strconv.Atoi(mux.Vars(r)["meeting_id"])
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "teacher" || jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		meeting, err := server.db.GetMeeting(meetingId)
		if err != nil {
			return
		}
		subject, err := server.db.GetSubject(meeting.SubjectID)
		if err != nil {
			return
		}
		if jwt["role"] == "teacher" && !(meeting.TeacherID == teacherId || subject.TeacherID == teacherId) {
			WriteForbiddenJWT(w)
			return
		}
		improvement := sql.Improvement{
			ID:        server.db.GetLastImprovementID(),
			StudentID: studentId,
			MeetingID: meetingId,
			Message:   r.FormValue("message"),
			TeacherID: teacherId,
		}
		server.db.InsertImprovement(improvement)
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}

func (server *httpImpl) GetImprovementsForUser(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	var studentId int
	if jwt["role"] == "student" {
		studentId = userId
	} else {
		studentId, err = strconv.Atoi(r.URL.Query().Get("studentId"))
	}
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}

	user, err := server.db.GetUser(userId)
	if err != nil {
		return
	}
	if user.Role == "parent" {
		var students []int
		err := json.Unmarshal([]byte(user.Users), &students)
		if err != nil {
			return
		}
		if !contains(students, studentId) {
			WriteForbiddenJWT(w)
			return
		}
	} else if user.Role == "teacher" {
		classes, err := server.db.GetClasses()
		if err != nil {
			return
		}
		var ok = false
		for i := 0; i < len(classes); i++ {
			class := classes[i]
			if class.Teacher != userId {
				continue
			}
			var students []int
			err := json.Unmarshal([]byte(class.Students), &students)
			if err != nil {
				return
			}
			if !contains(students, studentId) {
				continue
			}
			ok = true
		}
		if !ok {
			WriteForbiddenJWT(w)
			return
		}
	} else if jwt["role"] == "principal assistant" || jwt["role"] == "principal" || jwt["role"] == "admin" || jwt["role"] == "student" || jwt["role"] == "school psychologist" {
	} else {
		WriteForbiddenJWT(w)
		return
	}

	improvements, err := server.db.GetImprovementsForStudent(studentId)
	if err != nil {
		return
	}

	impJson := make([]Improvement, 0)

	for i := 0; i < len(improvements); i++ {
		improvement := improvements[i]
		user, err := server.db.GetUser(improvement.TeacherID)
		if err != nil {
			WriteJSON(w, Response{Data: "Could not find teacher", Success: false, Error: err.Error()}, http.StatusInternalServerError)
			return
		}
		meeting, err := server.db.GetMeeting(improvement.MeetingID)
		if err != nil {
			WriteJSON(w, Response{Data: "Could not find meeting", Success: false, Error: err.Error()}, http.StatusInternalServerError)
			return
		}
		impJson = append(impJson, Improvement{
			Improvement: improvement,
			TeacherName: user.Name,
			MeetingName: meeting.MeetingName,
		})
	}
	WriteJSON(w, Response{Data: impJson, Success: true}, http.StatusOK)
}
