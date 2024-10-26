package httphandlers

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/gorilla/mux"
	"net/http"
)

func (server *httpImpl) AssignUserToParent(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	userId := mux.Vars(r)["parent"]
	if err != nil {
		WriteBadRequest(w)
		return
	}
	parent, err := server.db.GetUser(userId)
	if err != nil {
		return
	}
	if parent.Role != PARENT {
		WriteJSON(w, Response{Data: "User isn't a parent", Success: false}, http.StatusConflict)
		return
	}
	studentId := mux.Vars(r)["student"]
	if err != nil {
		WriteBadRequest(w)
		return
	}
	student, err := server.db.GetUser(studentId)
	if err != nil {
		return
	}
	if student.Role != STUDENT {
		WriteJSON(w, Response{Data: "User isn't a student", Success: false}, http.StatusConflict)
		return
	}
	var users []string
	err = json.Unmarshal([]byte(parent.Users), &users)
	if err != nil {
		return
	}
	if helpers.Contains(users, studentId) {
		return
	}
	users = append(users, studentId)
	marshal, err := json.Marshal(users)
	if err != nil {
		return
	}
	parent.Users = string(marshal)
	err = server.db.UpdateUser(parent)
	if err != nil {
		return
	}
	WriteJSON(w, Response{
		Data: "OK", Success: true,
	}, http.StatusOK)
}

func (server *httpImpl) GetMyChildren(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == PARENT) {
		WriteForbiddenJWT(w)
		return
	}
	var parentId string
	if user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT {
		parentId = r.URL.Query().Get("parentId")
		if err != nil {
			return
		}
	} else {
		parentId = user.ID
	}
	parent, err := server.db.GetUser(parentId)
	if err != nil {
		return
	}
	var children []string
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
}

func (server *httpImpl) RemoveUserFromParent(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationToken(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	userId := mux.Vars(r)["parent"]
	if err != nil {
		WriteBadRequest(w)
		return
	}
	parent, err := server.db.GetUser(userId)
	if err != nil {
		return
	}
	if user.Role != PARENT {
		WriteJSON(w, Response{Data: "User isn't a parent", Success: false}, http.StatusConflict)
		return
	}
	studentId := mux.Vars(r)["student"]
	if err != nil {
		WriteBadRequest(w)
		return
	}
	var users []string
	err = json.Unmarshal([]byte(parent.Users), &users)
	if err != nil {
		return
	}
	for i := 0; i < len(users); i++ {
		if users[i] == studentId {
			users = helpers.Remove(users, i)
		}
	}
	marshal, err := json.Marshal(users)
	if err != nil {
		return
	}
	parent.Users = string(marshal)
	err = server.db.UpdateUser(parent)
	if err != nil {
		return
	}
	WriteJSON(w, Response{
		Data: "OK", Success: true,
	}, http.StatusOK)
}
