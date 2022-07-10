/// This file is a part of MeetPlan Proton, which is a part of MeetPlanBackend (https://github.com/MeetPlan/MeetPlanBackend).
///
/// Copyright (c) 2022, Mitja Ševerkar <mytja@protonmail.com> and The MeetPlan Team.
/// All rights reserved.
/// Use of this source code is governed by the GNU AGPLv3 license, that can be found in the LICENSE file.

/// POZOR!
/// Ta package vsebuje kar nekaj matematike in nerazumljive kode.
/// Se ne priporoča brati, če ni komentarjev, saj bo tako najbolje za vaše mentalno zdravje.
/// Avtor ni odgovoren za kakršnekoli (materialne, fizične, mentalne ipd.) poškodbe med interakcijo s to kodo.

package proton

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"go.uber.org/zap"
	"math/rand"
)

const PROTON_ALLOWED_WHILE_DEPTH = 1000
const PROTON_ALLOWED_FAIL_RATE = 500
const PROTON_ALLOWED_FAIL_RESET_RATE = 200
const PROTON_ALLOWED_HOLE_PATCHING_REPEAT_RATE = 200

type protonImpl struct {
	db     sql.SQL
	config ProtonConfig
	logger *zap.SugaredLogger
}

type Proton interface {
	ManageAbsences(meetingId int) ([]TeacherTier, error)

	NewProtonRule(rule ProtonRule) error
	GetProtonConfig() ProtonConfig

	GetAllRulesForTeacher(teacherId int) []ProtonRule
	CheckIfProtonConfigIsOk(timetable []ProtonMeeting) (bool, error)
	GetSubjectGroups() []ProtonRule
	SubjectHasDoubleHours(subjectId int) bool
	FindNonNormalHours(timetable []ProtonMeeting) []ProtonMeeting
	FindHoles(timetable []ProtonMeeting) []ProtonMeeting
	FillGapsInTimetable(timetable []ProtonMeeting) ([]ProtonMeeting, error)
	PatchTheHoles(timetable []ProtonMeeting, fullTimetable []ProtonMeeting) ([]ProtonMeeting, []ProtonMeeting)
	GetSubjectsOfClass(timetable []ProtonMeeting, classStudents []int, class sql.Class) ([]ProtonMeeting, error)
	GetSubjectsBeforeOrAfterClass() []int
	GetSubjectsWithStackedHours() []int
}

func NewProton(db sql.SQL, logger *zap.SugaredLogger) (Proton, error) {
	protonConfig, err := LoadConfig()
	return &protonImpl{db: db, config: protonConfig, logger: logger}, err
}

type TierGradingList struct {
	TeacherID          int
	HasMeetingBefore   bool
	HasMeetingLater    bool
	HasMeeting2HBefore bool
	HasMeeting2HLater  bool
	TeachesSameSubject bool
	Name               string
}

type TeacherTier struct {
	TeacherID   int
	Tier        int
	Name        string
	GradingList TierGradingList
}

func (p *protonImpl) ManageAbsences(meetingId int) ([]TeacherTier, error) {
	teachers, err := p.db.GetTeachers()
	if err != nil {
		return make([]TeacherTier, 0), err
	}
	originalMeeting, err := p.db.GetMeeting(meetingId)
	if err != nil {
		return make([]TeacherTier, 0), err
	}
	subject, err := p.db.GetSubject(originalMeeting.SubjectID)
	if err != nil {
		return make([]TeacherTier, 0), err
	}
	similarSubjects, err := p.db.GetSubjectsWithSpecificLongName(subject.LongName)
	if err != nil {
		return make([]TeacherTier, 0), err
	}
	var preferredTeachers = make([]int, 0)
	for i := 0; i < len(similarSubjects); i++ {
		subject := similarSubjects[i]
		if !contains(preferredTeachers, subject.TeacherID) {
			preferredTeachers = append(preferredTeachers, subject.TeacherID)
		}
	}
	var teacherTiers = make([]TierGradingList, 0)
	for i := 0; i < len(teachers); i++ {
		teacher := teachers[i]
		teacherMeetings, err := p.db.GetMeetingsForTeacherOnSpecificDate(teacher.ID, originalMeeting.Date)
		if err != nil {
			return make([]TeacherTier, 0), err
		}
		var teacherTier = TierGradingList{
			TeacherID:          teacher.ID,
			HasMeetingBefore:   false,
			HasMeetingLater:    false,
			TeachesSameSubject: false,
			Name:               teacher.Name,
		}
		var hasSameHour = false

		// This should not be impacted by when the teacher has meetings or not
		if contains(preferredTeachers, teacher.ID) {
			teacherTier.TeachesSameSubject = true
		}

		for n := 0; n < len(teacherMeetings); n++ {
			meeting := teacherMeetings[n]
			if meeting.Hour+1 == originalMeeting.Hour {
				teacherTier.HasMeetingBefore = true
			} else if meeting.Hour-1 == originalMeeting.Hour {
				teacherTier.HasMeetingLater = true
			} else if meeting.Hour == originalMeeting.Hour {
				hasSameHour = true
				break
			} else if meeting.Hour-2 == originalMeeting.Hour {
				teacherTier.HasMeeting2HLater = true
			} else if meeting.Hour+2 == originalMeeting.Hour {
				teacherTier.HasMeeting2HBefore = true
			}
		}
		if !hasSameHour {
			teacherTiers = append(teacherTiers, teacherTier)
		}
	}
	var recommendation = make([]TeacherTier, 0)
	for i := 0; i < len(teacherTiers); i++ {
		var tierGrade = 0
		teacherTier := teacherTiers[i]
		if teacherTier.TeachesSameSubject {
			tierGrade += 5
		}
		if teacherTier.HasMeeting2HLater {
			tierGrade += 1
		}
		if teacherTier.HasMeeting2HBefore {
			tierGrade += 1
		}
		if teacherTier.HasMeetingLater {
			tierGrade += 3
		}
		if teacherTier.HasMeetingBefore {
			tierGrade += 3
		}

		var skip = true

		for n := 0; n < len(recommendation); n++ {
			r := recommendation[n]
			if tierGrade > r.Tier {
				recommendation = insertTeacherTier(recommendation, n, TeacherTier{
					TeacherID:   teacherTier.TeacherID,
					Tier:        tierGrade,
					Name:        teacherTier.Name,
					GradingList: teacherTier,
				})
				skip = false
				break
			}
		}

		if skip {
			recommendation = append(recommendation, TeacherTier{
				TeacherID:   teacherTier.TeacherID,
				Tier:        tierGrade,
				Name:        teacherTier.Name,
				GradingList: teacherTier,
			})
		}
	}
	return recommendation, err
}

func (p *protonImpl) GetAllRulesForTeacher(teacherId int) []ProtonRule {
	protonRules := make([]ProtonRule, 0)

	for i := 0; i < len(p.config.Rules); i++ {
		rule := p.config.Rules[i]
		for n := 0; n < len(rule.Objects); n++ {
			object := rule.Objects[n]
			if object.Type == "teacher" && object.ObjectID == teacherId {
				protonRules = append(protonRules, rule)
				break
			}
		}
	}

	return protonRules
}

func (p *protonImpl) GetSubjectGroups() []ProtonRule {
	protonRules := make([]ProtonRule, 0)
	for i := 0; i < len(p.config.Rules); i++ {
		protonRule := p.config.Rules[i]
		if protonRule.RuleType == 2 {
			protonRules = append(protonRules, protonRule)
		}
	}
	return protonRules
}

// SubjectHasDoubleHours preverja, če ima predmet blok ure.
func (p *protonImpl) SubjectHasDoubleHours(subjectId int) bool {
	for i := 0; i < len(p.config.Rules); i++ {
		rule := p.config.Rules[i]
		if rule.RuleType == 4 {
			for n := 0; n < len(rule.Objects); n++ {
				object := rule.Objects[n]
				if object.Type == "subject" && object.ObjectID == subjectId {
					return true
				}
			}
		}
	}
	return false
}

func (p *protonImpl) FillGapsInTimetable(timetable []ProtonMeeting) ([]ProtonMeeting, error) {
	protonMeetings := make([]ProtonMeeting, 0)
	protonMeetings = append(protonMeetings, timetable...)

	// Pojdimo čez vseh 5 šolskih dni (pon, tor, sre, čet in pet) in preverimo za luknje pri vsakemu razredu.
	// Šolarji imajo generalno iste/podobne luknje, tako da ni potrebe po sestavljanju urnika za čisto vsakega učenca.
	// Izbirni predmeti ipd. so tako ali tako predure ali po koncu generalnega pouka.

	classes, err := p.db.GetClasses()
	if err != nil {
		return nil, err
	}

	subjectGroups := p.GetSubjectGroups()

	for i := 0; i < len(classes); i++ {
		depth := 0

		class := classes[i]

		var classStudents []int
		err := json.Unmarshal([]byte(class.Students), &classStudents)
		if err != nil {
			return nil, err
		}

		constantHoleLen := 0
		holeSame := 0

		for {
			timetable = make([]ProtonMeeting, 0)
			timetable = append(timetable, protonMeetings...)

			if depth >= (PROTON_ALLOWED_WHILE_DEPTH / 40) {
				p.logger.Debug("exiting due to exceeded allowed depth")
				return nil, errors.New("exceeded maximum allowed repeat depth")
			}

			var classTimetable = make([]ProtonMeeting, 0)

			// 1. del
			// Pridobimo vsa srečanja, ki so povezana z določenim razredom.
			t, err := p.GetSubjectsOfClass(timetable, classStudents, class)
			if err != nil {
				return nil, err
			}

			// 2. del
			// Zapolnimo luknje v urniku (z drugimi "bingljajočimi" urami) in ves čas preverjamo, če je vse v redu.
			nonNormal := p.FindNonNormalHours(classTimetable)
			holes := p.FindHoles(classTimetable)

			if len(holes) == 0 || len(nonNormal) == 0 {
				classTimetable, protonMeetings = p.PatchTheHoles(t, timetable)

				break
			}

			if len(holes) == constantHoleLen {
				holeSame++
			} else {
				p.logger.Debug("reset hole repeat counter")
				holeSame = 0
			}

			constantHoleLen = len(holes)

			if holeSame > 10 {
				break
			}

			h := rand.Intn(len(holes))
			n := rand.Intn(len(nonNormal))

			hole := holes[h]
			meeting := nonNormal[n]

			var subjectGroup = make([]int, 0)

			// TODO: Migriraj blok ure
			for x := 0; x < len(subjectGroups); x++ {
				group := subjectGroups[x]

				var ok = false

				for y := 0; y < len(group.Objects); y++ {
					object := group.Objects[y]
					if object.Type == "subject" && object.ObjectID == meeting.SubjectID {
						ok = true
					}
				}

				if !ok {
					continue
				}
				for y := 0; y < len(group.Objects); y++ {
					object := group.Objects[y]
					if object.Type == "subject" && !contains(subjectGroup, object.ObjectID) {
						subjectGroup = append(subjectGroup, object.ObjectID)
					}
				}
			}

			if len(subjectGroup) == 0 {
				subjectGroup = append(subjectGroup, meeting.SubjectID)
			}

			for y := 0; y < len(nonNormal); y++ {
				m := nonNormal[y]
				if m.Hour == meeting.Hour && m.DayOfTheWeek == meeting.DayOfTheWeek && contains(subjectGroup, m.SubjectID) {
					for x := 0; x < len(timetable); x++ {
						meeting := timetable[x]
						if meeting.ID == m.ID {
							timetable = remove(timetable, x)
						}
					}

					m.Hour = hole.Hour
					m.DayOfTheWeek = hole.DayOfTheWeek
					timetable = append(timetable, m)
				}
			}

			for x := 0; x < len(timetable); x++ {
				m := timetable[x]
				if m.ID == meeting.ID {
					timetable = remove(timetable, x)
				}
			}

			meeting.Hour = hole.Hour
			meeting.DayOfTheWeek = hole.DayOfTheWeek
			timetable = append(timetable, meeting)

			ok, err := p.CheckIfProtonConfigIsOk(timetable)
			if ok {
				p.logger.Debugw("successfully normalized proton timetable", "timetable", timetable, "protonMeetings", protonMeetings, "nonNormal", nonNormal, "classTimetable", classTimetable)

				protonMeetings = make([]ProtonMeeting, 0)
				protonMeetings = append(protonMeetings, timetable...)
			} else {
				p.logger.Debugw("failed to normalize proton-generated timetable", "error", err.Error(), "timetable", timetable, "protonTimetable", protonMeetings, "meeting", meeting)
			}

			depth++
		}
	}

	p.logger.Debug("successfully normalized the timetable")
	return protonMeetings, nil
}

func (p *protonImpl) GetSubjectsOfClass(timetable []ProtonMeeting, classStudents []int, class sql.Class) ([]ProtonMeeting, error) {
	var classTimetable = make([]ProtonMeeting, 0)

	for n := 0; n < len(timetable); n++ {
		meeting := timetable[n]
		subject, err := p.db.GetSubject(meeting.SubjectID)
		if err != nil {
			return nil, err
		}
		if subject.InheritsClass && class.ID == subject.ClassID {
			classTimetable = append(classTimetable, meeting)
		} else {
			var students []int
			err := json.Unmarshal([]byte(subject.Students), &students)
			if err != nil {
				return nil, err
			}
			var ok = false
			for x := 0; x < len(students); x++ {
				subjectStudent := students[x]
				if contains(classStudents, subjectStudent) {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}

			classTimetable = append(classTimetable, meeting)
		}
	}

	return classTimetable, nil
}

// GetSubjectsBeforeOrAfterClass retrieves all subjects, that are before or after the class (according to rule #3)
func (p *protonImpl) GetSubjectsBeforeOrAfterClass() []int {
	var subjects = make([]int, 0)
	rules := p.config.Rules
	for i := 0; i < len(rules); i++ {
		rule := rules[i]
		if rule.RuleType == 3 {
			for n := 0; n < len(rule.Objects); n++ {
				object := rule.Objects[n]
				if object.Type == "subject" && !contains(subjects, object.ObjectID) {
					subjects = append(subjects, object.ObjectID)
				}
			}
		}
	}
	return subjects
}

// GetSubjectsWithStackedHours retrieves all subjects, that have stacked hours (according to rule #4)
func (p *protonImpl) GetSubjectsWithStackedHours() []int {
	var subjects = make([]int, 0)
	rules := p.config.Rules
	for i := 0; i < len(rules); i++ {
		rule := rules[i]
		if rule.RuleType == 4 {
			for n := 0; n < len(rule.Objects); n++ {
				object := rule.Objects[n]
				if object.Type == "subject" && !contains(subjects, object.ObjectID) {
					subjects = append(subjects, object.ObjectID)
				}
			}
		}
	}
	return subjects
}

func (p *protonImpl) PatchTheHoles(timetable []ProtonMeeting, fullTimetable []ProtonMeeting) ([]ProtonMeeting, []ProtonMeeting) {
	// Go naredi neke čudne stvari, podobno kot Python. Pri Pythonu se, če bi shranil tole v nov variable (primer `a := b`)
	// bi se ustvaril pointer in bi bil vse samo en seznam. Tukaj je podobno, zato moramo ustvariti popolnoma nov seznam.
	stableFullTimetable := make([]ProtonMeeting, 0)
	stableFullTimetable = append(stableFullTimetable, fullTimetable...)

	var days = make([][]ProtonMeeting, 5)

	var k = make([][]int, 5)
	for i := 0; i < len(timetable); i++ {
		meeting := timetable[i]
		days[meeting.DayOfTheWeek] = append(days[meeting.DayOfTheWeek], meeting)
		if !contains(k[meeting.DayOfTheWeek], meeting.Hour) {
			k[meeting.DayOfTheWeek] = append(k[meeting.DayOfTheWeek], meeting.Hour)
		}
	}

	count := 0

	for {
		// We have to recalculate holes to make sure, we don't have holes that cover each other.
		if count > PROTON_ALLOWED_HOLE_PATCHING_REPEAT_RATE {
			p.logger.Debugw("aborting hole patching. this is a completely normal behaviour.", "days", days, "stableFullTimetable", stableFullTimetable)
			break
		}

		holes := p.FindHoles(timetable)

		if len(holes) == 0 {
			break
		}

		h := rand.Intn(len(holes))
		hole := holes[h]
		day := days[hole.DayOfTheWeek]

		var dayBackup = make([]ProtonMeeting, 0)
		dayBackup = append(dayBackup, day...)

		for n := 0; n < len(day); n++ {
			if day[n].Hour > hole.Hour {
				if hole.DayOfTheWeek == 4 {
					p.logger.Debugw("patching a hole", "day", day, "hole", hole, "meeting", day[n])
				}
				for x := 0; x < len(fullTimetable); x++ {
					m := fullTimetable[x]
					if m.ID == day[n].ID {
						fullTimetable = remove(fullTimetable, x)
						break
					}
				}
				days[hole.DayOfTheWeek][n].Hour--
				fullTimetable = append(fullTimetable, days[hole.DayOfTheWeek][n])

				ok, _ := p.CheckIfProtonConfigIsOk(fullTimetable)
				if ok {
					stableFullTimetable = make([]ProtonMeeting, 0)
					stableFullTimetable = append(stableFullTimetable, fullTimetable...)
				} else {
					days[hole.DayOfTheWeek][n].Hour++
					fullTimetable = make([]ProtonMeeting, 0)
					fullTimetable = append(fullTimetable, stableFullTimetable...)
				}
			}
		}

		ok, err := p.CheckIfProtonConfigIsOk(fullTimetable)
		if ok {
			p.logger.Debugw("successfully patched a hole", "hole", hole, "day", day)
			stableFullTimetable = make([]ProtonMeeting, 0)
			stableFullTimetable = append(stableFullTimetable, fullTimetable...)
		} else {
			p.logger.Debugw("failed to check proton config while patching holes", "err", err.Error(), "day", day, "hole", hole)

			fullTimetable = make([]ProtonMeeting, 0)
			fullTimetable = append(fullTimetable, stableFullTimetable...)

			days[hole.DayOfTheWeek] = dayBackup
		}

		count++
	}

	t := make([]ProtonMeeting, 0)
	for i := 0; i < len(days); i++ {
		t = append(t, days[i]...)
	}

	return t, stableFullTimetable
}

// FindHoles bo poiskal vse luknje vmes in z drugo metodo iz FindNonNormalHours poiskal vse nenormalne prazne ure.
func (p *protonImpl) FindHoles(timetable []ProtonMeeting) []ProtonMeeting {
	// Ne me vprašat zakaj sem to naredil.
	var status = make(map[int]map[int][]ProtonMeeting)

	beforeAfterSubjects := p.GetSubjectsBeforeOrAfterClass()

	for i := 0; i < len(timetable); i++ {
		meeting := timetable[i]
		if status[meeting.DayOfTheWeek] == nil {
			status[meeting.DayOfTheWeek] = make(map[int][]ProtonMeeting)
		}
		if status[meeting.DayOfTheWeek][meeting.Hour] == nil {
			status[meeting.DayOfTheWeek][meeting.Hour] = make([]ProtonMeeting, 0)
		}
		status[meeting.DayOfTheWeek][meeting.Hour] = append(status[meeting.DayOfTheWeek][meeting.Hour], meeting)
	}

	minHours := -1
	maxHours := -1

	for _, v := range status {
		// Pojdimo čez vsak dan
		if v == nil {
			continue
		}
		dayHours := 0
		for _, k := range v {
			// Pojdimo čez vsako uro
			if k == nil || len(k) == 0 {
				continue
			}

			// izločimo predure in poure iz seta
			var ok = true
			for i := 0; i < len(k); i++ {
				if !contains(beforeAfterSubjects, k[i].SubjectID) {
					ok = false
					break
				}
			}

			// če so vsi predmeti na to uro izbirni, preskočimo vse
			if ok {
				continue
			}

			dayHours++
		}
		if dayHours < minHours || minHours == -1 {
			minHours = dayHours
		}
		if dayHours > maxHours {
			maxHours = dayHours
		}
	}

	freeHours := make([]ProtonMeeting, 0)

	average := float32(minHours+maxHours) / 2.0

	// Iterate over each day
	for day := 0; day < 5; day++ {
		if status[day] == nil {
			// Samo preskoči ta dan
			continue
		}
		for hour := 1; hour <= 12; hour++ {
			if status[day][hour] != nil {
				continue
			}
			if float32(hour) < average {
				freeHours = append(freeHours, ProtonMeeting{Hour: hour, DayOfTheWeek: day})
				continue
			}
			for n := hour + 1; n <= 12; n++ {
				if status[day][n] != nil {
					freeHours = append(freeHours, ProtonMeeting{Hour: hour, DayOfTheWeek: day})
					break
				}
			}
		}
	}

	p.logger.Debugw("found holes", "holes", freeHours, "maxHours", maxHours, "minHours", minHours, "average", average)

	return freeHours
}

// FindNonNormalHours poišče vse nenormalne ("bingljajoče") ure z dvema metodama:
//
// 1. metoda – najdi vse bingljajoče ure z izračunom povprečnega števila ur vsak dan skupaj. Ta metoda deluje večino časa, a pri ekstremih, ko je en dan praktično prazen, na žalost ne.
//
// 2. metoda – najdi vse bingljajoče ure z izračunom povprečja med dnevom z največ šolskimi urami in dnevom z najmanj šolskimi urami. To deluje, kadar 1. metoda ne deluje.
//
// Primer:
//
// Ponedeljek ima samo štiri (navadne – tj. niso izbirni predmeti) šolske ure,
// torek pa ima sedem (navadnih) šolskih ur.
// V tem primeru bo poskušal poiskati ustrezno rešitev, tako, da bo poiskal vse bingljajoče ure – v našem primeru šesto in sedmo šolsko uro v torku.
func (p *protonImpl) FindNonNormalHours(timetable []ProtonMeeting) []ProtonMeeting {
	beforeAfterSubjects := p.GetSubjectsBeforeOrAfterClass()
	days := make(map[int][]int)
	for i := 0; i < len(timetable); i++ {
		meeting := timetable[i]
		if contains(beforeAfterSubjects, meeting.SubjectID) {
			continue
		}
		if days[meeting.DayOfTheWeek] == nil {
			days[meeting.DayOfTheWeek] = make([]int, 0)
		}
		if !contains(days[meeting.DayOfTheWeek], meeting.Hour) {
			days[meeting.DayOfTheWeek] = append(days[meeting.DayOfTheWeek], meeting.Hour)
		}
	}
	var totalHours = 0
	var totalDays = 0

	minHours := -1
	maxHours := -1

	for _, v := range days {
		if len(v) < minHours || minHours == -1 {
			minHours = len(v)
		}
		if len(v) > maxHours {
			maxHours = len(v)
		}
		totalHours += len(v)
		totalDays++
	}

	var meetings = make([]ProtonMeeting, 0)

	if minHours == -1 || maxHours == -1 {
		return meetings
	}

	average := float32((minHours + maxHours) / 2)

	if totalDays == 0 {
		return meetings
	}

	relation := float32(totalHours / totalDays)

	for i := 0; i < len(timetable); i++ {
		meeting := timetable[i]
		if contains(beforeAfterSubjects, meeting.SubjectID) {
			continue
		}
		if float32(meeting.Hour) > (relation-1) || float32(meeting.Hour) > average {
			meetings = append(meetings, meeting)
		}
	}

	p.logger.Debugw("found non-normal hours", "relation", relation, "totalHours", totalHours, "meetings", meetings, "totalDays", totalDays)

	return meetings
}

type ProtonMeeting struct {
	Hour         int
	DayOfTheWeek int
	SubjectName  string
	SubjectID    int
	ID           string
	TeacherID    int
	Week         int
	ClassID      []int
}

// CheckIfProtonConfigIsOk preverja, če je trenuten timetable v redu sestavljen
// Ta funkcija je temelj vsega našega sistema.
func (p *protonImpl) CheckIfProtonConfigIsOk(timetable []ProtonMeeting) (bool, error) {
	// Predpriprava
	subjectGroups := p.GetSubjectGroups()

	// 1. korak
	// Pojdimo čez vse učitelje in preverimo, da se nič ne prekriva in je urnik skladen z učiteljevimi urami.
	teachers, err := p.db.GetTeachers()
	if err != nil {
		return false, err
	}
	for i := 0; i < len(teachers); i++ {
		teacher := teachers[i]
		for t := 0; t < len(timetable); t++ {
			// Pojdimo čez vse ure in preverimo, če se kaka ujema z učiteljem. Če se, nadaljujemo s postopkom.
			meeting := timetable[t]

			subject1, err := p.db.GetSubject(meeting.SubjectID)
			if err != nil {
				return false, err
			}

			if meeting.TeacherID != teacher.ID {
				continue
			}

			var subjectsInGroup = make([]int, 0)

			var hoursToday = 1

			// Preverimo, če se kake ure prekrivajo.
			for n := 0; n < len(timetable); n++ {
				meeting2 := timetable[n]

				if meeting2.ID == meeting.ID {
					continue
				}

				if meeting2.DayOfTheWeek == meeting.DayOfTheWeek && meeting.SubjectID == meeting2.SubjectID && meeting.Week == meeting2.Week {
					hoursToday++
				}

				// Seveda moramo preveriti, če sta srečanji v isti proton skupini srečanj.
				var ok1 = false
				var ok2 = false
				for x := 0; x < len(subjectGroups); x++ {
					group := subjectGroups[x]
					for y := 0; y < len(group.Objects); y++ {
						if group.Objects[y].Type == "subject" && group.Objects[y].ObjectID == meeting.SubjectID {
							ok1 = true
							break
						}
					}
					if ok1 {
						for y := 0; y < len(group.Objects); y++ {
							if group.Objects[y].Type == "subject" {
								if meeting2.SubjectID == group.Objects[y].ObjectID {
									ok2 = true
								}
								if !contains(subjectsInGroup, group.Objects[y].ObjectID) {
									subjectsInGroup = append(subjectsInGroup, group.Objects[y].ObjectID)
								}
							}
						}
					}
				}

				if meeting2.DayOfTheWeek == meeting.DayOfTheWeek && meeting2.Hour == meeting.Hour && meeting.Week == meeting2.Week {
					// Preverimo, da se učencem ne prekriva

					subject2, err := p.db.GetSubject(meeting2.SubjectID)
					if err != nil {
						return false, err
					}
					if subject1.TeacherID == subject2.TeacherID {
						return false, errors.New("učitelj ne more učiti dveh predmetov ob istem času")
					}

					if subject1.InheritsClass && subject2.InheritsClass {
						if subject2.ClassID == subject1.ClassID {
							return false, errors.New(fmt.Sprintf("subjects %s and %s inherit the same class and thus cannot be made at same time", fmt.Sprint(subject1.ID), fmt.Sprint(subject2.ID)))
						}
						// V tem primeru nista isti razred, posledično se ne prekrivata
						continue
					}

					var students1 []int
					if subject1.InheritsClass {
						class, err := p.db.GetClass(subject1.ClassID)
						if err != nil {
							return false, err
						}
						err = json.Unmarshal([]byte(class.Students), &students1)
						if err != nil {
							return false, err
						}
					} else {
						err = json.Unmarshal([]byte(subject1.Students), &students1)
						if err != nil {
							return false, err
						}
					}

					var students2 []int
					if subject2.InheritsClass {
						class, err := p.db.GetClass(subject2.ClassID)
						if err != nil {
							return false, err
						}
						err = json.Unmarshal([]byte(class.Students), &students2)
						if err != nil {
							return false, err
						}
					} else {
						err = json.Unmarshal([]byte(subject2.Students), &students2)
						if err != nil {
							return false, err
						}
					}

					for s := 0; s < len(students1); s++ {
						student := students1[s]
						if contains(students2, student) {
							return false, errors.New(fmt.Sprintf("subjects %s and %s contain the same student %s and thus cannot be made at same time", fmt.Sprint(subject1.ID), fmt.Sprint(subject2.ID), fmt.Sprint(student)))
						}
					}
				}

				if !(ok1 && ok2) && (meeting2.DayOfTheWeek == meeting.DayOfTheWeek && meeting2.Hour == meeting.Hour && meeting.Week == meeting2.Week) {
					return false, errors.New(
						fmt.Sprintf(
							"srečanji %s (%s %s) in pa %s (%s %s) se prekrivata - ne morem ustvariti urnika.",
							fmt.Sprint(meeting),
							fmt.Sprint(meeting.Hour),
							fmt.Sprint(meeting.DayOfTheWeek),
							fmt.Sprint(meeting2),
							fmt.Sprint(meeting2.Hour),
							fmt.Sprint(meeting2.DayOfTheWeek),
						),
					)
				}
			}

			if hoursToday > 2 {
				return false, errors.New("ne moreta biti več kot dve uri istega predmeta na en dan")
			}

			if len(subjectsInGroup) == 0 {
				subjectsInGroup = append(subjectsInGroup, meeting.SubjectID)
			}

			//p.logger.Debugw("subjects in group", "group", subjectsInGroup, "subjectGroups", subjectGroups, "meeting", meeting)

			mmap := make(map[int]bool)

			for s := 0; s < len(subjectsInGroup); s++ {
				subject, err := p.db.GetSubject(subjectsInGroup[s])
				if err != nil {
					return false, errors.New(fmt.Sprintf("subject %s is not found in the database - %s", fmt.Sprint(subjectsInGroup[s]), err.Error()))
				}
				rules := p.GetAllRulesForTeacher(subject.TeacherID)
				mmap[subjectsInGroup[s]] = false
				if len(rules) == 0 {
					mmap[subjectsInGroup[s]] = true
				}
				for r := 0; r < len(rules); r++ {
					rule := rules[r]
					if rule.RuleType == 0 {
						// Polni dnevi učitelja na šoli
						for n := 0; n < len(rule.Objects); n++ {
							object := rule.Objects[n]
							if object.Type == "day" && object.ObjectID == meeting.DayOfTheWeek {
								mmap[subjectsInGroup[s]] = true
								break
							}
						}
					} else if rule.RuleType == 1 {
						// Ure učitelja na šoli

						// Dan je treba izvleči posebej
						// TODO: Seznam dni namesto integerja
						day := -1
						for k := 0; k < len(rule.Objects); k++ {
							object := rule.Objects[k]
							if object.Type == "day" {
								day = object.ObjectID
							}
						}
						if day == -1 {
							return false, errors.New("neveljavno pravilo brez dni - pravilo št. 1")
						}

						if meeting.DayOfTheWeek != day {
							continue
						}

						for n := 0; n < len(rule.Objects); n++ {
							object := rule.Objects[n]

							if object.Type == "hour" && object.ObjectID == meeting.Hour {
								mmap[subjectsInGroup[s]] = true
								break
							}
						}
					}
				}
			}
			//fmt.Println(mmap, meeting)
			for n, v := range mmap {
				if !v {
					return false, errors.New(fmt.Sprintf("srečanje s predmetom %s se ne ujema s tem, kdaj je učitelj na šoli", fmt.Sprint(n)))
				}
			}
		}
	}
	return true, nil
}

func (p *protonImpl) NewProtonRule(rule ProtonRule) error {
	config, err := AddNewRule(p.config, rule)
	if err != nil {
		return err
	}
	p.config = config
	return nil
}

func (p *protonImpl) GetProtonConfig() ProtonConfig {
	return p.config
}
