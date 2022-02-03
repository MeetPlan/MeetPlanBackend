package sql

type User struct {
	ID       int
	Email    string
	Password string `db:"pass"`
	Role     string
	Name     string
}

func (db *sqlImpl) GetUser(id int) (message User, err error) {
	err = db.db.Get(&message, "SELECT * FROM users WHERE id=$1", id)
	return message, err
}

func (db *sqlImpl) GetUserByEmail(email string) (message User, err error) {
	err = db.db.Get(&message, "SELECT * FROM users WHERE email=$1", email)
	return message, err
}

func (db *sqlImpl) InsertUser(user User) (err error) {
	_, err = db.db.NamedExec(
		"INSERT INTO users (id, email, pass, role, name) VALUES (:id, :email, :pass, :role, :name)",
		user)
	return err
}

func (db *sqlImpl) GetLastUserID() (id int) {
	err := db.db.Get(&id, "SELECT id FROM users WHERE id = (SELECT MAX(id) FROM users)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}
