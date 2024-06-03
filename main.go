package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type reservation struct {
	ID              int    `json:"reservering_id"`
	FirstName       string `json:"firstName" binding:"required"`
	LastName        string `json:"lastName" binding:"required"`
	PhoneNumber     string `json:"phoneNumber" binding:"required"`
	LicensePlate    string `json:"licensePlate" binding:"required"`
	DateOfDeparture string `json:"dateOfDeparture" binding:"required"`
	DateOfArrival   string `json:"dateOfArrival" binding:"required"`
}

var db *sql.DB

func initDatabase() {
	var err error
	dsn := "loek:Flint2013.@tcp(127.0.0.1:3306)/holidayparksdb"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	initDatabase()
	defer db.Close()

	router := gin.Default()

	router.GET("/reservations/licensePlate/:licensePlate", checkLicensePlate)
	router.POST("/reservation", createReservation)
	router.PATCH("/reservations/:id", updateReservation)
	router.DELETE("/reservations/:id", deleteReservation)

	router.Run("localhost:8080")
}

func checkLicensePlate(c *gin.Context) {
	licensePlate := c.Param("licensePlate")
	var exists bool

	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM reservations WHERE licensePlate = ?)", licensePlate).Scan(&exists)
	if err != nil {
		log.Printf("Error checking license plate: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error checking license plate"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"exists": exists})
}

func createReservation(c *gin.Context) {
	var newReservation reservation

	if err := c.BindJSON(&newReservation); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing fields"})
		return
	}

	insertQuery := `
	INSERT INTO reservations (firstName, lastName, phoneNumber, licensePlate, dateOfDeparture, dateOfArrival)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(insertQuery, newReservation.FirstName, newReservation.LastName, newReservation.PhoneNumber, newReservation.LicensePlate, newReservation.DateOfDeparture, newReservation.DateOfArrival)
	if err != nil {
		log.Printf("Error inserting reservation: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error inserting reservation"})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error getting last insert ID"})
		return
	}

	newReservation.ID = int(id)
	c.IndentedJSON(http.StatusCreated, newReservation)
}

func updateReservation(c *gin.Context) {
	id := c.Param("id")
	var updatedReservation reservation

	if err := c.BindJSON(&updatedReservation); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing fields"})
		return
	}

	updateQuery := `
	UPDATE reservations
	SET firstName = ?, lastName = ?, phoneNumber = ?, licensePlate = ?, dateOfDeparture = ?, dateOfArrival = ?
	WHERE id = ?
	`

	_, err := db.Exec(updateQuery, updatedReservation.FirstName, updatedReservation.LastName, updatedReservation.PhoneNumber, updatedReservation.LicensePlate, updatedReservation.DateOfDeparture, updatedReservation.DateOfArrival, id)
	if err != nil {
		log.Printf("Error updating reservation: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error updating reservation"})
		return
	}

	c.IndentedJSON(http.StatusOK, updatedReservation)
}

func deleteReservation(c *gin.Context) {
	id := c.Param("id")

	deleteQuery := "DELETE FROM reservations WHERE id = ?"
	_, err := db.Exec(deleteQuery, id)
	if err != nil {
		log.Printf("Error deleting reservation: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error deleting reservation"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "reservation deleted"})
}
