package proton

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/MeetPlan/MeetPlanBackend/proton/genetic"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/google/uuid"
	"strconv"
)

const MAX_HOURS_PER_DAY = 7

func (p *protonImpl) AssembleTimetable() ([]ProtonMeeting, error) {
	subjects, err := p.db.GetAllSubjects()
	if err != nil {
		return nil, err
	}

	classes, err := p.db.GetClasses()
	if err != nil {
		return nil, err
	}

	teachers, err := p.db.GetTeachers()
	if err != nil {
		return nil, err
	}

	daysSubjects := make(map[string][]int)

	for i := 0; i < len(teachers); i++ {
		subjects, err := p.db.GetAllSubjectsForTeacher(teachers[i].ID)
		if err != nil {
			return nil, err
		}
		teacherRules := p.GetAllRulesForTeacher(teachers[i].ID)
		for n := 0; n < len(teacherRules); n++ {
			teacherRule := teacherRules[n]
			for x := 0; x < len(teacherRule.Objects); x++ {
				object := teacherRule.Objects[x]
				if object.Type != "day" {
					continue
				}
				day, err := strconv.Atoi(object.ObjectID)
				if err != nil {
					return nil, err
				}
				for s := 0; s < len(subjects); s++ {
					if daysSubjects[subjects[s].ID] == nil {
						continue
					}
					daysSubjects[subjects[s].ID] = append(daysSubjects[subjects[s].ID], day)
				}
			}
		}
	}

	// Before & After class subjects will be treated differently
	//beforeAfterSubjects := p.GetSubjectsBeforeOrAfterClass()
	stackedSubjects := p.GetSubjectsWithStackedHours()

	meetings := make([]ProtonMeeting, 0)
	halves := make([]ProtonMeeting, 0)
	// TODO: Implement before/after class hours
	//beforeAfterClassW1 := make([]ProtonMeeting, 0)
	//beforeAfterClassW2 := make([]ProtonMeeting, 0)

	for i := 0; i < len(subjects); i++ {
		subject := subjects[i]
		var classId = make([]string, 0)
		if subject.InheritsClass {
			classId = append(classId, *subject.ClassID)
		} else {
			var students []string
			err := json.Unmarshal([]byte(subject.Students), &students)
			if err != nil {
				return nil, err
			}
			for i := 0; i < len(classes); i++ {
				var classStudents []string
				err := json.Unmarshal([]byte(classes[i].Students), &classStudents)
				if err != nil {
					return nil, err
				}
				for n := 0; n < len(students); n++ {
					if helpers.Contains(classStudents, students[n]) && !helpers.Contains(classId, classes[i].ID) {
						classId = append(classId, classes[i].ID)
					}
				}
			}
		}
		if subject.SelectedHours-float32(int(subject.SelectedHours)) == 0.5 {
			UUID, err := uuid.NewUUID()
			if err != nil {
				return nil, err
			}
			// halves don't need CommonID
			halves = append(halves, ProtonMeeting{
				ID:          UUID.String(),
				SubjectName: subject.Name,
				SubjectID:   subject.ID,
				ClassID:     classId,
				TeacherID:   subject.TeacherID,
				Week:        0,
				IsHalfHour:  true,
			})
		}
		for n := 0; n < int(subject.SelectedHours*2-1); n++ {
			CommonUUID, err := uuid.NewUUID()
			if err != nil {
				return nil, err
			}
			UUID, err := uuid.NewUUID()
			if err != nil {
				return nil, err
			}
			meetings = append(meetings, ProtonMeeting{
				ID:           UUID.String(),
				SubjectName:  subject.Name,
				SubjectID:    subject.ID,
				ClassID:      classId,
				TeacherID:    subject.TeacherID,
				Week:         0,
				IsHalfHour:   false,
				CommonID:     CommonUUID.String(),
				StackedHours: helpers.Contains(stackedSubjects, subject.ID),
			})
		}
	}

	timetable := make([]ProtonMeeting, 0)

	go func() {
		timetables := make([][]ProtonMeeting, 0)
		for {
			timetable := <-p.timetable
			fmt.Println("recieved a timetable", timetable)
			if timetable == nil {
				p.logger.Debug(timetables)
				return
			}
			timetables = append(timetables, timetable)
		}
	}()

	subjectGroups := p.GetSubjectGroups()

	//timetableStore := make([][]ProtonMeeting, 0)

	p.AssembleTimetableRecursive(1, 0, meetings, timetable, subjects, subjectGroups)

	// we shouldn't run the goroutine indefinitely
	p.timetable <- nil

	return nil, nil
}

func (p *protonImpl) AssembleTimetableRecursive(hour int, day int, meetings []ProtonMeeting, timetable []ProtonMeeting, subjects []sql.Subject, subjectGroups []genetic.ProtonRule) ([]string, error) {
	p.logger.Debug(len(meetings))

	if len(meetings) == 0 {
		// we send new timetable to channel where it's cached
		p.timetable <- timetable
		return make([]string, 0), nil
	}

	// TODO: Implement subject groups
	//subjectGroups := p.GetSubjectGroups()

	for i := 0; i < len(meetings); i++ {
		insertable := false
		collisions := make([]string, 0)
		for d := day; d < 5; d++ {
			h := 1
			if d == day {
				h = hour
			}
			for ; h <= 7; h++ {
				meetingsUnstable := make([]ProtonMeeting, 0)
				meetingsUnstable = append(meetingsUnstable, meetings...)
				meetingsUnstable[i].Hour = h
				meetingsUnstable[i].DayOfTheWeek = d
				meeting := meetingsUnstable[i]

				//if TimetableIsInStore(timetableStore, timetableUnstable) {
				//	p.logger.Debug("timetable is in store", timetableUnstable)
				//	continue
				//}

				//*timetableStore = append(*timetableStore, timetableUnstable)

				ok, err := p.CheckIfProtonConfigIsOk(timetable, meeting, subjects, subjectGroups)
				if ok != "" {
					// `ok` večino časa vsebuje UUID srečanja
					// to pomeni, da lahko zapustimo to funkcijo in jo zapuščamo DOKLER ne najdemo
					// srečanja, ki se prekriva s tem, posledično zamenjamo uro in dan
					if ok != "invalid" && !helpers.Contains(collisions, ok) {
						collisions = append(collisions, ok)
					}
					//p.logger.Debug("error while checking proton config", err)
					continue
				}

				timetableUnstable := make([]ProtonMeeting, 0)
				timetableUnstable = append(timetableUnstable, timetable...)

				timetableUnstable = append(timetableUnstable, meetingsUnstable[i])
				meetingsUnstable = helpers.Remove(meetingsUnstable, i)

				insertable = true

				collision, err := p.AssembleTimetableRecursive(hour, day, meetingsUnstable, timetableUnstable, subjects, subjectGroups)
				if len(collision) != 0 {
					if helpers.Contains(collision, meeting.ID) {
						p.logger.Debug("found collision on this recursive depth")
						continue
					} else {
						p.logger.Debug("returning due to collisions ", collision)
						return collision, err
					}
				}
			}
		}
		if !insertable {
			return collisions, errors.New(fmt.Sprintf("meeting %s is not insertable anywhere. cannot assemble the timetable", fmt.Sprint(meetings[i])))
		}
	}

	return []string{}, nil
}
