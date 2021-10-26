package db

const (
	saveVoteCountingStepResultQuery = "saveVoteCountingStepResult.sql"
	saveVoteCountingResultQuery     = "saveVoteCountingResult.sql"
	saveProofProposalQuery          = "saveProofProposal.sql"
	saveBlockProposalQuery          = "saveBlockProposal.sql"
	deleteVoteCountingOldDataQuery  = "deleteVoteCountingOldData.sql"

	blockRange = 7 * 24 * 60 * 3 / 2
)

func (a *postgresAccessor) SaveVoteCountingStepResult(value *VoteCountingStepResult) error {
	_, err := a.db.Exec(a.getQuery(saveVoteCountingStepResultQuery), &voteCountingStepResult{value})
	return getResultError(err)
}

func (a *postgresAccessor) SaveVoteCountingResult(value *VoteCountingResult) error {
	_, err := a.db.Exec(a.getQuery(saveVoteCountingResultQuery), &voteCountingResult{value})
	return getResultError(err)
}

func (a *postgresAccessor) SaveProofProposal(value *ProofProposal) error {
	_, err := a.db.Exec(a.getQuery(saveProofProposalQuery), &proofProposal{value})
	return getResultError(err)
}

func (a *postgresAccessor) SaveBlockProposal(value *BlockProposal) error {
	_, err := a.db.Exec(a.getQuery(saveBlockProposalQuery), &blockProposal{value})
	return getResultError(err)
}

func (a *postgresAccessor) DeleteVoteCountingOldData() error {
	_, err := a.db.Exec(a.getQuery(deleteVoteCountingOldDataQuery), blockRange)
	return getResultError(err)
}
