package server

import (
	"github.com/gorilla/mux"
	"github.com/idena-network/idena-indexer/core/api"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
	"io"
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
	router.Path(strings.ToLower("/Pool/{address}")).HandlerFunc(ri.pool)

	router.Path(strings.ToLower("/OnlineMiners/Count")).HandlerFunc(ri.onlineCount)

	router.Path(strings.ToLower("/Validators/Count")).HandlerFunc(ri.validatorsCount)
	router.Path(strings.ToLower("/Validators")).HandlerFunc(ri.validators)
	router.Path(strings.ToLower("/OnlineValidators/Count")).HandlerFunc(ri.onlineValidatorsCount)
	router.Path(strings.ToLower("/OnlineValidators")).HandlerFunc(ri.onlineValidators)

	router.Path(strings.ToLower("/SignatureAddress")).
		Queries("value", "{value}", "signature", "{signature}").
		HandlerFunc(ri.signatureAddress)

	router.Path(strings.ToLower("/UpgradeVoting")).HandlerFunc(ri.upgradeVoting)

	router.Path(strings.ToLower("/Now")).HandlerFunc(ri.now)

	router.Path(strings.ToLower("/MemPool/Transaction/{hash}")).HandlerFunc(ri.memPoolTransaction)
	router.Path(strings.ToLower("/MemPool/Transaction/{hash}/Raw")).HandlerFunc(ri.memPoolTransactionRaw)
	router.Path(strings.ToLower("/MemPool/Address/{address}/Transactions")).
		Queries("limit", "{limit}").
		HandlerFunc(ri.memPoolAddressTransactions)
	router.Path(strings.ToLower("/MemPool/Transactions")).
		Queries("limit", "{limit}").
		HandlerFunc(ri.memPoolTransactions)
	router.Path(strings.ToLower("/MemPool/Transactions/Count")).HandlerFunc(ri.memPoolTransactionsCount)
	router.Path(strings.ToLower("/MemPool/OracleVotingContractDeploys")).HandlerFunc(ri.memPoolOracleVotingContractDeploys)
	router.Path(strings.ToLower("/MemPool/Address/{address}/Contract/{contractAddress}/Txs")).HandlerFunc(ri.memPoolAddressContractTxs)

	router.Path(strings.ToLower("/Address/{address}/IdentityWithProof")).
		Queries("epoch", "{epoch:[0-9]+}").HandlerFunc(ri.identityWithProof)

	router.Path(strings.ToLower("/Staking")).HandlerFunc(ri.staking)
	router.Path(strings.ToLower("/StakingV2")).HandlerFunc(ri.stakingV2)

	router.Path(strings.ToLower("/Multisig/{address}")).HandlerFunc(ri.multisig)

	router.Path(strings.ToLower("/ForkCommittee/Count")).HandlerFunc(ri.forkCommitteeSize)

	router.Path(strings.ToLower("/Contract/{address}/Verify")).HandlerFunc(ri.verifyContract)
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

func (ri *routerInitializer) pool(w http.ResponseWriter, r *http.Request) {
	resp := ri.api.GetPool(mux.Vars(r)["address"])
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

func (ri *routerInitializer) memPoolTransaction(w http.ResponseWriter, r *http.Request) {
	hash := mux.Vars(r)["hash"]
	resp, err := ri.api.MemPoolTransaction(hash)
	WriteResponse(w, resp, err, ri.logger)
}

func (ri *routerInitializer) memPoolTransactionRaw(w http.ResponseWriter, r *http.Request) {
	hash := mux.Vars(r)["hash"]
	resp, err := ri.api.MemPoolTransactionRaw(hash)
	WriteResponse(w, resp, err, ri.logger)
}

func (ri *routerInitializer) memPoolAddressTransactions(w http.ResponseWriter, r *http.Request) {
	address := mux.Vars(r)["address"]
	count, _, err := ReadPaginatorParams(r.Form)
	if err != nil {
		WriteErrorResponse(w, err, ri.logger)
		return
	}
	resp, err := ri.api.MemPoolAddressTransactions(address, count)
	WriteResponse(w, resp, err, ri.logger)
}

func (ri *routerInitializer) memPoolTransactions(w http.ResponseWriter, r *http.Request) {
	count, _, err := ReadPaginatorParams(r.Form)
	if err != nil {
		WriteErrorResponse(w, err, ri.logger)
		return
	}
	resp, err := ri.api.MemPoolTransactions(count)
	WriteResponse(w, resp, err, ri.logger)
}

func (ri *routerInitializer) memPoolTransactionsCount(w http.ResponseWriter, r *http.Request) {
	resp, err := ri.api.MemPoolTransactionsCount()
	WriteResponse(w, resp, err, ri.logger)
}

func (ri *routerInitializer) memPoolOracleVotingContractDeploys(w http.ResponseWriter, r *http.Request) {
	author := mux.Vars(r)["author"]
	resp, err := ri.api.MemPoolOracleVotingContractDeploys(author)
	WriteResponse(w, resp, err, ri.logger)
}

func (ri *routerInitializer) memPoolAddressContractTxs(w http.ResponseWriter, r *http.Request) {
	address := mux.Vars(r)["address"]
	contractAddress := mux.Vars(r)["contractaddress"]
	resp, err := ri.api.MemPoolAddressContractTxs(address, contractAddress)
	WriteResponse(w, resp, err, ri.logger)
}

func (ri *routerInitializer) validatorsCount(w http.ResponseWriter, r *http.Request) {
	resp := ri.api.ValidatorsCount()
	WriteResponse(w, resp, nil, ri.logger)
}

func (ri *routerInitializer) validators(w http.ResponseWriter, r *http.Request) {
	count, continuationToken, err := ReadPaginatorParams(r.Form)
	if err != nil {
		WriteErrorResponse(w, err, ri.logger)
		return
	}
	resp, nextContinuationToken, err := ri.api.Validators(count, continuationToken)
	WriteResponsePage(w, resp, nextContinuationToken, err, ri.logger)
}

func (ri *routerInitializer) onlineValidatorsCount(w http.ResponseWriter, r *http.Request) {
	resp := ri.api.OnlineValidatorsCount()
	WriteResponse(w, resp, nil, ri.logger)
}

func (ri *routerInitializer) onlineValidators(w http.ResponseWriter, r *http.Request) {
	count, continuationToken, err := ReadPaginatorParams(r.Form)
	if err != nil {
		WriteErrorResponse(w, err, ri.logger)
		return
	}
	resp, nextContinuationToken, err := ri.api.OnlineValidators(count, continuationToken)
	WriteResponsePage(w, resp, nextContinuationToken, err, ri.logger)
}

func (ri *routerInitializer) identityWithProof(w http.ResponseWriter, r *http.Request) {
	epoch, err := ReadUint(mux.Vars(r), "epoch")
	if err != nil {
		WriteErrorResponse(w, err, ri.logger)
		return
	}
	address := mux.Vars(r)["address"]
	resp, err := ri.api.IdentityWithProof(epoch, address)
	WriteResponse(w, resp, err, ri.logger)
}

func (ri *routerInitializer) staking(w http.ResponseWriter, r *http.Request) {
	resp, err := ri.api.Staking()
	WriteResponse(w, resp.Weight, err, ri.logger)
}

func (ri *routerInitializer) stakingV2(w http.ResponseWriter, r *http.Request) {
	resp, err := ri.api.Staking()
	WriteResponse(w, resp, err, ri.logger)
}

func (ri *routerInitializer) multisig(w http.ResponseWriter, r *http.Request) {
	address := mux.Vars(r)["address"]
	resp, err := ri.api.Multisig(address)
	WriteResponse(w, resp, err, ri.logger)
}

func (ri *routerInitializer) forkCommitteeSize(w http.ResponseWriter, r *http.Request) {
	resp := ri.api.ForkCommitteeSize()
	WriteResponse(w, resp, nil, ri.logger)
}

func (ri *routerInitializer) verifyContract(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		WriteResponse(w, nil, errors.Wrap(err, "failed to read request data"), ri.logger)
		return
	}
	address := mux.Vars(r)["address"]
	usrErr, err := ri.api.VerifyContract(address, data)
	WriteResponseWithUserErr(w, nil, usrErr, err, ri.logger)
}
