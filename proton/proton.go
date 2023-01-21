package proton

import (
	"github.com/MeetPlan/MeetPlanBackend/proton/genetic"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"go.uber.org/zap"
)

type protonImpl struct {
	db                  sql.SQL
	config              genetic.ProtonConfig
	logger              *zap.SugaredLogger
	timetable           chan []ProtonMeeting
	percentageAssembled chan int
}

type Proton interface {
	ManageAbsences(meetingId string) ([]TeacherTier, error)

	NewProtonRule(rule genetic.ProtonRule) error
	GetProtonConfig() genetic.ProtonConfig

	// Suita funkcij, ki skrbijo za pravila

	GetAllRulesForTeacher(teacherId string) []genetic.ProtonRule
	GetSubjectGroups() []genetic.ProtonRule
	SubjectHasDoubleHours(subjectId string) bool
	CheckIfProtonConfigIsOk(timetable []ProtonMeeting, meeting ProtonMeeting, subjects []sql.Subject, subjectGroups []genetic.ProtonRule) (string, error)

	// FillGapsInTimetable(timetable []ProtonMeeting) ([]ProtonMeeting, error)

	// Post-procesirna suita funkcij

	//TimetablePostProcessing(stableTimetable []ProtonMeeting, class sql.Class, cancelPostProcessingBeforeDone bool) ([]ProtonMeeting, error)

	/*
		PatchTheHoles(timetable []ProtonMeeting, fullTimetable []ProtonMeeting) ([]ProtonMeeting, []ProtonMeeting)
		GetSubjectsOfClass(timetable []ProtonMeeting, classStudents []string, class sql.Class) ([]ProtonMeeting, error)
		GetSubjectsBeforeOrAfterClass() []string
		GetSubjectsWithStackedHours() []string
		FindNonNormalHours(timetable []ProtonMeeting) []ProtonMeeting
		PostProcessHolesAndNonNormalHours(classTimetable []ProtonMeeting, stableTimetable []ProtonMeeting) ([]ProtonMeeting, []ProtonMeeting)
		FindHoles(timetable []ProtonMeeting) [][]ProtonMeeting
		FindRelationalHoles(timetable []ProtonMeeting) []ProtonMeeting
		SwapMeetings(timetable []ProtonMeeting, fullTimetable []ProtonMeeting) ([]ProtonMeeting, []ProtonMeeting)
		PatchMistakes(timetable []ProtonMeeting, fullTimetable []ProtonMeeting) ([]ProtonMeeting, []ProtonMeeting)
	*/

	AssembleMeetingsFromProtonMeetings(timetable []ProtonMeeting, systemConfig sql.Config) ([]sql.Meeting, error)
	AssembleTimetableRecursive(hour int, day int, meetings []ProtonMeeting, timetable []ProtonMeeting, subjects []sql.Subject, subjectGroups []genetic.ProtonRule) ([]string, error)
	AssembleTimetable() ([]ProtonMeeting, error)

	FindIfHolesExist(timetable []ProtonMeeting) bool

	SaveConfig(config genetic.ProtonConfig)
	DeleteRule(ruleId string)
}

func NewProtonV2(db sql.SQL, logger *zap.SugaredLogger) (Proton, error) {
	protonConfig, err := genetic.LoadConfig()
	return &protonImpl{db: db, config: protonConfig, logger: logger, timetable: make(chan []ProtonMeeting), percentageAssembled: make(chan int)}, err
}
