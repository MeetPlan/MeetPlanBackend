package main

import (
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/httphandlers"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func main() {
	fmt.Println("Starting MeetPlan server...")

	var logger *zap.Logger
	var err error

	logger, err = zap.NewDevelopment()

	if err != nil {
		panic(err.Error())
		return
	}

	sugared := logger.Sugar()

	if _, err := os.Stat("MeetPlanDB"); os.IsNotExist(err) {
		os.Mkdir("MeetPlanDB", os.ModePerm)
	}

	db, err := sql.NewSQL("sqlite3", "MeetPlanDB/meetplan.db", sugared)
	db.Init()

	if err != nil {
		sugared.Fatal("Error while creating database: " + err.Error())
		return
	}

	httphandler := httphandlers.NewHTTPInterface(sugared, db)

	sugared.Info("Database created successfully")

	r := mux.NewRouter()
	r.HandleFunc("/user/new", httphandler.NewUser).Methods("POST")
	r.HandleFunc("/user/login", httphandler.Login).Methods("POST")
	// Get all classes for specific user
	r.HandleFunc("/user/get/classes", httphandler.GetAllClasses).Methods("GET")

	r.HandleFunc("/class/get/{class_id}/self_testing", httphandler.GetSelfTestingTeacher).Methods("GET")
	r.HandleFunc("/user/self_testing/patch/{class_id}/{student_id}", httphandler.PatchSelfTesting).Methods("PATCH")
	r.HandleFunc("/user/self_testing/get_results", httphandler.GetTestingResults).Methods("GET")
	r.HandleFunc("/user/self_testing/get_results/pdf/{test_id}", httphandler.GetPDFSelfTestingReportStudent).Methods("GET")

	r.HandleFunc("/class/new", httphandler.NewClass).Methods("POST")
	r.HandleFunc("/class/get/{id}", httphandler.GetClass).Methods("GET")
	r.HandleFunc("/class/get/{id}", httphandler.DeleteClass).Methods("DELETE")
	// Get all classes in database
	r.HandleFunc("/classes/get", httphandler.GetClasses).Methods("GET")
	r.HandleFunc("/class/get/{class_id}/add_user/{user_id}", httphandler.AssignUserToClass).Methods("PATCH")
	r.HandleFunc("/class/get/{class_id}/remove_user/{user_id}", httphandler.RemoveUserFromClass).Methods("DELETE")

	r.HandleFunc("/users/get", httphandler.GetAllUsers).Methods("GET")
	r.HandleFunc("/teachers/get", httphandler.GetTeachers).Methods("GET")
	r.HandleFunc("/students/get", httphandler.GetStudents).Methods("GET")
	r.HandleFunc("/user/role/update/{id}", httphandler.ChangeRole).Methods("PATCH")
	r.HandleFunc("/user/delete/{id}", httphandler.DeleteUser).Methods("DELETE")

	r.HandleFunc("/timetable/get", httphandler.GetTimetable).Methods("GET")

	r.HandleFunc("/meetings/new", httphandler.NewMeeting).Methods("POST")
	r.HandleFunc("/meetings/new/{id}", httphandler.PatchMeeting).Methods("PATCH")
	r.HandleFunc("/meetings/new/{id}", httphandler.DeleteMeeting).Methods("DELETE")

	r.HandleFunc("/meeting/get/{meeting_id}", httphandler.GetMeeting).Methods("GET")
	r.HandleFunc("/meeting/get/{meeting_id}/absences", httphandler.GetAbsencesTeacher).Methods("GET")
	r.HandleFunc("/meeting/get/{meeting_id}/grades", httphandler.GetGradesForMeeting).Methods("GET")

	r.HandleFunc("/meeting/absence/{absence_id}", httphandler.PatchAbsence).Methods("PATCH")

	r.HandleFunc("/grades/new/{meeting_id}", httphandler.NewGrade).Methods("POST")

	r.HandleFunc("/grade/get/{grade_id}", httphandler.PatchGrade).Methods("PATCH")
	r.HandleFunc("/grade/get/{grade_id}", httphandler.DeleteGrade).Methods("DELETE")

	r.HandleFunc("/subjects/get", httphandler.GetSubjects).Methods("GET")
	r.HandleFunc("/subjects/new", httphandler.NewSubject).Methods("POST")

	r.HandleFunc("/subject/get/{subject_id}", httphandler.GetSubject).Methods("GET")
	r.HandleFunc("/subject/get/{subject_id}", httphandler.DeleteSubject).Methods("DELETE")
	r.HandleFunc("/subject/get/{subject_id}/homework", httphandler.NewHomework).Methods("POST")
	r.HandleFunc("/subject/get/{subject_id}/homework", httphandler.GetAllHomeworksForSpecificSubject).Methods("GET")
	r.HandleFunc("/subject/get/{subject_id}/add_user/{user_id}", httphandler.AssignUserToSubject).Methods("PATCH")
	r.HandleFunc("/subject/get/{subject_id}/remove_user/{user_id}", httphandler.RemoveUserFromSubject).Methods("DELETE")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // All origins
		AllowedHeaders: []string{"Authorization"},
		AllowedMethods: []string{"POST", "GET", "DELETE", "PATCH", "PUT"},
	})

	host := os.Getenv("MP_HOST")
	if host == "" {
		host = "127.0.0.1:8000"
	}

	err = http.ListenAndServe(host, c.Handler(r))
	if err != nil {
		sugared.Fatal(err.Error())
	}

	sugared.Info("Done serving...")
}
