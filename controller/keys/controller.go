package keys

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/pakohan/craftdoor/model"
	"github.com/pakohan/craftdoor/model/key"
	"github.com/pakohan/craftdoor/service"
)

type controller struct {
	m model.Model
	s *service.Service
}

// New initializes a new router
func New(r *mux.Router, m model.Model, s *service.Service) {
	c := controller{
		m: m,
		s: s,
	}

	// POST requests.
	r.Methods(http.MethodPost).Path("/new").HandlerFunc(c.register)
	r.Methods(http.MethodPost).HandlerFunc(c.create)

	// GET requests.
	r.Methods(http.MethodGet).HandlerFunc(c.list)
	// TODO(duckworthd): Implement a method for fetching details about a single key.

	// PUT requests.
	r.Methods(http.MethodPut).Path("/{id}").HandlerFunc(c.update)

	// DELETE requests.
	r.Methods(http.MethodDelete).Path("/{id}").HandlerFunc(c.delete)
}

func (c *controller) list(w http.ResponseWriter, r *http.Request) {
	res, err := c.m.KeyModel.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Printf("err encoding response: %s", err.Error())
	}
}

func (c *controller) create(w http.ResponseWriter, r *http.Request) {
	t := key.Key{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.m.KeyModel.Create(r.Context(), &t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(t)
	if err != nil {
		log.Printf("err encoding response: %s", err.Error())
		return
	}
}

func (c *controller) update(w http.ResponseWriter, r *http.Request) {
	t := key.Key{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse Key ID.
	base := 10
	bitSize := 64
	t.ID, err = strconv.ParseInt(mux.Vars(r)["id"], base, bitSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse Member ID.
	memberID, err := strconv.ParseInt(r.Form.Get("member_id"), base, bitSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	t.MemberID = &memberID

	// Update database.
	err = c.m.KeyModel.Update(r.Context(), &t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (c *controller) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.m.KeyModel.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (c *controller) register(w http.ResponseWriter, r *http.Request) {
	// Populate what fields one can from JSON.
	t := key.Key{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Read tag's UUID from the tag reader.
	state, err := c.s.ReadNextTag(5 * time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !state.IsTagAvailable {
		http.Error(w, "RFID tag not found. Is the tag in front of the reader?", http.StatusInternalServerError)
	}
	if state.TagInfo.ID == "" {
		http.Error(w, "RFID tag's ID is empty. This is an internal error and should not happen...", http.StatusInternalServerError)
	}
	t.UUID = state.TagInfo.ID

	// Insert new tag into database.
	err = c.m.KeyModel.Create(r.Context(), &t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate response.
	err = json.NewEncoder(w).Encode(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
