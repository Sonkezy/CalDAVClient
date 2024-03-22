package swagger

import (
	"encoding/json"
	"fmt"
	"log"
	caldav "mycaldav/pkg/caldav_client"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func RunSwagger(client *caldav.CaldavClient) {
	log.Printf("Server started")

	var GetCalendarsNames = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		names := client.GetCalendarsNames()
		namesBytes, err := json.Marshal(names)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(namesBytes)
		w.WriteHeader(http.StatusOK)
	}
	var GetCalendars = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		events, err := client.GetCalendars()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		eventsBytes, err := json.Marshal(events)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(eventsBytes)
		w.WriteHeader(http.StatusOK)
	}
	var PutEvent = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		var event caldav.EventClient
		err := json.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		client.PutEvent(event)
		eventsBytes, err := json.Marshal(event)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		w.Write(eventsBytes)
		w.WriteHeader(http.StatusOK)
	}
	var routes = Routes{
		Route{
			"Index",
			"GET",
			"/",
			Index,
		},

		Route{
			"GetCalendarsNames",
			strings.ToUpper("Get"),
			"/GetCalendarsNames",
			GetCalendarsNames,
		},
		Route{
			"GetCalendars",
			strings.ToUpper("Get"),
			"/GetCalendars",
			GetCalendars,
		},
		Route{
			"PutEvent",
			strings.ToUpper("Post"),
			"/PutEvent",
			PutEvent,
		},
	}

	router := NewRouter(routes)
	log.Fatal(http.ListenAndServe(":8080", router))
}

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter(routes Routes) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

func Index(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Hello World!")
}
