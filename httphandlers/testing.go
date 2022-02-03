package httphandlers

import (
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	Error   interface{} `json:"error"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

func (server *httpImpl) GetSelfTestingTeacher(w http.ResponseWriter, r *http.Request) {
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
}

func (server *httpImpl) PatchSelfTesting(w http.ResponseWriter, r *http.Request) {
	studentId, err := strconv.Atoi(mux.Vars(r)["student_id"])
	if err != nil {
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
				TeacherID: 0,
				ClassID:   0,
				Result:    r.FormValue("result"),
			}
			err := server.db.InsertTestingResult(results)
			if err != nil {
				WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
				return
			}
		} else {
			WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
			return
		}
	}
	newr := r.FormValue("result")
	if newr == results.Result {
		results.Result = ""
	} else {
		results.Result = newr
	}
	err = server.db.UpdateTestingResult(results)
	if err != nil {
		WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Success: true, Data: results}, http.StatusOK)
}

func (server *httpImpl) GetPDFSelfTestingReportStudent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["test_id"])
	if err != nil {
		return
	}

	test, err := server.db.GetTestingResultByID(id)
	if err != nil {
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

	m.Row(40, func() {
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
		m.Col(4, func() {
			m.Text("Rezultat testiranja: ", props.Text{
				Size: 15,
				Top:  12,
			})
			m.Text(test.Result, props.Text{Size: 20, Top: 20})
		})
		m.ColSpace(4)
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
			m.Text(fmt.Sprintf(" Enolicni identifikator testiranja: %s", fmt.Sprint(test.ID)), props.Text{
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
			m.Text(fmt.Sprintf(" Enolicni identifikator osebe: %s", fmt.Sprint(student.ID)), props.Text{
				Size: 15,
				Top:  45,
			})
		})
		m.Col(6, func() {
			m.Text("Izdal MeetPlanCA", props.Text{
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

	output, err := m.Output()
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	w.Write(output.Bytes())
}
