package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type event struct {
	ID          int64  `json:"ID"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
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

func createEvent(w http.ResponseWriter, r *http.Request) {
	var newEvent event
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to create an event.")
		return
	}

	json.Unmarshal(reqBody, &newEvent)

	stmt, err := db.Prepare("INSERT INTO event (description, title) VALUES (?,?)")
	if err != nil {
		panic(err.Error())
	}

	_, err = stmt.Exec(newEvent.Title, newEvent.Description)
	if err != nil {
		panic(err.Error())
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newEvent)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	var newUser user
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Kindly enter data with the user first name and last name only in order to create a user.")
		return
	}

	json.Unmarshal(reqBody, &newUser)

	stmt, err := db.Prepare("INSERT INTO user (first_name, last_name) VALUES (?,?)")
	if err != nil {
		panic(err.Error())
	}

	res, err := stmt.Exec(newUser.FirstName, newUser.LastName)
	if err != nil {
		panic(err.Error())
	}

	userID, err := res.LastInsertId()
	if err != nil {
		panic(err.Error())
	}

	newUser.ID = userID

	json.NewEncoder(w).Encode(newUser)
}

func getOneEvent(w http.ResponseWriter, r *http.Request) {
	var e event
	eventID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Invalid eventID: The event id must be of type int.")
		return
	}

	result := db.QueryRow("SELECT * FROM event WHERE ID = ?")
	err = result.Scan(&e.ID, &e.Description, &e.Title)
	if err != nil {
		panic(err.Error())
	}

	json.NewEncoder(w).Encode(e)
}

func getOneUser(w http.ResponseWriter, r *http.Request) {
	var u user
	userID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Invalid userID: The user id must be of type int.")
		return
	}

	result := db.QueryRow("SELECT * FROM user WHERE ID = ?", userID)
	err = result.Scan(&u.ID, &u.FirstName, &u.LastName)
	if err != nil {
		panic(err.Error())
	}
	json.NewEncoder(w).Encode(u)
}

func getAllEvents(w http.ResponseWriter, r *http.Request) {
	var eventList []event
	result, err := db.Query("SELECT * FROM event")
	if err != nil {
		panic(err.Error())
	}

	for result.Next() {
		var e event
		result.Scan(&e.ID, &e.Description, &e.Title)
		eventList = append(eventList, e)
	}

	json.NewEncoder(w).Encode(eventList)
}

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var userList []user
	var eventList []event

	userIDIndexMap := make(map[int64]int)
	eventIDIndexMap := make(map[int64]int)

	result, err := db.Query("SELECT * FROM user")
	if err != nil {
		panic(err.Error())
	}

	defer result.Close()

	userIndex := 0
	for result.Next() {
		var u user
		err := result.Scan(&u.ID, &u.FirstName, &u.LastName)
		if err != nil {
			panic(err.Error())
		}
		userList = append(userList, u)

		userIDIndexMap[u.ID] = userIndex
		userIndex++
	}

	eventsResult, err := db.Query("SELECT * FROM event")
	if err != nil {
		panic(err.Error())
	}

	eventIndex := 0
	for eventsResult.Next() {
		var e event
		err := eventsResult.Scan(&e.ID, &e.Description, &e.Title)
		if err != nil {
			panic(err.Error())
		}
		eventList = append(eventList, e)
		eventIDIndexMap[e.ID] = eventIndex
		eventIndex++
	}

	eventSubscriptionResults, err := db.Query("SELECT * FROM event_subscriptions")
	if err != nil {
		panic(err.Error())
	}

	for eventSubscriptionResults.Next() {
		var es eventSubscription
		err := eventSubscriptionResults.Scan(&es.userID, &es.eventID)
		if err == nil {
			panic(err.Error())
		}

		ui := userIDIndexMap[es.userID]
		ei := eventIDIndexMap[es.eventID]
		uel := userList[ui].Events
		e := eventList[ei]
		uel = append(uel, e)
	}

	json.NewEncoder(w).Encode(userList)
}

func updateEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(mux.Vars(r)["id"])
	var updateEvent event
	var originalEvent event

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Invalid eventID: The event id must be of type int.")
		return
	}

	result := db.QueryRow("SELECT * FROM event WHERE ID = ?", eventID)
	err = result.Scan(&originalEvent.ID, &originalEvent.Description, &originalEvent.Title)
	if err != nil {
		panic(err.Error())
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
		return
	}
	json.Unmarshal(reqBody, &updateEvent)

	if updateEvent.Description == "" {
		updateEvent.Description = originalEvent.Description
	}

	if updateEvent.Title == "" {
		updateEvent.Title = originalEvent.Title
	}

	stmt, err := db.Prepare("UPDATE event SET description = ?, title = ? WHERE ID = ?")
	if err != nil {
		panic(err.Error())
	}

	_, err = stmt.Exec(updateEvent.Description, updateEvent.Title, updateEvent.ID)
	if err != nil {
		panic(err.Error())
	}

	json.NewEncoder(w).Encode(updateEvent)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Invalid userID: The user id must be of type int.")
		return
	}

	var updateUser user
	var originalUser user

	result := db.QueryRow("SELECT * FROM user WHERE ID = ?", userID)
	err = result.Scan(&originalUser.ID, &originalUser.FirstName, &originalUser.LastName)
	if err != nil {
		panic(err.Error())
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Kindly enter data with the event first name and last name only in order to update")
		return
	}
	json.Unmarshal(reqBody, &updateUser)

	if updateUser.FirstName == "" {
		updateUser.FirstName = originalUser.FirstName
	}

	if updateUser.LastName == "" {
		updateUser.LastName = originalUser.LastName
	}

	stmt, err := db.Prepare("UPDATE event SET first_name = ?, last_name = ? WHERE ID = ?")
	if err != nil {
		panic(err.Error())
	}

	_, err = stmt.Exec(updateUser.FirstName, updateUser.LastName, updateUser.ID)
	if err != nil {
		panic(err.Error())
	}

}

func addUserToEvent(w http.ResponseWriter, r *http.Request) {
	//TODO:: Finish implementing method
	userID, userErr := strconv.Atoi(mux.Vars(r)["userId"])
	if userErr != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Invalid userID: The user id must be of type int.")
		return
	}

	eventID, eventErr := strconv.Atoi(mux.Vars(r)["eventId"])
	if eventErr != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Invalid eventID: The event id must be of type int.")
		return
	}

}

func deleteEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Invalid eventID: The event id must be of type int.")
		return
	}

	stmt, err := db.Prepare("DELETE FROM event WHERE ID = ?")
	if err != nil {
		panic(err.Error())
	}

	_, err = stmt.Exec(eventID)
	if err != nil {
		panic(err.Error())
	}

	fmt.Fprintf(w, "Event with id of %v successfully deleted.", eventID)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Invalid userID: The user id must be of type int.")
		return
	}

	stmt, err := db.Prepare("DELETE FROM user WHERE ID = ?")
	if err != nil {
		panic(err.Error())
	}

	_, err = stmt.Exec(userID)
	if err != nil {
		panic(err.Error())
	}

	fmt.Fprintf(w, "User with id of %v successfully deleted.", userID)
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
