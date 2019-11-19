package server

import (
	"github.com/gorilla/mux"
	"github.com/idena-network/idena-indexer/core/api"
	"github.com/idena-network/idena-indexer/log"
	"net/http"
	"strings"
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
		HandlerFunc(ri.onlineIdentities)

	router.Path(strings.ToLower("/OnlineIdentity/{address}")).HandlerFunc(ri.onlineIdentity)

	router.Path(strings.ToLower("/OnlineMiners/Count")).HandlerFunc(ri.onlineCount)
}

func (ri *routerInitializer) onlineIdentitiesCount(w http.ResponseWriter, r *http.Request) {
	resp := ri.api.GetOnlineIdentitiesCount()
	WriteResponse(w, resp, nil, ri.logger)
}

func (ri *routerInitializer) onlineIdentities(w http.ResponseWriter, r *http.Request) {
	startIndex, count, err := ReadPaginatorParams(mux.Vars(r))
	if err != nil {
		WriteErrorResponse(w, err, ri.logger)
		return
	}
	resp := ri.api.GetOnlineIdentities(startIndex, count)
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
