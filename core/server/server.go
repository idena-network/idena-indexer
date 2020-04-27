package server

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

func NewServer(
	port int,
	maxReqCount int,
	timeout time.Duration,
	logger log.Logger,
) *Server {
	return &Server{
		port: port,
		limiter: &reqLimiter{
			queue:   make(chan struct{}, maxReqCount),
			timeout: timeout,
		},
		log: logger,
	}
}

type Server struct {
	port    int
	counter int
	limiter *reqLimiter
	log     log.Logger
	mutex   sync.Mutex
}

type reqLimiter struct {
	queue   chan struct{}
	timeout time.Duration
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
		var urlToLog *url.URL
		if !strings.Contains(strings.ToLower(r.URL.Path), "/search") {
			urlToLog = r.URL
		}
		s.log.Debug("Got api request", "reqId", reqId, "url", urlToLog, "from", GetIP(r))
		defer s.log.Debug("Completed api request", "reqId", reqId)

		if err := s.limiter.takeResource(); err != nil {
			s.log.Error("Unable to handle API request", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		defer s.limiter.releaseResource()

		err := r.ParseForm()
		if err != nil {
			s.log.Error("Unable to parse API request", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for name, value := range r.Form {
			r.Form[strings.ToLower(name)] = value
		}
		r.URL.Path = strings.ToLower(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (limiter *reqLimiter) takeResource() error {
	var ok bool
	select {
	case limiter.queue <- struct{}{}:
		ok = true
	case <-time.After(limiter.timeout):
	}
	if !ok {
		return errors.New("timeout while waiting for resource")
	}
	return nil
}

func (limiter *reqLimiter) releaseResource() {
	<-limiter.queue
}
