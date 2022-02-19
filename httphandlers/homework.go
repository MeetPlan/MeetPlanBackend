package httphandlers

import (
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type Homework struct {
	sql.Homework
	Students []sql.StudentHomeworkJSON
}

func (server *httpImpl) NewHomework(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "student" {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		return
	}
	meetingId, err := strconv.Atoi(mux.Vars(r)["meeting_id"])
	if err != nil {
		return
	}
	meeting, err := server.db.GetMeeting(meetingId)
	if err != nil {
		return
	}
	if jwt["role"] == "teacher" {
		if meeting.TeacherID != userId {
			WriteForbiddenJWT(w)
			return
		}
	}
	homework := sql.Homework{
		ID:          server.db.GetLastHomeworkID(),
		TeacherID:   userId,
		SubjectID:   meeting.SubjectID,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		// TODO: We do not support this currently
		ToDate:   "",
		FromDate: "",
	}
	err = server.db.InsertHomework(homework)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusCreated)
}

func (server *httpImpl) GetAllHomeworksForSpecificSubject(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "student" {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		return
	}
	meetingId, err := strconv.Atoi(mux.Vars(r)["meeting_id"])
	if err != nil {
		return
	}
	if jwt["role"] == "teacher" {
		meeting, err := server.db.GetMeeting(meetingId)
		if err != nil {
			return
		}
		if meeting.TeacherID != userId {
			WriteForbiddenJWT(w)
			return
		}
	}
	homework, err := server.db.GetHomeworkForSubject(meetingId)
	if err != nil {
		return
	}
	var homeworkJson = make([]Homework, 0)
	for i := 0; i < len(homework); i++ {
		h, err := server.db.GetStudentsHomeworkByHomeworkID(homework[i].ID, meetingId)
		if err != nil {
			return
		}
		homeworkJson = append(homeworkJson, Homework{
			Homework: homework[i],
			Students: h,
		})
	}
	WriteJSON(w, Response{Data: homeworkJson, Success: true}, http.StatusOK)
}

// GetHomeworkData TODO: Not used yet
func (server *httpImpl) GetHomeworkData(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "student" {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		return
	}
	homeworkId, err := strconv.Atoi(mux.Vars(r)["homework_id"])
	if err != nil {
		return
	}
	homework, err := server.db.GetHomework(homeworkId)
	if err != nil {
		return
	}
	if jwt["role"] == "teacher" {
		if homework.TeacherID != userId {
			WriteForbiddenJWT(w)
			return
		}
	}
	h, err := server.db.GetStudentsHomeworkByHomeworkID(homeworkId, -1)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Data: Homework{
		Homework: homework,
		Students: h,
	}, Success: true}, http.StatusOK)
}

func (server *httpImpl) PatchHomeworkForStudent(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "student" {
		WriteForbiddenJWT(w)
		return
	}
	teacherId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		return
	}
	// Maybe we will use it sometime, you never know
	_, err = strconv.Atoi(mux.Vars(r)["meeting_id"])
	if err != nil {
		return
	}
	homeworkId, err := strconv.Atoi(mux.Vars(r)["homework_id"])
	if err != nil {
		return
	}
	userId, err := strconv.Atoi(mux.Vars(r)["student_id"])
	if err != nil {
		return
	}
	homework, err := server.db.GetHomework(homeworkId)
	if err != nil {
		return
	}
	if jwt["role"] == "teacher" {
		if homework.TeacherID != teacherId {
			WriteForbiddenJWT(w)
			return
		}
	}
	h, err := server.db.GetStudentHomeworkForUser(homeworkId, userId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			h = sql.StudentHomework{
				ID:         server.db.GetLastStudentHomeworkID(),
				UserID:     userId,
				HomeworkID: homeworkId,
				Status:     "",
			}
			err = server.db.InsertStudentHomework(h)
			if err != nil {
				return
			}
		} else {
			return
		}
	}
	h.Status = r.FormValue("status")
	err = server.db.UpdateStudentHomework(h)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}
