package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/idena-network/idena-indexer/explorer/db"
	"github.com/idena-network/idena-indexer/log"
	"net/http"
	"strconv"
	"strings"
)

type Server interface {
	Start()
}

func NewServer(port int, db db.Accessor, logger log.Logger) Server {
	server := &httpServer{
		port:     port,
		api:      newApi(db),
		handlers: make(map[string]handler),
		log:      logger,
	}
	server.handlers[strings.ToLower("/api/Epochs")] = server.epochs
	server.handlers[strings.ToLower("/api/Epoch")] = server.epoch
	server.handlers[strings.ToLower("/api/EpochBlocks")] = server.epochBlocks
	server.handlers[strings.ToLower("/api/EpochTxs")] = server.epochTxs
	server.handlers[strings.ToLower("/api/BlockTxs")] = server.blockTxs
	server.handlers[strings.ToLower("/api/EpochFlips")] = server.epochFlips
	server.handlers[strings.ToLower("/api/EpochInvites")] = server.epochInvites
	server.handlers[strings.ToLower("/api/EpochIdentities")] = server.epochIdentities
	server.handlers[strings.ToLower("/api/Flip")] = server.flip
	server.handlers[strings.ToLower("/api/Identity")] = server.identity
	return server
}

type httpServer struct {
	port     int
	handlers map[string]handler
	api      *api
	log      log.Logger
}

type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  *RespError  `json:"error,omitempty"`
}

type RespError struct {
	Message string `json:"message"`
}

func caselessMatcher(next http.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.ToLower(r.URL.Path)
		next.ServeHTTP(w, r)
	}
}

func (s *httpServer) Start() {
	mux := http.NewServeMux()
	http.HandleFunc("/", caselessMatcher(mux))
	for path := range s.handlers {
		mux.HandleFunc(path, s.handleRequest)
	}
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
	if err != nil {
		panic(err)
	}
}

func (s *httpServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("Got api request", "r", r)
	err := r.ParseForm()
	if err != nil {
		s.log.Error(fmt.Sprintf("Unable to parse API request: %v", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path := r.URL.Path
	handler := s.handlers[strings.ToLower(path)]
	if handler == nil {
		s.log.Error(fmt.Sprintf("Theres is no API handler for path %v", path))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	resp := handler(r)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		s.log.Error(fmt.Sprintf("Unable to write API response: %v", err))
		return
	}
}

type handler func(r *http.Request) Response

func (s *httpServer) epochs(r *http.Request) Response {
	return getResponse(s.api.epochs())
}

func (s *httpServer) epoch(r *http.Request) Response {
	epoch, err := readUintParameter(r, "epoch")
	if err != nil {
		return getErrorResponse(err)
	}
	return getResponse(s.api.epoch(epoch))
}

func (s *httpServer) epochBlocks(r *http.Request) Response {
	epoch, err := readUintParameter(r, "epoch")
	if err != nil {
		return getErrorResponse(err)
	}
	return getResponse(s.api.epochBlocks(epoch))
}

func (s *httpServer) epochTxs(r *http.Request) Response {
	epoch, err := readUintParameter(r, "epoch")
	if err != nil {
		return getErrorResponse(err)
	}
	return getResponse(s.api.epochTxs(epoch))
}

func (s *httpServer) epochFlips(r *http.Request) Response {
	epoch, err := readUintParameter(r, "epoch")
	if err != nil {
		return getErrorResponse(err)
	}
	return getResponse(s.api.epochFlips(epoch))
}

func (s *httpServer) epochInvites(r *http.Request) Response {
	epoch, err := readUintParameter(r, "epoch")
	if err != nil {
		return getErrorResponse(err)
	}
	return getResponse(s.api.epochInvites(epoch))
}

func (s *httpServer) epochIdentities(r *http.Request) Response {
	epoch, err := readUintParameter(r, "epoch")
	if err != nil {
		return getErrorResponse(err)
	}
	return getResponse(s.api.epochIdentities(epoch))
}

func (s *httpServer) blockTxs(r *http.Request) Response {
	height, err := readUintParameter(r, "height")
	if err != nil {
		return getErrorResponse(err)
	}
	return getResponse(s.api.blockTxs(height))
}

func (s *httpServer) flip(r *http.Request) Response {
	hash, err := readStrParameter(r, "hash")
	if err != nil {
		return getErrorResponse(err)
	}
	return getResponse(s.api.flip(hash))
}

func (s *httpServer) identity(r *http.Request) Response {
	address, err := readStrParameter(r, "address")
	if err != nil {
		return getErrorResponse(err)
	}
	return getResponse(s.api.identity(address))
}

func getErrorMsgResponse(errMsg string) Response {
	return Response{
		Error: &RespError{
			Message: errMsg,
		},
	}
}

func getErrorResponse(err error) Response {
	return getErrorMsgResponse(err.Error())
}

func getResponse(result interface{}, err error) Response {
	if err != nil {
		return getErrorResponse(err)
	}
	return Response{
		Result: result,
	}
}

func readUintParameter(r *http.Request, name string) (uint64, error) {
	pValue, err := readStrParameter(r, name)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseUint(pValue, 10, 64)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("wrong value %s=%v", name, pValue))
	}
	return value, nil
}

func readStrParameter(r *http.Request, name string) (string, error) {
	values := r.Form[name]
	if len(values) == 0 {
		return "", errors.New(fmt.Sprintf("parameter '%s' is absent", name))
	}
	return values[0], nil
}
