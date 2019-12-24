package server

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/idena-network/idena-indexer/log"
	"net/http"
	"strings"
	"sync"
)

func NewServer(port int, logger log.Logger) *Server {
	return &Server{
		port: port,
		log:  logger,
	}
}

type Server struct {
	port    int
	log     log.Logger
	counter int
	mutex   sync.Mutex
}

func (s *Server) Start(routerInitializers ...RouterInitializer) {
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	for _, ri := range routerInitializers {
		ri.InitRouter(apiRouter)
	}

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port),
		handlers.CORS(originsOk, headersOk, methodsOk)(s.requestFilter(apiRouter)))
	if err != nil {
		panic(err)
	}
}

func (s *Server) generateReqId() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	id := s.counter
	s.counter++
	return id
}

func (s *Server) requestFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqId := s.generateReqId()
		s.log.Debug("Got api request", "reqId", reqId, "url", r.URL, "from", r.RemoteAddr)
		defer s.log.Debug("Completed api request", "reqId", reqId)
		err := r.ParseForm()
		if err != nil {
			s.log.Error("Unable to parse API request", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		r.URL.Path = strings.ToLower(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
