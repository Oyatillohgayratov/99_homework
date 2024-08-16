package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Task struct {
    ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
    Title       string             `json:"title,omitempty" bson:"title,omitempty"`
    Description string             `json:"description,omitempty" bson:"description,omitempty"`
    Status      string             `json:"status,omitempty" bson:"status,omitempty"`
    CreatedAt   time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
}

var client *mongo.Client

func CreateTask(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var task Task
    _ = json.NewDecoder(r.Body).Decode(&task)
    task.ID = primitive.NewObjectID()
    task.CreatedAt = time.Now()
    collection := client.Database("taskdb").Collection("tasks")
    _, err := collection.InsertOne(context.TODO(), task)
    if err != nil {
        http.Error(w, "Could not create task", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(task)
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var tasks []Task
    collection := client.Database("taskdb").Collection("tasks")
    cur, err := collection.Find(context.TODO(), bson.M{})
    if err != nil {
        http.Error(w, "Could not fetch tasks", http.StatusInternalServerError)
        return
    }
    defer cur.Close(context.TODO())

    for cur.Next(context.TODO()) {
        var task Task
        err := cur.Decode(&task)
        if err != nil {
            log.Fatal(err)
        }
        tasks = append(tasks, task)
    }

    if err := cur.Err(); err != nil {
        log.Fatal(err)
    }

    json.NewEncoder(w).Encode(tasks)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    id, _ := primitive.ObjectIDFromHex(params["id"])
    var task Task
    collection := client.Database("taskdb").Collection("tasks")
    err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&task)
    if err != nil {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(task)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    id, _ := primitive.ObjectIDFromHex(params["id"])
    var task Task
    _ = json.NewDecoder(r.Body).Decode(&task)
    collection := client.Database("taskdb").Collection("tasks")
    update := bson.M{"$set": task}
    _, err := collection.UpdateOne(context.TODO(), bson.M{"_id": id}, update)
    if err != nil {
        http.Error(w, "Could not update task", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(task)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    id, _ := primitive.ObjectIDFromHex(params["id"])
    collection := client.Database("taskdb").Collection("tasks")
    _, err := collection.DeleteOne(context.TODO(), bson.M{"_id": id})
    if err != nil {
        http.Error(w, "Could not delete task", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode("Task deleted")
}

func main() {
    var err error
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    client, err = mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }
    err = client.Ping(context.TODO(), nil)
    if err != nil {
        log.Fatal(err)
    }

    router := mux.NewRouter()
    router.HandleFunc("/tasks", CreateTask).Methods("POST")
    router.HandleFunc("/tasks", GetTasks).Methods("GET")
    router.HandleFunc("/tasks/{id}", GetTask).Methods("GET")
    router.HandleFunc("/tasks/{id}", UpdateTask).Methods("PUT")
    router.HandleFunc("/tasks/{id}", DeleteTask).Methods("DELETE")

    fmt.Println("Server started at :8000")
    log.Fatal(http.ListenAndServe(":8000", router))
}
