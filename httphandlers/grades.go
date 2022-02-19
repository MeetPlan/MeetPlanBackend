package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

type PeriodGrades struct {
	Period  int
	Grades  []sql.Grade
	Total   int
	Average float64
}

type UserGradeTable struct {
	ID      int
	Name    string
	Periods []PeriodGrades
}

type GradeTableResponse struct {
	Users       []UserGradeTable
	TeacherName string
}

func (server *httpImpl) GetGradesForMeeting(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "student" {
		WriteForbiddenJWT(w)
		return
	}
	teacherId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	meetingId, err := strconv.Atoi(mux.Vars(r)["meeting_id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	meeting, err := server.db.GetMeeting(meetingId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	subject, err := server.db.GetSubject(meeting.SubjectID)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if jwt["role"] == "teacher" && subject.TeacherID != teacherId {
		WriteForbiddenJWT(w)
		return
	}
	teacher, err := server.db.GetUser(teacherId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
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
		err = json.Unmarshal([]byte(subject.Students), &users)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
	}
	var usergrades = make([]UserGradeTable, 0)
	for i := 0; i < len(users); i++ {
		var period1 = make([]sql.Grade, 0)
		var period2 = make([]sql.Grade, 0)
		grades, err := server.db.GetGradesForUser(users[i])
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		for n := 0; n < len(grades); n++ {
			grade := grades[n]
			if grade.Period == 1 {
				period1 = append(period1, grade)
			} else if grade.Period == 2 {
				period2 = append(period2, grade)
			}
		}
		user, err := server.db.GetUser(users[i])
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}

		var firstPeriodTotal = 0
		for n := 0; n < len(period1); n++ {
			firstPeriodTotal += period1[n].Grade
		}

		var secondPeriodTotal = 0
		for n := 0; n < len(period2); n++ {
			secondPeriodTotal += period2[n].Grade
		}

		var firstAverage = 0.0
		if len(period1) != 0 {
			firstAverage = float64(firstPeriodTotal) / float64(len(period1))
		}

		var secondAverage = 0.0
		if len(period2) != 0 {
			secondAverage = float64(secondPeriodTotal) / float64(len(period2))
		}

		var periods = make([]PeriodGrades, 0)
		periods = append(periods, PeriodGrades{
			Period:  1,
			Grades:  period1,
			Total:   firstPeriodTotal,
			Average: firstAverage,
		})
		periods = append(periods, PeriodGrades{
			Period:  2,
			Grades:  period2,
			Total:   secondPeriodTotal,
			Average: secondAverage,
		})
		usergrades = append(usergrades, UserGradeTable{
			ID:      user.ID,
			Name:    user.Name,
			Periods: periods,
		})
	}
	WriteJSON(w, Response{
		Success: true,
		Data: GradeTableResponse{
			Users:       usergrades,
			TeacherName: teacher.Name,
		},
	}, http.StatusOK)
}

func (server *httpImpl) NewGrade(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "student" {
		WriteForbiddenJWT(w)
		return
	}

	teacherId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	meetingId, err := strconv.Atoi(mux.Vars(r)["meeting_id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	meeting, err := server.db.GetMeeting(meetingId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	subject, err := server.db.GetSubject(meeting.SubjectID)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if jwt["role"] == "teacher" && subject.TeacherID != teacherId {
		WriteForbiddenJWT(w)
		return
	}

	userId, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	grade, err := strconv.Atoi(r.FormValue("grade"))
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	period, err := strconv.Atoi(r.FormValue("period"))
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	isWritten := r.FormValue("is_written")
	isWrittenBool := false
	if isWritten == "true" {
		isWrittenBool = true
	}

	g := sql.Grade{
		ID:          server.db.GetLastGradeID(),
		UserID:      userId,
		TeacherID:   teacherId,
		SubjectID:   subject.ID,
		Grade:       grade,
		Date:        time.Now().String(),
		IsWritten:   isWrittenBool,
		Period:      period,
		Description: "",
	}

	err = server.db.InsertGrade(g)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusCreated)

}

func (server *httpImpl) PatchGrade(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "student" {
		WriteForbiddenJWT(w)
		return
	}

	teacherId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	gradeId, err := strconv.Atoi(mux.Vars(r)["grade_id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	grade, err := server.db.GetGrade(gradeId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if jwt["role"] == "teacher" && grade.TeacherID != teacherId {
		WriteForbiddenJWT(w)
		return
	}

	ngrade, err := strconv.Atoi(r.FormValue("grade"))
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	period, err := strconv.Atoi(r.FormValue("period"))
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	isWritten := r.FormValue("is_written")
	isWrittenBool := false
	if isWritten == "true" {
		isWrittenBool = true
	}

	grade.Grade = ngrade
	grade.Period = period
	grade.IsWritten = isWrittenBool

	err = server.db.UpdateGrade(grade)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusCreated)

}

func (server *httpImpl) DeleteGrade(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "student" {
		WriteForbiddenJWT(w)
		return
	}

	teacherId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	gradeId, err := strconv.Atoi(mux.Vars(r)["grade_id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	grade, err := server.db.GetGrade(gradeId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if jwt["role"] == "teacher" && grade.TeacherID != teacherId {
		WriteForbiddenJWT(w)
		return
	}

	err = server.db.DeleteGrade(gradeId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusCreated)
}
