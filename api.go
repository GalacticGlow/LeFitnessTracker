package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
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

type apiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func dbConnect() *sql.DB {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := sql.Open("postgres", os.Getenv("CONN_STR"))
	if err != nil {
		log.Fatal("Cant connect to database")
	}

	createTable(db)

	fmt.Println("Successfully connected to db and created table")
	return db
}

func createTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS workouts (
    	date DATE PRIMARY KEY,
    	workout_type TEXT NOT NULL,
    	exercise_data JSONB NOT NULL
	)`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Cant create table")
	}
}

func addWorkout(db *sql.DB, date string, workoutType string, workoutData string) (string, error) {
	query := "INSERT INTO workouts VALUES ($1, $2, $3)" // date, workout type, json workout data
	_, err := db.Exec(query, date, workoutType, workoutData)
	if err != nil {
		// Check if the error is a duplicate key violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return "", fmt.Errorf("Workout for this date already exists")
		}
		return "", fmt.Errorf("Cant add workout: %w", err)
	}
	return date, nil
}

func removeWorkout(db *sql.DB, date string) (string, error) {
	query := "DELETE FROM workouts WHERE date = $1"
	res, err := db.Exec(query, date)
	if err != nil {
		return "", fmt.Errorf("Cant remove workout")
	}

	result, err := res.RowsAffected()
	if err != nil || result != 1 {
		return "", fmt.Errorf("Couldnt remove workout, rows were not affected")
	}
	fmt.Println("Removed workout successfully", result)
	return date, nil
}

func listWorkouts(db *sql.DB) ([]workout, error) {
	query := "SELECT * FROM workouts"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("Workouts could not be found to list, query didnt work")
	}
	var workouts []workout
	for rows.Next() {
		var workoutDate string
		var workoutType string
		var workoutData string
		if err := rows.Scan(&workoutDate, &workoutType, &workoutData); err != nil {
			return nil, fmt.Errorf("Cant list workouts")
		}
		workouts = append(workouts, workout{workoutDate, workoutType, workoutData})
	}
	return workouts, nil
}

func getWorkout(db *sql.DB, date string) (workout, error) {
	query := "SELECT * FROM workouts WHERE date = $1"
	row := db.QueryRow(query, date)
	var workoutDate string
	var workoutType string
	var workoutData string
	err := row.Scan(&workoutDate, &workoutType, &workoutData)
	if err == sql.ErrNoRows {
		return workout{workoutDate, workoutType, workoutData}, fmt.Errorf("No workout found")
	} else if err != nil {
		return workout{}, err
	}
	return workout{workoutDate, workoutType, workoutData}, nil
}

func updateWorkout(db *sql.DB, date string, workoutData string) (string, error) {
	query := "UPDATE workouts SET exercise_data = $1 WHERE date = $2"
	res, err := db.Exec(query, workoutData, date)
	if err != nil {
		return "", fmt.Errorf("Cant update workout %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		return "", fmt.Errorf("Cant update workout rows affected")
	}
	fmt.Println("Updated workout successfully", rows)
	return date, nil
}

func listWorkoutsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		workouts, err := listWorkouts(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, apiResponse{
				Success: false, Error: err.Error(),
			})
			return
		}
		c.IndentedJSON(http.StatusOK, apiResponse{
			Success: true,
			Data:    workouts,
		})
	}
}

func getWorkoutHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		date := c.Param("date")
		workout, err := getWorkout(db, date)
		if err != nil {
			status := http.StatusInternalServerError
			if err.Error() == "No workout found" {
				status = http.StatusNotFound
			}
			c.IndentedJSON(status, apiResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		c.IndentedJSON(http.StatusOK, apiResponse{
			Success: true,
			Data:    workout,
		})
	}
}

func addWorkoutHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var workout workout
		if err := c.BindJSON(&workout); err != nil {
			c.JSON(http.StatusBadRequest, apiResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		date, err := addWorkout(db, workout.Date, workout.Wtype, workout.Data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, apiResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusCreated, apiResponse{
			Success: true,
			Data:    date,
		})
	}
}

func removeWorkoutHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		date := c.Param("date")
		dateRemoved, err := removeWorkout(db, date)
		if err != nil {
			status := http.StatusInternalServerError
			if err.Error() == "Cant remove workout" {
				status = http.StatusNotFound
			}
			c.JSON(status, apiResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		c.IndentedJSON(http.StatusOK, apiResponse{
			Success: true,
			Data:    dateRemoved,
		})
	}
}

func updateWorkoutHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var w workout
		if err := c.BindJSON(&w); err != nil {
			c.JSON(http.StatusBadRequest, apiResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		date := c.Param("date")

		if _, err := updateWorkout(db, date, w.Data); err != nil {
			c.JSON(http.StatusInternalServerError, apiResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		var updated workout
		er := db.QueryRow("SELECT date, workout_type, exercise_data FROM workouts WHERE date=$1", date).
			Scan(&updated.Date, &updated.Wtype, &updated.Data)
		if er != nil {
			c.JSON(http.StatusInternalServerError, apiResponse{
				Success: false,
				Error:   er.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, apiResponse{
			Success: true,
			Data:    updated,
		})
	}
}

func main() {
	db := dbConnect()
	//addWorkout(db, "2025-05-31", "vertical_push_pull", `{"26.05_hpp" : [{"ex_name" : "Flat dumbbell press", "sets" : 3, "reps" : 12, "weight" : 14, "notes" : "Last set 15"}, {"ex_name" : "Inclined db rows", "sets" : 5, "reps" : 69, "weight" : 140, "notes" : "Last set 75"}]}`)
	fmt.Println(listWorkouts(db))

	router := gin.Default()

	router.Static("/frontend", "./frontend")
	router.GET("/", func(c *gin.Context) {
		c.File("./frontend/index.html")
	})

	router.Use(cors.Default())
	router.GET("/workout/:date", getWorkoutHandler(db))
	router.GET("/allworkouts", listWorkoutsHandler(db))
	router.POST("/addworkout", addWorkoutHandler(db))
	router.DELETE("/removeworkout/:date", removeWorkoutHandler(db))
	router.PATCH("/updateworkout/:date", updateWorkoutHandler(db))
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
//curl -X POST localhost:6942/updateworkout/2025-05-30 -H "Content-Type:application/json" -d "{\"Data\":\"{\\\"30.05_vpp\\\":{\\\"ex_name\\\":\\\"Flat dumbbell press\\\",\\\"sets\\\":3,\\\"reps\\\":12,\\\"weight\\\":14,\\\"notes\\\":\\\"Last set 15\\\"}}\"}
