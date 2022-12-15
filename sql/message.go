package sql

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
)

type Message struct {
	ID              string
	CommunicationID string `db:"communication_id"`
	UserID          string `db:"user_id"`
	Body            string
	Seen            string
	DateCreated     string `db:"date_created"`

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (db *sqlImpl) GetMessage(id string) (message Message, err error) {
	err = db.db.Get(&message, "SELECT * FROM message WHERE id=$1", id)
	return message, err
}

func (db *sqlImpl) GetCommunicationMessages(communicationId string) (messages []Message, err error) {
	err = db.db.Select(&messages, "SELECT * FROM message WHERE communication_id=$1 ORDER BY id ASC", communicationId)
	return messages, err
}

func (db *sqlImpl) InsertMessage(message Message) (err error) {
	_, err = db.db.NamedExec(
		"INSERT INTO message (communication_id, body, seen, date_created, user_id) VALUES (:communication_id, :body, :seen, :date_created, :user_id)",
		message)
	return err
}

func (db *sqlImpl) UpdateMessage(message Message) error {
	_, err := db.db.NamedExec(
		"UPDATE message SET body=:body, seen=:seen WHERE id=:id",
		message)
	return err
}

func (db *sqlImpl) GetAllMessages() (messages []Message, err error) {
	err = db.db.Select(&messages, "SELECT * FROM message ORDER BY id ASC")
	return messages, err
}

func (db *sqlImpl) GetAllUnreadMessages(userId string) (messages []Message, err error) {
	err = db.db.Select(&messages, "SELECT * FROM message ORDER BY id ASC")
	var unread = make([]Message, 0)
	for i := 0; i < len(messages); i++ {
		message := messages[i]
		var users []string
		err := json.Unmarshal([]byte(message.Seen), &users)
		if err != nil {
			return make([]Message, 0), err
		}
		var communicationUsers []string
		communication, err := db.GetCommunication(message.CommunicationID)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(communication.People), &communicationUsers)
		if err != nil {
			return make([]Message, 0), err
		}
		if helpers.Contains(communicationUsers, userId) && !helpers.Contains(users, userId) {
			unread = append(unread, message)
		}
	}
	return unread, err
}

func (db *sqlImpl) DeleteMessage(ID string) error {
	_, err := db.db.Exec("DELETE FROM message WHERE id=$1", ID)
	return err
}
