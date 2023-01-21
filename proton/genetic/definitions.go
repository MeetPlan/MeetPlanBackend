package genetic

type Meeting struct {
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

type MeetingDefinition struct {
	ID           string
	ClassID      []string
	Duration     int
	IsHalfHour   bool
	StackedHours bool
	StudentIDs   []string
	Groups       []int
	SubjectID    string
	TeacherID    string
	ClassroomIDs []string
	Classrooms   []int
}

type Classroom struct {
	Name string
	Type string
}

type Data struct {
	Meetings   map[int]MeetingDefinition
	Groups     map[string]int
	Classrooms map[int]Classroom
}

type XY struct {
	X any
	Y any
}

type XYZ struct {
	X int
	Y int
	Z int
}
