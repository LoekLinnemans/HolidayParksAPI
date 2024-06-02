package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID            int    `json:"id"`
	Voornaam      string `json:"voornaam" binding:"required"`
	Achternaam    string `json:"achternaam" binding:"required"`
	Kenteken      string `json:"kenteken" binding:"required"`
	Vertrekdatum  string `json:"vertrekdatum" binding:"required"`
	Aankomstdatum string `json:"aankomstdatum" binding:"required"`
}

var db *sql.DB

func initDatabase() {
	var err error
	dsn := "user:password@tcp(127.0.0.1:3306)/userdb"
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

	router.GET("/users/kenteken/:kenteken", checkLicensePlate)
	router.POST("/gebruiker", createUser)
	router.PATCH("/users/:id", updateUser)
	router.DELETE("/users/:id", deleteUser)

	router.Run("localhost:8080")
}

func checkLicensePlate(c *gin.Context) {
	kenteken := c.Param("kenteken")
	var exists bool

	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE kenteken = ?)", kenteken).Scan(&exists)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"exists": exists})
}

func createUser(c *gin.Context) {
	var newUser User

	if err := c.BindJSON(&newUser); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing fields"})
		return
	}

	insertQuery := `
	INSERT INTO users (voornaam, achternaam, kenteken, vertrekdatum, aankomstdatum)
	VALUES (?, ?, ?, ?, ?)
	`

	result, err := db.Exec(insertQuery, newUser.Voornaam, newUser.Achternaam, newUser.Kenteken, newUser.Vertrekdatum, newUser.Aankomstdatum)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}

	newUser.ID = int(id)
	c.IndentedJSON(http.StatusCreated, newUser)
}

func updateUser(c *gin.Context) {
	id := c.Param("id")
	var updatedUser User

	if err := c.BindJSON(&updatedUser); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing fields"})
		return
	}

	updateQuery := `
	UPDATE users
	SET voornaam = ?, achternaam = ?, kenteken = ?, vertrekdatum = ?, aankomstdatum = ?
	WHERE id = ?
	`

	_, err := db.Exec(updateQuery, updatedUser.Voornaam, updatedUser.Achternaam, updatedUser.Kenteken, updatedUser.Vertrekdatum, updatedUser.Aankomstdatum, id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}

	c.IndentedJSON(http.StatusOK, updatedUser)
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")

	deleteQuery := "DELETE FROM users WHERE id = ?"
	_, err := db.Exec(deleteQuery, id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "user deleted"})
}
