package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/dchest/uniuri"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
	"github.com/signintech/gopdf"
	"net/http"
	"os"
	"strconv"
	"time"
)

func (server *httpImpl) Login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	pass := r.FormValue("pass")
	// Check if password is valid
	user, err := server.db.GetUserByEmail(email)

	if user.Role == "unverified" {
		WriteJSON(w, Response{Data: "You are unverified. You cannot login until the school administrator confirms you.", Success: false}, http.StatusForbidden)
		return
	}

	hashCorrect := sql.CheckHash(pass, user.Password)
	if !hashCorrect {
		WriteJSON(w, Response{Data: "Hashes don't match...", Success: false}, http.StatusForbidden)
		return
	}

	// Extract JWT
	jwt, err := sql.GetJWTFromUserPass(email, user.Role, user.ID)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: jwt, Success: true}, http.StatusOK)
}

func (server *httpImpl) NewUser(w http.ResponseWriter, r *http.Request) {
	if server.config.BlockRegistrations {
		j := GetAuthorizationJWT(r)
		if j == "" {
			WriteForbiddenJWT(w)
			return
		}
		jwt, err := sql.CheckJWT(j)
		if err != nil {
			WriteForbiddenJWT(w)
			return
		}
		if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		} else {
			WriteForbiddenJWT(w)
			return
		}
	}
	email := r.FormValue("email")
	pass := r.FormValue("pass")
	name := r.FormValue("name")
	if email == "" || pass == "" || name == "" {
		WriteJSON(w, Response{Data: "Bad Request. A parameter isn't provided", Success: false}, http.StatusBadRequest)
		return
	}
	// Check if user is already in DB
	var userCreated = true
	_, err := server.db.GetUserByEmail(email)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			userCreated = false
		} else {
			WriteJSON(w, Response{Error: err.Error(), Data: "Could not retrieve user from database", Success: false}, http.StatusInternalServerError)
			return
		}
	}
	if userCreated == true {
		WriteJSON(w, Response{Data: "User is already in database", Success: false}, http.StatusUnprocessableEntity)
		return
	}

	password, err := sql.HashPassword(pass)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to hash your password", Success: false}, http.StatusInternalServerError)
		return
	}

	var role = "unverified"

	isAdmin := !server.db.CheckIfAdminIsCreated()
	if isAdmin {
		role = "admin"
	}

	user := sql.User{
		ID:                     server.db.GetLastUserID(),
		Email:                  email,
		Password:               password,
		Role:                   role,
		Name:                   name,
		BirthCertificateNumber: "",
		Birthday:               "",
		CityOfBirth:            "",
		CountryOfBirth:         "",
		Users:                  "[]",
	}

	err = server.db.InsertUser(user)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to commit new user to database", Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: "Success", Success: true}, http.StatusCreated)
}

func (server *httpImpl) PatchUser(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		userId, err := strconv.Atoi(mux.Vars(r)["user_id"])
		if err != nil {
			WriteForbiddenJWT(w)
			return
		}

		user, err := server.db.GetUser(userId)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve used from database", Success: false}, http.StatusInternalServerError)
			return
		}
		if r.FormValue("birthday") != "" {
			user.Birthday = r.FormValue("birthday")
		}
		if r.FormValue("country_of_birth") != "" {
			user.CountryOfBirth = r.FormValue("country_of_birth")
		}
		if r.FormValue("city_of_birth") != "" {
			user.CityOfBirth = r.FormValue("city_of_birth")
		}
		if r.FormValue("email") != "" {
			user.Email = r.FormValue("email")
		}
		if r.FormValue("birth_certificate_number") != "" {
			user.BirthCertificateNumber = r.FormValue("birth_certificate_number")
		}
		if r.FormValue("name") != "" {
			user.Name = r.FormValue("name")
		}
		if r.FormValue("is_passing") != "" {
			isPassing, err := strconv.ParseBool(r.FormValue("is_passing"))
			if err != nil {
				WriteBadRequest(w)
				return
			}
			user.IsPassing = isPassing
		}
		err = server.db.UpdateUser(user)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to update user", Success: false}, http.StatusInternalServerError)
			return
		}
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) HasClass(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "teacher" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteForbiddenJWT(w)
			return
		}
		classes, err := server.db.GetClasses()
		if err != nil {
			return
		}
		var hasClass = false
		for i := 0; i < len(classes); i++ {
			if classes[i].Teacher == userId {
				hasClass = true
				break
			}
		}
		WriteJSON(w, Response{Data: hasClass, Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) GetUserData(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		return
	}
	var userId int
	if jwt["role"] == "student" {
		userId, err = strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteBadRequest(w)
			return
		}
	} else {
		userId, err = strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
	}
	user, err := server.db.GetUser(userId)
	if err != nil {
		return
	}

	var birthCertNum = ""
	if jwt["role"] == "admin" {
		birthCertNum = user.BirthCertificateNumber
	}

	ujson := UserJSON{
		Name:                   user.Name,
		ID:                     user.ID,
		Email:                  user.Email,
		Role:                   user.Role,
		BirthCertificateNumber: birthCertNum,
		Birthday:               user.Birthday,
		CityOfBirth:            user.CityOfBirth,
		CountryOfBirth:         user.CountryOfBirth,
		IsPassing:              user.IsPassing,
	}
	WriteJSON(w, Response{Data: ujson, Success: true}, http.StatusOK)
}

func (server *httpImpl) GetAbsencesUser(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	var studentId int
	if jwt["role"] == "student" {
		studentId, err = strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteBadRequest(w)
			return
		}
	} else {
		studentId, err = strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		teacherId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		if jwt["role"] == "teacher" {
			classes, err := server.db.GetClasses()
			if err != nil {
				WriteJSON(w, Response{Data: "Could not fetch classes", Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			var valid = false
			for i := 0; i < len(classes); i++ {
				class := classes[i]
				var users []int
				err := json.Unmarshal([]byte(class.Students), &users)
				if err != nil {
					WriteJSON(w, Response{Data: "Could not unmarshal students", Error: err.Error(), Success: false}, http.StatusInternalServerError)
					return
				}
				for j := 0; j < len(users); j++ {
					if users[j] == studentId && class.Teacher == teacherId {
						valid = true
						break
					}
				}
				if valid {
					break
				}
			}
			if !valid {
				WriteForbiddenJWT(w)
				return
			}
		} else if jwt["role"] == "parent" {
			if !server.config.ParentViewAbsences {
				WriteForbiddenJWT(w)
				return
			}
			parent, err := server.db.GetUser(teacherId)
			if err != nil {
				WriteJSON(w, Response{Data: "Could not fetch parent", Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			var students []int
			err = json.Unmarshal([]byte(parent.Users), &students)
			if err != nil {
				WriteJSON(w, Response{Data: "Could not unmarshal students", Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			if !helpers.Contains(students, studentId) {
				WriteForbiddenJWT(w)
				return
			}
		}
	}
	var absenceJson = make([]Absence, 0)
	absences, err := server.db.GetAbsencesForUser(studentId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			WriteJSON(w, Response{Data: absenceJson, Error: err.Error(), Success: true}, http.StatusOK)
			return
		}
		WriteJSON(w, Response{Data: "Could not fetch absences", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	for i := 0; i < len(absences); i++ {
		absence := absences[i]
		teacher, err := server.db.GetUser(absence.TeacherID)
		if err != nil {
			WriteJSON(w, Response{Data: "Could not fetch teacher", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		user, err := server.db.GetUser(absence.UserID)
		if err != nil {
			WriteJSON(w, Response{Data: "Could not fetch user", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		meeting, err := server.db.GetMeeting(absence.MeetingID)
		if err != nil {
			WriteJSON(w, Response{Data: "Could not fetch meeting", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		if absence.AbsenceType == "ABSENT" || absence.AbsenceType == "LATE" {
			absenceJson = append(absenceJson, Absence{
				Absence:     absence,
				TeacherName: teacher.Name,
				UserName:    user.Name,
				MeetingName: meeting.MeetingName,
			})
		}
	}
	WriteJSON(w, Response{Data: absenceJson, Success: true}, http.StatusOK)
}

func (server *httpImpl) GetAllClasses(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		return
	}

	var userId = make([]int, 0)
	var isTeacher = false
	if jwt["role"] == "admin" || jwt["role"] == "teacher" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		uid := r.URL.Query().Get("id")
		if uid == "" {
			u, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			userId = append(userId, u)
			isTeacher = true
		} else {
			u, err := strconv.Atoi(uid)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			userId = append(userId, u)
		}
	} else if jwt["role"] == "parent" {
		u, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		user, err := server.db.GetUser(u)
		if err != nil {
			return
		}
		err = json.Unmarshal([]byte(user.Users), &userId)
		if err != nil {
			return
		}
	} else {
		u, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		userId = append(userId, u)
	}

	classes, err := server.db.GetClasses()
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	var myclasses = make([]sql.Class, 0)
	var myClassesInt = make([]int, 0)

	for i := 0; i < len(classes); i++ {
		class := classes[i]
		if isTeacher {
			for n := 0; n < len(userId); n++ {
				if class.Teacher == userId[n] {
					myclasses = append(myclasses, class)
				}
			}
		} else {
			var students []int
			err := json.Unmarshal([]byte(class.Students), &students)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			server.logger.Debug(students, userId)
			for n := 0; n < len(students); n++ {
				for l := 0; l < len(userId); l++ {
					if students[n] == userId[l] && !helpers.Contains(myClassesInt, students[n]) {
						user, err := server.db.GetUser(students[n])
						if err != nil {
							return
						}
						var className = class.Name
						if jwt["role"] == "parent" {
							class.Name = fmt.Sprintf("%s - %s", class.Name, user.Name)
						}
						myclasses = append(myclasses, class)
						myClassesInt = append(myClassesInt, students[n])
						if jwt["role"] == "parent" {
							class.Name = className
						}
						break
					}
				}
			}
		}
	}
	WriteJSON(w, Response{Data: myclasses, Success: true}, http.StatusOK)
}

func (server *httpImpl) GetStudents(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" || jwt["role"] == "school psychologist" {
		students, err := server.db.GetStudents()
		if err != nil {
			return
		}
		var studentsJson = make([]UserJSON, 0)
		for i := 0; i < len(students); i++ {
			student := students[i]
			studentsJson = append(studentsJson, UserJSON{
				Name:                   student.Name,
				ID:                     student.ID,
				Email:                  student.Email,
				Role:                   student.Role,
				BirthCertificateNumber: student.BirthCertificateNumber,
				Birthday:               student.Birthday,
				CityOfBirth:            student.CityOfBirth,
				CountryOfBirth:         student.CountryOfBirth,
			})
		}
		WriteJSON(w, Response{Data: studentsJson, Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}

func (server *httpImpl) HasBirthday(w http.ResponseWriter, r *http.Request) {
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
	user, err := server.db.GetUser(userId)
	if err != nil {
		return
	}
	currentTime := time.Now()
	birthday, err := time.Parse("2006-01-02", user.Birthday)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to parse date", Success: false}, http.StatusInternalServerError)
		return
	}
	if currentTime.Before(birthday) {
		WriteJSON(w, Response{Data: "Invalid birthday", Success: false}, http.StatusConflict)
		return
	}
	_, tm, td := currentTime.Date()
	_, bm, bd := birthday.Date()
	WriteJSON(w, Response{Data: tm-bm == 0 && td-bd == 0, Success: true}, http.StatusOK)
}

func (server *httpImpl) CertificateOfSchooling(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}

	if !(jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" || jwt["role"] == "school psychologist") {
		WriteForbiddenJWT(w)
		return
	}

	userId, err := strconv.Atoi(mux.Vars(r)["user_id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	teacherId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteBadRequest(w)
		return
	}

	student, err := server.db.GetUser(userId)
	if err != nil {
		return
	}
	if student.Role != "student" {
		WriteForbiddenJWT(w)
		return
	}

	classes, err := server.db.GetClasses()
	if err != nil {
		return
	}

	var classId = -1

	for i := 0; i < len(classes); i++ {
		class := classes[i]
		var students []int
		err := json.Unmarshal([]byte(class.Students), &students)
		if err != nil {
			return
		}
		if helpers.Contains(students, userId) {
			classId = class.ID
			break
		}
	}

	if classId == -1 {
		return
	}

	class, err := server.db.GetClass(classId)
	if err != nil {
		return
	}

	var students []int
	err = json.Unmarshal([]byte(class.Students), &students)
	if err != nil {
		return
	}
	if !helpers.Contains(students, userId) {
		WriteForbiddenJWT(w)
		return
	}

	m := pdf.NewMaroto(consts.Portrait, consts.A4)

	m.AddUTF8Font("OpenSans", consts.Normal, "fonts/opensans.ttf")
	m.SetDefaultFontFamily("OpenSans")

	m.Row(40, func() {

		m.Col(3, func() {
			_ = m.FileImage("icons/school_logo.png", props.Rect{
				Center:  true,
				Percent: 80,
			})
		})

		m.ColSpace(1)

		m.Col(4, func() {
			m.Text("Potrdilo o šolanju", props.Text{
				Top:         12,
				Size:        25,
				Extrapolate: true,
			})
			m.Text("MeetPlan sistem", props.Text{
				Top:         23,
				Size:        13,
				Extrapolate: true,
			})
		})
		m.ColSpace(1)

		m.Col(3, func() {
			_ = m.FileImage("icons/country_coat_of_arms_black.png", props.Rect{
				Center:  true,
				Percent: 80,
			})
		})
	})

	m.Line(10)

	m.Row(40, func() {
		m.Text(fmt.Sprintf(
			"Učenec %s, rojen %s, %s, %s, v šolskem letu %s",
			student.Name, student.Birthday, student.CityOfBirth,
			student.CountryOfBirth, class.ClassYear,
		), props.Text{
			Top:         12,
			Size:        11,
			Extrapolate: true,
		})
		m.Text(fmt.Sprintf("obiskuje %s razred šole %s.",
			class.Name, server.config.SchoolName,
		), props.Text{
			Top:         16,
			Size:        11,
			Extrapolate: true,
		})
	})

	principal, err := server.db.GetPrincipal()
	if err != nil {
		return
	}

	m.Row(40, func() {
		m.ColSpace(1)
		m.Col(6, func() {
			m.Text("_________________________", props.Text{
				Top:  14,
				Size: 15,
			})
			m.Text(principal.Name, props.Text{
				Top:  14,
				Size: 15,
			})
			m.Text("digitalni podpis ravnatelja", props.Text{Top: 20, Size: 9})
		})
		m.Col(3, func() {
			m.Text("_________________________", props.Text{
				Top:  14,
				Size: 15,
			})
			m.Text("podpis ravnatelja", props.Text{Top: 20, Size: 9})
		})
	})

	m.Line(10)

	UUID := uuid.New().String()

	m.Row(40, func() {
		m.Text(fmt.Sprintf("Enolični identifikator dokumenta: %s", UUID), props.Text{
			Top:  14,
			Size: 10,
		})
	})

	output, err := m.Output()
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("documents/%s.pdf", UUID)

	err = helpers.Sign(output.Bytes(), filename, "cacerts/key-pair.p12", "")
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while signing", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	currentTime := time.Now().UnixMilli()

	document := sql.Document{
		ID:           UUID,
		ExportedBy:   teacherId,
		DocumentType: POTRDILO_O_SOLANJU,
		Timestamp:    int(currentTime),
		IsSigned:     true,
	}

	err = server.db.InsertDocument(document)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while inserting document into the database", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	file, err := os.ReadFile(filename)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while reading signed document", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	w.Write(file)
}

// TODO: Sign this document
func (server *httpImpl) GenerateNewUserCert(pdf *gopdf.GoPdf, userId int) (*gopdf.GoPdf, error) {
	user, err := server.db.GetUser(userId)
	if err != nil {
		return pdf, err
	}

	pdf.AddPage()
	rect := gopdf.Rect{H: 120, W: 120}

	err = pdf.Image("icons/meetplan.png", 50, 50, &rect)
	if err != nil {
		return pdf, err
	}

	newPassword := uniuri.NewLen(10)
	password, err := sql.HashPassword(newPassword)
	if err != nil {
		return pdf, err
	}

	user.Password = password

	const borderBase = 30

	pdf.SetX(250)
	pdf.SetY(120)
	pdf.SetFontSize(30)
	pdf.Text("MeetPlan")
	pdf.SetFontSize(18)
	pdf.SetX(250)
	pdf.SetY(140)
	pdf.Text("Pristopna izjava k MeetPlan sistemu")

	pdf.Line(20, 200, 575, 200)

	pdf.SetFontSize(13)
	pdf.SetX(borderBase)
	pdf.SetY(240)
	pdf.Text("Verjetno ste bili že obveščeni, da je vaša šola to leto izbrala drug sistem.")
	pdf.SetX(borderBase)
	pdf.SetY(255)
	pdf.Text("MeetPlan je popolnoma odprtokoden sistem, ki je popolnoma brezplačen za vse.")
	pdf.SetX(borderBase)
	pdf.SetY(270)
	pdf.Text("Ta izjava vsebuje vaše osebne podatke za dostop do MeetPlan sistema.")
	pdf.SetX(borderBase)
	pdf.SetY(285)
	pdf.Text("Priporočamo, da pri prvem vstopu v sistem to geslo tudi zamenjate.")
	pdf.SetX(borderBase)
	pdf.SetY(300)
	pdf.Text("Poleg spodaj naštetih podatkov zbiramo samo še matično številko osebe.")

	const differ = 150

	pdf.SetX(borderBase)
	pdf.SetY(350)
	pdf.SetFontSize(10)
	pdf.Text("uporabnik")
	pdf.SetFontSize(20)
	pdf.SetX(differ)
	pdf.Text(user.Name)

	pdf.SetX(borderBase)
	pdf.SetY(380)
	pdf.SetFontSize(10)
	pdf.Text("elektronski naslov")
	pdf.SetX(differ)
	pdf.SetFontSize(20)
	pdf.Text(user.Email)

	pdf.SetX(borderBase)
	pdf.SetY(410)
	pdf.SetFontSize(10)
	pdf.Text("geslo")
	pdf.SetX(differ)
	pdf.SetFontSize(20)
	pdf.Text(newPassword)

	pdf.SetX(borderBase)
	pdf.SetY(440)
	pdf.SetFontSize(10)
	pdf.Text("naziv v sistemu")
	pdf.SetX(differ)
	pdf.SetFontSize(20)
	pdf.Text(user.Role)

	pdf.SetX(borderBase)
	pdf.SetY(470)
	pdf.SetFontSize(10)
	pdf.Text("kraj rojstva")
	pdf.SetX(differ)
	pdf.SetFontSize(20)
	pdf.Text(user.CityOfBirth)

	pdf.SetX(borderBase)
	pdf.SetY(500)
	pdf.SetFontSize(10)
	pdf.Text("država rojstva")
	pdf.SetX(differ)
	pdf.SetFontSize(20)
	pdf.Text(user.CountryOfBirth)

	pdf.SetX(borderBase)
	pdf.SetY(530)
	pdf.SetFontSize(10)
	pdf.Text("datum rojstva")
	pdf.SetX(differ)
	pdf.SetFontSize(20)
	pdf.Text(user.Birthday)

	if server.config.Debug {
		pdf.SetX(borderBase)
		pdf.SetY(560)
		pdf.SetFontSize(10)
		pdf.Text("enolični identifikator")
		pdf.SetFontSize(20)
		pdf.SetX(differ)
		pdf.Text(fmt.Sprint(user.ID))
	}

	err = server.db.UpdateUser(user)
	return pdf, err
}

func (server *httpImpl) ResetPassword(w http.ResponseWriter, r *http.Request) {
	jwtData, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}

	if jwtData["role"] == "admin" || jwtData["role"] == "principal" || jwtData["role"] == "principal assistant" {
		id, err := strconv.Atoi(mux.Vars(r)["user_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}

		pdf := &gopdf.GoPdf{}
		pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

		err = pdf.AddTTFFont("opensans", "fonts/opensans.ttf")
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}

		err = pdf.SetFont("opensans", "", 11)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}

		pdf, err = server.GenerateNewUserCert(pdf, id)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed at generating PDF", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}

		w.Write(pdf.GetBytesPdf())
	} else {
		WriteForbiddenJWT(w)
	}
}

func (server *httpImpl) ChangePassword(w http.ResponseWriter, r *http.Request) {
	jwtData, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}

	userId, err := strconv.Atoi(fmt.Sprint(jwtData["user_id"]))
	if err != nil {
		WriteBadRequest(w)
		return
	}

	user, err := server.db.GetUser(userId)
	if err != nil {
		return
	}

	oldPass := r.FormValue("oldPassword")
	if !sql.CheckHash(oldPass, user.Password) {
		WriteJSON(w, Response{Data: "Wrong password", Success: false}, http.StatusForbidden)
		return
	}

	password, err := sql.HashPassword(r.FormValue("password"))
	if err != nil {
		return
	}

	user.Password = password

	server.db.UpdateUser(user)

	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}
