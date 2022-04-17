package httphandlers

import (
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type UserJSON struct {
	Name                   string
	ID                     int
	Email                  string
	Role                   string
	BirthCertificateNumber string
	Birthday               string
	CityOfBirth            string
	CountryOfBirth         string
}

func (server *httpImpl) ChangeRole(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}

	currentUserId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		return
	}

	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		userId, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			return
		}
		user, err := server.db.GetUser(userId)
		if err != nil {
			return
		}
		if user.ID == currentUserId {
			WriteJSON(w, Response{Data: "Cannot change role to itself", Success: false}, http.StatusConflict)
			return
		}
		nrole := r.FormValue("role")
		if nrole == "" {
			WriteJSON(w, Response{Data: "Role is empty", Error: nrole, Success: false}, http.StatusBadRequest)
			return
		}

		currentUser, err := server.db.GetUser(currentUserId)
		if err != nil {
			return
		}

		if (currentUser.Role == "principal assistant" && nrole != "admin" && nrole != "principal" && nrole != "principal assistant") ||
			(currentUser.Role == "principal" && nrole != "admin" && nrole != "principal") ||
			(currentUser.Role == "admin" && nrole != "admin") {
			if currentUser.Role == "admin" && nrole == "principal" {
				_, err := server.db.GetPrincipal()
				if err != nil {
					if err.Error() == "sql: no rows in result set" {
						user.Role = nrole
					}
				} else {
					WriteJSON(w, Response{Data: "There already is a principal", Success: false}, http.StatusConflict)
					return
				}
			} else {
				user.Role = nrole
			}
		} else {
			WriteForbiddenJWT(w)
			return
		}
		err = server.db.UpdateUser(user)
		if err != nil {
			return
		}
		WriteJSON(w, Response{Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}

func (server *httpImpl) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	_, err := sql.CheckJWT(GetAuthorizationJWT(r))
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
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
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
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) DeleteUser(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
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
	} else {
		WriteForbiddenJWT(w)
		return
	}
}
