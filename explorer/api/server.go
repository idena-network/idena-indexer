package api

import (
	"github.com/gorilla/mux"
	"github.com/idena-network/idena-indexer/core/server"
	"github.com/idena-network/idena-indexer/explorer/db"
	"github.com/idena-network/idena-indexer/log"
	"net/http"
	"strings"
	"time"
)

type Server interface {
	Start()
	InitRouter(router *mux.Router)
}

func NewServer(port int, latestHours int, db db.Accessor, logger log.Logger) Server {
	return &httpServer{
		port:        port,
		db:          db,
		log:         logger,
		latestHours: latestHours,
	}
}

type httpServer struct {
	port        int
	latestHours int
	db          db.Accessor
	log         log.Logger
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
	// todo Currently indexer starts its own server for explorer
	//headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	//originsOk := handlers.AllowedOrigins([]string{"*"})
	//methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	//err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), handlers.CORS(originsOk, headersOk, methodsOk)(s.initHandler()))
	//if err != nil {
	//	panic(err)
	//}
}

func (s *httpServer) initHandler() http.Handler {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	s.InitRouter(api)
	return s.requestFilter(r)
}

func (s *httpServer) InitRouter(router *mux.Router) {
	router.Path(strings.ToLower("/Search")).
		Queries("value", "{value}").
		HandlerFunc(s.search)

	router.Path(strings.ToLower("/Coins")).
		HandlerFunc(s.coins)

	router.Path(strings.ToLower("/Epochs/Count")).HandlerFunc(s.epochsCount)
	router.Path(strings.ToLower("/Epochs")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochs)

	router.Path(strings.ToLower("/Epoch/Last")).HandlerFunc(s.lastEpoch)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}")).HandlerFunc(s.epoch)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Blocks/Count")).
		HandlerFunc(s.epochBlocksCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Blocks")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochBlocks)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Flips/Count")).
		HandlerFunc(s.epochFlipsCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Flips")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochFlips)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/FlipAnswersSummary")).HandlerFunc(s.epochFlipAnswersSummary)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/FlipStatesSummary")).HandlerFunc(s.epochFlipStatesSummary)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/FlipWrongWordsSummary")).HandlerFunc(s.epochFlipWrongWordsSummary)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identities/Count")).
		HandlerFunc(s.epochIdentitiesCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identities/Count")).
		Queries("state", "{state}").
		HandlerFunc(s.epochIdentitiesCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identities")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochIdentities)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/IdentityStatesSummary")).HandlerFunc(s.epochIdentityStatesSummary)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/InvitesSummary")).HandlerFunc(s.epochInvitesSummary)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Invites/Count")).
		HandlerFunc(s.epochInvitesCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Invites")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochInvites)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Txs/Count")).
		HandlerFunc(s.epochTxsCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Txs")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochTxs)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Coins")).
		HandlerFunc(s.epochCoins)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/RewardsSummary")).HandlerFunc(s.epochRewardsSummary)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Authors/Bad/Count")).HandlerFunc(s.epochBadAuthorsCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Authors/Bad")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochBadAuthors)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Authors/Good/Count")).HandlerFunc(s.epochGoodAuthorsCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Authors/Good")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochGoodAuthors)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Rewards/Count")).HandlerFunc(s.epochRewardsCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Rewards")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochRewards)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/IdentityRewards/Count")).HandlerFunc(s.epochIdentitiesRewardsCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/IdentityRewards")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochIdentitiesRewards)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/FundPayments")).HandlerFunc(s.epochFundPayments)

	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}")).HandlerFunc(s.epochIdentity)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/FlipsToSolve/Short")).
		HandlerFunc(s.epochIdentityShortFlipsToSolve)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/FlipsToSolve/Long")).
		HandlerFunc(s.epochIdentityLongFlipsToSolve)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/Answers/Short")).
		HandlerFunc(s.epochIdentityShortAnswes)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/Answers/Long")).
		HandlerFunc(s.epochIdentityLongAnswers)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/Flips")).HandlerFunc(s.epochIdentityFlips)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/ValidationTxs")).
		HandlerFunc(s.epochIdentityValidationTxs)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/Rewards")).
		HandlerFunc(s.epochIdentityRewards)

	router.Path(strings.ToLower("/Block/{id}")).HandlerFunc(s.block)
	router.Path(strings.ToLower("/Block/{id}/Txs/Count")).HandlerFunc(s.blockTxsCount)
	router.Path(strings.ToLower("/Block/{id}/Txs")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.blockTxs)
	router.Path(strings.ToLower("/Block/{id}/Coins")).
		HandlerFunc(s.blockCoins)

	router.Path(strings.ToLower("/Identity/{address}")).HandlerFunc(s.identity)
	router.Path(strings.ToLower("/Identity/{address}/Age")).HandlerFunc(s.identityAge)
	router.Path(strings.ToLower("/Identity/{address}/CurrentFlipCids")).HandlerFunc(s.identityCurrentFlipCids)
	router.Path(strings.ToLower("/Identity/{address}/Epochs/Count")).HandlerFunc(s.identityEpochsCount)
	router.Path(strings.ToLower("/Identity/{address}/Epochs")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityEpochs)
	router.Path(strings.ToLower("/Identity/{address}/FlipStates")).HandlerFunc(s.identityFlipStates)
	router.Path(strings.ToLower("/Identity/{address}/FlipQualifiedAnswers")).HandlerFunc(s.identityFlipRightAnswers)
	router.Path(strings.ToLower("/Identity/{address}/Invites/Count")).HandlerFunc(s.identityInvitesCount)
	router.Path(strings.ToLower("/Identity/{address}/Invites")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityInvites)
	router.Path(strings.ToLower("/Identity/{address}/Rewards/Count")).HandlerFunc(s.identityRewardsCount)
	router.Path(strings.ToLower("/Identity/{address}/Rewards")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityRewards)
	router.Path(strings.ToLower("/Identity/{address}/EpochRewards/Count")).HandlerFunc(s.identityEpochRewardsCount)
	router.Path(strings.ToLower("/Identity/{address}/EpochRewards")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityEpochRewards)

	router.Path(strings.ToLower("/Flip/{hash}")).HandlerFunc(s.flip)
	router.Path(strings.ToLower("/Flip/{hash}/Content")).HandlerFunc(s.flipContent)
	router.Path(strings.ToLower("/Flip/{hash}/Answers/Short/Count")).HandlerFunc(s.flipShortAnswersCount)
	router.Path(strings.ToLower("/Flip/{hash}/Answers/Short")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.flipShortAnswers)
	router.Path(strings.ToLower("/Flip/{hash}/Answers/Long/Count")).HandlerFunc(s.flipLongAnswersCount)
	router.Path(strings.ToLower("/Flip/{hash}/Answers/Long")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.flipLongAnswers)

	router.Path(strings.ToLower("/Transaction/{hash}")).HandlerFunc(s.transaction)

	router.Path(strings.ToLower("/Address/{address}")).HandlerFunc(s.address)
	router.Path(strings.ToLower("/Address/{address}/Txs/Count")).HandlerFunc(s.identityTxsCount)
	router.Path(strings.ToLower("/Address/{address}/Txs")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityTxs)
	router.Path(strings.ToLower("/Address/{address}/Penalties/Count")).HandlerFunc(s.addressPenaltiesCount)
	router.Path(strings.ToLower("/Address/{address}/Penalties")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.addressPenalties)
	router.Path(strings.ToLower("/Address/{address}/Flips/Count")).HandlerFunc(s.identityFlipsCount)
	router.Path(strings.ToLower("/Address/{address}/Flips")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityFlips)
	router.Path(strings.ToLower("/Address/{address}/MiningRewards/Count")).HandlerFunc(s.addressMiningRewardsCount)
	router.Path(strings.ToLower("/Address/{address}/MiningRewards")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.addressMiningRewards)
	router.Path(strings.ToLower("/Address/{address}/BlockMiningRewards/Count")).
		HandlerFunc(s.addressBlockMiningRewardsCount)
	router.Path(strings.ToLower("/Address/{address}/BlockMiningRewards")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.addressBlockMiningRewards)
	router.Path(strings.ToLower("/Address/{address}/States/Count")).
		HandlerFunc(s.addressStatesCount)
	router.Path(strings.ToLower("/Address/{address}/States")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.addressStates)
	router.Path(strings.ToLower("/Address/{address}/TotalLatestMiningReward")).
		HandlerFunc(s.addressTotalLatestMiningReward)
	router.Path(strings.ToLower("/Address/{address}/TotalLatestBurntCoins")).
		HandlerFunc(s.addressTotalLatestBurntCoins)
	router.Path(strings.ToLower("/Address/{address}/Authors/Bad/Count")).HandlerFunc(s.addressBadAuthorsCount)
	router.Path(strings.ToLower("/Address/{address}/Authors/Bad")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.addressBadAuthors)

	router.Path(strings.ToLower("/Balances/Count")).HandlerFunc(s.balancesCount)
	router.Path(strings.ToLower("/Balances")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.balances)

	router.Path(strings.ToLower("/TotalLatestMiningRewards/Count")).HandlerFunc(s.totalLatestMiningRewardsCount)
	router.Path(strings.ToLower("/TotalLatestMiningRewards")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.totalLatestMiningRewards)

	router.Path(strings.ToLower("/TotalLatestBurntCoins/Count")).HandlerFunc(s.totalLatestBurntCoinsCount)
	router.Path(strings.ToLower("/TotalLatestBurntCoins")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.totalLatestBurntCoins)
}

func (s *httpServer) search(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.Search(mux.Vars(r)["value"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) coins(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.Coins()
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochsCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.EpochsCount()
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochs(w http.ResponseWriter, r *http.Request) {
	startIndex, count, err := server.ReadPaginatorParams(mux.Vars(r))
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.Epochs(startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) lastEpoch(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.LastEpoch()
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epoch(w http.ResponseWriter, r *http.Request) {
	epoch, err := server.ToUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.Epoch(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochBlocksCount(w http.ResponseWriter, r *http.Request) {
	epoch, err := server.ToUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochBlocksCount(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochBlocks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochBlocks(epoch, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochFlipsCount(w http.ResponseWriter, r *http.Request) {
	epoch, err := server.ToUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochFlipsCount(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochFlips(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochFlips(epoch, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochFlipAnswersSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochFlipAnswersSummary(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochFlipStatesSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochFlipStatesSummary(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochFlipWrongWordsSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochFlipWrongWordsSummary(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochIdentitiesCount(w http.ResponseWriter, r *http.Request) {
	epoch, err := server.ToUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochIdentitiesCount(epoch, convertStates(r.Form["states[]"]))
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochIdentities(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochIdentities(epoch, convertStates(r.Form["states[]"]), startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func convertStates(formValues []string) []string {
	if len(formValues) == 0 {
		return nil
	}
	var res []string
	for _, formValue := range formValues {
		states := strings.Split(formValue, ",")
		for _, state := range states {
			if len(state) == 0 {
				continue
			}
			res = append(res, strings.ToUpper(state[0:1])+strings.ToLower(state[1:]))
		}
	}
	return res
}

func (s *httpServer) epochIdentityStatesSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochIdentityStatesSummary(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochInvitesSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochInvitesSummary(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochInvitesCount(w http.ResponseWriter, r *http.Request) {
	epoch, err := server.ToUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochInvitesCount(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochInvites(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochInvites(epoch, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochTxsCount(w http.ResponseWriter, r *http.Request) {
	epoch, err := server.ToUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochTxsCount(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochTxs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochTxs(epoch, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochCoins(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochCoins(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochRewardsSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochRewardsSummary(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochBadAuthorsCount(w http.ResponseWriter, r *http.Request) {
	epoch, err := server.ToUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochBadAuthorsCount(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochBadAuthors(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochBadAuthors(epoch, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochGoodAuthorsCount(w http.ResponseWriter, r *http.Request) {
	epoch, err := server.ToUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochGoodAuthorsCount(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochGoodAuthors(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochGoodAuthors(epoch, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochRewardsCount(w http.ResponseWriter, r *http.Request) {
	epoch, err := server.ToUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochRewardsCount(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochRewards(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochRewards(epoch, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochIdentitiesRewardsCount(w http.ResponseWriter, r *http.Request) {
	epoch, err := server.ToUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochIdentitiesRewardsCount(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochIdentitiesRewards(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochIdentitiesRewards(epoch, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochFundPayments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochFundPayments(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochIdentity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochIdentity(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochIdentityShortFlipsToSolve(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochIdentityShortFlipsToSolve(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochIdentityLongFlipsToSolve(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochIdentityLongFlipsToSolve(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochIdentityShortAnswes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochIdentityShortAnswers(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochIdentityLongAnswers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochIdentityLongAnswers(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochIdentityFlips(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochIdentityFlips(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochIdentityValidationTxs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochIdentityValidationTxs(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) epochIdentityRewards(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epoch, err := server.ToUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochIdentityRewards(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) block(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	vars := mux.Vars(r)
	height, err := server.ToUint(vars, "id")
	if err != nil {
		resp, err = s.db.BlockByHash(vars["id"])
	} else {
		resp, err = s.db.BlockByHeight(height)
	}
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) blockTxsCount(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	vars := mux.Vars(r)
	height, err := server.ToUint(vars, "id")
	if err != nil {
		resp, err = s.db.BlockTxsCountByHash(vars["id"])
	} else {
		resp, err = s.db.BlockTxsCountByHeight(height)
	}
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) blockTxs(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	vars := mux.Vars(r)
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	height, err := server.ToUint(vars, "id")
	if err != nil {
		resp, err = s.db.BlockTxsByHash(vars["id"], startIndex, count)
	} else {
		resp, err = s.db.BlockTxsByHeight(height, startIndex, count)
	}
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) blockCoins(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	vars := mux.Vars(r)
	height, err := server.ToUint(vars, "id")
	if err != nil {
		resp, err = s.db.BlockCoinsByHash(vars["id"])
	} else {
		resp, err = s.db.BlockCoinsByHeight(height)
	}
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identity(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.Identity(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityAge(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityAge(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityCurrentFlipCids(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityCurrentFlipCids(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityEpochsCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityEpochsCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityEpochs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.IdentityEpochs(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityFlipsCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityFlipsCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityFlips(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.IdentityFlips(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityFlipStates(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityFlipStates(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityFlipRightAnswers(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityFlipQualifiedAnswers(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityInvitesCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityInvitesCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityInvites(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.IdentityInvites(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityTxsCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityTxsCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityTxs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.IdentityTxs(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityRewardsCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityRewardsCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityRewards(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.IdentityRewards(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityEpochRewardsCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.IdentityEpochRewardsCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) identityEpochRewards(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.IdentityEpochRewards(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) flip(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.Flip(mux.Vars(r)["hash"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) flipContent(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.FlipContent(mux.Vars(r)["hash"])
	server.WriteResponse(w, resp, err, s.log)
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
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) flipAnswers(w http.ResponseWriter, r *http.Request, isShort bool) {
	vars := mux.Vars(r)
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.FlipAnswers(vars["hash"], isShort, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) address(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.Address(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressPenaltiesCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.AddressPenaltiesCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressPenalties(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.AddressPenalties(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressMiningRewardsCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.AddressMiningRewardsCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressMiningRewards(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.AddressMiningRewards(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressBlockMiningRewardsCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.AddressBlockMiningRewardsCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressBlockMiningRewards(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.AddressBlockMiningRewards(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressStatesCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.AddressStatesCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressStates(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.AddressStates(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressTotalLatestMiningReward(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.AddressTotalLatestMiningReward(s.getOffsetUTC(), mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressTotalLatestBurntCoins(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.AddressTotalLatestBurntCoins(s.getOffsetUTC(), mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressBadAuthorsCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.AddressBadAuthorsCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressBadAuthors(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startIndex, count, err := server.ReadPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.AddressBadAuthors(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) transaction(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.Transaction(mux.Vars(r)["hash"])
	server.WriteResponse(w, resp, err, s.log)

}

func (s *httpServer) balancesCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.BalancesCount()
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) balances(w http.ResponseWriter, r *http.Request) {
	startIndex, count, err := server.ReadPaginatorParams(mux.Vars(r))
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.Balances(startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) totalLatestMiningRewardsCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.TotalLatestMiningRewardsCount(s.getOffsetUTC())
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) totalLatestMiningRewards(w http.ResponseWriter, r *http.Request) {
	startIndex, count, err := server.ReadPaginatorParams(mux.Vars(r))
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.TotalLatestMiningRewards(s.getOffsetUTC(), startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) totalLatestBurntCoinsCount(w http.ResponseWriter, r *http.Request) {
	resp, err := s.db.TotalLatestBurntCoinsCount(s.getOffsetUTC())
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) totalLatestBurntCoins(w http.ResponseWriter, r *http.Request) {
	startIndex, count, err := server.ReadPaginatorParams(mux.Vars(r))
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.TotalLatestBurntCoins(s.getOffsetUTC(), startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) getOffsetUTC() time.Time {
	return time.Now().UTC().Add(-time.Hour * time.Duration(s.latestHours))
}
