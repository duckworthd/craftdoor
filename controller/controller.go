package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jpillora/ipfilter"
	"github.com/pakohan/craftdoor/config"
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
func New(cfg *config.Config, m model.Model, s *service.Service) http.Handler {
	r := mux.NewRouter()

	// Filter
	var handler http.Handler = handlers.CORS(
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
	)(r)

	// Filter by IP address.
	handler = ipfilter.Wrap(handler, ipfilter.Options{
		AllowedIPs:     []string{"192.168.0.0/24", "10.0.0.0/16"},
		BlockByDefault: true,
	})

	c := &controller{
		m:       m,
		s:       s,
		Handler: handler,
	}
	r.Path("/api").Methods(http.MethodGet).HandlerFunc(c.ReadNextTag)
	members.New(r.PathPrefix("/api/members").Subrouter(), m)
	keys.New(r.PathPrefix("/api/keys").Subrouter(), m, s)

	// Assume everything other route is a static asset.
	//
	// TODO(duckworthd): The webapp changes the URL when switching between tabs,
	// but refreshing the page results in a 404. Figure out how to fix this.
	fileServerPath := cfg.StaticAssetsDir
	fileServer := http.FileServer(http.Dir(fileServerPath))
	log.Printf("Serving static files from %s", fileServerPath)
	r.PathPrefix("/").Methods(http.MethodGet).Handler(fileServer)

	return c
}

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
