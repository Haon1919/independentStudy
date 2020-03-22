package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type event struct {
	ID          int    `json:"ID"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
}

type user struct {
	ID        int     `json:"ID"`
	FirstName string  `json:"FirstName"`
	LastName  string  `json:"LastName"`
	Events    []event `json:"Events"`
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
	if len(users) != 0 {
		newEvent.ID = events[len(events)-1].ID + 1
	} else {
		newEvent.ID = 1
	}
	events = append(events, newEvent)
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newEvent)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var newUser user
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Kindly enter data with the user first name and last name only in order to create a user.")
		return
	}

	json.Unmarshal(reqBody, &newUser)
	if len(users) != 0 {
		newUser.ID = users[len(users)-1].ID + 1
	} else {
		newUser.ID = 1
	}
	users = append(users, newUser)
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newUser)
}

func getOneEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Invalid eventID: The event id must be of type int.")
		return
	}

	for _, singleEvent := range events {
		if eventID == singleEvent.ID {
			json.NewEncoder(w).Encode(singleEvent)
		}
	}
}

func getOneUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Invalid userID: The user id must be of type int.")
		return
	}

	for _, singleUser := range users {
		if userID == singleUser.ID {
			json.NewEncoder(w).Encode(singleUser)
		}
	}
}

func getAllEvents(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(events)
}

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(users)
}

func updateEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Invalid eventID: The event id must be of type int.")
		return
	}

	var updateEvent event

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
		return
	}
	json.Unmarshal(reqBody, &updateEvent)

	for i, singleEvent := range events {
		if singleEvent.ID == eventID {
			singleEvent.Title = updateEvent.Title
			singleEvent.Description = updateEvent.Description
			events[i] = singleEvent
			json.NewEncoder(w).Encode(singleEvent)
		}
	}
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Invalid userID: The user id must be of type int.")
		return
	}

	var updateUser user

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Kindly enter data with the event first name and last name only in order to update")
		return
	}
	json.Unmarshal(reqBody, &updateUser)

	for i, singleUser := range users {
		if singleUser.ID == userID {
			if updateUser.FirstName != "" {
				singleUser.FirstName = updateUser.FirstName
			}
			if updateUser.LastName != "" {
				singleUser.LastName = updateUser.LastName
			}
			users[i] = singleUser
			json.NewEncoder(w).Encode(singleUser)
		}
	}
}

func addUserToEvent(w http.ResponseWriter, r *http.Request) {
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

	var targetUser *user
	var targetEvent event

	for i := 0; i < len(users); i++ {
		if userID == users[i].ID {
			targetUser = &users[i]
			break
		}
	}

	if targetUser.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Invalid userID: No user exists with an id of %v.", userID)
		return
	}

	for _, singleEvent := range events {
		if eventID == singleEvent.ID {
			targetEvent = singleEvent
			break
		}
	}
	if targetEvent.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Invalid eventID: No event exists with an id of %v.", eventID)
		return
	}
	for _, userEvent := range targetUser.Events {
		if userEvent.ID == eventID {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "User with id %v already is scheduled for an event with id %v", userID, eventID)
			return
		}
	}
	targetUser.Events = append(targetUser.Events, targetEvent)

	json.NewEncoder(w).Encode(targetUser)
}

func deleteEvent(w http.ResponseWriter, r *http.Request) {
	var numberOfEvents int = len(events)

	eventID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Invalid eventID: The event id must be of type int.")
		return
	}

	if numberOfEvents > 0 {
		for i, singleEvent := range events {
			if singleEvent.ID == eventID {
				events = append(events[:i], events[i+1:]...)
				fmt.Fprintf(w, "The event with id %v has been successfully deleted.", eventID)
			}
		}
	}

	if numberOfEvents == len(events) {
		fmt.Fprintf(w, "Invalid eventID: No event exists with an id of %v.", eventID)
	}
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	var numberOfUsers int = len(users)

	userID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Invalid userID: The user id must be of type int.")
		return
	}

	if numberOfUsers > 0 {
		for i, singleUser := range events {
			if singleUser.ID == userID {
				users = append(users[:i], users[i+1:]...)
				fmt.Fprintf(w, "The user with id %v has been successfully deleted.", userID)
			}
		}
	}

	if numberOfUsers == len(users) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Invalid userID: No user exists with an id of %v.", userID)
		return
	}
}

func main() {
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
