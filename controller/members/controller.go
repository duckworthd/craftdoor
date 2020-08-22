package members

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pakohan/craftdoor/model"
)

type controller struct {
	m model.Model
}

// New initializes a new router
func New(r *mux.Router, m model.Model) {
	c := controller{
		m: m,
	}
	// POST requests.
	r.Methods(http.MethodPost).HandlerFunc(c.create)

	// GET requests.
	r.Methods(http.MethodGet).Path("/{id}").HandlerFunc(c.get)
	r.Methods(http.MethodGet).HandlerFunc(c.list)

	// PUT requests.
	r.Methods(http.MethodPut).Path("/{id}").HandlerFunc(c.update)

	// DELETE requests.
	r.Methods(http.MethodDelete).Path("/{id}").HandlerFunc(c.delete)
}

func (c *controller) create(w http.ResponseWriter, r *http.Request) {
	t := model.Member{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.m.MemberModel.Create(r.Context(), &t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *controller) list(w http.ResponseWriter, r *http.Request) {
	res, err := c.m.MemberModel.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *controller) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := c.m.MemberModel.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *controller) update(w http.ResponseWriter, r *http.Request) {
	t := model.Member{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t.ID, err = strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.m.MemberModel.Update(r.Context(), t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with new database entry.
	err = json.NewEncoder(w).Encode(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *controller) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.m.MemberModel.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
