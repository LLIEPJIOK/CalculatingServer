package database

import (
	"log"

	"github.com/LLIEPJIOK/calculating-server/internal/expression"
	_ "github.com/lib/pq"
)

func createOperationTimeTableIfNotExists() {
	_, err := dataBase.Exec(`
		CREATE TABLE IF NOT EXISTS "operations_time" (
			key CHARACTER VARYING PRIMARY KEY,
			value INT
		)`)
	if err != nil {
		log.Fatal("error creating expressions table:", err)
	}
}

func GetOperationTimesFromDatabase() map[string]int64 {
	rows, err := dataBase.Query(`
		SELECT * 
		FROM "operations_time"`)
	if err != nil {
		log.Fatal("error getting data from database:", err)
	}

	operationTimes := make(map[string]int64)
	for rows.Next() {
		var key string
		var val int64
		err := rows.Scan(&key, &val)
		if err != nil {
			log.Fatal("error in getting data from database:", err)
		}
		operationTimes[key] = val
	}
	if err := rows.Err(); err != nil {
		log.Fatal("error in getting data from database:", err)
	}
	return operationTimes
}

func UpdateOperationsTime(timePlus, timeMinus, timeMultiply, timeDivide int64) {
	expression.UpdateOperationsTime(timePlus, timeMinus, timeMultiply, timeDivide)

	for key, val := range expression.GetOperationTimes() {
		_, err := dataBase.Exec(`
			UPDATE "operations_time"
			SET value = $2
			WHERE key = $1`,
			key, val)
		if err != nil {
			log.Fatal("error updating operation:", err)
		}
	}
}