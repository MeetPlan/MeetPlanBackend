package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type ClassJSON struct {
	Students       []UserJSON
	ID             int
	TeacherID      int
	TeacherName    string
	ClassYear      string
	SOK            int
	EOK            int
	LastSchoolDate int
}

func (server *httpImpl) NewClass(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}

	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {

		className := r.FormValue("name")
		teacherIdStr := fmt.Sprint(r.FormValue("teacher_id"))

		teacherId, err := strconv.Atoi(teacherIdStr)
		if err != nil {
			WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
			return
		}

		class := sql.Class{ID: server.db.GetLastClassID(), Name: className, Teacher: teacherId, ClassYear: r.FormValue("class_year")}
		server.logger.Debug(class)
		err = server.db.InsertClass(class)
		if err != nil {
			WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
			return
		}
		WriteJSON(w, Response{Success: true, Data: class.ID}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) GetClasses(w http.ResponseWriter, r *http.Request) {
	_, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	classes, err := server.db.GetClasses()
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if classes == nil {
		classes = make([]sql.Class, 0)
	}
	WriteJSON(w, Response{Success: true, Data: classes}, http.StatusOK)
}

func (server *httpImpl) PatchClass(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		classId, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		sok, err := strconv.Atoi(r.FormValue("sok"))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		eok, err := strconv.Atoi(r.FormValue("eok"))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		lastDate, err := strconv.Atoi(r.FormValue("last_date"))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		class, err := server.db.GetClass(classId)
		if err != nil {
			return
		}
		class.ClassYear = r.FormValue("class_year")
		class.SOK = sok
		class.EOK = eok
		class.LastSchoolDate = lastDate
		err = server.db.UpdateClass(class)
		if err != nil {
			return
		}
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) GetClass(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "teacher" || jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		classId, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}

		class, err := server.db.GetClass(classId)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}

		var students []int
		err = json.Unmarshal([]byte(class.Students), &students)

		var studentsJson = make([]UserJSON, 0)

		for i := 0; i < len(students); i++ {
			student, err := server.db.GetUser(students[i])
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			sjson := UserJSON{ID: student.ID, Role: student.Role, Email: student.Email, Name: student.Name}
			studentsJson = append(studentsJson, sjson)
		}

		teacher, err := server.db.GetUser(class.Teacher)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}

		c := ClassJSON{
			ID:             class.ID,
			Students:       studentsJson,
			TeacherID:      class.Teacher,
			TeacherName:    teacher.Name,
			ClassYear:      class.ClassYear,
			SOK:            class.SOK,
			EOK:            class.EOK,
			LastSchoolDate: class.LastSchoolDate,
		}

		WriteJSON(w, Response{Success: true, Data: c}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}

func (server *httpImpl) AssignUserToClass(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		classId, err := strconv.Atoi(mux.Vars(r)["class_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		userId, err := strconv.Atoi(mux.Vars(r)["user_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		class, err := server.db.GetClass(classId)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		var m []int
		err = json.Unmarshal([]byte(class.Students), &m)
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
		class.Students = string(s)

		err = server.db.UpdateClass(class)
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

func (server *httpImpl) RemoveUserFromClass(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		classId, err := strconv.Atoi(mux.Vars(r)["class_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		userId, err := strconv.Atoi(mux.Vars(r)["user_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		class, err := server.db.GetClass(classId)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		var m []int
		err = json.Unmarshal([]byte(class.Students), &m)
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
		class.Students = string(s)

		err = server.db.UpdateClass(class)
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

func (server *httpImpl) DeleteClass(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		classId, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		err = server.db.DeleteClass(classId)
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
