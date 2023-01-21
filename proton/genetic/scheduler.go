package genetic

import (
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"strconv"
)

func InitialPopulation(data Data, matrix [][]*int, free []XY, filled map[int][]XY, groupsEmptySpace map[int][]int, teachersEmptySpace map[string][]int, subjectsOrder map[XY]XYZ) {
	for index, meeting := range data.Meetings {
		fmt.Println("Currently on index", index)
		ind := 0
		for {
			startField := free[ind]
			startTime, err := strconv.Atoi(fmt.Sprint(startField.X))
			if err != nil {
				return
			}
			endTime := startTime + meeting.Duration - 1
			if startTime%12 > endTime%12 {
				ind++
				continue
			}

			found := false
			// check if whole block for the class is free
			for i := 1; i < meeting.Duration; i++ {
				field := XY{X: i + startTime, Y: startField.Y}
				for n := 0; n < len(free); n++ {
					if free[n] == field {
						found = true
						ind++
						break
					}
				}
				if found {
					break
				}
			}

			y, err := strconv.Atoi(fmt.Sprint(startField.Y))
			if err != nil {
				panic(err)
			}

			if !helpers.Contains(meeting.Classrooms, y) {
				ind += 1
				continue
			}

			if !found {
				continue
			}

			// groups
			for _, groupIndex := range meeting.Groups {
				// add order of the subjects for group
				InsertOrder(&subjectsOrder, meeting.SubjectID, meeting.Groups, "P", startTime)
				// add times of the class for group
				for i := 0; i < meeting.Duration; i++ {
					groupsEmptySpace[groupIndex] = append(groupsEmptySpace[groupIndex], i+startTime)
				}
			}

			for i := 0; i < meeting.Duration; i++ {
				xy := XY{X: i + startTime, Y: startField.Y}
				filled[index] = make([]XY, 0)
				filled[index] = append(filled[index], xy)
				removed := 0
				for n, x := range free {
					if xy != x {
						continue
					}
					helpers.Remove(free, n-removed)
					removed++
				}
				teachersEmptySpace[meeting.TeacherID] = append(teachersEmptySpace[meeting.TeacherID], i+startTime)
			}
			break
		}
	}

	// fill the matrix
	for index, fieldsList := range filled {
		for _, field := range fieldsList {
			x, _ := strconv.Atoi(fmt.Sprint(field.X))
			y, _ := strconv.Atoi(fmt.Sprint(field.Y))
			matrix[x][y] = &index
		}
	}
}
