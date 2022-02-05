package httphandlers

import (
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type UserJSON struct {
	Name  string
	ID    int
	Email string
	Role  string
}

func (server *httpImpl) ChangeRole(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		return
	}
	if jwt["role"] != "admin" {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return
	}
	user, err := server.db.GetUser(userId)
	if err != nil {
		return
	}
	nrole := r.FormValue("role")
	if nrole == "" {
		WriteJSON(w, Response{Data: "Role is empty", Error: nrole, Success: false}, http.StatusBadRequest)
		return
	}
	user.Role = nrole
	err = server.db.UpdateUser(user)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true}, http.StatusOK)
}

func (server *httpImpl) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		return
	}
	if jwt["role"] != "admin" {
		WriteForbiddenJWT(w)
		return
	}
	users, err := server.db.GetAllUsers()
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	var usersjson = make([]UserJSON, 0)
	for i := 0; i < len(users); i++ {
		user := users[i]
		m := UserJSON{ID: user.ID, Email: user.Email, Role: user.Role, Name: user.Name}
		usersjson = append(usersjson, m)
	}
	WriteJSON(w, Response{Data: usersjson, Success: true}, http.StatusOK)

}

func (server *httpImpl) GetTeachers(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] != "admin" {
		WriteForbiddenJWT(w)
		return
	}
	users, err := server.db.GetTeachers()
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	var usersjson = make([]UserJSON, 0)
	for i := 0; i < len(users); i++ {
		user := users[i]
		m := UserJSON{ID: user.ID, Email: user.Email, Role: user.Role, Name: user.Name}
		usersjson = append(usersjson, m)
	}
	WriteJSON(w, Response{Data: usersjson, Success: true}, http.StatusOK)
}

func (server *httpImpl) DeleteUser(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		return
	}
	if jwt["role"] != "admin" {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return
	}
	err = server.db.DeleteUser(userId)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}
