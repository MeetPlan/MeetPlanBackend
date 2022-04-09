package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (server *httpImpl) AssignUserToParent(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] != "admin" {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(mux.Vars(r)["parent"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	user, err := server.db.GetUser(userId)
	if err != nil {
		return
	}
	if user.Role != "parent" {
		WriteJSON(w, Response{Data: "User isn't a parent", Success: false}, http.StatusConflict)
		return
	}
	studentId, err := strconv.Atoi(mux.Vars(r)["student"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	student, err := server.db.GetUser(studentId)
	if err != nil {
		return
	}
	if student.Role != "student" {
		WriteJSON(w, Response{Data: "User isn't a student", Success: false}, http.StatusConflict)
		return
	}
	var users []int
	err = json.Unmarshal([]byte(user.Users), &users)
	if err != nil {
		return
	}
	if contains(users, studentId) {
		return
	}
	users = append(users, studentId)
	marshal, err := json.Marshal(users)
	if err != nil {
		return
	}
	user.Users = string(marshal)
	err = server.db.UpdateUser(user)
	if err != nil {
		return
	}
	WriteJSON(w, Response{
		Data: "OK", Success: true,
	}, http.StatusOK)
}

func (server *httpImpl) GetMyChildren(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "parent" {
		var parentId int
		if jwt["role"] == "admin" {
			parentId, err = strconv.Atoi(r.URL.Query().Get("parentId"))
			if err != nil {
				return
			}
		} else {
			parentId, err = strconv.Atoi(fmt.Sprint(jwt["user_id"]))
			if err != nil {
				return
			}
		}
		parent, err := server.db.GetUser(parentId)
		if err != nil {
			return
		}
		var children []int
		err = json.Unmarshal([]byte(parent.Users), &children)
		if err != nil {
			return
		}
		var childrenJson = make([]UserJSON, 0)
		for i := 0; i < len(children); i++ {
			user, err := server.db.GetUser(children[i])
			if err != nil {
				return
			}
			childrenJson = append(childrenJson, UserJSON{
				Name:     user.Name,
				ID:       user.ID,
				Email:    user.Email,
				Role:     user.Role,
				Birthday: user.Birthday,
			})
		}
		WriteJSON(w, Response{Data: childrenJson, Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}

func (server *httpImpl) RemoveUserFromParent(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] != "admin" {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(mux.Vars(r)["parent"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	user, err := server.db.GetUser(userId)
	if err != nil {
		return
	}
	if user.Role != "parent" {
		WriteJSON(w, Response{Data: "User isn't a parent", Success: false}, http.StatusConflict)
		return
	}
	studentId, err := strconv.Atoi(mux.Vars(r)["student"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	var users []int
	err = json.Unmarshal([]byte(user.Users), &users)
	if err != nil {
		return
	}
	for i := 0; i < len(users); i++ {
		if users[i] == studentId {
			users = remove(users, i)
		}
	}
	marshal, err := json.Marshal(users)
	if err != nil {
		return
	}
	user.Users = string(marshal)
	err = server.db.UpdateUser(user)
	if err != nil {
		return
	}
	WriteJSON(w, Response{
		Data: "OK", Success: true,
	}, http.StatusOK)
}
