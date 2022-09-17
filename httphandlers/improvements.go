package httphandlers

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
)

type Improvement struct {
	sql.Improvement
	TeacherName string
	MeetingName string
}

func (server *httpImpl) NewImprovement(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	studentId := mux.Vars(r)["student_id"]
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	meetingId := mux.Vars(r)["meeting_id"]
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == TEACHER || user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == SCHOOL_PSYCHOLOGIST) {
		WriteForbiddenJWT(w)
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
	if user.Role == TEACHER && !(meeting.TeacherID == user.ID || subject.TeacherID == user.ID) {
		WriteForbiddenJWT(w)
		return
	}
	improvement := sql.Improvement{

		StudentID: studentId,
		MeetingID: meetingId,
		Message:   r.FormValue("message"),
		TeacherID: user.ID,
	}
	server.db.InsertImprovement(improvement)
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}

func (server *httpImpl) GetImprovementsForUser(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	var studentId string
	if user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == PARENT || user.Role == TEACHER || user.Role == SCHOOL_PSYCHOLOGIST {
		studentId = r.URL.Query().Get("studentId")
		if err != nil {
			WriteForbiddenJWT(w)
			return
		}
		if user.Role == PARENT {
			var students []string
			err := json.Unmarshal([]byte(user.Users), &students)
			if err != nil {
				return
			}
			if !helpers.Contains(students, studentId) {
				WriteForbiddenJWT(w)
				return
			}
		} else if user.Role == TEACHER {
			classes, err := server.db.GetClasses()
			if err != nil {
				return
			}
			var ok = false
			for i := 0; i < len(classes); i++ {
				class := classes[i]
				if class.Teacher != user.ID {
					continue
				}
				var students []string
				err := json.Unmarshal([]byte(class.Students), &students)
				if err != nil {
					return
				}
				if !helpers.Contains(students, studentId) {
					continue
				}
				ok = true
			}
			if !ok {
				WriteForbiddenJWT(w)
				return
			}
		}
	} else if user.Role == STUDENT {
		studentId = user.ID
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
