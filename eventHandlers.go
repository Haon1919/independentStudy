package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

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

	res, err := stmt.Exec(newEvent.Title, newEvent.Description)
	if err != nil {
		panic(err.Error())
	}

	eventID, err := res.LastInsertId()
	if err != nil {
		panic(err.Error())
	}

	newEvent.ID = eventID

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newEvent)
}

func getOneEvent(w http.ResponseWriter, r *http.Request) {
	var e event
	eventID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Invalid eventID: The event id must be of type int.")
		return
	}

	result := db.QueryRow("SELECT * FROM event WHERE ID = ?", eventID)
	err = result.Scan(&e.ID, &e.Description, &e.Title, &e.StartTime, &e.EndTime)
	if err != nil {
		panic(err.Error())
	}

	json.NewEncoder(w).Encode(e)
}

func getAllEvents(w http.ResponseWriter, r *http.Request) {
	var eventList []event
	result, err := db.Query("SELECT * FROM event")
	if err != nil {
		panic(err.Error())
	}

	for result.Next() {
		var e event
		result.Scan(&e.ID, &e.Description, &e.Title, &e.StartTime, &e.EndTime)
		eventList = append(eventList, e)
	}

	json.NewEncoder(w).Encode(eventList)
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
	err = result.Scan(&originalEvent.ID, &originalEvent.Description, &originalEvent.Title, &originalEvent.StartTime, &originalEvent.EndTime)
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

	if updateEvent.StartTime == "" {
		updateEvent.StartTime = originalEvent.StartTime
	}

	if updateEvent.EndTime == "" {
		updateEvent.EndTime = originalEvent.EndTime
	}

	stmt, err := db.Prepare("UPDATE event SET description = ?, title = ?, start_time = ?, end_time = ? WHERE ID = ?")
	if err != nil {
		panic(err.Error())
	}

	updateEvent.ID = int64(eventID)

	_, err = stmt.Exec(updateEvent.Description, updateEvent.Title, updateEvent.StartTime, updateEvent.EndTime, eventID)
	if err != nil {
		panic(err.Error())
	}

	json.NewEncoder(w).Encode(updateEvent)
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

	if !eventConflictScan(userID, eventID) {
		stmt, err := db.Prepare("INSERT INTO event_subscriptions VALUES(?,?)")
		if err != nil {
			panic(err.Error())
		}

		_, err = stmt.Exec(userID, eventID)
		if err != nil {
			panic(err.Error())
		}
	} else {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "Event can not be schedualed due to conflict with another event.")
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
