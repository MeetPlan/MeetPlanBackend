package sql

import "encoding/json"

type Message struct {
	ID              int
	CommunicationID int `db:"communication_id"`
	UserID          int `db:"user_id"`
	Body            string
	Seen            string
	DateCreated     string `db:"date_created"`
}

func (db *sqlImpl) GetMessage(id int) (message Message, err error) {
	err = db.db.Get(&message, "SELECT * FROM message WHERE id=$1", id)
	return message, err
}

func (db *sqlImpl) GetCommunicationMessages(communicationId int) (messages []Message, err error) {
	err = db.db.Select(&messages, "SELECT * FROM message WHERE communication_id=$1", communicationId)
	return messages, err
}

func (db *sqlImpl) InsertMessage(message Message) (err error) {
	_, err = db.db.NamedExec(
		"INSERT INTO message (id, communication_id, body, seen, date_created, user_id) VALUES (:id, :communication_id, :body, :seen, :date_created, :user_id)",
		message)
	return err
}

func (db *sqlImpl) UpdateMessage(message Message) error {
	_, err := db.db.NamedExec(
		"UPDATE message SET body=:body, seen=:seen WHERE id=:id",
		message)
	return err
}

func (db *sqlImpl) GetLastMessageID() (id int) {
	err := db.db.Get(&id, "SELECT id FROM message WHERE id = (SELECT MAX(id) FROM message)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) GetAllMessages() (messages []Message, err error) {
	err = db.db.Select(&messages, "SELECT * FROM message")
	return messages, err
}

func (db *sqlImpl) GetAllUnreadMessages(userId int) (messages []Message, err error) {
	err = db.db.Select(&messages, "SELECT * FROM message")
	var unread = make([]Message, 0)
	for i := 0; i < len(messages); i++ {
		message := messages[i]
		var users []int
		err := json.Unmarshal([]byte(message.Seen), &users)
		if err != nil {
			return make([]Message, 0), err
		}
		var communicationUsers []int
		communication, err := db.GetCommunication(message.CommunicationID)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(communication.People), &communicationUsers)
		if err != nil {
			return make([]Message, 0), err
		}
		if contains(communicationUsers, userId) && !contains(users, userId) {
			unread = append(unread, message)
		}
	}
	return unread, err
}

func (db *sqlImpl) DeleteMessage(ID int) error {
	_, err := db.db.Exec("DELETE FROM message WHERE id=$1", ID)
	return err
}
