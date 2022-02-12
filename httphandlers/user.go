package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"net/http"
	"strconv"
)

func (server *httpImpl) Login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	pass := r.FormValue("pass")
	// Check if password is valid
	user, err := server.db.GetUserByEmail(email)
	hashCorrect := sql.CheckHash(pass, user.Password)
	if !hashCorrect {
		WriteJSON(w, Response{Data: "Hashes don't match...", Success: false}, http.StatusForbidden)
		return
	}

	// Extract JWT
	jwt, err := sql.GetJWTFromUserPass(email, user.Role, user.ID)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: jwt, Success: true}, http.StatusOK)
}

func (server *httpImpl) NewUser(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	pass := r.FormValue("pass")
	name := r.FormValue("name")
	if email == "" || pass == "" || name == "" {
		WriteJSON(w, Response{Data: "Bad Request. A parameter isn't provided", Success: false}, http.StatusBadRequest)
		return
	}
	// Check if user is already in DB
	var userCreated = true
	_, err := server.db.GetUserByEmail(email)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			userCreated = false
		} else {
			WriteJSON(w, Response{Error: err.Error(), Data: "Could not retrieve user from database", Success: false}, http.StatusInternalServerError)
			return
		}
	}
	if userCreated == true {
		WriteJSON(w, Response{Data: "User is already in database", Success: false}, http.StatusUnprocessableEntity)
		return
	}

	password, err := sql.HashPassword(pass)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to hash your password", Success: false}, http.StatusInternalServerError)
		return
	}

	var role = "student"

	isAdmin := !server.db.CheckIfAdminIsCreated()
	if isAdmin {
		role = "admin"
	}

	user := sql.User{ID: server.db.GetLastUserID(), Email: email, Password: password, Role: role, Name: name}

	err = server.db.InsertUser(user)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to commit new user to database", Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: "Success", Success: true}, http.StatusCreated)
}

func (server *httpImpl) GetAllClasses(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		return
	}

	var userId int
	if jwt["role"] == "admin" || jwt["role"] == "teacher" {
		uid := r.URL.Query().Get("id")
		if uid == "" {
			userId, err = strconv.Atoi(fmt.Sprint(jwt["user_id"]))
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
		} else {
			userId, err = strconv.Atoi(uid)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
		}
	} else {
		userId, err = strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
	}

	classes, err := server.db.GetClasses()
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	var myclasses = make([]sql.Class, 0)

	for i := 0; i < len(classes); i++ {
		class := classes[i]
		var students []int
		err := json.Unmarshal([]byte(class.Students), &students)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		for n := 0; n < len(students); n++ {
			if students[n] == userId {
				myclasses = append(myclasses, class)
				break
			}
		}
	}
	WriteJSON(w, Response{Data: myclasses, Success: true}, http.StatusOK)
}

func (server *httpImpl) GetStudents(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" {
		students, err := server.db.GetStudents()
		if err != nil {
			return
		}
		WriteJSON(w, Response{Data: students, Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}
