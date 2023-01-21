package genetic

import "encoding/json"

func InsertOrder(subjectsOrder *map[XY]XYZ, subject string, group []int, tip string, startTime int) {
	/*
		We need to simulate Python inheritance using pointers over the map.
		Inserts start time of the class for given subject, group and type of class.
	*/
	marshal, err := json.Marshal(group)
	if err != nil {
		return
	}
	// torej, go očitno ne pusti []int v map keyu, tooooooorej rabimo vse skupi encodat v JSON string.
	times := (*subjectsOrder)[XY{X: subject, Y: string(marshal)}]
	if tip == "P" {
		times.X = startTime
	} else if tip == "V" {
		times.Y = startTime
	} else {
		times.Z = startTime
	}
	(*subjectsOrder)[XY{X: subject, Y: string(marshal)}] = times
}

func GenerateMatrix(numOfColumns int) ([][]*int, []XY) {
	matrix := make([][]*int, 5*8)
	for y := 0; y < 5*8; y++ {
		matrix[y] = make([]*int, numOfColumns)
		for x := 0; x < numOfColumns; x++ {
			matrix[y][x] = nil
		}
	}

	free := make([]XY, 0)

	// initialise free dict as all the fields from matrix
	for i := 0; i < len(matrix); i++ {
		for j := 0; j < len(matrix[i]); j++ {
			free = append(free, XY{X: i, Y: j})
		}
	}
	return matrix, free
}
