package main

import "time"

const (
	Stamp = "Jan _2 15:04:05"
)

func eventConflictScan(userID int, eventID int) bool {
	var e event

	eventResult := db.QueryRow("SELECT * FROM event WHERE ID = ?", eventID)
	err = eventResult.Scan(&e.ID, &e.Description, &e.Title, &e.StartTime, &e.EndTime)
	if err != nil {
		panic(err.Error())
	}

	st, err := time.Parse(Stamp, e.StartTime)
	if err != nil {
		panic(err.Error())
	}

	et, err := time.Parse(Stamp, e.EndTime)
	if err != nil {
		panic(err.Error())
	}

	result, err := db.Query("SELECT * FROM event JOIN event_subscription ON event.ID = event_subscription.event_id WHERE event.user_id = ?", userID)
	if err != nil {
		panic(err.Error())
	}

	for result.Next() {
		var e event
		result.Scan(&e.ID, &e.Description, &e.Title, &e.StartTime, &e.EndTime)

		eventST, err := time.Parse(Stamp, e.StartTime)
		if err != nil {
			panic(err.Error())
		}

		eventET, err := time.Parse(Stamp, e.EndTime)
		if err != nil {
			panic(err.Error())
		}

		if (st.After(eventST) && st.Before(eventET)) || (et.After(eventST) && et.Before(eventET)) {
			return true
		}

	}
	return false
}
