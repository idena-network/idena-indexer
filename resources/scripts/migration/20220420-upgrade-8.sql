DROP TABLE IF EXISTS vote_counting_step_results_old;
DROP TABLE IF EXISTS vote_counting_step_result_votes_old;
DROP TABLE IF EXISTS vote_counting_results_old;
DROP TABLE IF EXISTS vote_counting_result_validators_original_old;
DROP TABLE IF EXISTS vote_counting_result_validators_addresses_old;
DROP TABLE IF EXISTS vote_counting_result_cert_votes_old;
DROP TABLE IF EXISTS proof_proposals_old;
DROP TABLE IF EXISTS block_proposals_old;
DROP TABLE IF EXISTS vote_counting_step_results;
DROP TABLE IF EXISTS vote_counting_step_result_votes;
DROP TABLE IF EXISTS vote_counting_results;
DROP TABLE IF EXISTS vote_counting_result_validators_original;
DROP TABLE IF EXISTS vote_counting_result_validators_addresses;
DROP TABLE IF EXISTS vote_counting_result_cert_votes;
DROP TABLE IF EXISTS proof_proposals;
DROP TABLE IF EXISTS block_proposals;

DROP PROCEDURE IF EXISTS save_vote_counting_step_result;
DROP PROCEDURE IF EXISTS save_vote_counting_result;
DROP PROCEDURE IF EXISTS save_proof_proposal;
DROP PROCEDURE IF EXISTS save_block_proposal;
DROP PROCEDURE IF EXISTS delete_old_vote_counting_data;

DROP TYPE tp_oracle_voting_contract_call_vote_proof CASCADE;
DROP TYPE tp_oracle_voting_contract_call_vote CASCADE;

ALTER TABLE oracle_voting_contract_call_vote_proofs
    ADD COLUMN discriminated boolean;
ALTER TABLE oracle_voting_contract_call_votes
    ADD COLUMN discriminated boolean;

DROP TYPE tp_total_epoch_reward CASCADE;

ALTER TABLE total_rewards
    ADD COLUMN staking numeric(30, 18);
ALTER TABLE total_rewards
    ADD COLUMN candidate numeric(30, 18);
ALTER TABLE total_rewards
    ADD COLUMN staking_share numeric(30, 18);
ALTER TABLE total_rewards
    ADD COLUMN candidate_share numeric(30, 18);

DROP PROCEDURE IF EXISTS save_epoch_result;
DROP PROCEDURE IF EXISTS save_epoch_rewards;

ALTER TABLE validation_reward_summaries
    ADD COLUMN candidate numeric(30, 18);
ALTER TABLE validation_reward_summaries
    ADD COLUMN candidate_missed numeric(30, 18);
ALTER TABLE validation_reward_summaries
    ADD COLUMN candidate_missed_reason smallint;
ALTER TABLE validation_reward_summaries
    ADD COLUMN staking numeric(30, 18);
ALTER TABLE validation_reward_summaries
    ADD COLUMN staking_missed numeric(30, 18);
ALTER TABLE validation_reward_summaries
    ADD COLUMN staking_missed_reason smallint;

ALTER TABLE delegatee_total_validation_rewards
    ADD COLUMN candidate_balance numeric(30, 18);
ALTER TABLE delegatee_total_validation_rewards
    ADD COLUMN staking_balance numeric(30, 18);

ALTER TABLE delegatee_validation_rewards
    ADD COLUMN candidate_balance numeric(30, 18);
ALTER TABLE delegatee_validation_rewards
    ADD COLUMN staking_balance numeric(30, 18);