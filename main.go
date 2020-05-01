package main

//TODO: When deleting a user make sure any event they are registered for is deleted
//TODO: When deleting an event throw an error stating that users are registered for the event if the start and end time have not elapsed;
//TODO: Add start times to the event table
//TODO: Dissallow users to be schedualed for an event if the start time or end time of the event to be schedualed overlap with another event they have scheduled
//TODO: Break main.go file into multiple handlers
//TODO: Break up larger handler functions into helper functions
//TODO: Unit test where possible

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type event struct {
	ID          int64  `json:"ID"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
	StartTime   string `json:"StartTime"`
	EndTime     string `json:"EndTime"`
}

type user struct {
	ID        int64   `json:"ID"`
	FirstName string  `json:"FirstName"`
	LastName  string  `json:"LastName"`
	Events    []event `json:"Events"`
}

type eventSubscription struct {
	userID  int64
	eventID int64
}

type allEvents []event
type allUsers []user

var events = allEvents{
	{
		ID:          1,
		Title:       "Introduction to Golang",
		Description: "Come join us for a chance to learn how golang works and get to try it out",
	},
}

var users = allUsers{
	{
		ID:        1,
		FirstName: "Noah",
		LastName:  "Shirey",
		Events:    []event{},
	},
}

var db *sql.DB
var err error

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

func main() {
	db, err = sql.Open("mysql", "root:root@tcp(localhost:3306)/event_scheduler")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/event", createEvent).Methods("POST")
	router.HandleFunc("/user", createUser).Methods("POST")
	router.HandleFunc("/events", getAllEvents).Methods("GET")
	router.HandleFunc("/users", getAllUsers).Methods("GET")
	router.HandleFunc("/events/{id}", getOneEvent).Methods("GET")
	router.HandleFunc("/users/{id}", getOneUser).Methods("GET")
	router.HandleFunc("/events/{id}", updateEvent).Methods("PATCH")
	router.HandleFunc("/users/{id}", updateUser).Methods("PATCH")
	router.HandleFunc("/schedule/user/{userId}/event/{eventId}", addUserToEvent).Methods("PATCH")
	router.HandleFunc("/events/{id}", deleteEvent).Methods("DELETE")
	router.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))
}
