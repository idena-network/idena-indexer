package types

var (
	timeLockCallMethods = map[uint8]string{0: "Transfer"}

	oracleVotingCallMethods = map[uint8]string{
		0: "Start",
		1: "VoteProof",
		2: "Vote",
		3: "Finish",
		4: "Prolong",
		5: "AddStake",
	}

	oracleLockCallMethods = map[uint8]string{0: "Push", 1: "CheckOracleVoting"}

	multisigCallMethods = map[uint8]string{
		0: "Add",
		1: "Send",
		2: "Push",
	}

	refundableOracleLockCallMethods = map[uint8]string{
		0: "Deposit",
		1: "Push",
		2: "Refund",
	}
)

const (
	contractTimeLock             = "TimeLock"
	contractOracleVoting         = "OracleVoting"
	contractOracleLock           = "OracleLock"
	contractMultisig             = "Multisig"
	contractRefundableOracleLock = "RefundableOracleLock"
)

func GetCallMethodName(contractType string, callMethodCode uint8) string {
	switch contractType {
	case contractTimeLock:
		return timeLockCallMethods[callMethodCode]
	case contractOracleVoting:
		return oracleVotingCallMethods[callMethodCode]
	case contractOracleLock:
		return oracleLockCallMethods[callMethodCode]
	case contractMultisig:
		return multisigCallMethods[callMethodCode]
	case contractRefundableOracleLock:
		return refundableOracleLockCallMethods[callMethodCode]
	}
	return ""
}
