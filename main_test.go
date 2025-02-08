package main

import (
	"bytes"
	"encoding/json"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&Person{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func TestPingRoute(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	r := setupRouter(db)

	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, but got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "pong" {
		t.Fatalf("Expected response body 'pong', but got %s", w.Body.String())
	}
}

func TestAddPerson(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	r := setupRouter(db)

	person := Person{
		Firstname: "John",
		Lastname:  "Doe",
	}
	jsonValue, _ := json.Marshal(person)
	req, _ := http.NewRequest("POST", "/person", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, but got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response["status"] != "person added" {
		t.Fatalf("Expected status 'person added', but got %s", response["status"])
	}
}

func TestGetPersons(t *testing.T) {
	db, err := setupTestDB()
	db.Exec("DELETE FROM people")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	r := setupRouter(db)

	person := Person{
		Firstname: "John",
		Lastname:  "Doe",
	}
	db.Create(&person)

	req, _ := http.NewRequest("GET", "/person", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, but got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	persons := response["persons"].([]interface{})
	if len(persons) != 1 {
		t.Fatalf("Expected 1 person, but got %d", len(persons))
	}
}

func TestDeletePerson(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	r := setupRouter(db)

	person := Person{
		Firstname: "John",
		Lastname:  "Doe",
	}
	db.Create(&person)

	deleteRequest := struct {
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
	}{
		Firstname: "John",
		Lastname:  "Doe",
	}
	jsonValue, _ := json.Marshal(deleteRequest)
	req, _ := http.NewRequest("DELETE", "/person", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, but got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response["status"] != "person deleted" {
		t.Fatalf("Expected status 'person deleted', but got %s", response["status"])
	}
}
