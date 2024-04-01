package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/LLIEPJIOK/calculating-server/internal/expression"
	"github.com/LLIEPJIOK/calculating-server/internal/user"
	_ "github.com/lib/pq"
)

const (
	host             = "localhost"
	port             = 5432
	databaseUser     = "postgres"
	password         = "123409874567"
	ExpressionDBName = "expressions_db"
	OperationTimeDB  = "operation_time_db"
)

var (
	dataBase *sql.DB
)

func rowsToExpressionsSlice(rows *sql.Rows) []*expression.Expression {
	var expressions []*expression.Expression
	for rows.Next() {
		var exp expression.Expression
		err := rows.Scan(&exp.Id, &exp.Exp, &exp.Result, &exp.Status, &exp.CreationTime, &exp.CalculationTime)
		if err != nil {
			log.Println("error in getting data from database:", err)
			return nil
		}
		expressions = append(expressions, &exp)
	}

	if err := rows.Err(); err != nil {
		log.Println("error in getting data from database:", err)
		return nil
	}
	return expressions
}

func createDatabaseIfNotExists() {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
		host, port, databaseUser, password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("error open database:", err)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		log.Fatal("error connecting to database:", err)
	}

	rows, err := db.Query(`
		SELECT 1 
		FROM pg_database 
		WHERE datname = $1`,
		ExpressionDBName)
	if err != nil {
		log.Fatal("error checking database existence:", err)
	}
	defer rows.Close()

	if !rows.Next() {
		_, err = db.Exec(`CREATE DATABASE ` + ExpressionDBName)
		if err != nil {
			log.Fatal("error creating database:", err)
		}
	}
}

func createExpressionsTableIfNotExists() {
	_, err := dataBase.Exec(`
		CREATE TABLE IF NOT EXISTS "expressions" (
			id SERIAL PRIMARY KEY,
			exp CHARACTER VARYING,
			result DOUBLE PRECISION,
			status CHARACTER VARYING,
			creation_time TIMESTAMP,
			calculation_time TIMESTAMP
		)`)
	if err != nil {
		log.Fatal("error creating expressions table:", err)
	}
}

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

func createUsersTableIfNotExists() {
	_, err := dataBase.Exec(`
		CREATE TABLE IF NOT EXISTS "users" (
			login VARCHAR(100) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			hash_password VARCHAR(100) NOT NULL
		)`)
	if err != nil {
		log.Fatal("error creating users table:", err)
	}
}

func InsertUserInDatabase(insertingUser *user.User) {
	_, err := dataBase.Exec(`
		INSERT INTO "users"(login, name, hash_password) 
		VALUES($1, $2, $3)`,
		insertingUser.Login, insertingUser.Name, insertingUser.HashPassword)
	if err != nil {
		log.Printf("error in insert %#v in database: %v\n", *insertingUser, err)
		return
	}
}

func createTablesIfNotExists() {
	createExpressionsTableIfNotExists()
	createOperationTimeTableIfNotExists()
	createUsersTableIfNotExists()
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

func InsertExpressionInBD(exp *expression.Expression) {
	_, err := dataBase.Exec(`
		INSERT INTO "expressions"(exp, result, status, creation_time, calculation_time) 
		VALUES($1, $2, $3, $4, $5)`,
		exp.Exp, exp.Result, exp.Status, exp.CreationTime, exp.CalculationTime)
	if err != nil {
		log.Printf("error in insert %#v in database: %v\n", *exp, err)
		return
	}

	err = dataBase.QueryRow(`SELECT LASTVAL()`).Scan(&exp.Id)
	if err != nil {
		log.Fatal(err)
	}
}

func GetExpressionById(id string) []*expression.Expression {
	var rows *sql.Rows
	var err error
	if id == "" {
		rows, err = dataBase.Query(`
			SELECT * 
			FROM "expressions" 
			ORDER BY id DESC`)
	} else {
		rows, err = dataBase.Query(`
			SELECT * 
			FROM "expressions" 
			WHERE CAST(id AS TEXT) LIKE '%' || $1 || '%'
			ORDER BY id ASC`, id)
	}
	if err != nil {
		log.Fatal("error in getting data from database:", err)
	}
	defer rows.Close()

	return rowsToExpressionsSlice(rows)
}

func GetUncalculatingExpressions() []*expression.Expression {
	rows, err := dataBase.Query(`
		SELECT * 
		FROM "expressions"
		WHERE status = 'calculating' OR status = 'in queue'`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	return rowsToExpressionsSlice(rows)
}

func UpdateStatus(exp *expression.Expression) {
	_, err := dataBase.Exec(`
		UPDATE "expressions" 
		SET status = $1 
		WHERE id = $2`,
		exp.Status, exp.Id)
	if err != nil {
		log.Printf("error in setting %#v in database: %v\n", *exp, err)
	}
}

func UpdateResult(exp *expression.Expression) {
	_, err := dataBase.Exec(`
		UPDATE "expressions" 
		SET result = $1 
		WHERE id = $2`,
		exp.Result, exp.Id)
	if err != nil {
		log.Printf("error in setting %#v in database: %v\n", *exp, err)
	}

	_, err = dataBase.Exec(`
		UPDATE "expressions" 
		SET calculation_time = $1 
		WHERE id = $2`,
		exp.CalculationTime, exp.Id)
	if err != nil {
		log.Printf("error in setting %#v in database: %v\n", *exp, err)
	}
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

func GetLastExpressions() []*expression.Expression {
	rows, err := dataBase.Query("SELECT * FROM \"expressions\" ORDER BY creation_time DESC LIMIT 10")
	if err != nil {
		log.Println("error in getting data from database:", err)
		return nil
	}
	defer rows.Close()

	lastInputs := make([]*expression.Expression, 0, 10)
	for rows.Next() {
		var exp expression.Expression
		err := rows.Scan(&exp.Id, &exp.Exp, &exp.Result, &exp.Status, &exp.CreationTime, &exp.CalculationTime)
		if err != nil {
			log.Println("error in getting data from database:", err)
			return nil
		}
		lastInputs = append(lastInputs, &exp)
	}

	if err := rows.Err(); err != nil {
		log.Println("error in getting data from database:", err)
		return nil
	}
	return lastInputs
}

func Close() {
	dataBase.Close()
}

func Configure() {
	createDatabaseIfNotExists()

	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, databaseUser, password, ExpressionDBName)
	var err error
	dataBase, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("error open database:", err)
	}
	if err = dataBase.Ping(); err != nil {
		log.Fatal("error connecting to database:", err)
	}

	createTablesIfNotExists()
}
