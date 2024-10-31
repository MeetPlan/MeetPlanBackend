package httphandlers

import (
	sql2 "database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

type GradingDate struct {
	Date     string
	Gradings []Meeting
}

type UserGrade struct {
	ID      string
	Name    string
	Surname string
	Grade   sql.Grade
}

type GradingTerm struct {
	GradingTerm sql.GradingTerm
	Users       []UserGrade
}

type Grading struct {
	Grading      sql.Grading
	GradingTerms []GradingTerm
}

func (server *httpImpl) NewGrading(w http.ResponseWriter, r *http.Request) {
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

	name := r.FormValue("name")
	if name == "" {
		WriteBadRequest(w)
		return
	}

	description := r.FormValue("description")

	// 0 = ustno
	// 1 = pisno
	// 2 = drugo
	gradingType, err := strconv.Atoi(r.FormValue("grading_type"))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	if gradingType < 0 || gradingType > 2 {
		WriteBadRequest(w)
		return
	}

	// 0 = celoletno
	// 1 = prvo ocenjevalno obdobje
	// 2 = drugo ocenjevalno obdobje
	period, err := strconv.Atoi(r.FormValue("period"))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	if period < 0 || period > 2 {
		WriteBadRequest(w)
		return
	}

	grading := sql.Grading{
		SubjectID:   meeting.SubjectID,
		TeacherID:   subject.TeacherID,
		Name:        name,
		Description: description,
		GradingType: gradingType,
		SchoolYear:  helpers.GetCurrentSchoolYear(),
		Period:      period,
	}

	err = server.db.InsertGrading(grading)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusCreated)
}

func (server *httpImpl) GetGradingsTeacher(w http.ResponseWriter, r *http.Request) {
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

	gradings, err := server.db.GetGradingsForSubject(subject.ID)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	students := server.db.GetStudentsFromSubject(&subject)

	gradingsJson := make([]Grading, len(gradings))
	for i, v := range gradings {
		gradingTerms, err := server.db.GetGradingTermsForGrading(v.ID)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		if gradingTerms == nil {
			gradingTerms = make([]sql.GradingTerm, 0)
		}

		gradingTermsJson := make([]GradingTerm, len(gradingTerms))

		for h, l := range gradingTerms {
			term := GradingTerm{
				GradingTerm: l,
				Users:       make([]UserGrade, 0),
			}
			for _, student := range students {
				getUser, err := server.db.GetUser(student)
				if err != nil {
					continue
				}
				grade, err := server.db.GetGradeForTermAndUser(l.ID, getUser.ID)
				term.Users = append(term.Users, UserGrade{
					ID:      student,
					Name:    getUser.Name,
					Surname: getUser.Surname,
					Grade:   grade,
				})
			}
			gradingTermsJson[h] = term
		}
		gradingsJson[i] = Grading{
			Grading:      v,
			GradingTerms: gradingTermsJson,
		}
	}

	WriteJSON(w, Response{Data: gradingsJson, Success: true}, http.StatusOK)
}

func (server *httpImpl) NewGradingTerm(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}

	gradingId := mux.Vars(r)["grading_id"]
	grading, err := server.db.GetGrading(gradingId)
	if err != nil {
		WriteJSON(w, Response{Data: "Error whilst fetching a grading", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	subject, err := server.db.GetSubject(grading.SubjectID)
	if err != nil {
		WriteJSON(w, Response{Data: "Error whilst fetching a subject", Error: err.Error(), Success: false}, http.StatusInternalServerError)
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

	date := r.FormValue("date")
	if date == "" {
		WriteBadRequest(w)
		return
	}
	dateParsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		WriteBadRequest(w)
		return
	}
	dateFmt := dateParsed.Format("2006-01-02")

	hour, err := strconv.Atoi(r.FormValue("hour"))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	if hour < 0 || hour > 12 {
		WriteBadRequest(w)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		name = grading.Name
	}

	description := r.FormValue("description")
	description = grading.Description

	// 1 = prvi rok
	// 2 = drugi rok
	term, err := strconv.Atoi(r.FormValue("term"))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	if term < 1 || term > 2 {
		WriteBadRequest(w)
		return
	}

	// 0 = upoštevaj obe oceni
	// 1 = upoštevaj boljšo oceno
	// 2 = upoštevaj zadnjo oceno
	gradeAutoselectType, err := strconv.Atoi(r.FormValue("grade_autoselect_type"))
	if err != nil && term == 2 {
		WriteBadRequest(w)
		return
	}
	if term == 1 {
		gradeAutoselectType = 0
	}
	if gradeAutoselectType < 0 || gradeAutoselectType > 2 {
		WriteBadRequest(w)
		return
	}

	gradingTerm := sql.GradingTerm{
		TeacherID:           subject.TeacherID,
		GradingID:           grading.ID,
		Date:                dateFmt,
		Hour:                hour,
		Name:                name,
		Description:         description,
		Term:                term,
		GradeAutoselectType: gradeAutoselectType,
	}

	err = server.db.InsertGradingTerm(gradingTerm)
	if err != nil {
		WriteJSON(w, Response{Data: "Error whilst inserting a grading", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusCreated)
}

func (server *httpImpl) PatchGrading(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}

	gradingId := mux.Vars(r)["grading_id"]
	grading, err := server.db.GetGrading(gradingId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if user.Role == TEACHER && grading.TeacherID != user.ID {
		WriteForbiddenJWT(w)
		return
	}

	// preveri, da zadeva ni arhivirana
	if grading.SchoolYear != helpers.GetCurrentSchoolYear() {
		WriteForbiddenJWT(w)
		return
	}

	teacherId := r.FormValue("teacher_id")
	if teacherId != "" && (user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		getUser, err := server.db.GetUser(teacherId)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		if getUser.Role != TEACHER && getUser.Role != PRINCIPAL_ASSISTANT {
			WriteBadRequest(w)
			return
		}
		grading.TeacherID = teacherId
	}

	name := r.FormValue("name")
	if name != "" {
		grading.Name = name
	}

	description := r.FormValue("description")
	grading.Description = description

	gradingType, err := strconv.Atoi(r.FormValue("grading_type"))
	if err == nil {
		if gradingType < 0 || gradingType > 3 {
			WriteBadRequest(w)
			return
		}
		grading.GradingType = gradingType
	}

	period, err := strconv.Atoi(r.FormValue("period"))
	if err == nil {
		if period < 0 || period > 3 {
			WriteBadRequest(w)
			return
		}
		grading.Period = period
	}

	err = server.db.UpdateGrading(grading)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}

func (server *httpImpl) DeleteGrading(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}

	gradingId := mux.Vars(r)["grading_id"]
	grading, err := server.db.GetGrading(gradingId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if user.Role == TEACHER && grading.TeacherID != user.ID {
		WriteForbiddenJWT(w)
		return
	}

	// preveri, da zadeva ni arhivirana
	if grading.SchoolYear != helpers.GetCurrentSchoolYear() {
		WriteForbiddenJWT(w)
		return
	}

	gradingTerms, err := server.db.GetGradingTermsForGrading(gradingId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	for _, v := range gradingTerms {
		err = server.db.DeleteGradesByTermID(v.ID)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
	}

	err = server.db.DeleteGradingTermsByGrading(gradingId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	err = server.db.DeleteGrading(gradingId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}

func (server *httpImpl) PatchGradingTerm(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}

	gradingTermId := mux.Vars(r)["grading_term_id"]
	gradingTerm, err := server.db.GetGradingTerm(gradingTermId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if user.Role == TEACHER && gradingTerm.TeacherID != user.ID {
		WriteForbiddenJWT(w)
		return
	}

	grading, err := server.db.GetGrading(gradingTerm.GradingID)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	// preveri, da zadeva ni arhivirana
	if grading.SchoolYear != helpers.GetCurrentSchoolYear() {
		WriteForbiddenJWT(w)
		return
	}

	subject, err := server.db.GetSubject(grading.SubjectID)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	students := server.db.GetStudentsFromSubject(&subject)
	for _, student := range students {
		studentUser, err := server.db.GetUser(student)
		if err != nil {
			continue
		}
		fv := r.FormValue(fmt.Sprintf("%s.grade", studentUser.ID))
		if fv == "null" {
			server.db.DeleteGradeByTermAndUser(gradingTermId, studentUser.ID)
			continue
		}
		grade, err := strconv.Atoi(fv)
		if err != nil {
			continue
		}
		if grade <= 0 || grade > 5 {
			continue
		}
		gradeDb, err := server.db.GetGradeForTermAndUser(gradingTermId, studentUser.ID)
		if errors.Is(err, sql2.ErrNoRows) {
			now := time.Now()

			// če je grading.Period == 0, to pomeni, da je celoletno
			// v takem primeru prilagodimo obdobje na datum vpisa ocene
			period := grading.Period
			if grading.Period == 0 {
				if (now.Month() == time.January && time.Now().Day() <= 15) || (now.Month() >= 9) {
					period = 1
				} else {
					period = 2
				}
			}

			g := sql.Grade{
				UserID:      studentUser.ID,
				TeacherID:   user.ID,
				TermID:      &gradingTermId,
				SubjectID:   subject.ID,
				Grade:       grade,
				Date:        now.Format("2006-01-02"),
				IsWritten:   grading.GradingType == 1,
				IsFinal:     false,
				Period:      period,
				Description: fmt.Sprintf("%s; %s", grading.Name, grading.Description),
				CanPatch:    true,
			}
			server.db.InsertGrade(g)
			continue
		} else if err != nil {
			server.logger.Errorw("error while fetching a grade", "err", err)
			continue
		}
		if !gradeDb.CanPatch {
			continue
		}
		gradeDb.Grade = grade
		server.db.UpdateGrade(gradeDb)
	}

	teacherId := r.FormValue("teacher_id")
	if teacherId != "" && (user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		getUser, err := server.db.GetUser(teacherId)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		if getUser.Role != TEACHER && getUser.Role != PRINCIPAL_ASSISTANT {
			WriteBadRequest(w)
			return
		}
		gradingTerm.TeacherID = teacherId
	}

	hour, err := strconv.Atoi(r.FormValue("hour"))
	if err == nil {
		if hour < 0 || hour > 12 {
			WriteBadRequest(w)
			return
		}
		gradingTerm.Hour = hour
	}

	name := r.FormValue("name")
	if name != "" {
		gradingTerm.Name = name
	}

	description := r.FormValue("description")
	gradingTerm.Description = description

	// 0 = upoštevaj obe oceni
	// 1 = upoštevaj boljšo oceno
	// 2 = upoštevaj zadnjo oceno
	gradeAutoselectType, err := strconv.Atoi(r.FormValue("grade_autoselect_type"))
	if err == nil {
		if gradeAutoselectType < 0 || gradeAutoselectType > 2 {
			WriteBadRequest(w)
			return
		}
		gradingTerm.GradeAutoselectType = gradeAutoselectType
	}

	// 1 = prvi rok
	// 2 = drugi rok
	// 3 = tretji rok
	// 4 = četrti rok
	// 5 = peti rok
	term, err := strconv.Atoi(r.FormValue("term"))
	if err == nil {
		if term < 1 || term > 2 {
			WriteBadRequest(w)
			return
		}
		gradingTerm.Term = term
	}

	date := r.FormValue("date")
	if date != "" {
		dateParsed, err := time.Parse("2006-01-02", date)
		if err != nil {
			WriteBadRequest(w)
			return
		}
		dateFmt := dateParsed.Format("2006-01-02")
		gradingTerm.Date = dateFmt
	}

	err = server.db.UpdateGradingTerm(gradingTerm)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}

func (server *httpImpl) DeleteGradingTerm(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}

	gradingTermId := mux.Vars(r)["grading_term_id"]
	gradingTerm, err := server.db.GetGradingTerm(gradingTermId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if user.Role == TEACHER && gradingTerm.TeacherID != user.ID {
		WriteForbiddenJWT(w)
		return
	}

	grading, err := server.db.GetGrading(gradingTerm.GradingID)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	// preveri, da zadeva ni arhivirana
	if grading.SchoolYear != helpers.GetCurrentSchoolYear() {
		WriteForbiddenJWT(w)
		return
	}

	err = server.db.DeleteGradesByTermID(gradingTermId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	err = server.db.DeleteGradingTerm(gradingTermId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}

func (server *httpImpl) GetGradingTermsForGrading(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}

	gradingId := mux.Vars(r)["grading_id"]
	grading, err := server.db.GetGrading(gradingId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	subject, err := server.db.GetSubject(grading.SubjectID)
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

	gradingTerms, err := server.db.GetGradingTermsForGrading(subject.ID)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: gradingTerms, Success: true}, http.StatusCreated)
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
		teacherId = user.ID
	} else {
		studentId = user.ID
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
