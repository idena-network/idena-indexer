package server

import (
	"github.com/gorilla/mux"
	"github.com/idena-network/idena-indexer/core/api"
	"github.com/idena-network/idena-indexer/log"
	"net/http"
	"strings"
	"time"
)

type RouterInitializer interface {
	InitRouter(router *mux.Router)
}

type routerInitializer struct {
	api    *api.Api
	logger log.Logger
}

func NewRouterInitializer(api *api.Api, logger log.Logger) RouterInitializer {
	return &routerInitializer{
		api:    api,
		logger: logger,
	}
}

func (ri *routerInitializer) InitRouter(router *mux.Router) {
	router.Path(strings.ToLower("/OnlineIdentities/Count")).HandlerFunc(ri.onlineIdentitiesCount)
	router.Path(strings.ToLower("/OnlineIdentities")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(ri.onlineIdentitiesOld)
	router.Path(strings.ToLower("/OnlineIdentities")).
		HandlerFunc(ri.onlineIdentities)

	router.Path(strings.ToLower("/OnlineIdentity/{address}")).HandlerFunc(ri.onlineIdentity)

	router.Path(strings.ToLower("/OnlineMiners/Count")).HandlerFunc(ri.onlineCount)

	router.Path(strings.ToLower("/SignatureAddress")).
		Queries("value", "{value}", "signature", "{signature}").
		HandlerFunc(ri.signatureAddress)

	router.Path(strings.ToLower("/UpgradeVoting")).HandlerFunc(ri.upgradeVoting)

	router.Path(strings.ToLower("/Now")).HandlerFunc(ri.now)
}

func (ri *routerInitializer) onlineIdentitiesCount(w http.ResponseWriter, r *http.Request) {
	resp := ri.api.GetOnlineIdentitiesCount()
	WriteResponse(w, resp, nil, ri.logger)
}

func (ri *routerInitializer) onlineIdentities(w http.ResponseWriter, r *http.Request) {
	count, continuationToken, err := ReadPaginatorParams(r.Form)
	if err != nil {
		WriteErrorResponse(w, err, ri.logger)
		return
	}
	resp, nextContinuationToken, err := ri.api.GetOnlineIdentities(count, continuationToken)
	WriteResponsePage(w, resp, nextContinuationToken, err, ri.logger)
}

// Deprecated
func (ri *routerInitializer) onlineIdentitiesOld(w http.ResponseWriter, r *http.Request) {
	startIndex, count, err := ReadOldPaginatorParams(mux.Vars(r))
	if err != nil {
		WriteErrorResponse(w, err, ri.logger)
		return
	}
	resp := ri.api.GetOnlineIdentitiesOld(startIndex, count)
	WriteResponse(w, resp, nil, ri.logger)
}

func (ri *routerInitializer) onlineIdentity(w http.ResponseWriter, r *http.Request) {
	resp := ri.api.GetOnlineIdentity(mux.Vars(r)["address"])
	WriteResponse(w, resp, nil, ri.logger)
}

func (ri *routerInitializer) onlineCount(w http.ResponseWriter, r *http.Request) {
	resp := ri.api.GetOnlineCount()
	WriteResponse(w, resp, nil, ri.logger)
}

func (ri *routerInitializer) signatureAddress(w http.ResponseWriter, r *http.Request) {
	value := mux.Vars(r)["value"]
	signature := mux.Vars(r)["signature"]
	resp, err := ri.api.SignatureAddress(value, signature)
	WriteResponse(w, resp, err, ri.logger)
}

func (ri *routerInitializer) upgradeVoting(w http.ResponseWriter, r *http.Request) {
	resp := ri.api.UpgradeVoting()
	WriteResponse(w, resp, nil, ri.logger)
}

func (ri *routerInitializer) now(w http.ResponseWriter, r *http.Request) {
	WriteResponse(w, time.Now().UTC().Truncate(time.Second), nil, ri.logger)
}
