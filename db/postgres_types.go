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
	var isFirst bool
	var convertAndAddAnswer = func(isShort bool, flipCid string, answer Answer) {
		convertedAnswer := postgresAnswer{
			address: answer.Address,
			answer:  answer.Answer,
			point:   answer.Point,
			isShort: isShort,
			grade:   answer.Grade,
		}
		if isFirst {
			convertedAnswer.flipCid = flipCid
			isFirst = false
		}
		convertedAnswers = append(convertedAnswers, convertedAnswer)
	}
	for _, s := range stats {
		isFirst = true
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
}

func (v postgresEpochIdentity) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v)",
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
	var isFirst bool
	var convertAndAddFlipToSolve = func(address string, flipCid string, isShort bool) {
		converted := postgresFlipToSolve{
			cid:     flipCid,
			isShort: isShort,
		}
		if isFirst {
			converted.address = address
			isFirst = false
		}
		convertedFlipsToSolve = append(convertedFlipsToSolve, converted)
	}
	for _, identity := range identities {
		isFirst = true
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
