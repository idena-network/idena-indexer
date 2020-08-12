package server

import (
	"encoding/json"
	"fmt"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// mock type for swagger
type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  *RespError  `json:"error,omitempty"`
} // @Name Response

type ResponsePage struct {
	Result            interface{} `json:"result,omitempty"`
	ContinuationToken *string     `json:"continuationToken,omitempty"`
	Error             *RespError  `json:"error,omitempty"`
} // @Name ResponsePage

type RespError struct {
	Message string `json:"message"`
} // @Name Error

func WriteErrorResponse(w http.ResponseWriter, err error, logger log.Logger) {
	WriteResponse(w, nil, err, logger)
}

func WriteResponse(w http.ResponseWriter, result interface{}, err error, logger log.Logger) {
	WriteResponsePage(w, result, nil, err, logger)
}

func WriteResponsePage(w http.ResponseWriter, result interface{}, continuationToken *string, err error, logger log.Logger) {
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(getResponse(result, continuationToken, err))
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to write API response: %v", err))
		return
	}
}

func getResponse(result interface{}, continuationToken *string, err error) ResponsePage {
	if err != nil {
		return getErrorResponse(err)
	}
	return ResponsePage{
		Result:            result,
		ContinuationToken: continuationToken,
	}
}

func getErrorResponse(err error) ResponsePage {
	return getErrorMsgResponse(err.Error())
}

func getErrorMsgResponse(errMsg string) ResponsePage {
	return ResponsePage{
		Error: &RespError{
			Message: errMsg,
		},
	}
}

func ReadUint(vars map[string]string, name string) (uint64, error) {
	value, err := strconv.ParseUint(vars[name], 10, 64)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("wrong value %s=%v", name, vars[name]))
	}
	return value, nil
}

// Deprecated
func ReadOldPaginatorParams(vars map[string]string) (uint64, uint64, error) {
	startIndex, err := ReadUint(vars, "skip")
	if err != nil {
		return 0, 0, err
	}
	count, err := ReadUint(vars, "limit")
	if err != nil {
		return 0, 0, err
	}
	if count > 100 {
		return 0, 0, errors.Errorf("too big value limit=%d", count)
	}
	return startIndex, count, nil
}

func ReadPaginatorParams(params url.Values) (uint64, *string, error) {
	var continuationToken *string
	if v := params.Get("continuationtoken"); len(v) > 0 {
		continuationToken = &v
	}
	count, err := ReadUintUrlValue(params, "limit")
	if err != nil {
		return 0, nil, err
	}
	if count > 100 {
		return 0, nil, errors.Errorf("too big value limit=%d", count)
	}
	return count, continuationToken, nil
}

func ReadUintUrlValue(params url.Values, name string) (uint64, error) {
	value, err := strconv.ParseUint(params.Get(name), 10, 64)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("wrong value %s=%v", name, params.Get(name)))
	}
	return value, nil
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

func WriteTextPlainResponse(w http.ResponseWriter, result string, err error, logger log.Logger) {
	var bytes []byte
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		bytes = []byte(err.Error())
	} else {
		bytes = []byte(result)
	}
	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write(bytes); err != nil {
		logger.Error(fmt.Sprintf("Unable to write API response: %v", err))
	}
}
