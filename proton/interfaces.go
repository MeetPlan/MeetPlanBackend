package proton

type ProtonMeeting struct {
	Hour         int
	DayOfTheWeek int
	SubjectName  string
	SubjectID    string
	ID           string
	TeacherID    string
	Week         int
	ClassID      []string
	IsHalfHour   bool
	CommonID     string // CommonID is used to distinguish and connect meetings in different Weeks that should be displayed at the same time
	StackedHours bool
}
