package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
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
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
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
		if helpers.Contains(people, user.ID) {
			communicationsJson = append(communicationsJson, communication)
		}
	}
	WriteJSON(w, Response{Success: true, Data: communicationsJson}, http.StatusOK)
}

func (server *httpImpl) GetCommunication(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
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
	if !helpers.Contains(people, user.ID) {
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
		currentUser, err := server.db.GetUser(message.UserID)
		if err != nil {
			return
		}
		messagesJson = append(messagesJson, MessageJson{
			Message:  message,
			UserName: currentUser.Name,
		})
		var users []int
		err = json.Unmarshal([]byte(message.Seen), &users)
		if err != nil {
			return
		}
		if !helpers.Contains(users, user.ID) {
			users = append(users, user.ID)
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
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
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
	if !helpers.Contains(people, user.ID) {
		WriteForbiddenJWT(w)
		return
	}
	message := sql.Message{
		ID:              server.db.GetLastMessageID(),
		CommunicationID: communicationId,
		UserID:          user.ID,
		Body:            r.FormValue("body"),
		Seen:            fmt.Sprintf("[%s]", fmt.Sprint(user.ID)),
		DateCreated:     time.Now().String(),
	}
	err = server.db.InsertMessage(message)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true, Data: "OK"}, http.StatusCreated)
}

func (server *httpImpl) NewCommunication(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	var people []int
	err = json.Unmarshal([]byte(r.FormValue("users")), &people)
	if err != nil {
		return
	}
	if !helpers.Contains(people, user.ID) {
		people = append(people, user.ID)
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
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	messages, err := server.db.GetAllUnreadMessages(user.ID)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true, Data: messages}, http.StatusCreated)
}

func (server *httpImpl) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	messageId, err := strconv.Atoi(mux.Vars(r)["message_id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	message, err := server.db.GetMessage(messageId)
	if err != nil {
		return
	}
	if message.UserID != user.ID {
		WriteForbiddenJWT(w)
		return
	}
	err = server.db.DeleteMessage(messageId)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}

func (server *httpImpl) EditMessage(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	messageId, err := strconv.Atoi(mux.Vars(r)["message_id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	message, err := server.db.GetMessage(messageId)
	if err != nil {
		return
	}
	if message.UserID != user.ID {
		WriteForbiddenJWT(w)
		return
	}
	message.Body = r.FormValue("body")
	err = server.db.UpdateMessage(message)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}
