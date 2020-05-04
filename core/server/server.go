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
			queue:             make(chan struct{}, maxReqCount),
			adjacentDataQueue: make(chan struct{}, 1),
			timeout:           timeout,
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
		lowerUrlPath := strings.ToLower(r.URL.Path)
		if !strings.Contains(lowerUrlPath, "/search") {
			urlToLog = r.URL
		}
		s.log.Debug("Got api request", "reqId", reqId, "url", urlToLog, "from", GetIP(r))
		defer s.log.Debug("Completed api request", "reqId", reqId)

		if err := s.limiter.takeResource(lowerUrlPath); err != nil {
			s.log.Error("Unable to handle API request", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		defer s.limiter.releaseResource(lowerUrlPath)

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

type reqLimiter struct {
	queue             chan struct{}
	adjacentDataQueue chan struct{}
	timeout           time.Duration
	mutex             sync.Mutex
}

func (limiter *reqLimiter) takeResource(lowerUrlPath string) error {
	var ok bool
	queue := limiter.getQueueByUrlPath(lowerUrlPath)
	select {
	case queue <- struct{}{}:
		ok = true
	case <-time.After(limiter.timeout):
	}
	if !ok {
		return errors.New("timeout while waiting for resource")
	}
	return nil
}

func (limiter *reqLimiter) releaseResource(lowerUrlPath string) {
	queue := limiter.getQueueByUrlPath(lowerUrlPath)
	<-queue
}

func (limiter *reqLimiter) getQueueByUrlPath(lowerUrlPath string) chan struct{} {
	if strings.Contains(lowerUrlPath, "/adjacent") {
		return limiter.adjacentDataQueue
	}
	return limiter.queue
}
