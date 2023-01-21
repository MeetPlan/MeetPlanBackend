package proton

import (
	"errors"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/MeetPlan/MeetPlanBackend/proton/genetic"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"strconv"
)

// CheckIfProtonConfigIsOk preverja, če je trenuten timetable v redu sestavljen (v skladu z vsemi pravili).
// Ta funkcija je temelj vsega našega sistema.
func (p *protonImpl) CheckIfProtonConfigIsOk(timetable []ProtonMeeting, meeting ProtonMeeting, subjects []sql.Subject, subjectGroups []genetic.ProtonRule) (string, error) {
	// 1. korak
	// Pojdimo čez vse učitelje in preverimo, da se nič ne prekriva in je urnik skladen z učiteljevimi urami.
	var subjectsInGroup = make([]string, 0)
	var hoursToday = 1

	var subject1 sql.Subject
	for n := 0; n < len(subjects); n++ {
		if subjects[n].ID == meeting.SubjectID {
			subject1 = subjects[n]
			break
		}
	}

	for n := 0; n < len(timetable); n++ {
		meeting2 := timetable[n]

		if meeting2.DayOfTheWeek == meeting.DayOfTheWeek && meeting.SubjectID == meeting2.SubjectID && meeting.Week == meeting2.Week {
			hoursToday++
		}

		if !(meeting.DayOfTheWeek == meeting2.DayOfTheWeek && meeting.Hour == meeting2.Hour && meeting.Week == meeting2.Week) {
			continue
		}

		if meeting2.ID == meeting.ID {
			continue
		}

		if meeting.SubjectID == meeting2.SubjectID {
			return meeting2.ID, errors.New(
				fmt.Sprintf(
					"srečanja %s (%s %s) in pa %s (%s %s) se prekrivata zaradi skupnega predmeta - ne morem ustvariti urnika.",
					helpers.FmtSanitize(meeting),
					helpers.FmtSanitize(meeting.Hour),
					helpers.FmtSanitize(meeting.DayOfTheWeek),
					helpers.FmtSanitize(meeting2),
					helpers.FmtSanitize(meeting2.Hour),
					helpers.FmtSanitize(meeting2.DayOfTheWeek),
				),
			)
		}

		// Preverimo, da se učencem in učiteljem ne prekriva
		var subject2 sql.Subject
		for i := 0; i < len(subjects); i++ {
			if subjects[i].ID == meeting2.SubjectID {
				subject2 = subjects[i]
				break
			}
		}

		if subject1.InheritsClass && subject2.InheritsClass {
			if *subject2.ClassID == *subject1.ClassID {
				return meeting2.ID, errors.New(fmt.Sprintf("subjects %s and %s inherit the same class and thus cannot be made at same time", helpers.FmtSanitize(subject1.ID), helpers.FmtSanitize(subject2.ID)))
			}
			// V tem primeru nista isti razred, posledično se ne prekrivata
			continue
		}

		for s := 0; s < len(meeting.ClassID); s++ {
			class := meeting.ClassID[s]
			if helpers.Contains(meeting2.ClassID, class) {
				return meeting2.ID, errors.New(fmt.Sprintf("subjects %s and %s contain the same class %s and thus cannot be made at same time", helpers.FmtSanitize(subject1.ID), helpers.FmtSanitize(subject2.ID), helpers.FmtSanitize(class)))
			}
		}

		if subject1.TeacherID == subject2.TeacherID {
			return meeting2.ID, errors.New("učitelj ne more učiti dveh predmetov ob istem času")
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
						if !helpers.Contains(subjectsInGroup, group.Objects[y].ObjectID) {
							subjectsInGroup = append(subjectsInGroup, group.Objects[y].ObjectID)
						}
					}
				}
			}
		}

		if !(ok1 && ok2) {
			return meeting2.ID, errors.New(
				fmt.Sprintf(
					"srečanji %s (%s %s) in pa %s (%s %s) se prekrivata - ne morem ustvariti urnika.",
					helpers.FmtSanitize(meeting),
					helpers.FmtSanitize(meeting.Hour),
					helpers.FmtSanitize(meeting.DayOfTheWeek),
					helpers.FmtSanitize(meeting2),
					helpers.FmtSanitize(meeting2.Hour),
					helpers.FmtSanitize(meeting2.DayOfTheWeek),
				),
			)
		}
	}

	if hoursToday > 2 {
		return "invalid", errors.New("ne moreta biti več kot dve uri istega predmeta na en dan")
	}

	if len(subjectsInGroup) == 0 {
		subjectsInGroup = append(subjectsInGroup, meeting.SubjectID)
	}

	//p.logger.Debugw("subjects in group", "group", subjectsInGroup, "subjectGroups", subjectGroups, "meeting", meeting)

	mmap := make(map[string]bool)

	for s := 0; s < len(subjectsInGroup); s++ {
		var subject sql.Subject
		for n := 0; n < len(subjects); n++ {
			if subjects[n].ID == subjectsInGroup[s] {
				subject = subjects[n]
				break
			}
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
					if object.Type == "day" && object.ObjectID == fmt.Sprint(meeting.DayOfTheWeek) {
						mmap[subjectsInGroup[s]] = true
						break
					}
				}
			} else if rule.RuleType == 1 {
				// Ure učitelja na šoli

				// Dan je treba izvleči posebej
				// TODO: Seznam dni namesto integerja
				day := -1
				var err error
				for k := 0; k < len(rule.Objects); k++ {
					object := rule.Objects[k]
					if object.Type == "day" {
						day, err = strconv.Atoi(object.ObjectID)
						if err != nil {
							p.logger.Error(err.Error())
							break
						}
					}
				}
				if day == -1 {
					return "invalid", errors.New("neveljavno pravilo brez dni - pravilo št. 1")
				}

				if meeting.DayOfTheWeek != day {
					continue
				}

				for n := 0; n < len(rule.Objects); n++ {
					object := rule.Objects[n]

					if object.Type == "hour" && object.ObjectID == fmt.Sprint(meeting.Hour) {
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
			return "invalid", errors.New(fmt.Sprintf("srečanje s predmetom %s se ne ujema s tem, kdaj je učitelj na šoli", helpers.FmtSanitize(n)))
		}
	}
	return "", nil
}
