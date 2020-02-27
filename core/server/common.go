package server

import (
	"encoding/json"
	"fmt"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"strings"
)

type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  *RespError  `json:"error,omitempty"`
}

type RespError struct {
	Message string `json:"message"`
}

func WriteErrorResponse(w http.ResponseWriter, err error, logger log.Logger) {
	WriteResponse(w, nil, err, logger)
}

func WriteResponse(w http.ResponseWriter, result interface{}, err error, logger log.Logger) {
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(getResponse(result, err))
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to write API response: %v", err))
		return
	}
}

func getResponse(result interface{}, err error) Response {
	if err != nil {
		return getErrorResponse(err)
	}
	return Response{
		Result: result,
	}
}

func getErrorResponse(err error) Response {
	return getErrorMsgResponse(err.Error())
}

func getErrorMsgResponse(errMsg string) Response {
	return Response{
		Error: &RespError{
			Message: errMsg,
		},
	}
}

func ToUint(vars map[string]string, name string) (uint64, error) {
	value, err := strconv.ParseUint(vars[name], 10, 64)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("wrong value %s=%v", name, vars[name]))
	}
	return value, nil
}

func ReadPaginatorParams(vars map[string]string) (uint64, uint64, error) {
	startIndex, err := ToUint(vars, "skip")
	if err != nil {
		return 0, 0, err
	}
	count, err := ToUint(vars, "limit")
	if err != nil {
		return 0, 0, err
	}
	if count > 100 {
		return 0, 0, errors.Errorf("too big value limit=%d", count)
	}
	return startIndex, count, nil
}

func GetIP(r *http.Request) string {
	header := r.Header.Get("X-Forwarded-For")
	if len(header) > 0 {
		return strings.Split(header, ", ")[0]
	}
	if strings.Contains(r.RemoteAddr, ":") {
		return strings.Split(r.RemoteAddr, ":")[0]
	}
	return r.RemoteAddr
}
