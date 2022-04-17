package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Error   interface{} `json:"error"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

func (server *httpImpl) GetSelfTestingTeacher(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "teacher" || jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		classId, err := strconv.Atoi(mux.Vars(r)["class_id"])
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
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) PatchSelfTesting(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "teacher" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		studentId, err := strconv.Atoi(mux.Vars(r)["student_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}

		classId, err := strconv.Atoi(mux.Vars(r)["class_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}

		teacherId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteBadRequest(w)
			return
		}

		dt := time.Now()
		date := dt.Format("02-01-2006")

		results, err := server.db.GetTestingResult(date, studentId)
		if err != nil {
			server.logger.Debug(err)
			if err.Error() == "sql: no rows in result set" {
				results = sql.Testing{
					Date:      date,
					ID:        server.db.GetLastTestingID(),
					UserID:    studentId,
					TeacherID: teacherId,
					ClassID:   classId,
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
		ntid, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		results.TeacherID = ntid
		err = server.db.UpdateTestingResult(results)
		if err != nil {
			WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
			return
		}

		WriteJSON(w, Response{Success: true, Data: results}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}

func (server *httpImpl) GetPDFSelfTestingReportStudent(w http.ResponseWriter, r *http.Request) {
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

	id, err := strconv.Atoi(mux.Vars(r)["test_id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}

	test, err := server.db.GetTestingResultByID(id)
	if err != nil {
		return
	}

	if jwtData["role"] == "student" {
		if test.UserID != userId {
			WriteForbiddenJWT(w)
			return
		}
	} else if jwtData["role"] == "parent" {
		var users []int
		json.Unmarshal([]byte(user.Users), &users)
		if !contains(users, test.UserID) {
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

	jwt, err, expiration := sql.GetJWTForTestingResult(test.UserID, test.Result, test.ID, test.Date)
	if err != nil {
		return
	}

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
		m.Col(4, func() {
			m.QrCode(jwt, props.Rect{
				Center:  true,
				Percent: 100,
			})
		})
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
			m.Text(fmt.Sprintf(" Datum veljavnosti testa: %s", expiration), props.Text{
				Size: 15,
				Top:  25,
			})
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

	output, err := m.Output()
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	w.Write(output.Bytes())
}

func (server *httpImpl) GetTestingResults(w http.ResponseWriter, r *http.Request) {
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
	results, err := server.db.GetAllTestingsForUser(userId)
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
