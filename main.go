package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func initDatabase() {
	DbServername := os.Getenv("DB_SERVERNAME")
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	dsn := dbUsername + ":" + dbPassword + "@tcp(" + DbServername + ":" + dbPort + ")/" + dbName

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("SET time_zone = 'Europe/Amsterdam'")
	if err != nil {
		log.Fatalf("Error setting timezone: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error when trying to ping database: %v", err)
	}
}

type reservation struct {
	Reservation_id  int    `json:"reservation_id"`
	FirstName       string `json:"firstName" binding:"required"`
	LastName        string `json:"lastName" binding:"required"`
	PhoneNumber     string `json:"phoneNumber" binding:"required"`
	LicensePlate    string `json:"licensePlate" binding:"required"`
	DateOfDeparture string `json:"dateOfDeparture" binding:"required"`
	DateOfArrival   string `json:"dateOfArrival" binding:"required"`
}

func main() {
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error when opening log.txt: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.Printf("Log output set to file")

	initDatabase()
	defer db.Close()

	router := gin.Default()

	router.GET("/reservations/licensePlate/:licensePlate", checkLicensePlate)
	router.POST("/reservation", createReservation)
	router.PATCH("/reservations/:reservation_id", updateReservation)
	router.DELETE("/reservations/:reservation_id", deleteReservation)

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "8080"
	}

	router.Run(":" + dbPort)
}

func checkLicensePlate(c *gin.Context) {
	licensePlate := c.Param("licensePlate")
	var exists bool

	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM reservations WHERE licensePlate = ?)", licensePlate).Scan(&exists)
	if err != nil {
		log.Printf("Error checking license plate: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error checking license plate"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exists": exists})
}

func createReservation(c *gin.Context) {
	var newReservation reservation

	if err := c.BindJSON(&newReservation); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing fields"})
		return
	}
	exists, err := reservationExists(newReservation.LicensePlate)
	if err != nil {
		log.Printf("Error checking reservation existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error checking reservation existence"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"message": "Reservation already exists"})
		return
	}

	insertQuery := `
	INSERT INTO reservations (firstName, lastName, phoneNumber, licensePlate, dateOfDeparture, dateOfArrival)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(insertQuery, newReservation.FirstName, newReservation.LastName, newReservation.PhoneNumber, newReservation.LicensePlate, newReservation.DateOfDeparture, newReservation.DateOfArrival)
	if err != nil {
		log.Printf("Error inserting reservation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error inserting reservation"})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting last insert ID"})
		return
	}

	newReservation.Reservation_id = int(id)
	c.JSON(http.StatusCreated, newReservation)
}

func reservationExists(licensePlate string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM reservations WHERE licensePlate = ?)", licensePlate).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func updateReservation(c *gin.Context) {
	id := c.Param("reservation_id")
	var updatedReservation reservation

	if err := c.BindJSON(&updatedReservation); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing fields"})
		return
	}

	updateQuery := `
	UPDATE reservations
	SET firstName = ?, lastName = ?, phoneNumber = ?, licensePlate = ?, dateOfDeparture = ?, dateOfArrival = ?
	WHERE reservation_id = ?
	`

	_, err := db.Exec(updateQuery, updatedReservation.FirstName, updatedReservation.LastName, updatedReservation.PhoneNumber, updatedReservation.LicensePlate, updatedReservation.DateOfDeparture, updatedReservation.DateOfArrival, id)
	if err != nil {
		log.Printf("Error updating reservation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error updating reservation"})
		return
	}

	c.JSON(http.StatusOK, updatedReservation)
}

func deleteReservation(c *gin.Context) {
	id := c.Param("reservation_id")

	deleteQuery := "DELETE FROM reservations WHERE reservation_id = ?"
	_, err := db.Exec(deleteQuery, id)
	if err != nil {
		log.Printf("Error deleting reservation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error deleting reservation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reservation deleted"})
}
