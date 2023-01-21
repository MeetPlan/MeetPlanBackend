package proton

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/MeetPlan/MeetPlanBackend/proton/genetic"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"time"
)

func (p *protonImpl) GetAllRulesForTeacher(teacherId string) []genetic.ProtonRule {
	protonRules := make([]genetic.ProtonRule, 0)

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

func (p *protonImpl) GetSubjectGroups() []genetic.ProtonRule {
	protonRules := make([]genetic.ProtonRule, 0)
	for i := 0; i < len(p.config.Rules); i++ {
		protonRule := p.config.Rules[i]
		if protonRule.RuleType == 2 {
			protonRules = append(protonRules, protonRule)
		}
	}
	return protonRules
}

// SubjectHasDoubleHours preverja, če ima predmet blok ure.
func (p *protonImpl) SubjectHasDoubleHours(subjectId string) bool {
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

func (p *protonImpl) GetSubjectsOfClass(timetable []ProtonMeeting, classStudents []string, class sql.Class) ([]ProtonMeeting, error) {
	var classTimetable = make([]ProtonMeeting, 0)

	for n := 0; n < len(timetable); n++ {
		meeting := timetable[n]
		subject, err := p.db.GetSubject(meeting.SubjectID)
		if err != nil {
			return nil, err
		}
		if subject.InheritsClass && class.ID == *subject.ClassID {
			classTimetable = append(classTimetable, meeting)
		} else {
			var students []string
			err := json.Unmarshal([]byte(subject.Students), &students)
			if err != nil {
				return nil, err
			}
			var ok = false
			for x := 0; x < len(students); x++ {
				subjectStudent := students[x]
				if helpers.Contains(classStudents, subjectStudent) {
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
func (p *protonImpl) GetSubjectsBeforeOrAfterClass() []string {
	var subjects = make([]string, 0)
	rules := p.config.Rules
	for i := 0; i < len(rules); i++ {
		rule := rules[i]
		if rule.RuleType == 3 {
			for n := 0; n < len(rule.Objects); n++ {
				object := rule.Objects[n]
				if object.Type == "subject" && !helpers.Contains(subjects, object.ObjectID) {
					subjects = append(subjects, object.ObjectID)
				}
			}
		}
	}
	return subjects
}

// GetSubjectsWithStackedHours retrieves all subjects, that have stacked hours (according to rule #4)
func (p *protonImpl) GetSubjectsWithStackedHours() []string {
	var subjects = make([]string, 0)
	rules := p.config.Rules
	for i := 0; i < len(rules); i++ {
		rule := rules[i]
		if rule.RuleType == 4 {
			for n := 0; n < len(rule.Objects); n++ {
				object := rule.Objects[n]
				if object.Type == "subject" && !helpers.Contains(subjects, object.ObjectID) {
					subjects = append(subjects, object.ObjectID)
				}
			}
		}
	}
	return subjects
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
		if helpers.Contains(beforeAfterSubjects, meeting.SubjectID) {
			continue
		}
		if days[meeting.DayOfTheWeek] == nil {
			days[meeting.DayOfTheWeek] = make([]int, 0)
		}
		if !helpers.Contains(days[meeting.DayOfTheWeek], meeting.Hour) {
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
		if helpers.Contains(beforeAfterSubjects, meeting.SubjectID) {
			continue
		}
		if len(meetings) == 0 {
			meetings = append(meetings, meeting)
			continue
		}
		if meeting.Hour > int(relation) || meeting.Hour > int(average) {
			// Sortiramo po prioriteti (zadnje ure so vedno prve)
			for n := 0; n < len(meetings); n++ {
				if meetings[n].Hour < meeting.Hour {
					meetings = helpers.Insert(meetings, n, meeting)
					break
				}
			}
		}
	}

	p.logger.Debugw("found non-normal hours", "relation", helpers.FmtSanitize(relation), "totalHours", helpers.FmtSanitize(totalHours), "meetings", helpers.FmtSanitize(meetings), "totalDays", helpers.FmtSanitize(totalDays), "average", helpers.FmtSanitize(average))

	return meetings
}

// FindHoles bo poiskal vse luknje vmes.
func (p *protonImpl) FindHoles(timetable []ProtonMeeting) [][]ProtonMeeting {
	// Ne me vprašat zakaj sem to naredil.
	var status = make(map[int]map[int][]ProtonMeeting)

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

	freeHours := make([][]ProtonMeeting, 5)

	// Iterate over each day
	for day := 0; day < 5; day++ {
		freeHours[day] = make([]ProtonMeeting, 0)

		if status[day] == nil {
			// Samo preskoči ta dan
			continue
		}

		for hour := 1; hour <= 15; hour++ {
			if status[day][hour] != nil {
				continue
			}
			// Preverimo, če je po tej uri še kaj.
			// V primeru, da je, dodamo to kot luknjo
			// Zgodi se lahko, da se naša prva ura zgenerira na 12. uro, in če je blok ura, se zgenerira druga ura na 13. uro in je zato ne ujamemo.
			for n := hour + 1; n <= PROTON_MAX_AFTER_CLASS_HOUR+1; n++ {
				if status[day][n] != nil {
					freeHours[day] = append(freeHours[day], ProtonMeeting{Hour: hour, DayOfTheWeek: day})
					break
				}
			}
		}
	}

	p.logger.Debugw("found holes", "freeHours", helpers.FmtSanitize(freeHours))

	return freeHours
}

func (p *protonImpl) FindRelationalHoles(timetable []ProtonMeeting) []ProtonMeeting {
	var status = make(map[int]map[int][]ProtonMeeting)

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

	relationalHoles := make([]ProtonMeeting, 0)

	maxHour := 0
	minHour := -1

	// Iterate over each day
	for day := 0; day < 5; day++ {
		if status[day] == nil {
			// Samo preskoči ta dan
			continue
		}

		hCount := 0

		for hour := 1; hour <= PROTON_MAX_AFTER_CLASS_HOUR; hour++ {
			if status[day][hour] != nil {
				continue
			}

			hasHourAfter := false

			// Preverimo, če je po tej uri še kaj.
			// V primeru, da je, dodamo to kot luknjo
			// Zgodi se lahko, da se naša prva ura zgenerira na 12. uro, in če je blok ura, se zgenerira druga ura na 13. uro in je zato ne ujamemo.
			for n := hour + 1; n <= PROTON_MAX_AFTER_CLASS_HOUR+1; n++ {
				if status[day][n] != nil {
					hasHourAfter = true
					break
				}
			}

			if hasHourAfter {
				continue
			}

			hCount++
		}

		if hCount > maxHour {
			maxHour = hCount
		}
		if minHour == -1 || hCount < minHour {
			minHour = hCount
		}
	}

	// Preverimo za vse bingljajoče (neprave) luknje, ki nam niso ravno všeč.
	if minHour == -1 {
		return relationalHoles
	}

	relation := float32(minHour+maxHour) / 2
	for day := 0; day < 5; day++ {
		if status[day] == nil {
			// Samo preskoči ta dan
			continue
		}

		for hour := 1; hour <= int(relation); hour++ {
			if status[day][hour] != nil {
				continue
			}
			// Preverimo, če je po tej uri še kaj.
			// V primeru, da je, dodamo to kot luknjo
			// Zgodi se lahko, da se naša prva ura zgenerira na 12. uro, in če je blok ura, se zgenerira druga ura na 13. uro in je zato ne ujamemo.
			foundHourAfter := false
			for n := hour + 1; n <= 15; n++ {
				if status[day][n] != nil {
					foundHourAfter = true
					break
				}
			}

			if foundHourAfter {
				continue
			}

			relationalHoles = append(relationalHoles, ProtonMeeting{Hour: hour, DayOfTheWeek: day})
		}
	}

	p.logger.Debugw("found relational holes", "relationalHoles", helpers.FmtSanitize(relationalHoles))

	return relationalHoles
}

func OrderMeetingsByDay(timetable []ProtonMeeting) [][][]ProtonMeeting {
	k := make([][][]ProtonMeeting, 5)
	for i := 0; i < len(timetable); i++ {
		meeting := timetable[i]
		if meeting.Hour < 1 {
			continue
		}
		if k[meeting.DayOfTheWeek] == nil {
			// Večja številka, da ne dobimo overflowa
			k[meeting.DayOfTheWeek] = make([][]ProtonMeeting, 15)
		}
		if k[meeting.DayOfTheWeek][meeting.Hour] == nil {
			k[meeting.DayOfTheWeek][meeting.Hour] = make([]ProtonMeeting, 0)
		}
		k[meeting.DayOfTheWeek][meeting.Hour] = append(k[meeting.DayOfTheWeek][meeting.Hour], meeting)
	}

	return k
}

func (p *protonImpl) FindIfHolesExist(timetable []ProtonMeeting) bool {
	holes := p.FindHoles(timetable)
	for i := 0; i < len(holes); i++ {
		if len(holes[i]) != 0 {
			return true
		}
	}
	return false
}

func OrderMeetingsByWeek(timetable []ProtonMeeting) [][]ProtonMeeting {
	weeks := make([][]ProtonMeeting, 2)
	for i := 0; i < len(timetable); i++ {
		meeting := timetable[i]
		if weeks[meeting.Week] == nil {
			weeks[meeting.Week] = make([]ProtonMeeting, 0)
		}
		weeks[meeting.Week] = append(weeks[meeting.Week], meeting)
	}
	return weeks
}

func (p *protonImpl) AssembleMeetingsFromProtonMeetings(timetable []ProtonMeeting, systemConfig sql.Config) ([]sql.Meeting, error) {
	subjects, err := p.db.GetAllSubjects()
	if err != nil {
		return nil, err
	}
	classes, err := p.db.GetClasses()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(timetable); i++ {
		ok, err := p.CheckIfProtonConfigIsOk(timetable, timetable[i], subjects, p.GetSubjectGroups())
		if ok != "" {
			return nil, err
		}
	}

	m := make(map[string]*time.Time)

	lastSchoolDate := time.Unix(0, 0)

	// Preverimo za naše zadnje šolske dni in kateri je zadnji
	for i := 0; i < len(timetable); i++ {
		meeting := timetable[i]
		if m[meeting.SubjectID] == nil {
			subject, err := p.db.GetSubject(meeting.SubjectID)
			if err != nil {
				return nil, err
			}

			if subject.InheritsClass {
				class, err := p.db.GetClass(*subject.ClassID)
				if err != nil {
					return nil, err
				}
				t := time.Unix(int64(class.LastSchoolDate), 0)
				m[meeting.SubjectID] = &t

				if t.After(lastSchoolDate) {
					lastSchoolDate = t
				}

				continue
			}

			var subjectStudents []string
			err = json.Unmarshal([]byte(subject.Students), &subjectStudents)
			if err != nil {
				return nil, err
			}

			for n := 0; n < len(classes); n++ {
				class := classes[n]

				var students []string
				err := json.Unmarshal([]byte(class.Students), &students)
				if err != nil {
					return nil, err
				}

				var ok = false
				for x := 0; x < len(students); x++ {
					if helpers.Contains(subjectStudents, students[x]) {
						ok = true
						break
					}
				}

				if !ok {
					continue
				}

				t := time.Unix(int64(class.LastSchoolDate), 0)

				if t.After(lastSchoolDate) {
					lastSchoolDate = t
				}

				if m[meeting.SubjectID] == nil || t.After(*m[meeting.SubjectID]) {
					m[meeting.SubjectID] = &t
				}
			}
		}
	}

	p.logger.Info("last school date", lastSchoolDate)

	weeks := OrderMeetingsByWeek(timetable)

	newTimetable := make([]sql.Meeting, 0)

	currentTime := time.Now()
	firstSchoolDay, err := time.Parse("2006-01-02", fmt.Sprintf("%s-%s-%s", helpers.FmtSanitize(currentTime.Year()), "09", "01"))
	if err != nil {
		return nil, err
	}
	if firstSchoolDay.Weekday() == time.Sunday || firstSchoolDay.Weekday() == time.Saturday {
		// preskoči nekaj dni
		firstSchoolDay = firstSchoolDay.AddDate(0, 0, 1)
		if firstSchoolDay.Weekday() == time.Saturday {
			firstSchoolDay = firstSchoolDay.AddDate(0, 0, 1)
		}
	}
	firstMonday := firstSchoolDay.AddDate(0, 0, (-int(firstSchoolDay.Weekday()))+1)

	p.logger.Debugw(
		"calculated first school day",
		"schoolDay", firstSchoolDay,
		"currentTime", currentTime,
		"firstMonday", firstMonday,
		"firstSchoolWeekday", int(firstSchoolDay.Weekday()),
	)

	firstWeek := true
	weekCount := 0

	for {
		triggeredEndOfSchool := false

		for w := 0; w < len(weeks); w++ {
			week := weeks[w]

			for i := 0; i < len(week); i++ {
				meeting := week[i]

				if firstWeek && meeting.DayOfTheWeek < (int(firstSchoolDay.Weekday())-1) {
					continue
				}

				date := firstMonday.AddDate(0, 0, weekCount*7+meeting.DayOfTheWeek)
				p.logger.Debug(date)

				if (date.Day() == lastSchoolDate.Day() && date.Month() == lastSchoolDate.Month() && date.Year() == lastSchoolDate.Year()) || date.After(lastSchoolDate) {
					p.logger.Info("triggered last school date")
					triggeredEndOfSchool = true
				}

				nowDate := date.Format("02-01-2006")

				vacationDate := date.Format("2006-01-02")

				if helpers.Contains(systemConfig.SchoolFreeDays, vacationDate) {
					p.logger.Debugw("skipped meeting due to vacation", "date", vacationDate, "meeting", helpers.FmtSanitize(meeting))
					continue
				}

				subject, err := p.db.GetSubject(meeting.SubjectID)
				if err != nil {
					return nil, err
				}

				newTimetable = append(newTimetable, sql.Meeting{
					MeetingName:         meeting.SubjectName,
					TeacherID:           meeting.TeacherID,
					SubjectID:           meeting.SubjectID,
					Hour:                meeting.Hour,
					Location:            subject.Location,
					Date:                nowDate,
					IsMandatory:         true,
					URL:                 "",
					Details:             "",
					IsSubstitution:      false,
					IsGrading:           false,
					IsWrittenAssessment: false,
					IsTest:              false,
					IsCorrectionTest:    false,
					IsBeta:              true,
				})
			}
			firstWeek = false
			weekCount++

			if triggeredEndOfSchool {
				break
			}
		}

		if triggeredEndOfSchool {
			p.logger.Debug("triggered end of school")
			break
		}
	}

	return newTimetable, nil
}

func (p *protonImpl) NewProtonRule(rule genetic.ProtonRule) error {
	config, err := genetic.AddNewRule(p.config, rule)
	if err != nil {
		return err
	}
	p.config = config
	return nil
}

func (p *protonImpl) SaveConfig(config genetic.ProtonConfig) {
	err := genetic.SaveConfig(config)
	if err != nil {
		return
	}
	p.config = config
}

func (p *protonImpl) DeleteRule(ruleId string) {
	for i := 0; i < len(p.config.Rules); i++ {
		if p.config.Rules[i].ID == ruleId {
			p.config.Rules = helpers.Remove(p.config.Rules, i)
			genetic.SaveConfig(p.config)
			return
		}
	}
}

func (p *protonImpl) GetProtonConfig() genetic.ProtonConfig {
	return p.config
}

func TimetableIsInStore(timetableStore *[][]ProtonMeeting, timetable []ProtonMeeting) bool {
	for i := 0; i < len(*timetableStore); i++ {
		timetable1 := (*timetableStore)[i]
		if len(timetable1) != len(timetable) {
			continue
		}
		for n := 0; n < len(timetable1); n++ {
			ok := true
			meeting1 := timetable1[n]
			for x := 0; x < len(timetable); x++ {
				meeting2 := timetable[x]
				if meeting2.ID != meeting1.ID {
					continue
				}
				if !(meeting1.Week == meeting2.Week && meeting1.Hour == meeting2.Hour && meeting1.DayOfTheWeek == meeting2.DayOfTheWeek) {
					ok = false
					break
				}
				if x == len(timetable)-1 {
					return true
				}
			}
			if !ok {
				break
			}
		}
	}
	return false
}
