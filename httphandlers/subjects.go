package httphandlers

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type Subject struct {
	sql.Subject
	TeacherName     string
	User            []UserJSON
	RealizationDone float32
	TeacherID       int
}

func (server *httpImpl) GetSubjects(w http.ResponseWriter, r *http.Request) {
	_, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	subjects, err := server.db.GetAllSubjects()
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if subjects == nil {
		subjects = make([]sql.Subject, 0)
	}
	WriteJSON(w, Response{Success: true, Data: subjects}, http.StatusOK)
}

func (server *httpImpl) NewSubject(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		teacherId, err := strconv.Atoi(r.FormValue("teacher_id"))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		classId := r.FormValue("class_id")
		var inheritsClass = false
		var classIdInt = -1
		if classId != "" {
			inheritsClass = true
			classIdInt, err = strconv.Atoi(classId)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
		}
		realization, err := strconv.ParseFloat(r.FormValue("realization"), 32)
		if err != nil {
			WriteBadRequest(w)
			return
		}
		var students = make([]int, 0)
		studentsJson, err := json.Marshal(students)
		nSubject := sql.Subject{
			ID:            server.db.GetLastSubjectID(),
			TeacherID:     teacherId,
			Name:          r.FormValue("name"),
			LongName:      r.FormValue("long_name"),
			InheritsClass: inheritsClass,
			ClassID:       classIdInt,
			Students:      string(studentsJson),
			Realization:   float32(realization),
			SelectedHours: 1.0,
		}
		err = server.db.InsertSubject(nSubject)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusCreated)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) GetSubject(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" || jwt["role"] == "school psychologist" {
		subjectId, err := strconv.Atoi(mux.Vars(r)["subject_id"])
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		subject, err := server.db.GetSubject(subjectId)
		if err != nil {
			server.logger.Debug(err, subjectId)
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		var students []int
		if subject.InheritsClass {
			class, err := server.db.GetClass(subject.ClassID)
			if err != nil {
				server.logger.Debug(err, subject)
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			err = json.Unmarshal([]byte(class.Students), &students)
			if err != nil {
				server.logger.Debug(err, subject)
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
		} else {
			err = json.Unmarshal([]byte(subject.Students), &students)
			if err != nil {
				server.logger.Debug(err, subject)
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
		}

		var studentsJson = make([]UserJSON, 0)

		for i := 0; i < len(students); i++ {
			student, err := server.db.GetUser(students[i])
			if err != nil {
				server.logger.Debug(err, subject)
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			studentsJson = append(studentsJson, UserJSON{
				Name:  student.Name,
				ID:    student.ID,
				Email: student.Email,
				Role:  student.Role,
			})
		}

		teacher, err := server.db.GetUser(subject.TeacherID)
		if err != nil {
			server.logger.Debug(err, subject)
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}

		WriteJSON(w, Response{
			Data: Subject{
				Subject:     subject,
				TeacherName: teacher.Name,
				User:        studentsJson,
			},
			Success: true,
		}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) AssignUserToSubject(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		subjectId, err := strconv.Atoi(mux.Vars(r)["subject_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		userId, err := strconv.Atoi(mux.Vars(r)["user_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		subject, err := server.db.GetSubject(subjectId)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		var m []int
		err = json.Unmarshal([]byte(subject.Students), &m)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		for i := 0; i < len(m); i++ {
			if m[i] == userId {
				WriteJSON(w, Response{Data: "User is already in this class", Success: false}, http.StatusConflict)
				return
			}
		}
		m = append(m, userId)

		s, err := json.Marshal(m)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		subject.Students = string(s)

		err = server.db.UpdateSubject(subject)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) RemoveUserFromSubject(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		subjectId, err := strconv.Atoi(mux.Vars(r)["subject_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		userId, err := strconv.Atoi(mux.Vars(r)["user_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		subject, err := server.db.GetSubject(subjectId)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		var m []int
		err = json.Unmarshal([]byte(subject.Students), &m)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		for i := 0; i < len(m); i++ {
			if m[i] == userId {
				m = remove(m, i)
				break
			}
		}

		s, err := json.Marshal(m)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		subject.Students = string(s)

		err = server.db.UpdateSubject(subject)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) DeleteSubject(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		subjectId, err := strconv.Atoi(mux.Vars(r)["subject_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}

		subject, err := server.db.GetSubject(subjectId)
		if err != nil {
			return
		}
		err = server.db.DeleteSubject(subject)
		if err != nil {
			return
		}
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) PatchSubjectName(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		subjectId, err := strconv.Atoi(mux.Vars(r)["subject_id"])
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to parse subjectId", Error: err.Error(), Success: false}, http.StatusBadRequest)
			return
		}
		subject, err := server.db.GetSubject(subjectId)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to retrieve subject", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		selectedHours, err := strconv.ParseFloat(r.FormValue("selected_hours"), 32)
		if err != nil {
			WriteBadRequest(w)
			return
		}
		realization, err := strconv.ParseFloat(r.FormValue("realization"), 32)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to parse realization", Error: err.Error(), Success: false}, http.StatusBadRequest)
			return
		}
		subject.LongName = r.FormValue("long_name")
		subject.Realization = float32(realization)
		subject.SelectedHours = float32(selectedHours)
		err = server.db.UpdateSubject(subject)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to update subject", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}
