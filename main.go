package main

import (
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/httphandlers"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"net/http"
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

	db, err := sql.NewSQL("sqlite3", "meetplan.db", sugared)
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

	r.HandleFunc("/class/get/{class_id}/self_testing", httphandler.GetSelfTestingTeacher).Methods("GET")
	r.HandleFunc("/user/self_testing/patch/{student_id}", httphandler.PatchSelfTesting).Methods("PATCH")
	r.HandleFunc("/user/self_testing/get_results/pdf/{test_id}", httphandler.GetPDFSelfTestingReportStudent).Methods("GET")

	r.HandleFunc("/class/new", httphandler.NewClass).Methods("POST")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // All origins
		AllowedMethods: []string{"POST", "GET", "DELETE", "PATCH", "PUT"},
	})

	err = http.ListenAndServe("127.0.0.1:8000", c.Handler(r))
	if err != nil {
		sugared.Fatal(err.Error())
	}

	sugared.Info("Done serving...")
}
