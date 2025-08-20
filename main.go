package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

type workout struct {
	Date  string `json:"date"`
	Wtype string `json:"wtype"`
	Data  string `json:"data"`
}

func dbConnect() *sql.DB {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := sql.Open("postgres", os.Getenv("CONN_STR"))
	if err != nil {
		log.Fatal(err)
	}

	createTable(db)

	fmt.Println("Successfully connected to db and created table")
	return db
}

func createTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS workouts (
    	date DATE NOT NULL,
    	workout_type TEXT NOT NULL,
    	exercise_data JSONB NOT NULL
	)`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func addWorkout(db *sql.DB, date string, workoutType string, workoutData string) string {
	query := "INSERT INTO workouts VALUES ($1, $2, $3)" //date, workout type and json workout data
	_, err := db.Exec(query, date, workoutType, workoutData)
	if err != nil {
		log.Fatal(err)
	}
	return date
}

func removeWorkout(db *sql.DB, date string) string {
	query := "DELETE FROM workouts WHERE date = $1"
	res, err := db.Exec(query, date)
	if err != nil {
		log.Fatal(err)
	}

	result, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Removed workout successfully", result)
	return date
}

func listWorkouts(db *sql.DB) []workout {
	query := "SELECT * FROM workouts"
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	var workouts []workout
	for rows.Next() {
		var workoutDate string
		var workoutType string
		var workoutData string
		if err := rows.Scan(&workoutDate, &workoutType, &workoutData); err != nil {
			log.Fatal(err)
		}
		workouts = append(workouts, workout{workoutDate, workoutType, workoutData})
	}
	return workouts
}

func getWorkout(db *sql.DB, date string) workout {
	query := "SELECT * FROM workouts WHERE date = $1"
	row := db.QueryRow(query, date)
	var workoutDate string
	var workoutType string
	var workoutData string
	err := row.Scan(&workoutDate, &workoutType, &workoutData)
	if err != nil {
		log.Fatal(err)
	}
	return workout{workoutDate, workoutType, workoutData}
}

func listWorkoutsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		workouts := listWorkouts(db)
		c.IndentedJSON(http.StatusOK, workouts)
	}
}

func getWorkoutHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		date := c.Param("date")
		workout := getWorkout(db, date)
		c.JSON(http.StatusOK, workout)
	}
}

func addWorkoutHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var workout workout
		if err := c.BindJSON(&workout); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		date := addWorkout(db, workout.Date, workout.Wtype, workout.Data)
		c.JSON(http.StatusCreated, date)
	}
}

func removeWorkoutHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		date := c.Param("date")
		dateRemoved := removeWorkout(db, date)
		c.JSON(http.StatusOK, dateRemoved)
	}
}

func main() {
	db := dbConnect()
	//addWorkout(db, "2025-05-26", "horiz_push_pull", `{"26.05_hpp" : {"ex_name" : "Flat dumbbell press", "sets" : 3, "reps" : 12, "weight" : 14, "notes" : "Last set 15"}}`)
	fmt.Println(listWorkouts(db))

	router := gin.Default()
	router.GET("/workout/:date", getWorkoutHandler(db))
	router.GET("/allworkouts", listWorkoutsHandler(db))
	router.POST("/addworkout", addWorkoutHandler(db))
	router.DELETE("/removeworkout/:date", removeWorkoutHandler(db))
	err := router.Run("localhost:6942")
	if err != nil {
		log.Fatal(err)
	}
}

/*
{
"Date": "2025-05-26",
"Wtype": "horiz_push_pull",
"Data": "{\"26.05_hpp\": {\"ex_name\": \"Incline Dumbbell Rows\", \"sets\": 3, \"reps\": 12, \"weight\": 14, \"notes\": \"Last set 15\"}}"
}
*/
//curl -X POST localhost:6942/addworkout -H "Content-Type:application/josn" -d "{\"Date\":\"2025-05-26\",\"Wtype\":\"horiz_push_pull\",\"Data\":\"{\\\"26.05_hpp\\\":{\\\"ex_name\\\":\\\"Flat dumbbell press\\\",\\\"sets\\\":3,\\\"reps\\\":12,\\\"weight\\\":14,\\\"notes\\\":\\\"Last set 15\\\"}}\"}"
//"2025-05-26"
