package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func dbConnect() *sql.DB {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := sql.Open("postgres", os.Getenv("CONN_STR"))
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	fmt.Println("Successfully connected to db")
	return db
}

func main() {
	router := gin.Default()
	err := router.Run("localhost:6969")
	if err != nil {
		log.Fatal(err)
	}
}
