package httphandlers

import (
	sql2 "database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/dchest/uniuri"
	"github.com/gorilla/mux"
	"github.com/signintech/gopdf"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Grade struct {
	ID          string
	UserID      string
	TeacherID   string
	TermID      *string
	Counts      bool
	SubjectID   string
	Grade       int
	Date        string
	IsWritten   bool
	IsFinal     bool
	Period      int
	Description string
	CanPatch    bool
	CreatedAt   string
	UpdatedAt   string
}

type PeriodGrades struct {
	Period  int
	Grades  []Grade
	Total   int
	Average float64
}

type UserGradeTable struct {
	ID       string
	Name     string
	Surname  string
	Average  float64
	Final    int
	IsGraded bool
	Periods  []PeriodGrades
}

type SubjectPosition struct {
	X                      float64
	Y                      float64
	Name                   string
	IsThirdLanguage        bool
	IsDynamicallyAllocated bool
}

type SubjectGradesResponse struct {
	Subjects []UserGradeTable
}

type GradeTableResponse struct {
	Users       []UserGradeTable
	TeacherName string
}

func MakeGradeFromSQLGrade(grade *sql.Grade) Grade {
	return Grade{
		ID:          grade.ID,
		UserID:      grade.UserID,
		TeacherID:   grade.TeacherID,
		TermID:      grade.TermID,
		SubjectID:   grade.SubjectID,
		Grade:       grade.Grade,
		Date:        grade.Date,
		IsWritten:   grade.IsWritten,
		IsFinal:     grade.IsFinal,
		Period:      grade.Period,
		Description: grade.Description,
		CanPatch:    grade.CanPatch,
		CreatedAt:   grade.CreatedAt,
		UpdatedAt:   grade.UpdatedAt,
		Counts:      true,
	}
}

func (server *httpImpl) TransformGradesCountable(grades []sql.Grade) []Grade {
	if len(grades) == 0 {
		return make([]Grade, 0)
	}
	if len(grades) == 1 {
		return []Grade{MakeGradeFromSQLGrade(&(grades)[0])}
	}
	terms := make([]sql.GradingTerm, len(grades))
	firstGrade := -1
	for i, v := range grades {
		if v.TermID == nil {
			continue
		}
		term, err := server.db.GetGradingTerm(*v.TermID)
		if err != nil {
			continue
		}
		terms[i] = term
		if term.Term != 1 {
			continue
		}
		firstGrade = i
		continue
	}

	g := make([]Grade, len(grades))
	for i, v := range grades {
		g[i] = MakeGradeFromSQLGrade(&v)
	}

	if firstGrade == -1 {
		return g
	}

	for i, v := range grades {
		if i == firstGrade {
			continue
		}
		if terms[i].GradeAutoselectType == 0 {
			// obdrži obe
			// po defaultu je firstGradeGrade.Counts == true
			// če se to kdaj spremeni, se to spremeni, če imamo dve konfliktni pravili (npr. 0 in 1)
			// takrat pa še sam ne vem, kaj narediti, tako da ja ...
			continue
		}
		if terms[i].GradeAutoselectType == 1 {
			// štej boljšo
			if g[firstGrade].Grade > v.Grade {
				// prva je bila boljša od trenutne, obdrži prvo
				g[firstGrade].Counts = true
				g[i].Counts = false
			} else {
				// zadnja je bila boljša ali enaka prvi, obdrži zadnjo
				g[firstGrade].Counts = false
			}
		}
		if terms[i].GradeAutoselectType == 2 {
			// štej zadnjo
			g[firstGrade].Counts = false
		}
	}

	return g
}

func (server *httpImpl) GetGradesForMeeting(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
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
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	subject, err := server.db.GetSubject(meeting.SubjectID)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if user.Role == TEACHER && subject.TeacherID != user.ID {
		WriteForbiddenJWT(w)
		return
	}
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
		err = json.Unmarshal([]byte(subject.Students), &users)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
	}
	var usergrades = make([]UserGradeTable, 0)
	for i := 0; i < len(users); i++ {
		var period1 = make([]Grade, 0)
		var period2 = make([]Grade, 0)
		var final = 0
		grades, err := server.db.GetGradesForUser(users[i])
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}

		gm := make(map[string][]sql.Grade)
		for _, v := range grades {
			gradingId := ""
			if v.TermID != nil {
				termId := *v.TermID
				term, err := server.db.GetGradingTerm(termId)
				if err != nil {
					WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
					return
				}
				gradingId = term.GradingID
			}
			_, exists := gm[gradingId]
			if !exists {
				gm[gradingId] = make([]sql.Grade, 0)
			}
			gm[gradingId] = append(gm[gradingId], v)
		}

		for gradingId, g := range gm {
			var grades []Grade
			if gradingId == "" {
				grades = make([]Grade, 0)
				for _, v := range g {
					grades = append(grades, MakeGradeFromSQLGrade(&v))
				}
			} else {
				grades = server.TransformGradesCountable(g)
			}
			for n := 0; n < len(grades); n++ {
				grade := grades[n]

				if grade.SubjectID == subject.ID {
					if grade.IsFinal {
						final = grade.Grade
					} else if grade.Period == 1 {
						period1 = append(period1, grade)
					} else if grade.Period == 2 {
						period2 = append(period2, grade)
					}
				}
			}
		}
		user, err := server.db.GetUser(users[i])
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}

		var firstPeriodTotal = 0
		var firstPeriodNumber = 0
		for n := 0; n < len(period1); n++ {
			if !period1[n].Counts {
				continue
			}
			firstPeriodTotal += period1[n].Grade
			firstPeriodNumber++
		}

		var secondPeriodTotal = 0
		var secondPeriodNumber = 0
		for n := 0; n < len(period2); n++ {
			if !period2[n].Counts {
				continue
			}
			secondPeriodTotal += period2[n].Grade
			secondPeriodNumber++
		}

		var firstAverage = 0.0
		if len(period1) != 0 {
			firstAverage = float64(firstPeriodTotal) / float64(firstPeriodNumber)
		}

		var secondAverage = 0.0
		if len(period2) != 0 {
			secondAverage = float64(secondPeriodTotal) / float64(secondPeriodNumber)
		}

		var avg = 0.0
		if !(len(period1) == 0 && len(period2) == 0) {
			avg = float64(secondPeriodTotal+firstPeriodTotal) / float64(firstPeriodNumber+secondPeriodNumber)
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
			ID:       user.ID,
			Name:     user.Name,
			Surname:  user.Surname,
			Periods:  periods,
			IsGraded: subject.IsGraded,
			Average:  avg,
			Final:    final,
		})
	}
	WriteJSON(w, Response{
		Success: true,
		Data: GradeTableResponse{
			Users:       usergrades,
			TeacherName: user.Name,
		},
	}, http.StatusOK)
}

func (server *httpImpl) NewGrade(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
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
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	subject, err := server.db.GetSubject(meeting.SubjectID)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if !subject.IsGraded {
		WriteJSON(w, Response{Data: "Subject isn't graded. Cannot write any grades.", Success: false}, http.StatusConflict)
		return
	}
	if user.Role == TEACHER && subject.TeacherID != user.ID {
		WriteForbiddenJWT(w)
		return
	}

	userId := r.FormValue("user_id")
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	var hasFinal = true
	_, err = server.db.CheckIfFinal(userId, meeting.SubjectID)
	if err != nil {
		if errors.Is(err, sql2.ErrNoRows) {
			hasFinal = false
		} else {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
	}
	if hasFinal {
		WriteForbiddenJWT(w)
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
	isWritten, err := strconv.ParseBool(r.FormValue("is_written"))
	if err != nil {
		WriteBadRequest(w)
		return
	}

	isFinal, err := strconv.ParseBool(r.FormValue("is_final"))
	if err != nil {
		WriteBadRequest(w)
		return
	}

	canPatch, err := strconv.ParseBool(r.FormValue("can_patch"))
	if err != nil {
		WriteBadRequest(w)
		return
	}

	if isFinal {
		grades, err := server.db.GetGradesForUserInSubject(userId, meeting.SubjectID)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		for i := 0; i < len(grades); i++ {
			if grades[i].IsFinal {
				WriteForbiddenJWT(w)
				return
			}
		}
	}

	g := sql.Grade{
		UserID:      userId,
		TeacherID:   user.ID,
		SubjectID:   subject.ID,
		Grade:       grade,
		Date:        time.Now().String(),
		IsWritten:   isWritten,
		Period:      period,
		Description: r.FormValue("description"),
		IsFinal:     isFinal,
		CanPatch:    canPatch,
	}

	err = server.db.InsertGrade(g)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusCreated)
}

func (server *httpImpl) PatchGrade(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	gradeId := mux.Vars(r)["grade_id"]
	if err != nil {
		WriteBadRequest(w)
		return
	}
	grade, err := server.db.GetGrade(gradeId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	var hasFinal = true
	_, err = server.db.CheckIfFinal(grade.UserID, grade.SubjectID)
	if err != nil {
		if errors.Is(err, sql2.ErrNoRows) {
			hasFinal = false
		} else {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
	}
	if hasFinal {
		WriteForbiddenJWT(w)
		return
	}
	if user.Role == TEACHER && grade.TeacherID != user.ID {
		WriteForbiddenJWT(w)
		return
	}
	if grade.IsFinal {
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

	isWritten, err := strconv.ParseBool(r.FormValue("is_written"))
	if err != nil {
		WriteBadRequest(w)
		return
	}

	grade.Description = r.FormValue("description")
	grade.Grade = ngrade
	grade.Period = period
	grade.IsWritten = isWritten
	grade.TeacherID = user.ID

	err = server.db.UpdateGrade(grade)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusCreated)
}

func (server *httpImpl) DeleteGrade(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	gradeId := mux.Vars(r)["grade_id"]
	if err != nil {
		WriteBadRequest(w)
		return
	}
	grade, err := server.db.GetGrade(gradeId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	var hasFinal = true
	_, err = server.db.CheckIfFinal(grade.UserID, grade.SubjectID)
	if err != nil {
		if errors.Is(err, sql2.ErrNoRows) {
			hasFinal = false
		} else {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
	}
	if hasFinal {
		WriteForbiddenJWT(w)
		return
	}
	if user.Role == TEACHER && grade.TeacherID != user.ID {
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

func (server *httpImpl) GetMyGrades(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == STUDENT || user.Role == TEACHER || user.Role == PARENT || user.Role == PRINCIPAL ||
		user.Role == PRINCIPAL_ASSISTANT || user.Role == ADMIN || user.Role == SCHOOL_PSYCHOLOGIST) {
		WriteForbiddenJWT(w)
		return
	}

	var studentId string
	var teacherId string
	if user.Role == TEACHER || user.Role == PARENT || user.Role == PRINCIPAL ||
		user.Role == PRINCIPAL_ASSISTANT || user.Role == ADMIN || user.Role == SCHOOL_PSYCHOLOGIST {
		if user.Role == PARENT {
			if !server.config.ParentViewGrades {
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
	if user.Role == TEACHER {
		classes, err := server.db.GetClasses()
		if err != nil {
			return
		}
		var valid = false
		for i := 0; i < len(classes); i++ {
			class := classes[i]
			var users []string
			err := json.Unmarshal([]byte(class.Students), &users)
			if err != nil {
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
			return
		}
		var children []string
		json.Unmarshal([]byte(parent.Users), &children)
		if !helpers.Contains(children, studentId) {
			WriteForbiddenJWT(w)
			return
		}
	}
	userGrades, err := server.db.GetGradesForUser(studentId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	gm := make(map[string][]sql.Grade)
	for _, v := range userGrades {
		gradingId := ""
		if v.TermID != nil {
			term, err := server.db.GetGradingTerm(*v.TermID)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			gradingId = term.GradingID
		}
		_, exists := gm[gradingId]
		if !exists {
			gm[gradingId] = make([]sql.Grade, 0)
		}
		gm[gradingId] = append(gm[gradingId], v)
	}

	allGrades := make([]Grade, 0)
	for gradingId, grades := range gm {
		if gradingId == "" {
			for _, v := range grades {
				allGrades = append(allGrades, MakeGradeFromSQLGrade(&v))
			}
			continue
		}

		allGrades = append(allGrades, server.TransformGradesCountable(grades)...)
	}

	subjects, err := server.db.GetAllSubjectsForUser(studentId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	var subjectsResponse = make([]UserGradeTable, 0)
	for i := 0; i < len(subjects); i++ {
		subject := subjects[i]
		var periods = make([]PeriodGrades, 0)
		var total = 0
		var gradesCount = 0
		var final = 0
		for n := 1; n <= 2; n++ {
			var gradesPeriod = make([]Grade, 0)
			var iGradeCount = 0
			var iTotal = 0
			for _, grade := range allGrades {
				if grade.SubjectID == subject.ID && grade.IsFinal {
					final = grade.Grade
				} else if grade.SubjectID == subject.ID && grade.Period == n {
					gradesPeriod = append(gradesPeriod, grade)
					if grade.Counts {
						gradesCount++
						total += grade.Grade
						// No, I don't mean you - Apple. i => internal
						iGradeCount++
						iTotal += grade.Grade
					}
				}
			}
			var avg = 0.0
			if iTotal != 0 && iGradeCount != 0 {
				avg = float64(iTotal) / float64(iGradeCount)
			}
			period := PeriodGrades{
				Period:  n,
				Grades:  gradesPeriod,
				Total:   iTotal,
				Average: avg,
			}
			periods = append(periods, period)
		}
		var avg = 0.0
		if total != 0 && gradesCount != 0 {
			avg = float64(total) / float64(gradesCount)
		}
		grades := UserGradeTable{
			ID:       subject.ID,
			Name:     subject.Name,
			Average:  avg,
			Periods:  periods,
			IsGraded: subject.IsGraded,
			Final:    final,
		}
		subjectsResponse = append(subjectsResponse, grades)
	}
	WriteJSON(w, Response{Data: SubjectGradesResponse{
		Subjects: subjectsResponse,
	}, Success: true}, http.StatusOK)
}

func (server *httpImpl) PrintCertificateOfEndingClass(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	const x1 = 210
	const yb = 18
	const x2 = 485
	var subjectsPosition = []SubjectPosition{
		// Pos 1
		{
			X:    x1,
			Y:    yb * 1,
			Name: "slovenščina",
		},
		{
			X:    x2,
			Y:    yb * 1,
			Name: "kemija",
		},

		// Pos 2
		{
			X:    x1,
			Y:    yb * 2,
			Name: "matematika",
		},
		{
			X:    x2,
			Y:    yb * 2,
			Name: "biologija",
		},

		// Pos 3
		{
			X:                      x1,
			Y:                      yb * 3,
			Name:                   "tretji jezik",
			IsThirdLanguage:        true,
			IsDynamicallyAllocated: true,
		},
		{
			X:    x2,
			Y:    yb * 3,
			Name: "naravoslovje",
		},

		// Pos 4
		{
			X:    x1,
			Y:    yb * 4,
			Name: "likovna umetnost",
		},
		{
			X:    x2,
			Y:    yb * 4,
			Name: "naravoslovje in tehnika",
		},

		// Pos 5
		{
			X:    x1,
			Y:    yb * 5,
			Name: "glasbena umetnost",
		},
		{
			X:    x2,
			Y:    yb * 5,
			Name: "tehnika in tehnologija",
		},

		// Pos 6
		{
			X:    x1,
			Y:    yb * 6,
			Name: "družba",
		},
		{
			X:    x2,
			Y:    yb * 6,
			Name: "gospodinjstvo",
		},

		// Pos 7
		{
			X:    x1,
			Y:    yb * 7,
			Name: "geografija",
		},
		{
			X:    x2,
			Y:    yb * 7,
			Name: "šport",
		},

		// Pos 8
		{
			X:    x1,
			Y:    yb * 8,
			Name: "zgodovina",
		},
		{
			X:                      x2,
			Y:                      yb * 8,
			Name:                   "",
			IsDynamicallyAllocated: true,
		},

		// Pos 9
		{
			X:    x1,
			Y:    yb * 9.5,
			Name: "domovinska in državljanska kultura in etika",
		},
		{
			X:                      x2,
			Y:                      yb * 9,
			Name:                   "",
			IsDynamicallyAllocated: true,
		},

		// Pos 10
		{
			X:    x1,
			Y:    yb * 11,
			Name: "spoznavanje okolja",
		},
		{
			X:                      x2,
			Y:                      yb * 10,
			Name:                   "",
			IsDynamicallyAllocated: true,
		},

		// Pos 11
		{
			X:    x1,
			Y:    yb * 12,
			Name: "fizika",
		},
		{
			X:                      x2,
			Y:                      yb * 11,
			Name:                   "",
			IsDynamicallyAllocated: true,
		},

		// Pos 12 (right)
		{
			X:                      x2,
			Y:                      yb * 12,
			Name:                   "",
			IsDynamicallyAllocated: true,
		},
	}

	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}

	studentId := mux.Vars(r)["student_id"]
	if err != nil {
		WriteBadRequest(w)
		return
	}
	classes, err := server.db.GetClasses()
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	var class *sql.Class
	for i := 0; i < len(classes); i++ {
		if user.Role == TEACHER && classes[i].Teacher != user.ID {
			continue
		}
		var users []string
		err := json.Unmarshal([]byte(classes[i].Students), &users)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		if helpers.Contains(users, studentId) {
			class = &classes[i]
		}
	}

	if class == nil {
		WriteJSON(w, Response{Data: "Class is nil", Success: false}, http.StatusInternalServerError)
		return
	}

	if user.Role == TEACHER {
		var valid = false
		for i := 0; i < len(classes); i++ {
			class := classes[i]
			var users []string
			err := json.Unmarshal([]byte(class.Students), &users)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			if helpers.Contains(users, studentId) && class.Teacher == user.ID {
				valid = true
				break
			}
		}
		if !valid {
			WriteForbiddenJWT(w)
			return
		}
	}

	student, err := server.db.GetUser(studentId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	subjects, err := server.db.GetAllSubjectsForUser(studentId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	if server.config.Debug || r.URL.Query().Get("useDocument") == "true" {
		// Import page 1
		tpl1 := pdf.ImportPage("officialdocs/spričevalo.pdf", 1, "/MediaBox")

		// Draw pdf onto page
		pdf.UseImportedTemplate(tpl1, 0, 0, 595, 0)
	}

	err = pdf.AddTTFFont("opensans", "fonts/opensans.ttf")
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	err = pdf.SetFont("opensans", "", 11)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	// School info
	pdf.SetX(50)
	pdf.SetY(132)
	pdf.Cell(nil, server.config.SchoolName)
	pdf.SetX(50)
	pdf.SetY(157)
	pdf.Cell(nil, fmt.Sprintf("%s, %s %s, %s", server.config.SchoolAddress, fmt.Sprint(server.config.SchoolPostCode), server.config.SchoolCity, server.config.SchoolCountry))

	// Student info
	pdf.SetX(50)
	pdf.SetY(270)
	pdf.Cell(nil, student.Name)

	pdf.SetY(300)
	pdf.SetX(50)
	pdf.Cell(nil, student.Birthday)
	pdf.SetX(215)
	pdf.Cell(nil, fmt.Sprintf("%s, %s", student.CityOfBirth, student.CountryOfBirth))

	pdf.SetX(50)
	pdf.SetY(332)
	pdf.Cell(nil, student.BirthCertificateNumber)
	pdf.SetX(215)
	pdf.Cell(nil, class.Name)
	pdf.SetX(430)
	pdf.Cell(nil, class.ClassYear)

	var subjectsAlreadyIn = make([]string, 0)

	for i := 0; i < len(subjectsPosition); i++ {
		var found = -1
		var name = ""
		for n := 0; n < len(subjects); n++ {
			if !subjects[n].IsGraded {
				continue
			}
			if subjectsPosition[i].IsThirdLanguage {
				if subjects[n].LongName == "angleščina" || subjects[n].LongName == "madžarščina" || subjects[n].LongName == "italijanščina" {
					name = subjects[n].LongName
					found = n
					break
				}
			} else {
				if subjectsPosition[i].IsDynamicallyAllocated && !helpers.Contains(subjectsAlreadyIn, subjects[n].LongName) {
					name = subjects[n].LongName
					found = n
					break
				} else if subjectsPosition[i].Name == subjects[n].LongName {
					name = subjectsPosition[i].Name
					found = n
					break
				}
			}
		}
		var grade = "/"
		if found != -1 {
			grades, err := server.db.GetGradesForUserInSubject(studentId, subjects[found].ID)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			final := 0
			for x := 0; x < len(grades); x++ {
				if grades[x].IsFinal {
					final = grades[x].Grade
				}
			}
			if final == 0 {
				grade = "NEOCENJEN"
			} else {
				grade = fmt.Sprint(final)
			}
		} else {
			// Student doesn't have this subject
			grade = "/"
		}
		if grade == "5" {
			grade = fmt.Sprintf("odlično %s", grade)
		} else if grade == "4" {
			grade = fmt.Sprintf("prav dobro %s", grade)
		} else if grade == "3" {
			grade = fmt.Sprintf("dobro %s", grade)
		} else if grade == "2" {
			grade = fmt.Sprintf("zadostno %s", grade)
		} else if grade == "1" {
			grade = fmt.Sprintf("nezadostno %s", grade)
		}
		server.logger.Debug(name, grade)
		pdf.SetY(subjectsPosition[i].Y + 372)
		if subjectsPosition[i].IsDynamicallyAllocated {
			pdf.SetX(subjectsPosition[i].X - 175)
			pdf.Cell(nil, name)
		}
		pdf.SetX(subjectsPosition[i].X - float64(len(grade)/3)*6)
		pdf.Cell(nil, grade)
		subjectsAlreadyIn = append(subjectsAlreadyIn, name)
	}

	pdf.SetLineWidth(2)
	pdf.SetLineType("full")

	const lineY = 640

	if student.IsPassing {
		pdf.Line(190, lineY, 335, lineY)
	} else {
		pdf.Line(340, lineY, 412, lineY)
	}

	pdf.SetX(70)
	pdf.SetY(669)
	pdf.Cell(nil, fmt.Sprint(class.SOK))

	pdf.SetX(150)
	pdf.Cell(nil, fmt.Sprint(class.EOK))

	UUID := uniuri.NewLen(10)

	lastDate := time.UnixMilli(int64(class.LastSchoolDate * 1000))
	year, month, day := lastDate.Date()
	pdf.SetX(50)
	pdf.SetY(725)
	pdf.Cell(nil, fmt.Sprintf("%s.%s.%s", fmt.Sprint(day), fmt.Sprint(int(month)), fmt.Sprint(year)))
	pdf.SetX(390)
	pdf.Cell(nil, fmt.Sprintf("00/%s/%s", fmt.Sprint(year), UUID))

	teacher, err := server.db.GetUser(class.Teacher)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	pdf.SetX(50)
	pdf.SetY(770)
	pdf.Cell(nil, teacher.Name)

	principal, err := server.db.GetPrincipal()
	if err != nil {
		return
	}

	pdf.SetX(390)
	pdf.Cell(nil, principal.Name)

	output := pdf.GetBytesPdf()

	filename := fmt.Sprintf("documents/%s.pdf", UUID)

	err = helpers.Sign(output, filename, "cacerts/key-pair.p12", "")
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while signing", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	document := sql.Document{
		ID:           UUID,
		ExportedBy:   user.ID,
		DocumentType: SPRICEVALO,
		IsSigned:     true,
	}
	err = server.db.InsertDocument(document)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while inserting Document into database", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	file, err := os.ReadFile(filename)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while reading signed document", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	w.Write(file)

}
