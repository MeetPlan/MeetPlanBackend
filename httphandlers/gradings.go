package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"net/http"
	"strconv"
	"time"
)

type GradingDate struct {
	Date     string
	Gradings []Meeting
}

func insertGradingDate(a []GradingDate, index int, value GradingDate) []GradingDate {
	if len(a) == index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a
}

func (server *httpImpl) GetMyGradings(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "student" || jwt["role"] == "teacher" || jwt["role"] == "parent" || jwt["role"] == "admin" || jwt["role"] == "principal assistant" || jwt["role"] == "principal" || jwt["role"] == "school psychologist" {
		var studentId int
		var teacherId int
		if jwt["role"] == "teacher" || jwt["role"] == "parent" || jwt["role"] == "admin" || jwt["role"] == "principal assistant" || jwt["role"] == "principal" || jwt["role"] == "school psychologist" {
			if jwt["role"] == "parent" {
				if !server.config.ParentViewGradings {
					WriteForbiddenJWT(w)
					return
				}
			}
			studentId, err = strconv.Atoi(r.URL.Query().Get("studentId"))
			if err != nil {
				WriteBadRequest(w)
				return
			}
			teacherId, err = strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		} else {
			studentId, err = strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		}
		if err != nil {
			WriteBadRequest(w)
			return
		}
		if jwt["role"] == "teacher" {
			classes, err := server.db.GetClasses()
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve classes for teacher", Success: false}, http.StatusInternalServerError)
				return
			}
			var valid = false
			for i := 0; i < len(classes); i++ {
				class := classes[i]
				var users []int
				err := json.Unmarshal([]byte(class.Students), &users)
				if err != nil {
					WriteJSON(w, Response{Error: err.Error(), Data: "Failed to unmarshal class users", Success: false}, http.StatusInternalServerError)
					return
				}
				for j := 0; j < len(users); j++ {
					if users[j] == studentId && class.Teacher == teacherId {
						valid = true
						break
					}
				}
				if valid {
					break
				}
			}
			if !valid {
				WriteForbiddenJWT(w)
				return
			}
		} else if jwt["role"] == "parent" {
			parent, err := server.db.GetUser(teacherId)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve parent", Success: false}, http.StatusInternalServerError)
				return
			}
			var children []int
			err = json.Unmarshal([]byte(parent.Users), &children)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Data: "Failed to unmarshal parent's students", Success: false}, http.StatusInternalServerError)
				return
			}
			if !contains(children, studentId) {
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
					dates = insertGradingDate(dates, n, GradingDate{
						Date:     gradings[i].Date,
						Gradings: []Meeting{gradings[i]},
					})
					added = true
				}
			}
			if !added {
				dates = insertGradingDate(dates, 0, GradingDate{
					Date:     gradings[i].Date,
					Gradings: []Meeting{gradings[i]},
				})
			}
		}
		for i, j := 0, len(dates)-1; i < j; i, j = i+1, j-1 {
			dates[i], dates[j] = dates[j], dates[i]
		}
		WriteJSON(w, Response{Data: dates, Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}
