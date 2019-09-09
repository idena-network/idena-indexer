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
	router.Path(strings.ToLower("/Activities/Count")).HandlerFunc(ri.activitiesCount)
	router.Path(strings.ToLower("/Activities")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(ri.activities)

	router.Path(strings.ToLower("/Activity/{address}")).HandlerFunc(ri.activity)

	router.Path(strings.ToLower("/CurrentPenalties/Count")).HandlerFunc(ri.currentPenaltiesCount)
	router.Path(strings.ToLower("/CurrentPenalties")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(ri.currentPenalties)

	router.Path(strings.ToLower("/CurrentPenalty/{address}")).HandlerFunc(ri.currentPenalty)
}

func (ri *routerInitializer) activitiesCount(w http.ResponseWriter, r *http.Request) {
	resp := ri.api.GetLastActivitiesCount()
	WriteResponse(w, resp, nil, ri.logger)
}

func (ri *routerInitializer) activities(w http.ResponseWriter, r *http.Request) {
	startIndex, count, err := ReadPaginatorParams(mux.Vars(r))
	if err != nil {
		WriteErrorResponse(w, err, ri.logger)
		return
	}
	resp := ri.api.GetLastActivities(startIndex, count)
	WriteResponse(w, resp, nil, ri.logger)
}

func (ri *routerInitializer) activity(w http.ResponseWriter, r *http.Request) {
	resp := ri.api.GetLastActivity(mux.Vars(r)["address"])
	WriteResponse(w, resp, nil, ri.logger)
}

func (ri *routerInitializer) currentPenaltiesCount(w http.ResponseWriter, r *http.Request) {
	resp := ri.api.GetCurrentPenaltiesCount()
	WriteResponse(w, resp, nil, ri.logger)
}

func (ri *routerInitializer) currentPenalties(w http.ResponseWriter, r *http.Request) {
	startIndex, count, err := ReadPaginatorParams(mux.Vars(r))
	if err != nil {
		WriteErrorResponse(w, err, ri.logger)
		return
	}
	resp := ri.api.GetCurrentPenalties(startIndex, count)
	WriteResponse(w, resp, nil, ri.logger)
}

func (ri *routerInitializer) currentPenalty(w http.ResponseWriter, r *http.Request) {
	resp := ri.api.GetCurrentPenalty(mux.Vars(r)["address"])
	WriteResponse(w, resp, nil, ri.logger)
}
