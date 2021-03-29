package api

import (
	"encoding/hex"
	"github.com/gorilla/mux"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-indexer/core/server"
	"github.com/idena-network/idena-indexer/explorer/monitoring"
	service2 "github.com/idena-network/idena-indexer/explorer/service"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Server interface {
	Start()
	InitRouter(router *mux.Router)
}

func NewServer(
	port int,
	latestHours int,
	activeAddrHours int,
	frozenBalanceAddrs []string,
	getDumpLink func() string,
	service Service,
	contractsService service2.Contracts,
	logger log.Logger,
	pm monitoring.PerformanceMonitor,
) Server {
	var lowerFrozenBalanceAddrs []string
	for _, frozenBalanceAddr := range frozenBalanceAddrs {
		lowerFrozenBalanceAddrs = append(lowerFrozenBalanceAddrs, strings.ToLower(frozenBalanceAddr))
	}
	return &httpServer{
		port:               port,
		service:            service,
		contractsService:   contractsService,
		log:                logger,
		latestHours:        latestHours,
		activeAddrHours:    activeAddrHours,
		frozenBalanceAddrs: lowerFrozenBalanceAddrs,
		getDumpLink:        getDumpLink,
		pm:                 pm,
	}
}

type httpServer struct {
	port               int
	latestHours        int
	activeAddrHours    int
	frozenBalanceAddrs []string
	service            Service
	contractsService   service2.Contracts
	log                log.Logger
	pm                 monitoring.PerformanceMonitor
	counter            int
	mutex              sync.Mutex
	getDumpLink        func() string
}

func (s *httpServer) generateReqId() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	id := s.counter
	s.counter++
	return id
}

func (s *httpServer) Start() {
	// Currently indexer starts its own server for explorer
}

func (s *httpServer) InitRouter(router *mux.Router) {
	router.Path(strings.ToLower("/DumpLink")).HandlerFunc(s.dumpLink)

	router.Path(strings.ToLower("/Search")).
		Queries("value", "{value}").
		HandlerFunc(s.search)

	router.Path(strings.ToLower("/Coins")).
		HandlerFunc(s.coins)
	router.Path(strings.ToLower("/Txt/TotalSupply")).
		HandlerFunc(s.txtTotalSupply)

	router.Path(strings.ToLower("/CirculatingSupply")).
		Queries("format", "{format}").
		HandlerFunc(s.circulatingSupply)
	router.Path(strings.ToLower("/CirculatingSupply")).
		HandlerFunc(s.circulatingSupply)

	router.Path(strings.ToLower("/Txt/CirculatingSupply")).
		HandlerFunc(s.txtCirculatingSupply)

	router.Path(strings.ToLower("/ActiveAddresses/Count")).
		HandlerFunc(s.activeAddressesCount)

	router.Path(strings.ToLower("/Upgrades")).
		HandlerFunc(s.upgrades)

	router.Path(strings.ToLower("/Epochs/Count")).HandlerFunc(s.epochsCount)
	router.Path(strings.ToLower("/Epochs")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochsOld)
	router.Path(strings.ToLower("/Epochs")).HandlerFunc(s.epochs)

	router.Path(strings.ToLower("/Epoch/Last")).HandlerFunc(s.lastEpoch)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}")).HandlerFunc(s.epoch)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Blocks/Count")).
		HandlerFunc(s.epochBlocksCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Blocks")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochBlocksOld)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Blocks")).
		HandlerFunc(s.epochBlocks)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Flips/Count")).
		HandlerFunc(s.epochFlipsCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Flips")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochFlipsOld)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Flips")).
		HandlerFunc(s.epochFlips)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/FlipAnswersSummary")).HandlerFunc(s.epochFlipAnswersSummary)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/FlipStatesSummary")).HandlerFunc(s.epochFlipStatesSummary)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/FlipWrongWordsSummary")).HandlerFunc(s.epochFlipWrongWordsSummary)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identities/Count")).
		HandlerFunc(s.epochIdentitiesCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identities")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochIdentitiesOld)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identities")).
		HandlerFunc(s.epochIdentities)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/IdentityStatesSummary")).HandlerFunc(s.epochIdentityStatesSummary)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/IdentityStatesInterimSummary")).HandlerFunc(s.epochIdentityStatesInterimSummary)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/InvitesSummary")).HandlerFunc(s.epochInvitesSummary)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/InviteStatesSummary")).HandlerFunc(s.epochInviteStatesSummary)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Invites/Count")).
		HandlerFunc(s.epochInvitesCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Invites")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochInvitesOld)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Invites")).
		HandlerFunc(s.epochInvites)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Txs/Count")).
		HandlerFunc(s.epochTxsCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Txs")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochTxsOld)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Txs")).
		HandlerFunc(s.epochTxs)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Coins")).
		HandlerFunc(s.epochCoins)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/RewardsSummary")).HandlerFunc(s.epochRewardsSummary)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Authors/Bad/Count")).HandlerFunc(s.epochBadAuthorsCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Authors/Bad")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochBadAuthorsOld)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Authors/Bad")).
		HandlerFunc(s.epochBadAuthors)
	//router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Rewards/Count")).HandlerFunc(s.epochRewardsCount)
	//router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Rewards")).
	//	Queries("skip", "{skip}", "limit", "{limit}").
	//	HandlerFunc(s.epochRewards)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/IdentityRewards/Count")).HandlerFunc(s.epochIdentitiesRewardsCount)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/IdentityRewards")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.epochIdentitiesRewardsOld)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/IdentityRewards")).
		HandlerFunc(s.epochIdentitiesRewards)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/FundPayments")).HandlerFunc(s.epochFundPayments)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/RewardBounds")).HandlerFunc(s.epochRewardBounds)

	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}")).HandlerFunc(s.epochIdentity)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/FlipsToSolve/Short")).
		HandlerFunc(s.epochIdentityShortFlipsToSolve)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/FlipsToSolve/Long")).
		HandlerFunc(s.epochIdentityLongFlipsToSolve)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/Answers/Short")).
		HandlerFunc(s.epochIdentityShortAnswers)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/Answers/Long")).
		HandlerFunc(s.epochIdentityLongAnswers)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/Flips")).HandlerFunc(s.epochIdentityFlips)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/ValidationTxs")).
		HandlerFunc(s.epochIdentityValidationTxs)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/Rewards")).
		HandlerFunc(s.epochIdentityRewards)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/RewardedFlips")).
		HandlerFunc(s.epochIdentityFlipsWithRewardFlag)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/ReportRewards")).
		HandlerFunc(s.epochIdentityReportedFlipRewards)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/Authors/Bad")).
		HandlerFunc(s.epochIdentityBadAuthor)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/RewardedInvites")).
		HandlerFunc(s.epochIdentityInvitesWithRewardFlag)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/SavedInviteRewards")).
		HandlerFunc(s.epochIdentitySavedInviteRewards)
	router.Path(strings.ToLower("/Epoch/{epoch:[0-9]+}/Identity/{address}/AvailableInvites")).
		HandlerFunc(s.epochIdentityAvailableInvites)

	router.Path(strings.ToLower("/Block/{id}")).HandlerFunc(s.block)
	router.Path(strings.ToLower("/Block/{id}/Txs/Count")).HandlerFunc(s.blockTxsCount)
	router.Path(strings.ToLower("/Block/{id}/Txs")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.blockTxsOld)
	router.Path(strings.ToLower("/Block/{id}/Txs")).
		HandlerFunc(s.blockTxs)
	router.Path(strings.ToLower("/Block/{id}/Coins")).
		HandlerFunc(s.blockCoins)

	router.Path(strings.ToLower("/Identity/{address}")).HandlerFunc(s.identity)
	router.Path(strings.ToLower("/Identity/{address}/Age")).HandlerFunc(s.identityAge)
	router.Path(strings.ToLower("/Identity/{address}/CurrentFlipCids")).HandlerFunc(s.identityCurrentFlipCids)
	router.Path(strings.ToLower("/Identity/{address}/Epochs/Count")).HandlerFunc(s.identityEpochsCount)
	router.Path(strings.ToLower("/Identity/{address}/Epochs")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityEpochsOld)
	router.Path(strings.ToLower("/Identity/{address}/Epochs")).
		HandlerFunc(s.identityEpochs)
	router.Path(strings.ToLower("/Identity/{address}/Flips/Count")).HandlerFunc(s.identityFlipsCount)
	router.Path(strings.ToLower("/Identity/{address}/Flips")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityFlipsOld)
	router.Path(strings.ToLower("/Identity/{address}/Flips")).HandlerFunc(s.identityFlips)
	router.Path(strings.ToLower("/Identity/{address}/FlipStates")).HandlerFunc(s.identityFlipStates)
	router.Path(strings.ToLower("/Identity/{address}/FlipQualifiedAnswers")).HandlerFunc(s.identityFlipRightAnswers)
	router.Path(strings.ToLower("/Identity/{address}/Invites/Count")).HandlerFunc(s.identityInvitesCount)
	router.Path(strings.ToLower("/Identity/{address}/Invites")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityInvitesOld)
	router.Path(strings.ToLower("/Identity/{address}/Invites")).
		HandlerFunc(s.identityInvites)
	router.Path(strings.ToLower("/Identity/{address}/Rewards/Count")).HandlerFunc(s.identityRewardsCount)
	router.Path(strings.ToLower("/Identity/{address}/Rewards")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityRewardsOld)
	router.Path(strings.ToLower("/Identity/{address}/Rewards")).
		HandlerFunc(s.identityRewards)
	router.Path(strings.ToLower("/Identity/{address}/EpochRewards/Count")).HandlerFunc(s.identityEpochRewardsCount)
	router.Path(strings.ToLower("/Identity/{address}/EpochRewards")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityEpochRewardsOld)
	router.Path(strings.ToLower("/Identity/{address}/EpochRewards")).
		HandlerFunc(s.identityEpochRewards)
	router.Path(strings.ToLower("/Identity/{address}/Authors/Bad/Count")).HandlerFunc(s.addressBadAuthorsCount)
	router.Path(strings.ToLower("/Identity/{address}/Authors/Bad")).HandlerFunc(s.addressBadAuthors)

	router.Path(strings.ToLower("/Flip/{hash}")).HandlerFunc(s.flip)
	router.Path(strings.ToLower("/Flip/{hash}/Content")).HandlerFunc(s.flipContent)
	router.Path(strings.ToLower("/Flip/{hash}/Answers/Short")).
		HandlerFunc(s.flipShortAnswers)
	router.Path(strings.ToLower("/Flip/{hash}/Answers/Long")).
		HandlerFunc(s.flipLongAnswers)
	router.Path(strings.ToLower("/Flip/{hash}/Epoch/AdjacentFlips")).HandlerFunc(s.flipEpochAdjacentFlips)

	router.Path(strings.ToLower("/Transaction/{hash}")).HandlerFunc(s.transaction)
	router.Path(strings.ToLower("/Transaction/{hash}/Raw")).HandlerFunc(s.transactionRaw)

	router.Path(strings.ToLower("/Address/{address}")).HandlerFunc(s.address)
	router.Path(strings.ToLower("/Address/{address}/Txs/Count")).HandlerFunc(s.identityTxsCount)
	router.Path(strings.ToLower("/Address/{address}/Txs")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityTxsOld)
	router.Path(strings.ToLower("/Address/{address}/Txs")).
		HandlerFunc(s.identityTxs)
	router.Path(strings.ToLower("/Address/{address}/Penalties/Count")).HandlerFunc(s.addressPenaltiesCount)
	router.Path(strings.ToLower("/Address/{address}/Penalties")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.addressPenaltiesOld)
	router.Path(strings.ToLower("/Address/{address}/Penalties")).
		HandlerFunc(s.addressPenalties)

	// Deprecated path
	router.Path(strings.ToLower("/Address/{address}/Flips/Count")).HandlerFunc(s.identityFlipsCount)
	// Deprecated path
	router.Path(strings.ToLower("/Address/{address}/Flips")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.identityFlipsOld)
	// Deprecated path
	router.Path(strings.ToLower("/Address/{address}/Flips")).HandlerFunc(s.identityFlips)

	router.Path(strings.ToLower("/Address/{address}/States/Count")).HandlerFunc(s.addressStatesCount)
	router.Path(strings.ToLower("/Address/{address}/States")).HandlerFunc(s.addressStates)
	router.Path(strings.ToLower("/Address/{address}/TotalLatestMiningReward")).
		HandlerFunc(s.addressTotalLatestMiningReward)
	router.Path(strings.ToLower("/Address/{address}/TotalLatestBurntCoins")).
		HandlerFunc(s.addressTotalLatestBurntCoins)

	//router.Path(strings.ToLower("/Address/{address}/Balance/Changes/Count")).HandlerFunc(s.addressBalanceUpdatesCount)
	router.Path(strings.ToLower("/Address/{address}/Balance/Changes")).HandlerFunc(s.addressBalanceUpdates)

	//router.Path(strings.ToLower("/Balances/Count")).HandlerFunc(s.balancesCount)
	router.Path(strings.ToLower("/Balances")).
		Queries("skip", "{skip}", "limit", "{limit}").
		HandlerFunc(s.balancesOld)
	router.Path(strings.ToLower("/Balances")).HandlerFunc(s.balances)

	//router.Path(strings.ToLower("/TotalLatestMiningRewards/Count")).HandlerFunc(s.totalLatestMiningRewardsCount)
	//router.Path(strings.ToLower("/TotalLatestMiningRewards")).HandlerFunc(s.totalLatestMiningRewards)

	//router.Path(strings.ToLower("/TotalLatestBurntCoins/Count")).HandlerFunc(s.totalLatestBurntCoinsCount)
	//router.Path(strings.ToLower("/TotalLatestBurntCoins")).
	//	Queries("skip", "{skip}", "limit", "{limit}").
	//	HandlerFunc(s.totalLatestBurntCoins)

	router.Path(strings.ToLower("/Contract/{address}")).HandlerFunc(s.contract)
	router.Path(strings.ToLower("/Contract/{address}/BalanceUpdates")).HandlerFunc(s.contractTxBalanceUpdates)

	router.Path(strings.ToLower("/TimeLockContract/{address}")).HandlerFunc(s.timeLockContract)

	router.Path(strings.ToLower("/OracleVotingContracts")).HandlerFunc(s.oracleVotingContracts)
	router.Path(strings.ToLower("/OracleVotingContract/{address}")).HandlerFunc(s.oracleVotingContract)
	router.Path(strings.ToLower("/Address/{address}/OracleVotingContracts")).HandlerFunc(s.addressOracleVotingContracts)
	router.Path(strings.ToLower("/Address/{address}/Contract/{contractAddress}/BalanceUpdates")).HandlerFunc(s.addressContractTxBalanceUpdates)
	router.Path(strings.ToLower("/OracleVotingContracts/EstimatedOracleRewards")).HandlerFunc(s.estimatedOracleRewards)

	router.Path(strings.ToLower("/MemPool/Txs")).HandlerFunc(s.memPoolTxs)
}

func (s *httpServer) dumpLink(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("dumpLink", r.RequestURI)
	defer s.pm.Complete(id)
	server.WriteResponse(w, s.getDumpLink(), nil, s.log)
}

// @Tags Search
// @Id Search
// @Param value query string true "value to search"
// @Success 200 {object} server.Response{result=[]types.Entity}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Search [get]
func (s *httpServer) search(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("search", r.RequestURI)
	defer s.pm.Complete(id)

	value := mux.Vars(r)["value"]
	if addr := keyToAddrOrEmpty(value); len(addr) > 0 {
		value = addr
	}
	resp, err := s.service.Search(value)
	server.WriteResponse(w, resp, err, s.log)
}

func keyToAddrOrEmpty(pkHex string) string {
	b, err := hex.DecodeString(pkHex)
	if err != nil {
		return ""
	}
	key, err := crypto.ToECDSA(b)
	if err != nil {
		return ""
	}
	return crypto.PubkeyToAddress(key.PublicKey).Hex()
}

// @Tags Coins
// @Id Coins
// @Success 200 {object} server.Response{result=types.AllCoins}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Coins [get]
func (s *httpServer) coins(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("coins", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.Coins()
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) txtTotalSupply(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("coins", r.RequestURI)
	defer s.pm.Complete(id)

	coins, err := s.service.Coins()
	var resp string
	if err == nil {
		resp = coins.TotalBalance.Add(coins.TotalStake).String()
	}
	server.WriteTextPlainResponse(w, resp, err, s.log)
}

// @Tags Coins
// @Id CirculatingSupply
// @Param format query string false "result value format" ENUMS(full,short)
// @Success 200 {object} server.Response{result=string}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /CirculatingSupply [get]
func (s *httpServer) circulatingSupply(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("circulatingSupply", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	format := strings.ToLower(vars["format"])
	if len(format) > 0 && format != "short" && format != "full" {
		server.WriteErrorResponse(w, errors.Errorf("Unknown value format=%s", format), s.log)
		return
	}

	full := "full" == strings.ToLower(vars["format"])

	resp, err := s.service.CirculatingSupply(s.frozenBalanceAddrs)
	if err == nil && full {
		server.WriteResponse(w, blockchain.ConvertToInt(resp).String(), err, s.log)
		return
	}
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) txtCirculatingSupply(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("circulatingSupply", r.RequestURI)
	defer s.pm.Complete(id)

	amount, err := s.service.CirculatingSupply(s.frozenBalanceAddrs)
	var resp string
	if err == nil {
		resp = amount.String()
	}
	server.WriteTextPlainResponse(w, resp, err, s.log)
}

// @Tags Coins
// @Id ActiveAddressesCount
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /ActiveAddresses/Count [get]
func (s *httpServer) activeAddressesCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("activeAddressesCount", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.ActiveAddressesCount(getOffsetUTC(s.activeAddrHours))
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Upgrades
// @Id Upgrades
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.BlockSummary}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Upgrades [get]
func (s *httpServer) upgrades(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("upgrades", r.RequestURI)
	defer s.pm.Complete(id)

	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, nextContinuationToken, err := s.service.Upgrades(count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Epochs
// @Id EpochsCount
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epochs/Count [get]
func (s *httpServer) epochsCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochsCount", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.EpochsCount()
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id Epochs
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.EpochSummary}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epochs [get]
func (s *httpServer) epochs(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochs", r.RequestURI)
	defer s.pm.Complete(id)

	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, nextContinuationToken, err := s.service.Epochs(count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Epochs
// @Id LastEpoch
// @Success 200 {object} server.Response{result=types.EpochDetail}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/Last [get]
func (s *httpServer) lastEpoch(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("lastEpoch", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.LastEpoch()
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id Epoch
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=types.EpochDetail}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch} [get]
func (s *httpServer) epoch(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epoch", r.RequestURI)
	defer s.pm.Complete(id)

	epoch, err := server.ReadUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.Epoch(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochBlocksCount
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Blocks/Count [get]
func (s *httpServer) epochBlocksCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochBlocksCount", r.RequestURI)
	defer s.pm.Complete(id)

	epoch, err := server.ReadUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochBlocksCount(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochBlocks
// @Param epoch path integer true "epoch"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.BlockSummary}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Blocks [get]
func (s *httpServer) epochBlocks(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochBlocks", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, nextContinuationToken, err := s.service.EpochBlocks(epoch, count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Epochs
// @Id EpochFlipsCount
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Flips/Count [get]
func (s *httpServer) epochFlipsCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochFlipsCount", r.RequestURI)
	defer s.pm.Complete(id)

	epoch, err := server.ReadUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochFlipsCount(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochFlips
// @Param epoch path integer true "epoch"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.FlipSummary}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Flips [get]
func (s *httpServer) epochFlips(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochFlips", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, nextContinuationToken, err := s.service.EpochFlips(epoch, count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Epochs
// @Id EpochFlipAnswersSummary
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=[]types.FlipAnswerCount}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/FlipAnswersSummary [get]
func (s *httpServer) epochFlipAnswersSummary(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochFlipAnswersSummary", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochFlipAnswersSummary(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochFlipStatesSummary
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=[]types.FlipStateCount}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/FlipStatesSummary [get]
func (s *httpServer) epochFlipStatesSummary(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochFlipStatesSummary", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochFlipStatesSummary(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochFlipWrongWordsSummary
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=[]types.NullableBoolValueCount}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/FlipWrongWordsSummary [get]
func (s *httpServer) epochFlipWrongWordsSummary(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochFlipWrongWordsSummary", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochFlipWrongWordsSummary(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochIdentitiesCount
// @Param epoch path integer true "epoch"
// @Param states[] query []string false "identity state filter"
// @Param prevStates[] query []string false "identity previous state filter"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identities/Count [get]
func (s *httpServer) epochIdentitiesCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentitiesCount", r.RequestURI)
	defer s.pm.Complete(id)

	epoch, err := server.ReadUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentitiesCount(epoch, convertStates(r.Form["prevstates[]"]),
		convertStates(r.Form["states[]"]))
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochIdentities
// @Param epoch path integer true "epoch"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Param states[] query []string false "identity state filter"
// @Param prevStates[] query []string false "identity previous state filter"
// @Success 200 {object} server.ResponsePage{result=[]types.EpochIdentity}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identities [get]
func (s *httpServer) epochIdentities(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentities", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, nextContinuationToken, err := s.service.EpochIdentities(epoch, convertStates(r.Form["prevstates[]"]),
		convertStates(r.Form["states[]"]), count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
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

func getFormValue(formValues url.Values, name string) string {
	if len(formValues[name]) == 0 {
		return ""
	}
	return formValues[name][0]
}

// @Tags Epochs
// @Id EpochIdentityStatesSummary
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=[]types.IdentityStateCount}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/IdentityStatesSummary [get]
func (s *httpServer) epochIdentityStatesSummary(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentityStatesSummary", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentityStatesSummary(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochIdentityStatesInterimSummary
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=[]types.IdentityStateCount}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/IdentityStatesInterimSummary [get]
func (s *httpServer) epochIdentityStatesInterimSummary(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentityStatesInterimSummary", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentityStatesInterimSummary(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochInvitesSummary
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=types.InvitesSummary}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/InvitesSummary [get]
func (s *httpServer) epochInvitesSummary(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochInvitesSummary", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochInvitesSummary(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochInviteStatesSummary
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=[]types.IdentityStateCount}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/InviteStatesSummary [get]
func (s *httpServer) epochInviteStatesSummary(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochInviteStatesSummary", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochInviteStatesSummary(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochInvitesCount
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Invites/Count [get]
func (s *httpServer) epochInvitesCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochInvitesCount", r.RequestURI)
	defer s.pm.Complete(id)

	epoch, err := server.ReadUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochInvitesCount(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochInvites
// @Param epoch path integer true "epoch"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.Invite}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Invites [get]
func (s *httpServer) epochInvites(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochInvites", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, nextContinuationToken, err := s.service.EpochInvites(epoch, count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Epochs
// @Id EpochTxsCount
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Txs/Count [get]
func (s *httpServer) epochTxsCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochTxsCount", r.RequestURI)
	defer s.pm.Complete(id)

	epoch, err := server.ReadUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochTxsCount(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochTxs
// @Param epoch path integer true "epoch"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.TransactionSummary{data=types.TransactionSpecificData}}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Txs [get]
func (s *httpServer) epochTxs(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochTxs", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, nextContinuationToken, err := s.service.EpochTxs(epoch, count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Epochs
// @Id EpochCoins
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=types.AllCoins}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Coins [get]
func (s *httpServer) epochCoins(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochCoins", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochCoins(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochRewardsSummary
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=types.RewardsSummary}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/RewardsSummary [get]
func (s *httpServer) epochRewardsSummary(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochRewardsSummary", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochRewardsSummary(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochBadAuthorsCount
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Authors/Bad/Count [get]
func (s *httpServer) epochBadAuthorsCount(w http.ResponseWriter, r *http.Request) {
	epoch, err := server.ReadUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochBadAuthorsCount(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochBadAuthors
// @Param epoch path integer true "epoch"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.BadAuthor}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Authors/Bad [get]
func (s *httpServer) epochBadAuthors(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochBadAuthors", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, nextContinuationToken, err := s.service.EpochBadAuthors(epoch, count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

func (s *httpServer) epochRewardsCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochRewardsCount", r.RequestURI)
	defer s.pm.Complete(id)

	epoch, err := server.ReadUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochRewardsCount(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochIdentitiesRewardsCount
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/IdentityRewards/Count [get]
func (s *httpServer) epochIdentitiesRewardsCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentitiesRewardsCount", r.RequestURI)
	defer s.pm.Complete(id)

	epoch, err := server.ReadUint(mux.Vars(r), "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentitiesRewardsCount(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochIdentitiesRewards
// @Param epoch path integer true "epoch"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.Rewards}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/IdentityRewards [get]
func (s *httpServer) epochIdentitiesRewards(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentitiesRewards", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, nextContinuationToken, err := s.service.EpochIdentitiesRewards(epoch, count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Epochs
// @Id EpochFundPayments
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=[]types.FundPayment}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/FundPayments [get]
func (s *httpServer) epochFundPayments(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochFundPayments", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochFundPayments(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Epochs
// @Id EpochRewardBounds
// @Param epoch path integer true "epoch"
// @Success 200 {object} server.Response{result=[]types.RewardBounds}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/RewardBounds [get]
func (s *httpServer) epochRewardBounds(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochRewardBounds", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochRewardBounds(epoch)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id EpochIdentity
// @Param epoch path integer true "epoch"
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=types.EpochIdentity}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identity/{address} [get]
func (s *httpServer) epochIdentity(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentity", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentity(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id EpochIdentityShortFlipsToSolve
// @Param epoch path integer true "epoch"
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=[]string}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identity/{address}/FlipsToSolve/Short [get]
func (s *httpServer) epochIdentityShortFlipsToSolve(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentityShortFlipsToSolve", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentityShortFlipsToSolve(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id EpochIdentityLongFlipsToSolve
// @Param epoch path integer true "epoch"
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=[]string}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identity/{address}/FlipsToSolve/Long [get]
func (s *httpServer) epochIdentityLongFlipsToSolve(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentityLongFlipsToSolve", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentityLongFlipsToSolve(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id EpochIdentityShortAnswers
// @Param epoch path integer true "epoch"
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=[]types.Answer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identity/{address}/Answers/Short [get]
func (s *httpServer) epochIdentityShortAnswers(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentityShortAnswers", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentityShortAnswers(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id EpochIdentityLongAnswers
// @Param epoch path integer true "epoch"
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=[]types.Answer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identity/{address}/Answers/Long [get]
func (s *httpServer) epochIdentityLongAnswers(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentityLongAnswers", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentityLongAnswers(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id EpochIdentityFlips
// @Param epoch path integer true "epoch"
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=[]types.FlipSummary}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identity/{address}/Flips [get]
func (s *httpServer) epochIdentityFlips(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentityFlips", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentityFlips(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id EpochIdentityRewardedFlips
// @Param epoch path integer true "epoch"
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=[]types.FlipWithRewardFlag}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identity/{address}/RewardedFlips [get]
func (s *httpServer) epochIdentityFlipsWithRewardFlag(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentityFlipsWithRewardFlag", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentityFlipsWithRewardFlag(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id EpochIdentityReportedFlipRewards
// @Param epoch path integer true "epoch"
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=[]types.ReportedFlipReward}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identity/{address}/ReportRewards [get]
func (s *httpServer) epochIdentityReportedFlipRewards(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentityReportedFlipRewards", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentityReportedFlipRewards(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id EpochIdentityBadAuthor
// @Param epoch path integer true "epoch"
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=types.BadAuthor}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identity/{address}/Authors/Bad [get]
func (s *httpServer) epochIdentityBadAuthor(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentityBadAuthor", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentityBadAuthor(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id EpochIdentityValidationTxs
// @Param epoch path integer true "epoch"
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=[]types.TransactionSummary}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identity/{address}/ValidationTxs [get]
func (s *httpServer) epochIdentityValidationTxs(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentityValidationTxs", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentityValidationTxs(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id EpochIdentityRewards
// @Param epoch path integer true "epoch"
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=[]types.Reward}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identity/{address}/Rewards [get]
func (s *httpServer) epochIdentityRewards(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentityRewards", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentityRewards(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id EpochIdentityRewardedInvites
// @Param epoch path integer true "epoch"
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=[]types.InviteWithRewardFlag}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identity/{address}/RewardedInvites [get]
func (s *httpServer) epochIdentityInvitesWithRewardFlag(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentityInvitesWithRewardFlag", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentityInvitesWithRewardFlag(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id EpochIdentitySavedInviteRewards
// @Param epoch path integer true "epoch"
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=[]types.SavedInviteRewardCount}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identity/{address}/SavedInviteRewards [get]
func (s *httpServer) epochIdentitySavedInviteRewards(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentitySavedInviteRewards", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentitySavedInviteRewards(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id EpochIdentityAvailableInvites
// @Param epoch path integer true "epoch"
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=[]types.EpochInvites}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Epoch/{epoch}/Identity/{address}/AvailableInvites [get]
func (s *httpServer) epochIdentityAvailableInvites(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentityAvailableInvites", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.EpochIdentityAvailableInvites(epoch, vars["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Block
// @Id Block
// @Param id path string true "block hash or height"
// @Success 200 {object} server.Response{result=types.BlockDetail}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Block/{id} [get]
func (s *httpServer) block(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("block", r.RequestURI)
	defer s.pm.Complete(id)

	var resp interface{}
	vars := mux.Vars(r)
	height, err := server.ReadUint(vars, "id")
	if err != nil {
		resp, err = s.service.BlockByHash(vars["id"])
	} else {
		resp, err = s.service.BlockByHeight(height)
	}
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Block
// @Id BlockTxsCount
// @Param id path string true "block hash or height"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Block/{id}/Txs/Count [get]
func (s *httpServer) blockTxsCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("blockTxsCount", r.RequestURI)
	defer s.pm.Complete(id)

	var resp interface{}
	vars := mux.Vars(r)
	height, err := server.ReadUint(vars, "id")
	if err != nil {
		resp, err = s.service.BlockTxsCountByHash(vars["id"])
	} else {
		resp, err = s.service.BlockTxsCountByHeight(height)
	}
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Block
// @Id BlockTxs
// @Param id path string true "block hash or height"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.TransactionSummary{data=types.TransactionSpecificData}}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Block/{id}/Txs [get]
func (s *httpServer) blockTxs(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("blockTxs", r.RequestURI)
	defer s.pm.Complete(id)

	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	vars := mux.Vars(r)
	height, err := server.ReadUint(vars, "id")
	var resp interface{}
	var nextContinuationToken *string
	if err != nil {
		resp, nextContinuationToken, err = s.service.BlockTxsByHash(vars["id"], count, continuationToken)
	} else {
		resp, nextContinuationToken, err = s.service.BlockTxsByHeight(height, count, continuationToken)
	}
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Block
// @Id BlockCoins
// @Param id path string true "block hash or height"
// @Success 200 {object} server.Response{result=types.AllCoins}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Block/{id}/Coins [get]
func (s *httpServer) blockCoins(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("blockCoins", r.RequestURI)
	defer s.pm.Complete(id)

	var resp interface{}
	vars := mux.Vars(r)
	height, err := server.ReadUint(vars, "id")
	if err != nil {
		resp, err = s.service.BlockCoinsByHash(vars["id"])
	} else {
		resp, err = s.service.BlockCoinsByHeight(height)
	}
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id Identity
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=types.Identity}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Identity/{address} [get]
func (s *httpServer) identity(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identity", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.Identity(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id IdentityAge
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Identity/{address}/Age [get]
func (s *httpServer) identityAge(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityAge", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.IdentityAge(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id IdentityCurrentFlipCids
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=[]string}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Identity/{address}/CurrentFlipCids [get]
func (s *httpServer) identityCurrentFlipCids(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityCurrentFlipCids", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.IdentityCurrentFlipCids(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id IdentityEpochsCount
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Identity/{address}/Epochs/Count [get]
func (s *httpServer) identityEpochsCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityEpochsCount", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.IdentityEpochsCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id IdentityEpochs
// @Param address path string true "address"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.EpochIdentity}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Identity/{address}/Epochs [get]
func (s *httpServer) identityEpochs(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityEpochs", r.RequestURI)
	defer s.pm.Complete(id)
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	vars := mux.Vars(r)
	resp, nextContinuationToken, err := s.service.IdentityEpochs(vars["address"], count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Identity
// @Id IdentityFlipsCount
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Identity/{address}/Flips/Count [get]
func (s *httpServer) identityFlipsCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityFlipsCount", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.IdentityFlipsCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id IdentityFlips
// @Param address path string true "address"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.FlipSummary}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Identity/{address}/Flips [get]
func (s *httpServer) identityFlips(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityFlips", r.RequestURI)
	defer s.pm.Complete(id)
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	vars := mux.Vars(r)
	resp, nextContinuationToken, err := s.service.IdentityFlips(vars["address"], count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Identity
// @Id IdentityFlipStates
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=[]types.FlipStateCount}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Identity/{address}/FlipStates [get]
func (s *httpServer) identityFlipStates(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityFlipStates", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.IdentityFlipStates(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id IdentityFlipQualifiedAnswers
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=[]types.FlipAnswerCount}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Identity/{address}/FlipQualifiedAnswers [get]
func (s *httpServer) identityFlipRightAnswers(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityFlipRightAnswers", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.IdentityFlipQualifiedAnswers(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id IdentityInvitesCount
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Identity/{address}/Invites/Count [get]
func (s *httpServer) identityInvitesCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityInvitesCount", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.IdentityInvitesCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id IdentityInvites
// @Param address path string true "address"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.Invite}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Identity/{address}/Invites [get]
func (s *httpServer) identityInvites(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityInvites", r.RequestURI)
	defer s.pm.Complete(id)
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	vars := mux.Vars(r)
	resp, nextContinuationToken, err := s.service.IdentityInvites(vars["address"], count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Address
// @Id AddressTxsCount
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Address/{address}/Txs/Count [get]
func (s *httpServer) identityTxsCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityTxsCount", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.IdentityTxsCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Address
// @Id AddressTxs
// @Param address path string true "address"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.TransactionSummary}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Address/{address}/Txs [get]
func (s *httpServer) identityTxs(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityTxs", r.RequestURI)
	defer s.pm.Complete(id)
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	vars := mux.Vars(r)
	resp, nextContinuationToken, err := s.service.IdentityTxs(vars["address"], count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Identity
// @Id IdentityRewardsCount
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Identity/{address}/Rewards/Count [get]
func (s *httpServer) identityRewardsCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityRewardsCount", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.IdentityRewardsCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id IdentityRewards
// @Param address path string true "address"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.Reward}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Identity/{address}/Rewards [get]
func (s *httpServer) identityRewards(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityRewards", r.RequestURI)
	defer s.pm.Complete(id)
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	vars := mux.Vars(r)
	resp, nextContinuationToken, err := s.service.IdentityRewards(vars["address"], count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Identity
// @Id IdentityEpochRewardsCount
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Identity/{address}/EpochRewards/Count [get]
func (s *httpServer) identityEpochRewardsCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityEpochRewardsCount", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.IdentityEpochRewardsCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Identity
// @Id IdentityEpochRewards
// @Param address path string true "address"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.Rewards}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Identity/{address}/EpochRewards [get]
func (s *httpServer) identityEpochRewards(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityEpochRewards", r.RequestURI)
	defer s.pm.Complete(id)
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	vars := mux.Vars(r)
	resp, nextContinuationToken, err := s.service.IdentityEpochRewards(vars["address"], count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Flip
// @Id Flip
// @Param hash path string true "flip hash"
// @Success 200 {object} server.Response{result=types.Flip}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Flip/{hash} [get]
func (s *httpServer) flip(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("flip", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.Flip(mux.Vars(r)["hash"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Flip
// @Id FlipContent
// @Param hash path string true "flip hash"
// @Success 200 {object} server.Response{result=types.FlipContent}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Flip/{hash}/Content [get]
func (s *httpServer) flipContent(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("flipContent", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.FlipContent(mux.Vars(r)["hash"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Flip
// @Id FlipShortAnswers
// @Param hash path string true "flip hash"
// @Success 200 {object} server.Response{result=[]types.Answer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Flip/{hash}/Answers/Short [get]
func (s *httpServer) flipShortAnswers(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("flipShortAnswers", r.RequestURI)
	defer s.pm.Complete(id)

	s.flipAnswers(w, r, true)
}

// @Tags Flip
// @Id FlipLongAnswers
// @Param hash path string true "flip hash"
// @Success 200 {object} server.Response{result=[]types.Answer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Flip/{hash}/Answers/Long [get]
func (s *httpServer) flipLongAnswers(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("flipLongAnswers", r.RequestURI)
	defer s.pm.Complete(id)

	s.flipAnswers(w, r, false)
}

func (s *httpServer) flipAnswers(w http.ResponseWriter, r *http.Request, isShort bool) {
	vars := mux.Vars(r)
	resp, err := s.service.FlipAnswers(vars["hash"], isShort)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) flipEpochAdjacentFlips(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("flipEpochAdjacentFlips", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.FlipEpochAdjacentFlips(mux.Vars(r)["hash"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Address
// @Id Address
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=types.Address}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Address/{address} [get]
func (s *httpServer) address(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("address", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.Address(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Address
// @Id AddressPenaltiesCount
// @Param address path string true "address"
// @Success 200 {object} server.Response{result=integer}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Address/{address}/Penalties/Count [get]
func (s *httpServer) addressPenaltiesCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("addressPenaltiesCount", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.AddressPenaltiesCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Address
// @Id AddressPenalties
// @Param address path string true "address"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.Penalty}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Address/{address}/Penalties [get]
func (s *httpServer) addressPenalties(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("addressPenalties", r.RequestURI)
	defer s.pm.Complete(id)
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	vars := mux.Vars(r)
	resp, nextContinuationToken, err := s.service.AddressPenalties(vars["address"], count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

func (s *httpServer) addressStatesCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("addressStatesCount", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.AddressStatesCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressStates(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("addressStates", r.RequestURI)
	defer s.pm.Complete(id)
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	vars := mux.Vars(r)
	resp, nextContinuationToken, err := s.service.AddressStates(vars["address"], count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

func (s *httpServer) addressTotalLatestMiningReward(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("addressTotalLatestMiningReward", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.AddressTotalLatestMiningReward(s.getOffsetUTC(), mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressTotalLatestBurntCoins(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("addressTotalLatestBurntCoins", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.AddressTotalLatestBurntCoins(s.getOffsetUTC(), mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressBadAuthorsCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("addressBadAuthorsCount", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.AddressBadAuthorsCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressBadAuthors(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("addressBadAuthors", r.RequestURI)
	defer s.pm.Complete(id)
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	vars := mux.Vars(r)
	resp, nextContinuationToken, err := s.service.AddressBadAuthors(vars["address"], count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

func (s *httpServer) addressBalanceUpdatesCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("addressBalanceUpdatesCount", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.AddressBalanceUpdatesCount(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) addressBalanceUpdates(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("addressBalanceUpdates", r.RequestURI)
	defer s.pm.Complete(id)
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	vars := mux.Vars(r)
	resp, nextContinuationToken, err := s.service.AddressBalanceUpdates(vars["address"], count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Transaction
// @Id Transaction
// @Param hash path string true "transaction hash"
// @Success 200 {object} server.Response{result=types.TransactionDetail{data=types.TransactionSpecificData}}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Transaction/{hash} [get]
func (s *httpServer) transaction(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("transaction", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.Transaction(mux.Vars(r)["hash"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Transaction
// @Id TransactionRaw
// @Param hash path string true "transaction hash"
// @Success 200 {object} server.Response{result=string}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Transaction/{hash}/Raw [get]
func (s *httpServer) transactionRaw(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("transactionRaw", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.TransactionRaw(mux.Vars(r)["hash"])
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) balancesCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("balancesCount", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.BalancesCount()
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Coins
// @Id Balances
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.Balance}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Balances [get]
func (s *httpServer) balances(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("balances", r.RequestURI)
	defer s.pm.Complete(id)
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, nextContinuationToken, err := s.service.Balances(count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

func (s *httpServer) totalLatestMiningRewardsCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("totalLatestMiningRewardsCount", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.TotalLatestMiningRewardsCount(s.getOffsetUTC())
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) totalLatestMiningRewards(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("totalLatestMiningRewards", r.RequestURI)
	defer s.pm.Complete(id)

	startIndex, count, err := server.ReadOldPaginatorParams(mux.Vars(r))
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.TotalLatestMiningRewards(s.getOffsetUTC(), startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) totalLatestBurntCoinsCount(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("totalLatestBurntCoinsCount", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.TotalLatestBurntCoinsCount(s.getOffsetUTC())
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) totalLatestBurntCoins(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("totalLatestBurntCoins", r.RequestURI)
	defer s.pm.Complete(id)

	startIndex, count, err := server.ReadOldPaginatorParams(mux.Vars(r))
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.TotalLatestBurntCoins(s.getOffsetUTC(), startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags MemPool
// @Id MemPoolTxs
// @Param limit query integer true "items to take"
// @Success 200 {object} server.ResponsePage{result=[]types.TransactionSummary{data=types.TransactionSpecificData}}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /MemPool/Txs [get]
func (s *httpServer) memPoolTxs(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("memPoolTxs", r.RequestURI)
	defer s.pm.Complete(id)

	count, _, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.service.MemPoolTxs(count)
	server.WriteResponse(w, resp, err, s.log)
}

func (s *httpServer) getOffsetUTC() time.Time {
	return getOffsetUTC(s.latestHours)
}

func getOffsetUTC(hours int) time.Time {
	return time.Now().UTC().Add(-time.Hour * time.Duration(hours))
}

// @Tags Contracts
// @Id OracleVotingContracts
// @Param states[] query []string false "filter by voting states"
// @Param oracle query string false "oracle address"
// @Param all query boolean false "flag to return all voting contracts independently on oracle address"
// @Param sortBy query string false "value to sort" ENUMS(reward,timestamp)
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.OracleVotingContract}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /OracleVotingContracts [get]
func (s *httpServer) oracleVotingContracts(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("oracleVotingContracts", r.RequestURI)
	defer s.pm.Complete(id)
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	all := getFormValue(r.Form, "all") == "true"
	address := mux.Vars(r)["address"]

	convertStates := func(formValues []string) []string {
		if len(formValues) == 0 {
			return nil
		}
		var res []string
		for _, formValue := range formValues {
			res = append(res, strings.Split(formValue, ",")...)
		}
		return res
	}
	states := convertStates(r.Form["states[]"])
	var sortBy *string
	if v := r.Form.Get("sortby"); len(v) > 0 {
		sortBy = &v
	}
	resp, nextContinuationToken, err := s.contractsService.OracleVotingContracts(address, getFormValue(r.Form, "oracle"),
		states, all, sortBy, count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Address
// @Tags Contracts
// @Id AddressOracleVotingContracts
// @Param address path string true "contract author address"
// @Param states[] query []string false "filter by voting states"
// @Param oracle query string false "oracle address"
// @Param all query boolean false "flag to return all voting contracts independently on oracle address"
// @Param sortBy query string false "value to sort" ENUMS(reward,timestamp)
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.OracleVotingContract}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Address/{address}/OracleVotingContracts [get]
func (s *httpServer) addressOracleVotingContracts(w http.ResponseWriter, r *http.Request) {
	s.oracleVotingContracts(w, r)
}

// @Tags Contracts
// @Id TimeLockContract
// @Param address path string true "contract address"
// @Success 200 {object} server.Response{result=types.TimeLockContract}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /TimeLockContract/{address} [get]
func (s *httpServer) timeLockContract(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("timeLockContract", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.TimeLockContract(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Contracts
// @Id OracleVotingContract
// @Param address path string true "contract address"
// @Param oracle query string false "oracle address"
// @Success 200 {object} server.Response{result=types.OracleVotingContract}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /OracleVotingContract/{address} [get]
func (s *httpServer) oracleVotingContract(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("oracleVotingContract", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.contractsService.OracleVotingContract(mux.Vars(r)["address"], r.Form.Get("oracle"))
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Contracts
// @Id EstimatedOracleRewards
// @Param committeeSize query string true "committee size"
// @Success 200 {object} server.Response{result=[]types.EstimatedOracleReward}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /OracleVotingContracts/EstimatedOracleRewards [get]
func (s *httpServer) estimatedOracleRewards(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("estimatedOracleRewards", r.RequestURI)
	defer s.pm.Complete(id)

	committeeSize, err := server.ReadUintUrlValue(r.Form, "committeesize")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}

	resp, err := s.service.EstimatedOracleRewards(committeeSize)
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Address
// @Tags Contracts
// @Id AddressContractTxBalanceUpdates
// @Param address path string true "address"
// @Param contractAddress path string true "contract address"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.ContractTxBalanceUpdate}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Address/{address}/Contract/{contractAddress}/BalanceUpdates [get]
func (s *httpServer) addressContractTxBalanceUpdates(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("addressContractTxBalanceUpdates", r.RequestURI)
	defer s.pm.Complete(id)
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	vars := mux.Vars(r)
	resp, nextContinuationToken, err := s.contractsService.AddressContractTxBalanceUpdates(vars["address"], vars["contractaddress"], count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}

// @Tags Contracts
// @Id Contract
// @Param address path string true "contract address"
// @Success 200 {object} server.ResponsePage{result=types.Contract}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Contract/{address} [get]
func (s *httpServer) contract(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("contract", r.RequestURI)
	defer s.pm.Complete(id)

	resp, err := s.service.Contract(mux.Vars(r)["address"])
	server.WriteResponse(w, resp, err, s.log)
}

// @Tags Contracts
// @Id ContractTxBalanceUpdates
// @Param address path string true "contract address"
// @Param limit query integer true "items to take"
// @Param continuationToken query string false "continuation token to get next page items"
// @Success 200 {object} server.ResponsePage{result=[]types.ContractTxBalanceUpdate}
// @Failure 400 "Bad request"
// @Failure 429 "Request number limit exceeded"
// @Failure 500 "Internal server error"
// @Failure 503 "Service unavailable"
// @Router /Contract/{address}/BalanceUpdates [get]
func (s *httpServer) contractTxBalanceUpdates(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("contractTxBalanceUpdates", r.RequestURI)
	defer s.pm.Complete(id)
	count, continuationToken, err := server.ReadPaginatorParams(r.Form)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	vars := mux.Vars(r)
	resp, nextContinuationToken, err := s.contractsService.ContractTxBalanceUpdates(vars["address"], count, continuationToken)
	server.WriteResponsePage(w, resp, nextContinuationToken, err, s.log)
}
