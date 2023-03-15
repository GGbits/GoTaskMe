package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ravendb/ravendb-go-client"
	"log"
	"net/http"
	"time"
)

type task struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	DueDate     time.Time `json:"dueDate"`
	Completed   bool      `json:"completed"`
	Recur       time.Time `json:"recur"`
}

var apr1, err1 = time.Parse("2006-01-02 03:04:05", "2023-04-01 00:00:00")
var apr15, err15 = time.Parse("2006-01-02 03:04:05", "2023-04-15 00:00:00")

var tasks = []task{
	{Id: "1", Name: "Clean Bedroom", Description: "Clean your bedroom", Priority: 4, DueDate: apr1, Completed: false, Recur: apr15},
	{Id: "2", Name: "Vacuum", Description: "Vacuum the house", Priority: 3, DueDate: apr1, Completed: false},
}

func main() {
	// initialize database
	store, err := getDocumentStore("gtm")
	if err != nil {
		fmt.Printf("There was an issue initializing the database. The error is %s", err)
	}

	// add initial values
	for _, tk := range tasks {
		var t *task = &tk
		storeTask(store, t)
	}

	// spin up api
	router := gin.Default()
	// Use Docstore function to create middleware for grabbing Document store variable
	router.Use(DocStore(store))
	router.GET("/tasks", getTasks)
	router.POST("/tasks", postTasks)
	router.Run("localhost:8080")

}

// getTasks response with the list of all tasks as JSON.
func getTasks(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, tasks)
}

// postTasks adds a task from JSON received in the request body
func postTasks(c *gin.Context) {
	var newTask task
	var ntp *task
	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newTask); err != nil {
		return
	}

	// Add the new album to the slice.
	tasks = append(tasks, newTask)
	ntp = &newTask
	// Grab DocumentStore from Gin Context
	ds := c.MustGet("ds").(*ravendb.DocumentStore)
	storeTask(ds, ntp)
	c.IndentedJSON(http.StatusCreated, newTask)
}

func getDocumentStore(databaseName string) (*ravendb.DocumentStore, error) {
	serverNodes := []string{"http://localhost:8000"}
	store := ravendb.NewDocumentStore(serverNodes, databaseName)
	if err := store.Initialize(); err != nil {
		return nil, err
	}
	return store, nil
}

func storeTask(store *ravendb.DocumentStore, t *task) {
	session, err := store.OpenSession("gtm")
	if err != nil {
		log.Fatalf("store.OpenSession() faailed with %s", err)
	}
	err = session.Store(t)
	if err != nil {
		log.Fatalf("session.Store() failed with %s\n", err)
	}
	err = session.SaveChanges()
	if err != nil {
		log.Fatalf("session.SaveChanges() failed with %s\n", err)
	}
	session.Close()
}

// DocStore Middleware allowing document store variable to be accessed in gin context.
func DocStore(ds *ravendb.DocumentStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("ds", ds)
		c.Next()
	}
}
