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

	if DbServername == "" || dbUsername == "" || dbPassword == "" || dbName == "" || dbPort == "" {
		log.Fatal("One or more environment variables are not set correctly")
	}

	dsn := dbUsername + ":" + dbPassword + "@tcp(" + DbServername + ":" + dbPort + ")/" + dbName

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error when trying to ping database: %v", err)
	}
}

type reservation struct {
	ReservationID int    `json:"reservation_id"`
	Name          string `json:"name" binding:"required"`
	Surname       string `json:"surname" binding:"required"`
	Phone         string `json:"phone" binding:"required"`
	LicensePlate  string `json:"license_plate" binding:"required"`
	CheckOutDate  string `json:"check_out_date" binding:"required"`
	CheckInDate   string `json:"check_in_date" binding:"required"`
}

func main() {
	log.SetOutput(os.Stdout)
	log.Printf("Log output set to stdout")

	initDatabase()
	log.Printf("database initialized")

	router := gin.Default()

	router.GET("/reservations/licensePlate/:license_plate", checkLicensePlate)
	router.POST("/reservation", createReservation)
	router.PATCH("/reservations/:reservation_id", updateReservation)
	router.DELETE("/reservations/:reservation_id", deleteReservation)

	router.Run(":8080")
	defer db.Close()
}

func checkLicensePlate(c *gin.Context) {
	licensePlate := c.Param("license_plate")
	var exists bool

	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM reservations WHERE license_plate = ?)", licensePlate).Scan(&exists)
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
	INSERT INTO reservations (name, surname, phone, license_plate, check_out_date, check_in_date)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(insertQuery, newReservation.Name, newReservation.Surname, newReservation.Phone, newReservation.LicensePlate, newReservation.CheckOutDate, newReservation.CheckInDate)
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

	newReservation.ReservationID = int(id)
	c.JSON(http.StatusCreated, newReservation)
}

func reservationExists(licensePlate string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM reservations WHERE license_plate = ?)", licensePlate).Scan(&exists)
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
	SET name = ?, surname = ?, phone = ?, license_plate = ?, check_out_date = ?, check_in_date = ?
	WHERE reservation_id = ?
	`

	_, err := db.Exec(updateQuery, updatedReservation.Name, updatedReservation.Surname, updatedReservation.Phone, updatedReservation.LicensePlate, updatedReservation.CheckOutDate, updatedReservation.CheckInDate, id)
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
