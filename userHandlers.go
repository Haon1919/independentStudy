package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

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
		err := eventsResult.Scan(&e.ID, &e.Description, &e.Title, &e.StartTime, &e.EndTime)
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

	stmt, err := db.Prepare("UPDATE user SET first_name = ?, last_name = ? WHERE ID = ?")
	if err != nil {
		panic(err.Error())
	}

	_, err = stmt.Exec(updateUser.FirstName, updateUser.LastName, userID)
	if err != nil {
		panic(err.Error())
	}

	updateUser.ID = int64(userID)

	json.NewEncoder(w).Encode(updateUser)
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
