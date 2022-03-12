package sql

type Communication struct {
	ID          int
	People      string
	DateCreated string `db:"date_created"`
	Title       string
}

func (db *sqlImpl) GetCommunication(id int) (communication Communication, err error) {
	err = db.db.Get(&communication, "SELECT * FROM communication WHERE id=$1", id)
	return communication, err
}

func (db *sqlImpl) InsertCommunication(communication Communication) (err error) {
	_, err = db.db.NamedExec(
		"INSERT INTO communication (id, people, title, date_created) VALUES (:id, :people, :title, :date_created)",
		communication)
	return err
}

func (db *sqlImpl) UpdateCommunication(communication Communication) error {
	_, err := db.db.NamedExec(
		"UPDATE communication SET people=:people, title=:title WHERE id=:id",
		communication)
	return err
}

func (db *sqlImpl) GetLastCommunicationID() (id int) {
	err := db.db.Get(&id, "SELECT id FROM communication WHERE id = (SELECT MAX(id) FROM communication)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) GetCommunications() (communication []Communication, err error) {
	err = db.db.Select(&communication, "SELECT * FROM communication")
	return communication, err
}

func (db *sqlImpl) DeleteCommunication(ID int) error {
	_, err := db.db.Exec("DELETE FROM communication WHERE id=$1", ID)
	if err != nil {
		return err
	}
	_, err = db.db.Exec("DELETE FROM message WHERE communication_id=$1", ID)
	return err
}