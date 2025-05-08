package main

import (
	"fmt"
	"strings"
)

func generateQueryForExerciseMeasurements(
	exerciseMeasurements ChoosenExercisesMap,
	userId int64,
) (sqlStatement string, args []any) {
	var placeHolderCounter int
	var placeHolders []string
	for exerciseId, measurement := range exerciseMeasurements {
		var ph []string
		ph = []string{}
		for i := 0; i < 3; i++ {
			placeHolderCounter++
			ph = append(ph, "?")
		}

		queryPart := strings.Join(ph, ", ")
		placeHolders = append(placeHolders, fmt.Sprintf("(%s)", queryPart))

		args = append(args, userId, exerciseId, measurement)
	}
	sqlStatement = fmt.Sprintf(`
		INSERT INTO last_completed_measurement (user_id, exercise_id, measurement)
		VALUES 
			%s
		ON CONFLICT (user_id, exercise_id) DO UPDATE 
		SET 
			measurement = excluded.measurement`,
		strings.Join(placeHolders, ",\n"),
	)
	return sqlStatement, args
}
