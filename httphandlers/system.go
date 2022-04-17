package httphandlers

import (
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

func (server *httpImpl) GetSystemNotifications(w http.ResponseWriter, r *http.Request) {
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
		WriteJSON(w, Response{Data: "Could not retrieve user", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	notifications, err := server.db.GetAllNotifications()
	if err != nil {
		WriteJSON(w, Response{Data: "Could not fetch notifications", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if notifications == nil {
		notifications = make([]sql.NotificationSQL, 0)
	}
	currentTime := time.Now()
	birthday, err := time.Parse("2006-01-02", user.Birthday)
	if err == nil {
		if currentTime.Before(birthday) {
			WriteJSON(w, Response{Data: "Invalid birthday", Success: false}, http.StatusConflict)
			return
		}
		_, tm, td := currentTime.Date()
		_, bm, bd := birthday.Date()
		if tm-bm == 0 && td-bd == 0 {
			if jwt["role"] == "student" {
				notifications = append(notifications, sql.NotificationSQL{Notification: "\U0001F973 Kdo pa ima danes rojstni dan? Odgovor: Ti. Čeprav ne moremo urediti, da danes nimaš šole, ti ekipa MeetPlan sistema želi vse najboljše in čim boljše ocene v tem šolskem letu."})
			} else {
				notifications = append(notifications, sql.NotificationSQL{Notification: "\U0001F973 Kdo pa ima danes rojstni dan? Odgovor: Vi. Čeprav ne moremo urediti, da danes nimate službe, vam ekipa MeetPlan sistema želi vse najboljše. Biti učitelj je zelo plemenito delo in zato se vam zahvaljujemo. Še naprej širite svoje znanje na nove generacije."})
			}
		}
	}
	WriteJSON(w, Response{Data: notifications, Success: true}, http.StatusOK)
}

func (server *httpImpl) NewNotification(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {

		notification := sql.NotificationSQL{
			ID:           server.db.GetLastNotificationID(),
			Notification: r.FormValue("body"),
		}
		err = server.db.InsertNotification(notification)
		if err != nil {
			return
		}
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		atoi, err := strconv.Atoi(mux.Vars(r)["notification_id"])
		if err != nil {
			return
		}
		err = server.db.DeleteNotification(atoi)
		if err != nil {
			return
		}
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}
