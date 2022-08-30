package main

import (
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/httphandlers"
	"github.com/MeetPlan/MeetPlanBackend/proton"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
)

func main() {
	fmt.Println("Starting MeetPlan server...")

	var logger *zap.Logger
	var err error

	config, err := sql.GetConfig()
	if err != nil {
		panic("Error while retrieving config: " + err.Error())
		return
	}

	if config.Debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		panic(err.Error())
		return
	}

	sugared := logger.Sugar()

	if _, err := os.Stat("MeetPlanDB"); os.IsNotExist(err) {
		os.Mkdir("MeetPlanDB", os.ModePerm)
	}

	if _, err := os.Stat("documents"); os.IsNotExist(err) {
		os.Mkdir("documents", os.ModePerm)
	}

	db, err := sql.NewSQL(config.DatabaseName, config.DatabaseConfig, sugared)
	db.Init()

	sugared.Infof("Hardcoded commit is %s, meanwhile current version was at commit %s", sql.COMMIT_HASH, config.CommitHash)

	if config.CommitHash == "" && sql.COMMIT_HASH != "" {
		config.CommitHash = sql.COMMIT_HASH
		err := sql.SaveConfig(config)
		if err != nil {
			sugared.Fatalf("Failed while saving the config %s", err.Error())
		}
	}

	// MeetPlan Database automigration tool
	if config.CommitHash != "" && sql.COMMIT_HASH != "" && config.CommitHash != sql.COMMIT_HASH {
		sugared.Info("Attempting to do database automigration")

		r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL: config.RemoteRepository,
		})

		if err != nil {
			sugared.Fatal(err.Error())
			return
		}

		latestCommit, err := r.CommitObject(plumbing.NewHash(sql.COMMIT_HASH))
		if err != nil {
			sugared.Fatal(err.Error())
			return
		}
		currentCommit, err := r.CommitObject(plumbing.NewHash(config.CommitHash))
		if err != nil {
			sugared.Fatal(err.Error())
			return
		}

		patches, err := currentCommit.Patch(latestCommit)
		if err != nil {
			sugared.Fatal(err.Error())
			return
		}
		for _, v := range patches.FilePatches() {
			filePatch := v
			_, after := filePatch.Files()
			if after == nil {
				continue
			}
			match, err := regexp.MatchString("migrations/.*\\.sql", after.Path())
			if err != nil {
				sugared.Errorf("Failed while using regex to verify the migration %s with an error of %s", after.Path(), err.Error())
				continue
			}
			if !match {
				continue
			}

			sugared.Infof("Currently executing migrations on file %s", after.Path())

			object, err := r.BlobObject(after.Hash())
			if err != nil {
				sugared.Errorf("Failed while creating a blob object for %s with an error of %s", after.Path(), err.Error())
				continue
			}

			reader, err := object.Reader()
			if err != nil {
				sugared.Errorf("Failed while creating an reader object for %s with an error of %s", after.Path(), err.Error())
				continue
			}

			read, err := ioutil.ReadAll(reader)
			if err != nil {
				sugared.Errorf("Failed while reading a blob object %s with an error of %s", after.Path(), err.Error())
				continue
			}

			err = db.Exec(string(read))
			if err != nil {
				sugared.Fatalf("Failed while executing one of the migrations %s", err.Error())
			}

			sugared.Infof("Successfully executed %s from file %s on the database", string(read), after.Path())
		}

		config.CommitHash = sql.COMMIT_HASH
		err = sql.SaveConfig(config)
		if err != nil {
			sugared.Fatalf("Failed while saving the config %s", err.Error())
		}

		sugared.Info("Successfully executed the migrations and saved the configuration.")
	}

	if err != nil {
		sugared.Fatal("Error while creating database: ", err.Error())
		return
	}

	protonState, err := proton.NewProton(db, sugared)
	if err != nil {
		sugared.Fatal("Error while initializing Proton: ", err.Error())
		return
	}

	httphandler := httphandlers.NewHTTPInterface(sugared, db, config, protonState)

	sugared.Info("Database created successfully")

	r := mux.NewRouter()
	r.HandleFunc("/user/new", httphandler.NewUser).Methods("POST")
	r.HandleFunc("/user/login", httphandler.Login).Methods("POST")
	// Get all classes for specific user
	r.HandleFunc("/user/get/classes", httphandler.GetAllClasses).Methods("GET")
	r.HandleFunc("/user/get/password_change", httphandler.ChangePassword).Methods("PATCH")
	r.HandleFunc("/user/check/has/class", httphandler.HasClass).Methods("GET")
	r.HandleFunc("/user/get/data/{id}", httphandler.GetUserData).Methods("GET")
	r.HandleFunc("/user/get/data/{user_id}", httphandler.PatchUser).Methods("PATCH")
	r.HandleFunc("/user/get/password_reset/{user_id}", httphandler.ResetPassword).Methods("GET")
	r.HandleFunc("/user/get/homework/{id}", httphandler.GetUserHomework).Methods("GET")
	r.HandleFunc("/user/get/absences/{id}", httphandler.GetAbsencesUser).Methods("GET")
	r.HandleFunc("/user/get/improvements", httphandler.GetImprovementsForUser).Methods("GET")
	r.HandleFunc("/user/get/ending_certificate/{student_id}", httphandler.PrintCertificateOfEndingClass).Methods("GET")
	r.HandleFunc("/user/get/certificate_of_schooling/{user_id}", httphandler.CertificateOfSchooling).Methods("GET")
	r.HandleFunc("/user/get/unread_messages", httphandler.GetUnreadMessages).Methods("GET")

	r.HandleFunc("/user/get/absences/{student_id}/excuse/{absence_id}", httphandler.ExcuseAbsence).Methods("PATCH")

	r.HandleFunc("/class/get/{class_id}/self_testing", httphandler.GetSelfTestingTeacher).Methods("GET")
	r.HandleFunc("/user/self_testing/patch/{class_id}/{student_id}", httphandler.PatchSelfTesting).Methods("PATCH")
	r.HandleFunc("/user/self_testing/get_results", httphandler.GetTestingResults).Methods("GET")
	r.HandleFunc("/user/self_testing/get_results/pdf/{test_id}", httphandler.GetPDFSelfTestingReportStudent).Methods("GET")

	r.HandleFunc("/class/new", httphandler.NewClass).Methods("POST")
	r.HandleFunc("/class/get/{id}", httphandler.GetClass).Methods("GET")
	r.HandleFunc("/class/get/{id}", httphandler.PatchClass).Methods("PATCH")
	r.HandleFunc("/class/get/{id}", httphandler.DeleteClass).Methods("DELETE")
	// Get all classes in database
	r.HandleFunc("/classes/get", httphandler.GetClasses).Methods("GET")
	r.HandleFunc("/class/get/{class_id}/add_user/{user_id}", httphandler.AssignUserToClass).Methods("PATCH")
	r.HandleFunc("/class/get/{class_id}/remove_user/{user_id}", httphandler.RemoveUserFromClass).Methods("DELETE")

	r.HandleFunc("/users/get", httphandler.GetAllUsers).Methods("GET")
	r.HandleFunc("/meals/get", httphandler.GetMeals).Methods("GET")
	r.HandleFunc("/meal/get/{meal_id}", httphandler.EditMeal).Methods("PATCH")
	r.HandleFunc("/meal/get/{meal_id}", httphandler.DeleteMeal).Methods("DELETE")
	r.HandleFunc("/meals/new", httphandler.NewMeal).Methods("POST")
	r.HandleFunc("/meals/blocked", httphandler.MealsBlocked).Methods("GET")
	r.HandleFunc("/teachers/get", httphandler.GetTeachers).Methods("GET")
	r.HandleFunc("/students/get", httphandler.GetStudents).Methods("GET")
	r.HandleFunc("/user/role/update/{id}", httphandler.ChangeRole).Methods("PATCH")
	r.HandleFunc("/user/delete/{id}", httphandler.DeleteUser).Methods("DELETE")

	r.HandleFunc("/parent/{parent}/assign/student/{student}", httphandler.AssignUserToParent).Methods("PATCH")
	r.HandleFunc("/parent/{parent}/assign/student/{student}", httphandler.RemoveUserFromParent).Methods("DELETE")
	r.HandleFunc("/parents/get/students", httphandler.GetMyChildren).Methods("GET")
	r.HandleFunc("/parents/get/config", httphandler.ParentConfig).Methods("GET")

	r.HandleFunc("/order/new/{meal_id}", httphandler.NewOrder).Methods("POST")
	r.HandleFunc("/order/get/{meal_id}/block_unblock", httphandler.BlockUnblockOrder).Methods("PATCH")
	r.HandleFunc("/order/get/{meal_id}", httphandler.RemoveOrder).Methods("DELETE")
	r.HandleFunc("/order/get/{meal_id}/{user_id}", httphandler.RemoveSpecificOrder).Methods("DELETE")

	r.HandleFunc("/my/grades", httphandler.GetMyGrades).Methods("GET")
	r.HandleFunc("/my/gradings", httphandler.GetMyGradings).Methods("GET")

	r.HandleFunc("/timetable/get", httphandler.GetTimetable).Methods("GET")

	r.HandleFunc("/meetings/new", httphandler.NewMeeting).Methods("POST")
	r.HandleFunc("/meetings/new/{id}", httphandler.PatchMeeting).Methods("PATCH")
	r.HandleFunc("/meetings/new/{id}", httphandler.DeleteMeeting).Methods("DELETE")
	r.HandleFunc("/meetings/beta", httphandler.MigrateBetaMeetings).Methods("PATCH")
	r.HandleFunc("/meetings/beta", httphandler.DeleteBetaMeetings).Methods("DELETE")

	r.HandleFunc("/communications/get", httphandler.GetCommunications).Methods("GET")
	r.HandleFunc("/communication/get/{id}", httphandler.GetCommunication).Methods("GET")
	r.HandleFunc("/communication/get/{id}/message/new", httphandler.NewMessage).Methods("POST")
	r.HandleFunc("/communication/new", httphandler.NewCommunication).Methods("POST")

	r.HandleFunc("/message/get/{message_id}", httphandler.DeleteMessage).Methods("DELETE")
	r.HandleFunc("/message/get/{message_id}", httphandler.EditMessage).Methods("PATCH")

	r.HandleFunc("/meeting/get/{meeting_id}", httphandler.GetMeeting).Methods("GET")
	r.HandleFunc("/meeting/get/{meeting_id}/absences", httphandler.GetAbsencesTeacher).Methods("GET")
	r.HandleFunc("/meeting/get/{meeting_id}/users", httphandler.GetUsersForMeeting).Methods("GET")
	r.HandleFunc("/meeting/get/{meeting_id}/grades", httphandler.GetGradesForMeeting).Methods("GET")
	r.HandleFunc("/meeting/get/{meeting_id}/homework/{homework_id}/{student_id}", httphandler.PatchHomeworkForStudent).Methods("PATCH")
	r.HandleFunc("/meeting/get/{meeting_id}/improvement/new/{student_id}", httphandler.NewImprovement).Methods("POST")
	r.HandleFunc("/meeting/get/{meeting_id}/homework", httphandler.NewHomework).Methods("POST")
	r.HandleFunc("/meeting/get/{meeting_id}/homework", httphandler.GetAllHomeworksForSpecificSubject).Methods("GET")
	r.HandleFunc("/meeting/get/{meeting_id}/substitutions/proton", httphandler.ManageTeacherAbsences).Methods("GET")

	r.HandleFunc("/meeting/absence/{absence_id}", httphandler.PatchAbsence).Methods("PATCH")

	r.HandleFunc("/grades/new/{meeting_id}", httphandler.NewGrade).Methods("POST")

	r.HandleFunc("/grade/get/{grade_id}", httphandler.PatchGrade).Methods("PATCH")
	r.HandleFunc("/grade/get/{grade_id}", httphandler.DeleteGrade).Methods("DELETE")

	r.HandleFunc("/subjects/get", httphandler.GetSubjects).Methods("GET")
	r.HandleFunc("/subjects/new", httphandler.NewSubject).Methods("POST")

	r.HandleFunc("/subject/get/{subject_id}", httphandler.GetSubject).Methods("GET")
	r.HandleFunc("/subject/get/{subject_id}", httphandler.DeleteSubject).Methods("DELETE")
	r.HandleFunc("/subject/get/{subject_id}", httphandler.PatchSubjectName).Methods("PATCH")
	r.HandleFunc("/subject/get/{subject_id}/add_user/{user_id}", httphandler.AssignUserToSubject).Methods("PATCH")
	r.HandleFunc("/subject/get/{subject_id}/remove_user/{user_id}", httphandler.RemoveUserFromSubject).Methods("DELETE")

	r.HandleFunc("/admin/config/get", httphandler.GetConfig).Methods("GET")
	r.HandleFunc("/admin/config/get", httphandler.UpdateConfiguration).Methods("PATCH")

	r.HandleFunc("/system/notifications", httphandler.GetSystemNotifications).Methods("GET")
	r.HandleFunc("/system/notifications/new", httphandler.NewNotification).Methods("POST")
	r.HandleFunc("/notification/{notification_id}", httphandler.DeleteNotification).Methods("DELETE")

	r.HandleFunc("/proton/rule/new", httphandler.NewProtonRule).Methods("POST")
	r.HandleFunc("/proton/rules/get", httphandler.GetProtonRules).Methods("GET")
	r.HandleFunc("/proton/rule/get", httphandler.DeleteProtonRule).Methods("DELETE")

	r.HandleFunc("/proton/assemble/timetable", httphandler.AssembleTimetable).Methods("GET")
	r.HandleFunc("/proton/accept/timetable", httphandler.AcceptAssembledTimetable).Methods("POST")
	r.HandleFunc("/proton/timetable/manual_postprocessing", httphandler.ManualPostProcessRepeat).Methods("POST")

	r.HandleFunc("/documents/get", httphandler.FetchAllDocuments).Methods("GET")
	r.HandleFunc("/documents/get", httphandler.DeleteDocument).Methods("DELETE")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // All origins
		AllowedHeaders: []string{"Authorization"},
		AllowedMethods: []string{"POST", "GET", "DELETE", "PATCH", "PUT"},
	})

	err = http.ListenAndServe(config.Host, c.Handler(r))
	if err != nil {
		sugared.Fatal(err.Error())
	}

	sugared.Info("Done serving...")
}
