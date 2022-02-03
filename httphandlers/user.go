package httphandlers

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"net/http"
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
	jwt, err := sql.GetJWTFromUserPass(email)
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

	user := sql.User{ID: server.db.GetLastUserID(), Email: email, Password: password, Role: "student", Name: name}

	err = server.db.InsertUser(user)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to commit new user to database", Success: false}, http.StatusInternalServerError)
		return
	}

	classId := 0
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
	m = append(m, user.ID)

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

	WriteJSON(w, Response{Data: "Success", Success: true}, http.StatusCreated)
}
