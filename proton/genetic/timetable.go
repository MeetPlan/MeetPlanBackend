package genetic

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/google/uuid"
	"math/rand"
)

func (p *protonImpl) AssembleTimetable() ([]Meeting, error) {
	subjects, err := p.db.GetAllSubjects()
	if err != nil {
		return nil, err
	}

	classesSQL, err := p.db.GetClasses()
	if err != nil {
		return nil, err
	}

	allClassrooms := make([]string, 0)

	classList := make([]MeetingDefinition, 0)
	for _, subject := range subjects {
		var classId = make([]string, 0)
		var students []string
		if subject.InheritsClass {
			classId = append(classId, *subject.ClassID)
			for i := 0; i < len(classesSQL); i++ {
				if classesSQL[i].ID != *subject.ClassID {
					continue
				}
				err = json.Unmarshal([]byte(classesSQL[i].Students), &students)
				if err != nil {
					return nil, err
				}
			}
		} else {
			err = json.Unmarshal([]byte(subject.Students), &students)
			if err != nil {
				return nil, err
			}
			for i := 0; i < len(classesSQL); i++ {
				var classStudents []string
				err := json.Unmarshal([]byte(classesSQL[i].Students), &classStudents)
				if err != nil {
					return nil, err
				}
				for n := 0; n < len(students); n++ {
					if helpers.Contains(classStudents, students[n]) && !helpers.Contains(classId, classesSQL[i].ID) {
						classId = append(classId, classesSQL[i].ID)
					}
				}
			}
		}
		if !helpers.Contains(allClassrooms, subject.Location) {
			allClassrooms = append(allClassrooms, subject.Location)
		}
		for i := 0; i < int(subject.SelectedHours)/2; i++ {
			meeting := MeetingDefinition{
				ID:           uuid.NewString(),
				ClassID:      classId,
				Duration:     2,
				IsHalfHour:   false,
				StackedHours: false,
				StudentIDs:   students,
				SubjectID:    subject.ID,
				TeacherID:    subject.TeacherID,
				ClassroomIDs: []string{subject.Location},
				Classrooms:   []int{},
			}
			classList = append(classList, meeting)
		}
	}

	groupsEmptySpace := make(map[int][]int)
	teachersEmptySpace := make(map[string][]int)
	subjectsOrder := make(map[XY]XYZ)

	classes := make(map[int]MeetingDefinition)
	classrooms := make(map[int]Classroom)
	groups := make(map[string]int)
	teachers := make(map[string]int)

	for _, meeting := range classList {
		_, containsKey := teachersEmptySpace[meeting.TeacherID]
		if !containsKey {
			teachersEmptySpace[meeting.TeacherID] = make([]int, 0)
		}

		// groups == students
		for _, group := range meeting.StudentIDs {
			_, containsKey = groups[group]
			if !containsKey {
				groups[group] = len(groups)
				groupsEmptySpace[groups[group]] = make([]int, 0)
			}
		}

		_, containsKey = teachers[meeting.TeacherID]
		if !containsKey {
			teachers[meeting.TeacherID] = len(teachers)
		}
	}

	rand.Shuffle(len(classList), func(i, j int) {
		classList[i], classList[j] = classList[j], classList[i]
	})
	for _, cl := range classList {
		classes[len(classes)] = cl
	}

	for _, classroom := range allClassrooms {
		classrooms[len(classrooms)] = Classroom{Name: classroom, Type: "P"}
	}

	for i, cl := range classes {
		indexClassrooms := make([]int, 0)

		for index := range classrooms {
			// tuki za izbiro
			//if helpers.Contains(cl.ClassroomIDs, c) {
			indexClassrooms = append(indexClassrooms, index)
			//}
		}
		cl.Classrooms = indexClassrooms

		indexGroups := make([]int, 0)
		for name, index := range groups {
			//p.logger.Debug(name, index)
			if !helpers.Contains(cl.StudentIDs, name) {
				continue
			}
			xy := XY{
				X: cl.SubjectID,
				Y: index,
			}
			_, contains := subjectsOrder[xy]
			if !contains {
				subjectsOrder[xy] = XYZ{-1, -1, -1}
			}
			indexGroups = append(indexGroups, index)
		}
		cl.Groups = indexGroups

		classes[i] = cl
	}

	data := Data{
		Meetings:   classes,
		Groups:     groups,
		Classrooms: classrooms,
	}

	matrix, free := GenerateMatrix(len(allClassrooms))
	filled := make(map[int][]XY)
	p.logger.Debug(groups)
	//p.logger.Debug(groupsEmptySpace)
	//p.logger.Debug(teachersEmptySpace)
	//p.logger.Debug(subjectsOrder)
	p.logger.Debug(data.Meetings)
	InitialPopulation(data, matrix, free, filled, groupsEmptySpace, teachersEmptySpace, subjectsOrder)
	p.logger.Debug("done")

	return nil, nil
}
