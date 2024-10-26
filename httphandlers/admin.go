package httphandlers

import (
	sql2 "database/sql"
	"errors"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/gorilla/mux"
	"net/http"
)

type UserJSON struct {
	Name                   string
	ID                     string
	Email                  string
	Role                   string
	BirthCertificateNumber string
	Birthday               string
	CityOfBirth            string
	CountryOfBirth         string
	IsPassing              bool
	IsLocked               bool
}

const ADMIN = "admin"
const PRINCIPAL = "principal"
const PRINCIPAL_ASSISTANT = "principal assistant"
const SCHOOL_PSYCHOLOGIST = "school psychologist"
const TEACHER = "teacher"
const FOOD_ORGANIZER = "food organizer"
const PARENT = "parent"
const STUDENT = "student"
const UNVERIFIED = "unverified"

var roles = []string{
	ADMIN,
	PRINCIPAL,
	PRINCIPAL_ASSISTANT,
	SCHOOL_PSYCHOLOGIST,
	TEACHER,
	FOOD_ORGANIZER,
	PARENT,
	STUDENT,
	UNVERIFIED,
}

func (server *httpImpl) ChangeRole(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}

	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}

	userId := mux.Vars(r)["id"]
	if err != nil {
		return
	}
	selectedUser, err := server.db.GetUser(userId)
	if err != nil {
		return
	}
	if selectedUser.ID == user.ID {
		WriteJSON(w, Response{Data: "Cannot change role to itself", Success: false}, http.StatusConflict)
		return
	}
	nrole := r.FormValue("role")
	if nrole == "" {
		WriteJSON(w, Response{Data: "Role is empty", Error: nrole, Success: false}, http.StatusBadRequest)
		return
	}

	if !helpers.Contains(roles, nrole) {
		WriteJSON(w, Response{Data: "Invalid role", Success: false}, http.StatusBadRequest)
		return
	}

	if !((user.Role == PRINCIPAL_ASSISTANT && nrole != ADMIN && nrole != PRINCIPAL && nrole != PRINCIPAL_ASSISTANT) ||
		(user.Role == PRINCIPAL && nrole != ADMIN && nrole != PRINCIPAL) ||
		(user.Role == ADMIN && nrole != ADMIN)) {
		WriteForbiddenJWT(w)
		return
	}

	if user.Role == ADMIN && nrole == PRINCIPAL {
		_, err := server.db.GetPrincipal()
		if err != nil {
			if errors.Is(err, sql2.ErrNoRows) {
				selectedUser.Role = nrole
			}
		} else {
			WriteJSON(w, Response{Data: "There already is a principal", Success: false}, http.StatusConflict)
			return
		}
	} else {
		selectedUser.Role = nrole
	}

	err = server.db.UpdateUser(selectedUser)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true}, http.StatusOK)
}

func (server *httpImpl) LockUnlockUser(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}

	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}

	userId := mux.Vars(r)["id"]
	if err != nil {
		return
	}
	selectedUser, err := server.db.GetUser(userId)
	if err != nil {
		return
	}
	if selectedUser.ID == user.ID {
		WriteJSON(w, Response{Data: "Cannot lock yourself", Success: false}, http.StatusConflict)
		return
	}
	if !((user.Role == PRINCIPAL_ASSISTANT && selectedUser.Role != ADMIN && selectedUser.Role != PRINCIPAL && selectedUser.Role != PRINCIPAL_ASSISTANT) ||
		(user.Role == PRINCIPAL && selectedUser.Role != ADMIN && selectedUser.Role != PRINCIPAL) ||
		(user.Role == ADMIN && selectedUser.Role != ADMIN)) {
		WriteForbiddenJWT(w)
		return
	}
	selectedUser.IsLocked = !selectedUser.IsLocked

	err = server.db.UpdateUser(selectedUser)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true}, http.StatusOK)
}

func (server *httpImpl) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		return
	}
	users, err := server.db.GetAllUsers()
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	var usersjson = make([]UserJSON, 0)
	for i := 0; i < len(users); i++ {
		currentUser := users[i]
		// Only teachers (and above) should be able to access students' and parents' data. Students need to access teachers' data for the purpose of communication module.
		if !(user.Role == TEACHER || user.Role == SCHOOL_PSYCHOLOGIST || user.Role == PRINCIPAL_ASSISTANT || user.Role == PRINCIPAL || user.Role == ADMIN || user.Role == FOOD_ORGANIZER) && (currentUser.Role == STUDENT || currentUser.Role == PARENT) {
			continue
		}
		m := UserJSON{ID: currentUser.ID, Email: currentUser.Email, Role: currentUser.Role, Name: currentUser.Name, IsLocked: currentUser.IsLocked}
		usersjson = append(usersjson, m)
	}
	WriteJSON(w, Response{Data: usersjson, Success: true}, http.StatusOK)

}

func (server *httpImpl) GetTeachers(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
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
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}

	userId := mux.Vars(r)["id"]
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
