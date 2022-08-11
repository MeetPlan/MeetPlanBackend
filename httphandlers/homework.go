package httphandlers

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

type Homework struct {
	sql.Homework
	Students []sql.StudentHomeworkJSON
}

type HomeworkJSON struct {
	sql.Homework
	TeacherName string
	SubjectName string
	Status      string
}

type HomeworkPerDate struct {
	Date     string
	Homework []HomeworkJSON
}

func (server *httpImpl) NewHomework(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
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
	subject, err := server.db.GetSubject(meeting.SubjectID)
	if err != nil {
		return
	}
	if user.Role == TEACHER && !(subject.TeacherID == user.ID || meeting.TeacherID == user.ID) {
		WriteForbiddenJWT(w)
		return
	}
	currentTime := time.Now()
	homework := sql.Homework{
		ID:          server.db.GetLastHomeworkID(),
		TeacherID:   user.ID,
		SubjectID:   meeting.SubjectID,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		ToDate:      r.FormValue("to_date"),
		FromDate:    currentTime.Format("2006-01-02"),
	}
	err = server.db.InsertHomework(homework)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusCreated)
}

func (server *httpImpl) GetAllHomeworksForSpecificSubject(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
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
	subject, err := server.db.GetSubject(meeting.SubjectID)
	if err != nil {
		return
	}
	if user.Role == TEACHER && !(subject.TeacherID == user.ID || meeting.TeacherID == user.ID) {
		WriteForbiddenJWT(w)
		return
	}
	homework, err := server.db.GetHomeworkForSubject(meeting.SubjectID)
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
	for i, j := 0, len(homeworkJson)-1; i < j; i, j = i+1, j-1 {
		homeworkJson[i], homeworkJson[j] = homeworkJson[j], homeworkJson[i]
	}
	WriteJSON(w, Response{Data: homeworkJson, Success: true}, http.StatusOK)
}

// GetHomeworkData TODO: Not used yet
func (server *httpImpl) GetHomeworkData(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
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
	if user.Role == TEACHER && homework.TeacherID != user.ID {
		WriteForbiddenJWT(w)
		return
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
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == TEACHER || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
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
	subject, err := server.db.GetSubject(homework.SubjectID)
	if err != nil {
		return
	}
	if user.Role == TEACHER && !(subject.TeacherID == user.ID || homework.TeacherID == user.ID) {
		WriteForbiddenJWT(w)
		return
	}
	h, err := server.db.GetStudentHomeworkForUser(homeworkId, userId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			h = sql.StudentHomework{
				ID:         server.db.GetLastStudentHomeworkID(),
				UserID:     userId,
				HomeworkID: homeworkId,
				Status:     r.FormValue("status"),
			}
			err := server.db.InsertStudentHomework(h)
			if err == nil {
				WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
			}
		}
		return
	}
	h.Status = r.FormValue("status")
	err = server.db.UpdateStudentHomework(h)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}

// GetUserHomework TODO: Restrict teachers from retrieving just anybody's homework
func (server *httpImpl) GetUserHomework(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	var studentId int
	if user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == SCHOOL_PSYCHOLOGIST || user.Role == TEACHER {
		studentId, err = strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
			return
		}
		if user.Role == PARENT {
			if !server.config.ParentViewHomework {
				WriteForbiddenJWT(w)
				return
			}
			var children []int
			err := json.Unmarshal([]byte(user.Users), &children)
			if err != nil {
				WriteJSON(w, Response{Data: "Failed while unmarshalling parent's students", Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			if !helpers.Contains(children, studentId) {
				WriteForbiddenJWT(w)
				return
			}
		} else if user.Role == TEACHER {
			classes, err := server.db.GetClasses()
			if err != nil {
				WriteJSON(w, Response{Data: "Failed while fetching classes", Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			var ok = false
			for i := 0; i < len(classes); i++ {
				class := classes[i]
				if class.Teacher != user.ID {
					continue
				}
				var students []int
				err := json.Unmarshal([]byte(class.Students), &students)
				if err != nil {
					WriteJSON(w, Response{Data: "Failed while unmarshalling", Error: err.Error(), Success: false}, http.StatusInternalServerError)
					return
				}
				if helpers.Contains(students, studentId) {
					ok = true
					break
				}
			}
			if !ok {
				WriteForbiddenJWT(w)
				return
			}
		}
	} else if user.Role == STUDENT {
		studentId = user.ID
	} else {
		WriteForbiddenJWT(w)
		return
	}

	subjects, err := server.db.GetAllSubjectsForUser(studentId)
	if err != nil {
		WriteJSON(w, Response{Success: false, Error: err.Error(), Data: "Could not fetch subjects"}, http.StatusInternalServerError)
		return
	}
	var homeworkJson = make([]HomeworkPerDate, 0)
	for i := 0; i < len(subjects); i++ {
		subject := subjects[i]
		homeworkForSubject, err := server.db.GetHomeworkForSubject(subject.ID)
		if err != nil {
			WriteJSON(w, Response{Success: false, Error: err.Error(), Data: "Could not fetch homework for subject"}, http.StatusInternalServerError)
			return
		}
		for n := 0; n < len(homeworkForSubject); n++ {
			homework := homeworkForSubject[n]
			date := homework.FromDate
			var contains = false
			var containsAt = -1
			for x := 0; x < len(homeworkJson); x++ {
				if homeworkJson[x].Date == date {
					contains = true
					containsAt = 0
					break
				}
			}
			teacher, err := server.db.GetUser(homework.TeacherID)
			if err != nil {
				WriteJSON(w, Response{Success: false, Error: err.Error(), Data: "Could not fetch teacher"}, http.StatusInternalServerError)
				return
			}
			subject, err := server.db.GetSubject(homework.SubjectID)
			if err != nil {
				WriteJSON(w, Response{Success: false, Error: err.Error(), Data: "Could not fetch subject"}, http.StatusInternalServerError)
				return
			}
			var status = ""
			homeworkStatus, err := server.db.GetStudentHomeworkForUser(homework.ID, studentId)
			if err != nil {
				if err.Error() != "sql: no rows in result set" {
					WriteJSON(w, Response{Data: "Could not fetch homework status", Error: err.Error(), Success: false}, http.StatusInternalServerError)
					return
				} else {
					status = "NOT MANAGED"
				}
			}
			if status == "" {
				status = homeworkStatus.Status
			}
			if !contains {
				var hw = make([]HomeworkJSON, 0)
				hw = append(hw, HomeworkJSON{
					Homework:    homework,
					TeacherName: teacher.Name,
					SubjectName: subject.Name,
					Status:      status,
				})
				homeworkJson = append(homeworkJson, HomeworkPerDate{
					Date:     date,
					Homework: hw,
				})
			} else {
				homeworkJson[containsAt].Homework = append(homeworkJson[containsAt].Homework, HomeworkJSON{
					Homework:    homework,
					TeacherName: teacher.Name,
					SubjectName: subject.Name,
					Status:      status,
				})
			}
		}
	}
	for i, j := 0, len(homeworkJson)-1; i < j; i, j = i+1, j-1 {
		homeworkJson[i], homeworkJson[j] = homeworkJson[j], homeworkJson[i]
	}
	WriteJSON(w, Response{Data: homeworkJson, Success: true}, http.StatusOK)
}
