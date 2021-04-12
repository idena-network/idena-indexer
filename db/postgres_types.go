package db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
	"strings"
)

func (v *MiningReward) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v)", v.Address, v.Balance, v.Stake, v.Proposer), nil
}

func (v Balance) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v)", v.Address, v.Balance, v.Stake), nil
}

func (v *BalanceUpdate) Value() (driver.Value, error) {
	var txHash string
	if v.TxHash != nil {
		txHash = conversion.ConvertHash(*v.TxHash)
	}
	return fmt.Sprintf("(%v,%v,%v,%v,%v,%v,%v,%v,%v)",
		conversion.ConvertAddress(v.Address),
		blockchain.ConvertToFloat(v.BalanceOld),
		blockchain.ConvertToFloat(v.StakeOld),
		blockchain.ConvertToFloat(v.PenaltyOld),
		blockchain.ConvertToFloat(v.BalanceNew),
		blockchain.ConvertToFloat(v.StakeNew),
		blockchain.ConvertToFloat(v.PenaltyNew),
		txHash,
		v.Reason), nil
}

func (v Birthday) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)", v.Address, v.BirthEpoch), nil
}

func (v *MemPoolFlipKey) Value() (driver.Value, error) {
	if v == nil {
		return nil, nil
	}
	return fmt.Sprintf("(%v,%v)", v.Address, v.Key), nil
}

func (v Transaction) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v,%v,%v,%v,%v,%v,%v)",
		v.Hash, v.Type, v.From, v.To, v.Amount, v.Tips, v.MaxFee, v.Fee, v.Size, v.Raw), nil
}

func (v ActivationTxTransfer) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)", v.TxHash, v.BalanceTransfer), nil
}

func (v KillTxTransfer) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)", v.TxHash, v.StakeTransfer), nil
}

func (v KillInviteeTxTransfer) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)", v.TxHash, v.StakeTransfer), nil
}

func (v ActivationTx) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)", v.TxHash, v.InviteTxHash), nil
}

func (v KillInviteeTx) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)", v.TxHash, v.InviteTxHash), nil
}

func (v DeletedFlip) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)", v.TxHash, v.Cid), nil
}

func (v *BadAuthor) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)",
		v.Address,
		v.Reason,
	), nil
}

func (v *TotalRewards) Value() (driver.Value, error) {
	if v == nil {
		return nil, nil
	}
	return fmt.Sprintf("(%v,%v,%v,%v,%v,%v,%v,%v,%v)",
		v.Total,
		v.Validation,
		v.Flips,
		v.Invitations,
		v.FoundationPayouts,
		v.ZeroWalletFund,
		v.ValidationShare,
		v.FlipsShare,
		v.InvitationsShare,
	), nil
}

func (v *Reward) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v)",
		v.Address,
		v.Balance,
		v.Stake,
		v.Type,
	), nil
}

func (v *RewardedInvite) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)",
		v.TxHash,
		v.Type,
	), nil
}

func (v *SavedInviteRewards) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v)",
		v.Address,
		v.Type,
		v.Count,
	), nil
}

func (v *ReportedFlipReward) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v)",
		v.Address,
		v.Cid,
		v.Balance,
		v.Stake,
	), nil
}

func (p Penalty) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)", p.Address, p.Penalty), nil
}

func (v FlipWords) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v)", v.Cid, v.Word1, v.Word2, v.TxHash), nil
}

func (v *FailedFlipContent) Value() (driver.Value, error) {
	var timestamp int64
	if v.NextAttemptTimestamp != nil {
		timestamp = v.NextAttemptTimestamp.Int64()
	}
	return fmt.Sprintf("(%v,%v,%v)",
		v.Cid,
		v.AttemptsLimitReached,
		timestamp,
	), nil
}

func negativeIfNil(v *big.Int) decimal.Decimal {
	if v == nil {
		return decimal.NewFromInt(-1)
	}
	return blockchain.ConvertToFloat(v)
}

func negativeIfNilByte(v *byte) int64 {
	if v == nil {
		return -1
	}
	return int64(*v)
}

func negativeIfNilUint64(v *uint64) int64 {
	if v == nil {
		return -1
	}
	return int64(*v)
}

func (v *OracleVotingContract) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		conversion.ConvertAddress(v.ContractAddress),
		blockchain.ConvertToFloat(v.Stake),
		v.StartTime,
		v.VotingDuration,
		negativeIfNil(v.VotingMinPayment),
		hex.EncodeToString(v.Fact),
		v.State,
		v.PublicVotingDuration,
		v.WinnerThreshold,
		v.Quorum,
		v.CommitteeSize,
		v.OwnerFee,
	), nil
}

type oracleVotingContractCallStart struct {
	TxHash           string   `json:"txHash"`
	State            byte     `json:"state"`
	StartHeight      uint64   `json:"startBlockHeight"`
	Epoch            uint16   `json:"epoch"`
	VotingMinPayment *string  `json:"votingMinPayment,omitempty"`
	VrfSeed          bytes    `json:"vrfSeed"`
	Committee        []string `json:"committee,omitempty"`
}

func (v *OracleVotingContractCallStart) Value() (driver.Value, error) {
	res := oracleVotingContractCallStart{}
	res.TxHash = conversion.ConvertHash(v.TxHash)
	res.State = v.State
	res.StartHeight = v.StartHeight
	res.Epoch = v.Epoch
	if v.VotingMinPayment != nil {
		s := blockchain.ConvertToFloat(v.VotingMinPayment).String()
		res.VotingMinPayment = &s
	}
	res.VrfSeed = v.VrfSeed
	if len(v.Committee) > 0 {
		res.Committee = make([]string, len(v.Committee))
		for i, addr := range v.Committee {
			res.Committee[i] = conversion.ConvertAddress(addr)
		}
	}
	return json.Marshal(res)
}

func (v *OracleVotingContractCallVoteProof) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		hex.EncodeToString(v.VoteHash),
		negativeIfNilUint64(v.NewSecretVotesCount),
	), nil
}

func (v *OracleVotingContractCallVote) Value() (driver.Value, error) {
	var delegatee string
	if v.Delegatee != nil && !v.Delegatee.IsEmpty() {
		delegatee = conversion.ConvertAddress(*v.Delegatee)
	}
	return fmt.Sprintf("(%v,%v,%v,%v,%v,%v,%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		v.Vote,
		hex.EncodeToString(v.Salt),
		negativeIfNilUint64(v.OptionVotes),
		negativeIfNilUint64(v.OptionAllVotes),
		negativeIfNilUint64(v.SecretVotesCount),
		delegatee,
		negativeIfNilByte(v.PrevPoolVote),
		negativeIfNilUint64(v.PrevOptionVotes),
	), nil
}

func (v *OracleVotingContractCallFinish) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		v.State,
		negativeIfNilByte(v.Result),
		blockchain.ConvertToFloat(v.Fund),
		blockchain.ConvertToFloat(v.OracleReward),
		blockchain.ConvertToFloat(v.OwnerReward),
	), nil
}

type oracleVotingContractCallProlongation struct {
	TxHash             string   `json:"txHash"`
	Epoch              uint16   `json:"epoch"`
	StartHeight        *uint64  `json:"startBlockHeight"`
	VrfSeed            bytes    `json:"vrfSeed"`
	EpochWithoutGrowth *byte    `json:"epochWithoutGrowth"`
	ProlongVoteCount   *uint64  `json:"prolongVoteCount"`
	Committee          []string `json:"committee,omitempty"`
}

func (v *OracleVotingContractCallProlongation) Value() (driver.Value, error) {
	res := oracleVotingContractCallProlongation{}
	res.TxHash = conversion.ConvertHash(v.TxHash)
	res.Epoch = v.Epoch
	res.StartHeight = v.StartBlock
	res.VrfSeed = v.VrfSeed
	res.EpochWithoutGrowth = v.EpochWithoutGrowth
	res.ProlongVoteCount = v.ProlongVoteCount
	if len(v.Committee) > 0 {
		res.Committee = make([]string, len(v.Committee))
		for i, addr := range v.Committee {
			res.Committee[i] = conversion.ConvertAddress(addr)
		}
	}
	return json.Marshal(res)
}

func (v *OracleVotingContractCallAddStake) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v)",
		conversion.ConvertHash(v.TxHash),
	), nil
}

func (v *OracleVotingContractTermination) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		negativeIfNil(v.Fund),
		negativeIfNil(v.OracleReward),
		negativeIfNil(v.OwnerReward),
	), nil
}

func (v *OracleLockContract) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v,%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		conversion.ConvertAddress(v.ContractAddress),
		blockchain.ConvertToFloat(v.Stake),
		conversion.ConvertAddress(v.OracleVotingAddress),
		v.ExpectedValue,
		conversion.ConvertAddress(v.SuccessAddress),
		conversion.ConvertAddress(v.FailAddress),
	), nil
}

func (v *OracleLockContractCallPush) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		v.Success,
		v.OracleVotingResult,
		blockchain.ConvertToFloat(v.Transfer),
	), nil
}

func (v *OracleLockContractCallCheckOracleVoting) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)",
		conversion.ConvertHash(v.TxHash),
		negativeIfNilByte(v.OracleVotingResult),
	), nil
}

func (v *OracleLockContractTermination) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)",
		conversion.ConvertHash(v.TxHash),
		conversion.ConvertAddress(v.Dest),
	), nil
}

func (v *RefundableOracleLockContract) Value() (driver.Value, error) {
	var successAddress, failAddress string
	if v.SuccessAddress != nil {
		successAddress = conversion.ConvertAddress(*v.SuccessAddress)
	}
	if v.FailAddress != nil {
		failAddress = conversion.ConvertAddress(*v.FailAddress)
	}
	return fmt.Sprintf("(%v,%v,%v,%v,%v,%v,%v,%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		conversion.ConvertAddress(v.ContractAddress),
		blockchain.ConvertToFloat(v.Stake),
		conversion.ConvertAddress(v.OracleVotingAddress),
		v.ExpectedValue,
		successAddress,
		failAddress,
		v.RefundDelay,
		v.DepositDeadline,
		v.OracleVotingFee,
	), nil
}

func (v *RefundableOracleLockContractCallDeposit) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		blockchain.ConvertToFloat(v.OwnSum),
		blockchain.ConvertToFloat(v.Sum),
		blockchain.ConvertToFloat(v.Fee),
	), nil
}

func (v *RefundableOracleLockContractCallPush) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		v.State,
		v.OracleVotingExists,
		negativeIfNilByte(v.OracleVotingResult),
		blockchain.ConvertToFloat(v.Transfer),
		v.RefundBlock,
	), nil
}

func (v *RefundableOracleLockContractCallRefund) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		blockchain.ConvertToFloat(v.Balance),
		v.Coef,
	), nil
}

func (v *RefundableOracleLockContractTermination) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)",
		conversion.ConvertHash(v.TxHash),
		conversion.ConvertAddress(v.Dest),
	), nil
}

func (v *TimeLockContract) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		conversion.ConvertAddress(v.ContractAddress),
		blockchain.ConvertToFloat(v.Stake),
		v.Timestamp,
	), nil
}

func (v *TimeLockContractCallTransfer) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		conversion.ConvertAddress(v.Dest),
		blockchain.ConvertToFloat(v.Amount),
	), nil
}

func (v *TimeLockContractTermination) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)",
		conversion.ConvertHash(v.TxHash),
		conversion.ConvertAddress(v.Dest),
	), nil
}

func (v *MultisigContract) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		conversion.ConvertAddress(v.ContractAddress),
		blockchain.ConvertToFloat(v.Stake),
		v.MinVotes,
		v.MaxVotes,
		v.State,
	), nil
}

func (v *MultisigContractCallAdd) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		conversion.ConvertAddress(v.Address),
		negativeIfNilByte(v.NewState),
	), nil
}

func (v *MultisigContractCallSend) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		conversion.ConvertAddress(v.Dest),
		blockchain.ConvertToFloat(v.Amount),
	), nil
}

func (v *MultisigContractCallPush) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		conversion.ConvertAddress(v.Dest),
		blockchain.ConvertToFloat(v.Amount),
		v.VoteAddressCnt,
		v.VoteAmountCnt,
	), nil
}

func (v *MultisigContractTermination) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)",
		conversion.ConvertHash(v.TxHash),
		conversion.ConvertAddress(v.Dest),
	), nil
}

type contractTxBalanceUpdates struct {
	TxHash             string                     `json:"txHash"`
	ContractAddress    string                     `json:"contractAddress"`
	ContractCallMethod *uint8                     `json:"contractCallMethod,omitempty"`
	Updates            []*contractTxBalanceUpdate `json:"updates"`
}

type contractTxBalanceUpdate struct {
	Address    string           `json:"address"`
	BalanceOld *decimal.Decimal `json:"balanceOld"`
	BalanceNew *decimal.Decimal `json:"balanceNew"`
}

func (v *ContractTxBalanceUpdates) Value() (driver.Value, error) {
	res := contractTxBalanceUpdates{}
	res.TxHash = conversion.ConvertHash(v.TxHash)
	res.ContractAddress = conversion.ConvertAddress(v.ContractAddress)
	res.ContractCallMethod = v.ContractCallMethod
	res.Updates = make([]*contractTxBalanceUpdate, len(v.Updates))
	for i, update := range v.Updates {
		var balanceOld, balanceNew *decimal.Decimal
		if update.BalanceOld != nil {
			v := blockchain.ConvertToFloat(update.BalanceOld)
			balanceOld = &v
		}
		if update.BalanceNew != nil {
			v := blockchain.ConvertToFloat(update.BalanceNew)
			balanceNew = &v
		}
		res.Updates[i] = &contractTxBalanceUpdate{
			Address:    conversion.ConvertAddress(update.Address),
			BalanceOld: balanceOld,
			BalanceNew: balanceNew,
		}
	}
	return json.Marshal(res)
}

func (v *TxReceipt) Value() (driver.Value, error) {
	errorMsg := v.Error
	if len(errorMsg) > 50 {
		errorMsg = errorMsg[:50]
	}
	return fmt.Sprintf("(%v,%v,%v,%v,%v,%v)",
		conversion.ConvertHash(v.TxHash),
		v.Success,
		v.GasUsed,
		blockchain.ConvertToFloat(v.GasCost),
		v.Method,
		errorMsg,
	), nil
}

type flipContent struct {
	Cid    string  `json:"cid"`
	Pics   []bytes `json:"pics"`
	Orders [][]int `json:"orders,omitempty"`
	Icon   bytes   `json:"icon,omitempty"`
}

type bytes []byte

func (b bytes) MarshalText() ([]byte, error) {
	result := make([]byte, len(b)*2)
	hex.Encode(result[:], b)
	return result, nil
}

func (v *FlipContent) Value() (driver.Value, error) {
	fc := flipContent{
		Cid:  v.Cid,
		Icon: v.Icon,
	}
	for _, pic := range v.Pics {
		fc.Pics = append(fc.Pics, pic)
	}
	for i, answerOrders := range v.Orders {
		for j, order := range answerOrders {
			if j == 0 {
				fc.Orders = append(fc.Orders, make([]int, len(answerOrders)))
			}
			fc.Orders[i][j] = int(order)
		}
	}
	return json.Marshal(fc)
}

type postgresAddrBurntCoins struct {
	*BurntCoins
	address string
	getTxId func(hash string) (int64, error)
}

func (v postgresAddrBurntCoins) Value() (driver.Value, error) {
	var txId int64
	if len(v.TxHash) > 0 {
		var err error
		if txId, err = v.getTxId(v.TxHash); err != nil {
			return nil, errors.Wrap(err, "unable to create db value")
		}
	}
	return fmt.Sprintf("(%v,%v,%v,%v)", v.address, v.Amount, v.Reason, txId), nil
}

func getPostgresBurntCoins(v map[common.Address][]*BurntCoins, getTxId func(hash string) (int64, error)) []postgresAddrBurntCoins {
	var values []postgresAddrBurntCoins
	for addr, coins := range v {
		for i, c := range coins {
			var convertedAddr string
			if i == 0 {
				convertedAddr = conversion.ConvertAddress(addr)
			}
			values = append(values, postgresAddrBurntCoins{
				address:    convertedAddr,
				BurntCoins: c,
				getTxId:    getTxId,
			})
		}
	}
	return values
}

type postgresAddress struct {
	address     string
	isTemporary bool
}

func (v postgresAddress) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)", v.address, v.isTemporary), nil
}

type postgresAddressStateChange struct {
	address  string
	newState uint8
	txHash   string
}

func (v postgresAddressStateChange) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v)", v.address, v.newState, v.txHash), nil
}

func getPostgresAddressesAndAddressStateChangesArrays(addresses []Address) (addressesArray, addressStateChangesArray interface {
	driver.Valuer
	sql.Scanner
}) {
	var convertedAddresses []postgresAddress
	var convertedAddressStateChanges []postgresAddressStateChange
	for _, address := range addresses {
		convertedAddresses = append(convertedAddresses, postgresAddress{
			address:     address.Address,
			isTemporary: address.IsTemporary,
		})
		for _, addressStateChange := range address.StateChanges {
			convertedAddressStateChanges = append(convertedAddressStateChanges, postgresAddressStateChange{
				address:  address.Address,
				newState: addressStateChange.NewState,
				txHash:   addressStateChange.TxHash,
			})
		}
	}
	return pq.Array(convertedAddresses), pq.Array(convertedAddressStateChanges)
}

type txHashId struct {
	Hash string
	Id   int64
}

func (v *txHashId) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("wrong txHashId")
	}
	s := string(b)
	sArr := strings.Split(strings.Trim(s, "()"), ",")
	if len(sArr) != 2 {
		return errors.New("unknown txHashId: " + s)
	}
	id, err := strconv.ParseInt(sArr[1], 0, 64)
	if err != nil {
		return errors.Wrap(err, "invalid txHashId: "+s)
	}
	v.Hash, v.Id = sArr[0], id
	return nil
}

type postgresAnswer struct {
	flipCid string
	address string
	isShort bool
	answer  byte
	point   float32
	grade   byte
}

func (v postgresAnswer) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v,%v,%v)", v.flipCid, v.address, v.isShort, v.answer,
		v.point, v.grade), nil
}

type postgresFlipsState struct {
	flipCid string
	answer  byte
	status  byte
	grade   byte
}

func (v postgresFlipsState) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v)", v.flipCid, v.answer, v.status, v.grade), nil
}

func getFlipStatsArrays(stats []FlipStats) (answersArray, statesArray interface {
	driver.Valuer
	sql.Scanner
}, shortAnswerCountsByAddr, longAnswerCountsByAddr, wrongWordsFlipsCountsByAddr map[string]int) {
	shortAnswerCountsByAddr, longAnswerCountsByAddr, wrongWordsFlipsCountsByAddr = make(map[string]int), make(map[string]int), make(map[string]int)
	var convertedAnswers []postgresAnswer
	var convertedStates []postgresFlipsState
	var convertAndAddAnswer = func(isShort bool, flipCid string, answer Answer) {
		convertedAnswer := postgresAnswer{
			address: answer.Address,
			answer:  answer.Answer,
			point:   answer.Point,
			isShort: isShort,
			grade:   answer.Grade,
			flipCid: flipCid,
		}
		convertedAnswers = append(convertedAnswers, convertedAnswer)
	}
	for _, s := range stats {
		for _, answer := range s.ShortAnswers {
			convertAndAddAnswer(true, s.Cid, answer)
			shortAnswerCountsByAddr[answer.Address]++
		}
		for _, answer := range s.LongAnswers {
			convertAndAddAnswer(false, s.Cid, answer)
			longAnswerCountsByAddr[answer.Address]++
		}
		convertedStates = append(convertedStates, postgresFlipsState{
			flipCid: s.Cid,
			answer:  s.Answer,
			status:  s.Status,
			grade:   s.Grade,
		})
		const gradeReported = 1
		if s.Grade == byte(gradeReported) {
			wrongWordsFlipsCountsByAddr[s.Author]++
		}
	}
	return pq.Array(convertedAnswers), pq.Array(convertedStates), shortAnswerCountsByAddr, longAnswerCountsByAddr, wrongWordsFlipsCountsByAddr
}

type postgresEpochIdentity struct {
	address          string
	state            uint8
	shortPoint       float32
	shortFlips       uint32
	totalShortPoint  float32
	totalShortFlips  uint32
	longPoint        float32
	longFlips        uint32
	approved         bool
	missed           bool
	requiredFlips    uint8
	availableFlips   uint8
	madeFlips        uint8
	nextEpochInvites uint8
	birthEpoch       uint64
	shortAnswers     uint32
	longAnswers      uint32
	wrongWordsFlips  uint8
	delegateeAddress string
}

func (v postgresEpochIdentity) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v)",
		v.address,
		v.state,
		v.shortPoint,
		v.shortFlips,
		v.totalShortPoint,
		v.totalShortFlips,
		v.longPoint,
		v.longFlips,
		v.approved,
		v.missed,
		v.requiredFlips,
		v.availableFlips,
		v.madeFlips,
		v.nextEpochInvites,
		v.birthEpoch,
		v.shortAnswers,
		v.longAnswers,
		v.wrongWordsFlips,
		v.delegateeAddress,
	), nil
}

type postgresFlipToSolve struct {
	address string
	cid     string
	isShort bool
}

func (v postgresFlipToSolve) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v)",
		v.address,
		v.cid,
		v.isShort,
	), nil
}

func getEpochIdentitiesArrays(
	identities []EpochIdentity,
	shortAnswerCountsByAddr,
	longAnswerCountsByAdds,
	wrongWordsFlipsCountsByAddr map[string]int,
) (identitiesArray, flipsToSolveArray interface {
	driver.Valuer
	sql.Scanner
}) {
	var convertedIdentities []postgresEpochIdentity
	var convertedFlipsToSolve []postgresFlipToSolve
	var convertAndAddFlipToSolve = func(address string, flipCid string, isShort bool) {
		converted := postgresFlipToSolve{
			cid:     flipCid,
			isShort: isShort,
			address: address,
		}
		convertedFlipsToSolve = append(convertedFlipsToSolve, converted)
	}
	for _, identity := range identities {
		for _, flipCid := range identity.ShortFlipCidsToSolve {
			convertAndAddFlipToSolve(identity.Address, flipCid, true)
		}
		for _, flipCid := range identity.LongFlipCidsToSolve {
			convertAndAddFlipToSolve(identity.Address, flipCid, false)
		}
		var shortAnswers, longAnswers uint32
		if shortAnswerCountsByAddr != nil {
			shortAnswers = uint32(shortAnswerCountsByAddr[identity.Address])
		}
		if longAnswerCountsByAdds != nil {
			longAnswers = uint32(longAnswerCountsByAdds[identity.Address])
		}
		var wrongWordsFlips uint8
		if wrongWordsFlipsCountsByAddr != nil {
			wrongWordsFlips = uint8(wrongWordsFlipsCountsByAddr[identity.Address])
		}
		convertedIdentities = append(convertedIdentities, postgresEpochIdentity{
			address:          identity.Address,
			state:            identity.State,
			shortPoint:       identity.ShortPoint,
			shortFlips:       identity.ShortFlips,
			totalShortPoint:  identity.TotalShortPoint,
			totalShortFlips:  identity.TotalShortFlips,
			longPoint:        identity.LongPoint,
			longFlips:        identity.LongFlips,
			approved:         identity.Approved,
			missed:           identity.Missed,
			requiredFlips:    identity.RequiredFlips,
			availableFlips:   identity.AvailableFlips,
			madeFlips:        identity.MadeFlips,
			nextEpochInvites: identity.NextEpochInvites,
			birthEpoch:       identity.BirthEpoch,
			shortAnswers:     shortAnswers,
			longAnswers:      longAnswers,
			wrongWordsFlips:  wrongWordsFlips,
			delegateeAddress: identity.DelegateeAddress,
		})
	}
	return pq.Array(convertedIdentities), pq.Array(convertedFlipsToSolve)
}

type postgresRewardAge struct {
	address string
	age     uint16
}

func (v postgresRewardAge) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v)",
		v.address,
		v.age,
	), nil
}

func getRewardAgesArray(agesByAddress map[string]uint16) interface {
	driver.Valuer
	sql.Scanner
} {
	var converted []postgresRewardAge
	for address, age := range agesByAddress {
		converted = append(converted, postgresRewardAge{
			address: address,
			age:     age,
		})
	}
	return pq.Array(converted)
}

type rewardBounds struct {
	BoundType  byte            `json:"boundType"`
	MinAmount  decimal.Decimal `json:"minAmount"`
	MinAddress string          `json:"minAddress"`
	MaxAmount  decimal.Decimal `json:"maxAmount"`
	MaxAddress string          `json:"maxAddress"`
}

func (v *RewardBounds) Value() (driver.Value, error) {
	res := rewardBounds{}
	res.BoundType = v.Type
	res.MinAmount = blockchain.ConvertToFloat(v.Min.Amount)
	res.MinAddress = conversion.ConvertAddress(v.Min.Address)
	res.MaxAmount = blockchain.ConvertToFloat(v.Max.Amount)
	res.MaxAddress = conversion.ConvertAddress(v.Max.Address)
	return json.Marshal(res)
}

type data struct {
	DelegationSwitches []*delegationSwitch `json:"delegationSwitches,omitempty"`
}

func (v *data) Value() (driver.Value, error) {
	return json.Marshal(v)
}

type delegationSwitch struct {
	Delegator  string  `json:"delegator"`
	Delegatee  *string `json:"delegatee,omitempty"`
	BirthEpoch *uint16 `json:"birthEpoch,omitempty"`
}

func getData(delegationSwitches []*DelegationSwitch) *data {
	res := &data{}
	if len(delegationSwitches) > 0 {
		res.DelegationSwitches = make([]*delegationSwitch, 0, len(delegationSwitches))
		for _, incomingDelegationSwitch := range delegationSwitches {
			postgresDelegationSwitch := &delegationSwitch{
				Delegator:  conversion.ConvertAddress(incomingDelegationSwitch.Delegator),
				BirthEpoch: incomingDelegationSwitch.BirthEpoch,
			}
			if incomingDelegationSwitch.Delegatee != nil {
				delegatee := conversion.ConvertAddress(*incomingDelegationSwitch.Delegatee)
				postgresDelegationSwitch.Delegatee = &delegatee
			}
			res.DelegationSwitches = append(res.DelegationSwitches, postgresDelegationSwitch)
		}
	}
	return res
}

type restoredData struct {
	PoolSizes   []*poolSize   `json:"poolSizes,omitempty"`
	Delegations []*delegation `json:"delegations,omitempty"`
}

func (v *restoredData) Value() (driver.Value, error) {
	return json.Marshal(v)
}

type poolSize struct {
	Address string `json:"address"`
	Size    uint64 `json:"size"`
}

type delegation struct {
	Delegator  string  `json:"delegator"`
	Delegatee  string  `json:"delegatee"`
	BirthEpoch *uint16 `json:"birthEpoch,omitempty"`
}

func getRestoredData(data *RestoredData) *restoredData {
	if data == nil {
		return nil
	}
	res := &restoredData{}

	if len(data.PoolSizes) > 0 {
		res.PoolSizes = make([]*poolSize, 0, len(data.PoolSizes))
		for _, incomingPoolSize := range data.PoolSizes {
			res.PoolSizes = append(res.PoolSizes, &poolSize{
				Address: conversion.ConvertAddress(incomingPoolSize.Address),
				Size:    incomingPoolSize.Size,
			})
		}
	}

	if len(data.Delegations) > 0 {
		res.Delegations = make([]*delegation, 0, len(data.Delegations))
		for _, incomingDelegation := range data.Delegations {
			res.Delegations = append(res.Delegations, &delegation{
				Delegator:  conversion.ConvertAddress(incomingDelegation.Delegator),
				Delegatee:  conversion.ConvertAddress(incomingDelegation.Delegatee),
				BirthEpoch: incomingDelegation.BirthEpoch,
			})
		}
	}

	return res
}
