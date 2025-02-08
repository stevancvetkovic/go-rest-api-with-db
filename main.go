package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

func connectToDB() (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database. Retrying... (%d/5)\n", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Person{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

type Person struct {
	ID        uint      `json:"id"`
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func addPerson(db *gorm.DB, person *Person) error {
	result := db.Create(person)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func getPerson(db *gorm.DB, person *Person) (*Person, error) {
	var resulting *Person
	result := db.First(person)
	if result.Error != nil {
		return nil, result.Error
	}
	return resulting, nil
}

func updatePerson(db *gorm.DB, person *Person) error {
	result := db.Save(person)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func deletePerson(db *gorm.DB, person *Person) error {
	result := db.Delete(person)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func setupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", handlePing)
	r.POST("/person", handleAddPerson(db))
	r.GET("/person", handleGetPersons(db))
	r.DELETE("/person", handleDeletePerson(db))
	r.PUT("/person", handleUpdatePerson(db))

	return r
}

func handlePing(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func handleAddPerson(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var person Person
		if err := c.ShouldBindJSON(&person); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		existingPerson, err := getPerson(db, &person)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"status": "person already exists", "person": existingPerson})
			return
		}

		if err := addPerson(db, &person); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "person added", "person": person})
	}
}

func handleGetPersons(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var persons []Person
		result := db.Find(&persons)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"persons": persons})
	}
}

func handleDeletePerson(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var person Person
		if err := c.ShouldBindJSON(&person); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := getPerson(db, &person)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := deletePerson(db, &person); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "person deleted"})
	}
}

func handleUpdatePerson(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var person Person
		if err := c.ShouldBindJSON(&person); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := getPerson(db, &person)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := updatePerson(db, &person); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "person updated"})
	}
}

func main() {
	db, err := connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	r := setupRouter(db)
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Server Run Failed:", err)
	}
}
