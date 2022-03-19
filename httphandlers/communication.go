package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

type MessageJson struct {
	sql.Message
	UserName string
}

type CommunicationJson struct {
	sql.Communication
	Messages []MessageJson
}

func (server *httpImpl) GetCommunications(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	communications, err := server.db.GetCommunications()
	if err != nil {
		return
	}

	var communicationsJson = make([]sql.Communication, 0)

	for i := 0; i < len(communications); i++ {
		communication := communications[i]
		var people []int
		err := json.Unmarshal([]byte(communication.People), &people)
		if err != nil {
			return
		}
		if contains(people, userId) {
			communicationsJson = append(communicationsJson, communication)
		}
	}
	WriteJSON(w, Response{Success: true, Data: communicationsJson}, http.StatusOK)
}

func (server *httpImpl) GetCommunication(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	communicationId, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	communication, err := server.db.GetCommunication(communicationId)
	if err != nil {
		return
	}
	var people []int
	err = json.Unmarshal([]byte(communication.People), &people)
	if err != nil {
		return
	}
	if !contains(people, userId) {
		WriteForbiddenJWT(w)
		return
	}
	messages, err := server.db.GetCommunicationMessages(communicationId)
	if err != nil {
		return
	}
	if messages == nil {
		messages = make([]sql.Message, 0)
	}
	var messagesJson = make([]MessageJson, 0)
	for i := 0; i < len(messages); i++ {
		message := messages[i]
		user, err := server.db.GetUser(message.UserID)
		if err != nil {
			return
		}
		messagesJson = append(messagesJson, MessageJson{
			Message:  message,
			UserName: user.Name,
		})
		var users []int
		err = json.Unmarshal([]byte(message.Seen), &users)
		if err != nil {
			return
		}
		if !contains(users, userId) {
			users = append(users, userId)
			marshal, err := json.Marshal(users)
			if err != nil {
				return
			}
			message.Seen = string(marshal)
			// Not a fatal error, move on
			server.db.UpdateMessage(message)
		}

	}
	j := CommunicationJson{
		Communication: communication,
		Messages:      messagesJson,
	}
	WriteJSON(w, Response{Success: true, Data: j}, http.StatusOK)
}

func (server *httpImpl) NewMessage(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	communicationId, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	communication, err := server.db.GetCommunication(communicationId)
	if err != nil {
		return
	}
	var people []int
	err = json.Unmarshal([]byte(communication.People), &people)
	if err != nil {
		return
	}
	if !contains(people, userId) {
		WriteForbiddenJWT(w)
		return
	}
	message := sql.Message{
		ID:              server.db.GetLastMessageID(),
		CommunicationID: communicationId,
		UserID:          userId,
		Body:            r.FormValue("body"),
		Seen:            fmt.Sprintf("[%s]", fmt.Sprint(userId)),
		DateCreated:     time.Now().String(),
	}
	err = server.db.InsertMessage(message)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true, Data: "OK"}, http.StatusCreated)
}

func (server *httpImpl) NewCommunication(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	var people []int
	err = json.Unmarshal([]byte(r.FormValue("users")), &people)
	if err != nil {
		return
	}
	if !contains(people, userId) {
		people = append(people, userId)
	}
	users, err := json.Marshal(people)
	if err != nil {
		return
	}
	comm := sql.Communication{
		ID:          server.db.GetLastCommunicationID(),
		People:      string(users),
		DateCreated: time.Now().String(),
		Title:       r.FormValue("title"),
	}
	err = server.db.InsertCommunication(comm)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true, Data: "OK"}, http.StatusCreated)
}

func (server *httpImpl) GetUnreadMessages(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	messages, err := server.db.GetAllUnreadMessages(userId)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true, Data: messages}, http.StatusCreated)
}
