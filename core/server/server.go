package server

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/idena-network/idena-indexer/docs"
	"github.com/idena-network/idena-indexer/explorer/config"
	"github.com/idena-network/idena-indexer/log"
	"github.com/patrickmn/go-cache"
	httpSwagger "github.com/swaggo/http-swagger"
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
	reqsPerMinuteLimit int,
	description []byte,
) *Server {
	return &Server{
		port:        port,
		description: description,
		limiter: &reqLimiter{
			queue:               make(chan struct{}, maxReqCount),
			adjacentDataQueue:   make(chan struct{}, 1),
			timeout:             timeout,
			reqCountsByClientId: cache.New(time.Second*30, time.Minute*5),
			reqLimit:            reqsPerMinuteLimit / 2,
		},
		log: logger,
	}
}

type Server struct {
	port        int
	counter     int
	limiter     *reqLimiter
	log         log.Logger
	mutex       sync.Mutex
	description []byte
}

func (s *Server) Start(swaggerConfig config.SwaggerConfig, routerInitializers ...RouterInitializer) {
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()

	apiRouter.Path("").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(s.description)
		})

	for _, ri := range routerInitializers {
		ri.InitRouter(apiRouter)
	}
	if swaggerConfig.Enabled {
		docs.SwaggerInfo.Title = "Idena indexer API"
		docs.SwaggerInfo.Version = "0.1.0"
		docs.SwaggerInfo.Host = swaggerConfig.Host
		docs.SwaggerInfo.BasePath = swaggerConfig.BasePath
		apiRouter.PathPrefix("/swagger").Handler(httpSwagger.Handler(
			httpSwagger.URL("/api/swagger/doc.json"),
		))
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
		ip := GetIP(r)
		s.log.Debug("Got api request", "reqId", reqId, "url", urlToLog, "from", ip)
		if err := s.limiter.takeResource(ip, lowerUrlPath); err != nil {
			s.log.Error("Unable to handle API request", "reqId", reqId, "err", err)
			switch err {
			case errTimeout:
				w.WriteHeader(http.StatusServiceUnavailable)
				break
			case errReqLimitExceed:
				w.WriteHeader(http.StatusTooManyRequests)
				break
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
			WriteErrorResponse(w, err, s.log)
			return
		}
		defer s.limiter.releaseResource(lowerUrlPath)

		err := r.ParseForm()
		if err != nil {
			s.log.Error("Unable to parse API request", "reqId", reqId, "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer s.log.Debug("Completed api request", "reqId", reqId)
		for name, value := range r.Form {
			r.Form[strings.ToLower(name)] = value
		}
		r.URL.Path = strings.ToLower(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
