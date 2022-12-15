package sql

type User struct {
	ID                     string
	Email                  string
	Password               string `db:"pass"`
	Role                   string
	Name                   string
	BirthCertificateNumber string `db:"birth_certificate_number"`
	Birthday               string
	CityOfBirth            string `db:"city_of_birth"`
	CountryOfBirth         string `db:"country_of_birth"`
	Users                  string
	LoginToken             string `db:"login_token"`
	IsPassing              bool   `db:"is_passing"`
	IsLocked               bool   `db:"is_locked"`

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (db *sqlImpl) GetUser(id string) (user User, err error) {
	err = db.db.Get(&user, "SELECT * FROM users WHERE id=$1", id)
	return user, err
}

func (db *sqlImpl) GetUserByLoginToken(loginToken string) (user User, err error) {
	err = db.db.Get(&user, "SELECT * FROM users WHERE login_token=$1", loginToken)
	return user, err
}

func (db *sqlImpl) GetTeachers() (user []User, err error) {
	err = db.db.Select(&user, "SELECT * FROM users WHERE role='teacher' ORDER BY id ASC")
	return user, err
}

func (db *sqlImpl) GetPrincipal() (principal User, err error) {
	err = db.db.Get(&principal, "SELECT * FROM users WHERE role='principal' ORDER BY id ASC")
	return principal, err
}

func (db *sqlImpl) GetStudents() (message []User, err error) {
	err = db.db.Select(&message, "SELECT * FROM users WHERE role='student' ORDER BY id ASC")
	return message, err
}

func (db *sqlImpl) GetUserByEmail(email string) (user User, err error) {
	err = db.db.Get(&user, "SELECT * FROM users WHERE email=$1", email)
	return user, err
}

func (db *sqlImpl) InsertUser(user User) (err error) {
	_, err = db.db.NamedExec(
		"INSERT INTO users (email, pass, role, name, birth_certificate_number, city_of_birth, country_of_birth, birthday, users, is_passing, login_token, is_locked) VALUES (:email, :pass, :role, :name, :birth_certificate_number, :city_of_birth, :country_of_birth, :birthday, :users, :is_passing, :login_token, :is_locked)",
		user)
	return err
}

func (db *sqlImpl) CheckIfAdminIsCreated() bool {
	var users []User
	err := db.db.Select(&users, "SELECT * FROM users")
	if err != nil {
		// Return true, as we don't want all the kids, on some internal error to become administrators
		return true
	}
	return len(users) > 0
}

func (db *sqlImpl) GetAllUsers() (users []User, err error) {
	err = db.db.Select(&users, "SELECT * FROM users ORDER BY id ASC")
	return users, err
}

func (db *sqlImpl) UpdateUser(user User) error {
	_, err := db.db.NamedExec(
		"UPDATE users SET pass=:pass, name=:name, role=:role, email=:email, birth_certificate_number=:birth_certificate_number, city_of_birth=:city_of_birth, country_of_birth=:country_of_birth, birthday=:birthday, users=:users, is_passing=:is_passing, login_token=:login_token, is_locked=:is_locked WHERE id=:id",
		user)
	return err
}

func (db *sqlImpl) DeleteUser(ID string) error {
	db.DeleteAllTeacherHomeworks(ID)
	db.DeleteStudentHomeworkByStudentID(ID)
	db.DeleteGradesByTeacherID(ID)
	db.DeleteGradesByUserID(ID)
	db.DeleteUserCommunications(ID)
	db.DeleteAbsencesForUser(ID)
	db.DeleteAbsencesForTeacher(ID)
	db.DeleteTeacherClasses(ID)
	db.DeleteUserClasses(ID)
	db.DeleteMeetingsForTeacher(ID)
	db.DeleteUserSelfTesting(ID)
	db.DeleteTeacherSelfTesting(ID)
	db.DeleteStudentSubject(ID)

	_, err := db.db.Exec("DELETE FROM users WHERE id=$1", ID)
	return err
}
