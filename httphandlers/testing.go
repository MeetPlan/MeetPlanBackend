package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
	"net/http"
	"os"
	"strings"
	"time"
)

func (server *httpImpl) GetSelfTestingTeacher(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == TEACHER || user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == SCHOOL_PSYCHOLOGIST) {
		WriteForbiddenJWT(w)
		return
	}
	classId := mux.Vars(r)["class_id"]
	if err != nil {
		WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
		return
	}
	dt := time.Now()
	date := dt.Format("02-01-2006")
	results, err := server.db.GetTestingResults(date, classId)
	if err != nil {
		WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Success: true, Data: results}, http.StatusOK)
}

func (server *httpImpl) PatchSelfTesting(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == SCHOOL_PSYCHOLOGIST) {
		WriteForbiddenJWT(w)
		return
	}
	studentId := mux.Vars(r)["student_id"]
	if err != nil {
		WriteBadRequest(w)
		return
	}

	student, err := server.db.GetUser(studentId)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while fetching student from the database", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	if student.Role != STUDENT {
		WriteJSON(w, Response{Data: "Student isn't really a student", Success: false}, http.StatusBadRequest)
		return
	}

	classId := mux.Vars(r)["class_id"]
	if err != nil {
		WriteBadRequest(w)
		return
	}

	class, err := server.db.GetClass(classId)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while fetching student from the database", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	dt := time.Now()
	date := dt.Format("02-01-2006")

	results, err := server.db.GetTestingResult(date, studentId)
	if err != nil {
		server.logger.Debug(err)
		if err.Error() == "sql: no rows in result set" {
			results = sql.Testing{
				Date: date,

				UserID:    studentId,
				TeacherID: user.ID,
				ClassID:   class.ID,
				Result:    r.FormValue("result"),
			}
			err := server.db.InsertTestingResult(results)
			if err != nil {
				WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
				return
			}
			WriteJSON(w, Response{Success: true, Data: results}, http.StatusOK)
			return
		} else {
			WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
			return
		}
	} else {
		newr := r.FormValue("result")
		if newr == results.Result {
			results.Result = ""
		} else {
			results.Result = newr
		}
	}
	results.TeacherID = user.ID
	err = server.db.UpdateTestingResult(results)
	if err != nil {
		WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Success: true, Data: results}, http.StatusOK)
}

func (server *httpImpl) GetPDFSelfTestingReportStudent(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}

	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == TEACHER || user.Role == PARENT || user.Role == SCHOOL_PSYCHOLOGIST || user.Role == STUDENT) {
		WriteForbiddenJWT(w)
		return
	}
	id := mux.Vars(r)["test_id"]
	if err != nil {
		WriteBadRequest(w)
		return
	}

	test, err := server.db.GetTestingResultByID(id)
	if err != nil {
		return
	}

	if user.Role == STUDENT {
		if test.UserID != user.ID {
			WriteForbiddenJWT(w)
			return
		}
	} else if user.Role == PARENT {
		var users []string
		json.Unmarshal([]byte(user.Users), &users)
		if !helpers.Contains(users, test.UserID) {
			WriteForbiddenJWT(w)
			return
		}
	}

	if test.Result == "SE NE TESTIRA" {
		return
	}

	teacher, err := server.db.GetUser(test.TeacherID)
	if err != nil {
		return
	}

	student, err := server.db.GetUser(test.UserID)
	if err != nil {
		return
	}

	//jwt, err, expiration := sql.GetJWTForTestingResult(test.UserID, test.Result, test.ID, test.Date)
	//if err != nil {
	//	return
	//}

	m := pdf.NewMaroto(consts.Portrait, consts.A4)

	m.AddUTF8Font("OpenSans", consts.Normal, "fonts/opensans.ttf")
	m.SetDefaultFontFamily("OpenSans")

	m.Row(40, func() {
		m.Col(3, func() {
			_ = m.Base64Image(MeetPlanLogoBase64, consts.Png, props.Rect{
				Center:  true,
				Percent: 80,
			})
		})

		m.ColSpace(1)

		m.Col(5, func() {
			m.Text("MeetPlan", props.Text{
				Top:         12,
				Size:        30,
				Extrapolate: true,
			})
			m.Text("Rezultati samotestiranja", props.Text{
				Size: 20,
				Top:  22,
			})
		})
		m.ColSpace(4)
	})

	m.Line(10)

	m.Row(40, func() {
		m.Col(7, func() {
			m.Text("Rezultat testiranja: ", props.Text{
				Size: 15,
				Top:  12,
			})
			m.Text(test.Result, props.Text{Size: 20, Top: 20})
			if test.Result == "POZITIVEN" {
				m.Text(
					"Vaš test je bil pozitiven. Samoizolirajte se v čim manjšem možnem času. To potrdilo vam lahko s podpisom osebe, ki je izvajala testiranje, tudi služi kot dokaz za PCR testiranje.",
					props.Text{Top: 30},
				)
			} else if test.Result == "NEVELJAVEN" {
				m.Text(
					"Vaš test je bil neveljaven. Ponovite testiranje.",
					props.Text{Top: 30},
				)
			}
		})
		m.ColSpace(1)
		//m.Col(4, func() {
		//	m.QrCode(jwt, props.Rect{
		//		Center:  true,
		//		Percent: 100,
		//	})
		//})
	})

	m.Line(10)

	m.SetBorder(true)

	m.Row(60, func() {
		m.Col(6, func() {
			m.Text(fmt.Sprintf(" Enolični identifikator testiranja: %s", fmt.Sprint(test.ID)), props.Text{
				Size: 15,
				Top:  5,
			})
			m.Text(fmt.Sprintf(" Datum izvedbe testiranja: %s", test.Date), props.Text{
				Size: 15,
				Top:  15,
			})
			//m.Text(fmt.Sprintf(" Datum veljavnosti testa: %s", expiration), props.Text{
			//	Size: 15,
			//	Top:  25,
			//})
			m.Text(fmt.Sprintf(" Oseba: %s", student.Name), props.Text{
				Size: 15,
				Top:  35,
			})
			m.Text(fmt.Sprintf(" Enolični identifikator osebe: %s", fmt.Sprint(student.ID)), props.Text{
				Size: 15,
				Top:  45,
			})
		})
		m.Col(6, func() {
			m.Text("Izdal MeetPlan Certificate Authority", props.Text{
				Top:   25,
				Size:  15,
				Align: consts.Center,
			})
		})
	})

	m.SetBorder(false)

	m.Row(40, func() {})

	m.Row(40, func() {
		m.ColSpace(1)
		m.Col(6, func() {
			m.Text("_________________________", props.Text{
				Top:  14,
				Size: 15,
			})
			m.Text(teacher.Name, props.Text{
				Top:  14,
				Size: 15,
			})
			m.Text("digitalni podpis izvajalca testiranja", props.Text{Top: 20, Size: 9})
		})
		m.Col(3, func() {
			m.Text("_________________________", props.Text{
				Top:  14,
				Size: 15,
			})
			m.Text("podpis izvajalca testiranja", props.Text{Top: 20, Size: 9})
		})
	})

	m.Row(5, func() {
		m.Col(12, func() {
			m.Text("S podpisom tega dokumenta, potrjujem, da se je oseba, navedena zgoraj samotestirala in to sem to tudi potrdil(a).", props.Text{
				Top:  14,
				Size: 15,
			})
		})
	})

	UUID := uuid.New().String()

	m.Row(15, func() {
		m.Col(12, func() {
			m.Text(fmt.Sprintf("Edinstveni identifikator dokumenta: %s", UUID), props.Text{
				Top:  40,
				Size: 10,
			})
		})
	})

	m.Row(5, func() {
		m.Col(12, func() {
			m.Text(fmt.Sprintf("Izvozil/a: %s", user.Name), props.Text{
				Top:  12,
				Size: 10,
			})
		})
	})

	filename := fmt.Sprintf("documents/%s.pdf", UUID)

	output, err := m.Output()
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	err = helpers.Sign(output.Bytes(), filename, "cacerts/key-pair.p12", "")
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while signing", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	document := sql.Document{
		ID:           UUID,
		ExportedBy:   user.ID,
		DocumentType: POTRDILO_O_SAMOTESTIRANJU,
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

func (server *httpImpl) GetTestingResults(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	results, err := server.db.GetAllTestingsForUser(user.ID)
	if err != nil {
		return
	}

	var res = make([]sql.TestingJSON, 0)

	for i := 0; i < len(results); i++ {
		r := results[i]
		teacher, err := server.db.GetUser(r.TeacherID)
		if err != nil {
			return
		}
		expirationTime, err := time.Parse("02-01-2006", r.Date)
		if err != nil {
			return
		}
		expirationTime = expirationTime.Add(48 * time.Hour)
		etime := strings.Split(expirationTime.Format("02-01-2006"), " ")[0]
		j := sql.TestingJSON{IsDone: true, ID: r.ID, ClassID: r.ClassID, TeacherID: r.TeacherID, TeacherName: teacher.Name, UserID: r.UserID, Date: r.Date, Result: r.Result, ValidUntil: etime}
		res = append(res, j)
	}
	// Magic to reverse slice
	for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
		res[i], res[j] = res[j], res[i]
	}
	WriteJSON(w, Response{Data: res, Success: true}, http.StatusOK)
}
