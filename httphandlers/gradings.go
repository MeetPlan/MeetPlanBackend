package httphandlers

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"net/http"
	"time"
)

type GradingDate struct {
	Date     string
	Gradings []Meeting
}

func (server *httpImpl) GetMyGradings(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == STUDENT || user.Role == TEACHER || user.Role == PARENT || user.Role == ADMIN || user.Role == PRINCIPAL_ASSISTANT || user.Role == PRINCIPAL || user.Role == SCHOOL_PSYCHOLOGIST) {
		WriteForbiddenJWT(w)
		return
	}

	var studentId string
	var teacherId string
	if user.Role == TEACHER || user.Role == PARENT || user.Role == ADMIN || user.Role == PRINCIPAL_ASSISTANT || user.Role == PRINCIPAL || user.Role == SCHOOL_PSYCHOLOGIST {
		if user.Role == PARENT {
			if !server.config.ParentViewGradings {
				WriteForbiddenJWT(w)
				return
			}
		}
		studentId = r.URL.Query().Get("studentId")
		if err != nil {
			WriteBadRequest(w)
			return
		}
		teacherId = user.ID
	} else {
		studentId = user.ID
	}
	if err != nil {
		WriteBadRequest(w)
		return
	}
	if user.Role == TEACHER {
		classes, err := server.db.GetClasses()
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve classes for teacher", Success: false}, http.StatusInternalServerError)
			return
		}
		var valid = false
		for i := 0; i < len(classes); i++ {
			class := classes[i]
			var users []string
			err := json.Unmarshal([]byte(class.Students), &users)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Data: "Failed to unmarshal class users", Success: false}, http.StatusInternalServerError)
				return
			}
			if helpers.Contains(users, studentId) && class.Teacher == teacherId {
				valid = true
				break
			}
		}
		if !valid {
			WriteForbiddenJWT(w)
			return
		}
	} else if user.Role == PARENT {
		parent, err := server.db.GetUser(teacherId)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve parent", Success: false}, http.StatusInternalServerError)
			return
		}
		var children []string
		err = json.Unmarshal([]byte(parent.Users), &children)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to unmarshal parent's students", Success: false}, http.StatusInternalServerError)
			return
		}
		if !helpers.Contains(children, studentId) {
			WriteForbiddenJWT(w)
			return
		}
	}
	subjects, err := server.db.GetAllSubjectsForUser(studentId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	var gradings = make([]Meeting, 0)
	for i := 0; i < len(subjects); i++ {
		meetings, err := server.db.GetMeetingsForSubject(subjects[i].ID)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		for n := 0; n < len(meetings); n++ {
			meeting := meetings[n]
			if meeting.IsGrading || meeting.IsTest {
				user, err := server.db.GetUser(meeting.TeacherID)
				if err != nil {
					WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve user", Success: false}, http.StatusInternalServerError)
					return
				}
				gradings = append(gradings, Meeting{
					Meeting:     meeting,
					TeacherName: user.Name,
					SubjectName: subjects[i].Name,
				})
			}
		}
	}
	var dates = make([]GradingDate, 0)
	for i := 0; i < len(gradings); i++ {
		var added = false
		gradingDate, err := time.Parse("02-01-2006", gradings[i].Date)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to parse gradingDate", Success: false}, http.StatusInternalServerError)
			return
		}
		for n := 0; n < len(dates); n++ {
			date := dates[n]
			parsedDate, err := time.Parse("02-01-2006", date.Date)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Data: "Failed to parse parsedDate", Success: false}, http.StatusInternalServerError)
				return
			}
			if gradingDate.Equal(parsedDate) {
				dates[n].Gradings = append(dates[n].Gradings, gradings[i])
				added = true
			} else if gradingDate.After(parsedDate) {
				dates = helpers.Insert(dates, n, GradingDate{
					Date:     gradings[i].Date,
					Gradings: []Meeting{gradings[i]},
				})
				added = true
			}
		}
		if !added {
			dates = helpers.Insert(dates, 0, GradingDate{
				Date:     gradings[i].Date,
				Gradings: []Meeting{gradings[i]},
			})
		}
	}
	for i, j := 0, len(dates)-1; i < j; i, j = i+1, j-1 {
		dates[i], dates[j] = dates[j], dates[i]
	}
	WriteJSON(w, Response{Data: dates, Success: true}, http.StatusOK)
}
