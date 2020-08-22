package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pakohan/craftdoor/controller/keys"
	"github.com/pakohan/craftdoor/controller/members"
	"github.com/pakohan/craftdoor/model"
	"github.com/pakohan/craftdoor/service"
)

type controller struct {
	m model.Model
	s *service.Service

	http.Handler
}

// New returns a new http.Handler
func New(m model.Model, s *service.Service) http.Handler {
	r := mux.NewRouter()
	c := &controller{
		m: m,
		s: s,
		Handler: handlers.CORS(
			handlers.AllowedOrigins([]string{
				"http://localhost:8081",
				"http://localhost:8080",
			}),
			handlers.AllowedHeaders([]string{
				"Authorization",
				"Content-Type",
				"Accept",
				"Origin",
				"User-Agent",
				"DNT",
				"Cache-Control",
				"X-Mx-ReqToken",
				"Keep-Alive",
				"X-Requested-With",
				"If-Modified-Since",
			}),
			handlers.AllowedMethods([]string{
				"GET",
				"PUT",
				"POST",
				"DELETE",
				"HEAD",
			}),
		)(r),
	}
	r.Path("/").Methods(http.MethodGet).HandlerFunc(c.ReadNextTag)
	members.New(r.PathPrefix("/members").Subrouter(), m)
	keys.New(r.PathPrefix("/keys").Subrouter(), m, s)
	return c
}

// func (c *controller) returnState(w http.ResponseWriter, r *http.Request) {
// 	id, err := uuid.Parse(r.URL.Query().Get("id"))
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	state, err := c.s.WaitForChange(r.Context(), id)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	err = json.NewEncoder(w).Encode(state)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		log.Printf("err encoding response: %s", err)
// 		return
// 	}
// }

// ReadNextTag reads the next available RFID tag and returns its data.
func (c *controller) ReadNextTag(resp http.ResponseWriter, req *http.Request) {
	log.Printf("Attempting to read next available tag...")

	// TODO(duckworthd): Read timeout from query parameter "timeout_sec" if available.
	var timeout time.Duration = 5 * time.Second

	state, err := c.s.ReadNextTag(timeout)
	if err != nil {
		log.Printf("Failed in call to Service.ReadNextTag(): %s", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(resp).Encode(state)
	if err != nil {
		log.Printf("Failed to encode JSON: %s", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Succesfully return tag: %s", state.UUID)
}
