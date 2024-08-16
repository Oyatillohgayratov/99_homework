package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var testRouter *mux.Router
var testCollection *mongo.Collection

func TestMain(m *testing.M) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		panic(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}

	testCollection = client.Database("testdb").Collection("tasks")
	testRouter = mux.NewRouter()
	testRouter.HandleFunc("/tasks", CreateTask).Methods("POST")
	testRouter.HandleFunc("/tasks", GetTasks).Methods("GET")
	testRouter.HandleFunc("/tasks/{id}", GetTask).Methods("GET")
	testRouter.HandleFunc("/tasks/{id}", UpdateTask).Methods("PUT")
	testRouter.HandleFunc("/tasks/{id}", DeleteTask).Methods("DELETE")

	testCollection.DeleteMany(context.TODO(), bson.M{})

	m.Run()
}

func createTask(title, description, status string) *httptest.ResponseRecorder {
	task := Task{
		Title:       title,
		Description: description,
		Status:      status,
		CreatedAt:   time.Now(),
	}
	jsonTask, _ := json.Marshal(task)
	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonTask))
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)
	return rr
}

func TestCreateTask(t *testing.T) {
	rr := createTask("Test Task", "This is a test task", "pending")
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Noto'g'ri status kodi: oldi %v kerak %v", status, http.StatusOK)
	}

	var createdTask Task
	_ = json.NewDecoder(rr.Body).Decode(&createdTask)
	if createdTask.Title != "Test Task" {
		t.Errorf("Noto'g'ri vazifa: oldi %v kerak %v", createdTask.Title, "Test Task")
	}
}

func TestGetTasks(t *testing.T) {
	createTask("Task 1", "Description 1", "pending")
	createTask("Task 2", "Description 2", "completed")

	req, _ := http.NewRequest("GET", "/tasks", nil)
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Noto'g'ri status kodi: oldi %v kerak %v", status, http.StatusOK)
	}

	var tasks []Task
	_ = json.NewDecoder(rr.Body).Decode(&tasks)
	if len(tasks) != 2 {
		t.Errorf("Vazifalar soni noto'g'ri: oldi %v kerak %v", len(tasks), 2)
	}
}

func TestGetTask(t *testing.T) {
	rr := createTask("Task ID", "Description ID", "pending")
	var createdTask Task
	_ = json.NewDecoder(rr.Body).Decode(&createdTask)

	req, _ := http.NewRequest("GET", "/tasks/"+createdTask.ID.Hex(), nil)
	rr = httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Noto'g'ri status kodi: oldi %v kerak %v", status, http.StatusOK)
	}

	var fetchedTask Task
	_ = json.NewDecoder(rr.Body).Decode(&fetchedTask)
	if fetchedTask.ID != createdTask.ID {
		t.Errorf("Noto'g'ri vazifa: oldi %v kerak %v", fetchedTask.ID, createdTask.ID)
	}
}

func TestUpdateTask(t *testing.T) {
	rr := createTask("Task to Update", "Description to Update", "pending")
	var createdTask Task
	_ = json.NewDecoder(rr.Body).Decode(&createdTask)

	updatedTask := Task{
		Title:       "Updated Task",
		Description: "Updated Description",
		Status:      "completed",
	}
	jsonTask, _ := json.Marshal(updatedTask)
	req, _ := http.NewRequest("PUT", "/tasks/"+createdTask.ID.Hex(), bytes.NewBuffer(jsonTask))
	rr = httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Noto'g'ri status kodi: oldi %v kerak %v", status, http.StatusOK)
	}

	var fetchedTask Task
	_ = json.NewDecoder(rr.Body).Decode(&fetchedTask)
	if fetchedTask.Title != "Updated Task" {
		t.Errorf("Noto'g'ri yangilangan vazifa: oldi %v kerak %v", fetchedTask.Title, "Updated Task")
	}
}

func TestDeleteTask(t *testing.T) {
	rr := createTask("Task to Delete", "Description to Delete", "pending")
	var createdTask Task
	_ = json.NewDecoder(rr.Body).Decode(&createdTask)

	req, _ := http.NewRequest("DELETE", "/tasks/"+createdTask.ID.Hex(), nil)
	rr = httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Noto'g'ri status kodi: oldi %v kerak %v", status, http.StatusOK)
	}

	req, _ = http.NewRequest("GET", "/tasks/"+createdTask.ID.Hex(), nil)
	rr = httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Noto'g'ri status kodi: oldi %v kerak %v", status, http.StatusNotFound)
	}
}
