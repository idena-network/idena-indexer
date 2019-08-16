package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
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
		port: port,
		db:   db,
		log:  logger,
	}
	return server
}

type httpServer struct {
	port int
	db   db.Accessor
	log  log.Logger
}

type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  *RespError  `json:"error,omitempty"`
}

type RespError struct {
	Message string `json:"message"`
}

func (s *httpServer) requestFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.log.Debug("Got api request", "url", r.URL, "from", r.RemoteAddr)
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

func (s *httpServer) Start() {
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.initHandler())
	if err != nil {
		panic(err)
	}
}

func (s *httpServer) initHandler() http.Handler {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	api.Path(strings.ToLower("/Search")).
		Queries("value", "{value}").
		HandlerFunc(s.search)

	api.Path(strings.ToLower("/Coins")).
		HandlerFunc(s.coins)

	api.Path(strings.ToLower("/Epochs/Count")).HandlerFunc(s.epochsCount)
	api.Path(strings.ToLower("/Epochs")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochs)

	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}")).HandlerFunc(s.epoch)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Blocks/Count")).
		HandlerFunc(s.epochBlocksCount)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Blocks")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochBlocks)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Flips/Count")).
		HandlerFunc(s.epochFlipsCount)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Flips")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochFlips)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/FlipAnswersSummary")).HandlerFunc(s.epochFlipAnswersSummary)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/FlipStatesSummary")).HandlerFunc(s.epochFlipStatesSummary)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identities/Count")).
		HandlerFunc(s.epochIdentitiesCount)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identities")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochIdentities)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/IdentityStatesSummary")).HandlerFunc(s.epochIdentityStatesSummary)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/InvitesSummary")).HandlerFunc(s.epochInvitesSummary)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Invites/Count")).
		HandlerFunc(s.epochInvitesCount)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Invites")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochInvites)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Txs/Count")).
		HandlerFunc(s.epochTxsCount)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Txs")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochTxs)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Coins")).
		HandlerFunc(s.epochCoins)

	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}")).HandlerFunc(s.epochIdentity)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/FlipsToSolve/Short")).HandlerFunc(s.epochIdentityShortFlipsToSolve)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/FlipsToSolve/Long")).HandlerFunc(s.epochIdentityLongFlipsToSolve)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/Answers/Short")).HandlerFunc(s.epochIdentityShortAnswes)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/Answers/Long")).HandlerFunc(s.epochIdentityLongAnswers)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/Flips")).HandlerFunc(s.epochIdentityFlips)
	api.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/ValidationTxs")).HandlerFunc(s.epochIdentityValidationTxs)

	api.Path(strings.ToLower("/Block/{id}")).HandlerFunc(s.block)
	api.Path(strings.ToLower("/Block/{id}/Txs/Count")).HandlerFunc(s.blockTxsCount)
	api.Path(strings.ToLower("/Block/{id}/Txs")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.blockTxs)
	api.Path(strings.ToLower("/Block/{id}/Coins")).
		HandlerFunc(s.blockCoins)

	api.Path(strings.ToLower("/Identity/{address}")).HandlerFunc(s.identity)
	api.Path(strings.ToLower("/Identity/{address}/Age")).HandlerFunc(s.identityAge)
	api.Path(strings.ToLower("/Identity/{address}/CurrentFlipCids")).HandlerFunc(s.identityCurrentFlipCids)
	api.Path(strings.ToLower("/Identity/{address}/Epochs/Count")).HandlerFunc(s.identityEpochsCount)
	api.Path(strings.ToLower("/Identity/{address}/Epochs")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityEpochs)
	api.Path(strings.ToLower("/Identity/{address}/FlipStates")).HandlerFunc(s.identityFlipStates)
	api.Path(strings.ToLower("/Identity/{address}/FlipQualifiedAnswers")).HandlerFunc(s.identityFlipRightAnswers)
	api.Path(strings.ToLower("/Identity/{address}/Invites/Count")).HandlerFunc(s.identityInvitesCount)
	api.Path(strings.ToLower("/Identity/{address}/Invites")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityInvites)
	api.Path(strings.ToLower("/Identity/{address}/States/Count")).HandlerFunc(s.identityStatesCount)
	api.Path(strings.ToLower("/Identity/{address}/States")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityStates)

	api.Path(strings.ToLower("/Flip/{hash}")).HandlerFunc(s.flip)
	api.Path(strings.ToLower("/Flip/{hash}/Content")).HandlerFunc(s.flipContent)
	api.Path(strings.ToLower("/Flip/{hash}/Answers/Short/Count")).HandlerFunc(s.flipShortAnswersCount)
	api.Path(strings.ToLower("/Flip/{hash}/Answers/Short")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.flipShortAnswers)
	api.Path(strings.ToLower("/Flip/{hash}/Answers/Long/Count")).HandlerFunc(s.flipLongAnswersCount)
	api.Path(strings.ToLower("/Flip/{hash}/Answers/Long")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.flipLongAnswers)

	api.Path(strings.ToLower("/Transaction/{hash}")).HandlerFunc(s.transaction)

	api.Path(strings.ToLower("/Address/{address}")).HandlerFunc(s.address)
	api.Path(strings.ToLower("/Address/{address}/Txs/Count")).HandlerFunc(s.identityTxsCount)
	api.Path(strings.ToLower("/Address/{address}/Txs")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityTxs)

	api.Path(strings.ToLower("/Balances/Count")).HandlerFunc(s.balancesCount)
	api.Path(strings.ToLower("/Balances")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.balances)

	return s.requestFilter(r)
}

func (s *httpServer) search(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.Search(mux.Vars(r)["value"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) coins(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.Coins()
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochsCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.EpochsCount()
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochs(w http.ResponseWriter, r *http.Request) {
	startIndex, count, err := readPaginatorParams(mux.Vars(r))
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.Epochs(startIndex, count)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epoch(w http.ResponseWriter, r *http.Request) {
	epoch, err := toUint(mux.Vars(r), "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.Epoch(epoch)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochBlocksCount(w http.ResponseWriter, r *http.Request) {
	epoch, err := toUint(mux.Vars(r), "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochBlocksCount(epoch)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochBlocks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	startIndex, count, err := readPaginatorParams(vars)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochBlocks(epoch, startIndex, count)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochFlipsCount(w http.ResponseWriter, r *http.Request) {
	epoch, err := toUint(mux.Vars(r), "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochFlipsCount(epoch)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochFlips(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	startIndex, count, err := readPaginatorParams(vars)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochFlips(epoch, startIndex, count)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochFlipAnswersSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochFlipAnswersSummary(epoch)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochFlipStatesSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochFlipStatesSummary(epoch)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochIdentitiesCount(w http.ResponseWriter, r *http.Request) {
	epoch, err := toUint(mux.Vars(r), "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochIdentitiesCount(epoch)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochIdentities(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	startIndex, count, err := readPaginatorParams(vars)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochIdentities(epoch, startIndex, count)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochIdentityStatesSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochIdentityStatesSummary(epoch)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochInvitesSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochInvitesSummary(epoch)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochInvitesCount(w http.ResponseWriter, r *http.Request) {
	epoch, err := toUint(mux.Vars(r), "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochInvitesCount(epoch)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochInvites(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	startIndex, count, err := readPaginatorParams(vars)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochInvites(epoch, startIndex, count)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochTxsCount(w http.ResponseWriter, r *http.Request) {
	epoch, err := toUint(mux.Vars(r), "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochTxsCount(epoch)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochTxs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	startIndex, count, err := readPaginatorParams(vars)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochTxs(epoch, startIndex, count)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochCoins(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochCoins(epoch)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochIdentity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochIdentity(epoch, vars["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochIdentityShortFlipsToSolve(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochIdentityShortFlipsToSolve(epoch, vars["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochIdentityLongFlipsToSolve(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochIdentityLongFlipsToSolve(epoch, vars["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochIdentityShortAnswes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochIdentityShortAnswers(epoch, vars["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochIdentityLongAnswers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochIdentityLongAnswers(epoch, vars["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochIdentityFlips(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochIdentityFlips(epoch, vars["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) epochIdentityValidationTxs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := toUint(vars, "epoch")
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.EpochIdentityValidationTxs(epoch, vars["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) block(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	vars := mux.Vars(r)
	height, err := toUint(vars, "id")
	if err != nil {
		resp, err = s.db.BlockByHash(vars["id"])
	} else {
		resp, err = s.db.BlockByHeight(height)
	}
	s.writeResponse(w, resp, err)
}

func (s *httpServer) blockTxsCount(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	vars := mux.Vars(r)
	height, err := toUint(vars, "id")
	if err != nil {
		resp, err = s.db.BlockTxsCountByHash(vars["id"])
	} else {
		resp, err = s.db.BlockTxsCountByHeight(height)
	}
	s.writeResponse(w, resp, err)
}

func (s *httpServer) blockTxs(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	vars := mux.Vars(r)
	startIndex, count, err := readPaginatorParams(vars)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	height, err := toUint(vars, "id")
	if err != nil {
		resp, err = s.db.BlockTxsByHash(vars["id"], startIndex, count)
	} else {
		resp, err = s.db.BlockTxsByHeight(height, startIndex, count)
	}
	s.writeResponse(w, resp, err)
}

func (s *httpServer) blockCoins(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	vars := mux.Vars(r)
	height, err := toUint(vars, "id")
	if err != nil {
		resp, err = s.db.BlockCoinsByHash(vars["id"])
	} else {
		resp, err = s.db.BlockCoinsByHeight(height)
	}
	s.writeResponse(w, resp, err)
}

func (s *httpServer) identity(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.Identity(mux.Vars(r)["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) identityAge(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityAge(mux.Vars(r)["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) identityCurrentFlipCids(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityCurrentFlipCids(mux.Vars(r)["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) identityEpochsCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityEpochsCount(mux.Vars(r)["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) identityEpochs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startIndex, count, err := readPaginatorParams(vars)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.IdentityEpochs(vars["address"], startIndex, count)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) identityFlipStates(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityFlipStates(mux.Vars(r)["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) identityFlipRightAnswers(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityFlipQualifiedAnswers(mux.Vars(r)["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) identityInvitesCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityInvitesCount(mux.Vars(r)["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) identityInvites(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startIndex, count, err := readPaginatorParams(vars)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.IdentityInvites(vars["address"], startIndex, count)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) identityStatesCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityStatesCount(mux.Vars(r)["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) identityStates(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startIndex, count, err := readPaginatorParams(vars)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.IdentityStates(vars["address"], startIndex, count)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) identityTxsCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityTxsCount(mux.Vars(r)["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) identityTxs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startIndex, count, err := readPaginatorParams(vars)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.IdentityTxs(vars["address"], startIndex, count)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) flip(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.Flip(mux.Vars(r)["hash"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) flipContent(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.FlipContent(mux.Vars(r)["hash"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) flipShortAnswersCount(w http.ResponseWriter, r *http.Request) {
	s.flipAnswersCount(w, r, true)
}

func (s *httpServer) flipShortAnswers(w http.ResponseWriter, r *http.Request) {
	s.flipAnswers(w, r, true)
}

func (s *httpServer) flipLongAnswersCount(w http.ResponseWriter, r *http.Request) {
	s.flipAnswersCount(w, r, false)
}

func (s *httpServer) flipLongAnswers(w http.ResponseWriter, r *http.Request) {
	s.flipAnswers(w, r, false)
}

func (s *httpServer) flipAnswersCount(w http.ResponseWriter, r *http.Request, isShort bool) {
	resp, err := s.db.FlipAnswersCount(mux.Vars(r)["hash"], isShort)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) flipAnswers(w http.ResponseWriter, r *http.Request, isShort bool) {
	vars := mux.Vars(r)
	startIndex, count, err := readPaginatorParams(vars)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.FlipAnswers(vars["hash"], isShort, startIndex, count)
	s.writeResponse(w, resp, err)
}

func (s *httpServer) address(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.Address(mux.Vars(r)["address"])
	s.writeResponse(w, resp, err)
}

func (s *httpServer) transaction(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.Transaction(mux.Vars(r)["hash"])
	s.writeResponse(w, resp, err)

}

func (s *httpServer) balancesCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.BalancesCount()
	s.writeResponse(w, resp, err)
}

func (s *httpServer) balances(w http.ResponseWriter, r *http.Request) {
	startIndex, count, err := readPaginatorParams(mux.Vars(r))
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	resp, err := s.db.Balances(startIndex, count)
	s.writeResponse(w, resp, err)
}

func getErrorMsgResponse(errMsg string) Response {
	return Response{
		Error: &RespError{
			Message: errMsg,
		},
	}
}

func (s *httpServer) writeErrorResponse(w http.ResponseWriter, err error) {
	s.writeResponse(w, nil, err)
}

func (s *httpServer) writeResponse(w http.ResponseWriter, result interface{}, err error) {
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(getResponse(result, err))
	if err != nil {
		s.log.Error(fmt.Sprintf("Unable to write API response: %v", err))
		return
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

func toUint(vars map[string]string, name string) (uint64, error) {
	value, err := strconv.ParseUint(vars[name], 10, 64)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("wrong value %s=%v", name, vars[name]))
	}
	return value, nil
}

func readPaginatorParams(vars map[string]string) (uint64, uint64, error) {
	startIndex, err := toUint(vars, "skip")
	if err != nil {
		return 0, 0, err
	}
	count, err := toUint(vars, "limit")
	if err != nil {
		return 0, 0, err
	}
	return startIndex, count, nil
}
