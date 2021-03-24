ALTER TABLE epoch_identities
    ADD COLUMN delegatee_address_id bigint;

ALTER TABLE blocks
    ADD COLUMN pool_validators_count integer;

DROP TYPE tp_epoch_identity CASCADE;

DROP PROCEDURE save_block;

ALTER TABLE oracle_voting_contract_call_vote_proofs
    ADD COLUMN secret_votes_count bigint;

ALTER TABLE oracle_voting_contract_call_votes
    ADD COLUMN option_votes bigint;
ALTER TABLE oracle_voting_contract_call_votes
    ADD COLUMN option_all_votes bigint;
ALTER TABLE oracle_voting_contract_call_votes
    ADD COLUMN secret_votes_count bigint;
ALTER TABLE oracle_voting_contract_call_votes
    ADD COLUMN delegatee_address_id bigint;
ALTER TABLE oracle_voting_contract_call_votes
    ADD COLUMN prev_pool_vote smallint;
ALTER TABLE oracle_voting_contract_call_votes
    ADD COLUMN prev_option_votes bigint;

ALTER TABLE oracle_voting_contract_results
    ADD COLUMN all_votes_count bigint;
ALTER TABLE oracle_voting_contract_results_changes
    ADD COLUMN all_votes_count bigint;

DROP TYPE tp_oracle_voting_contract_call_vote_proof CASCADE;
DROP TYPE tp_oracle_voting_contract_call_vote CASCADE;

DROP PROCEDURE update_oracle_voting_contract_summaries;

ALTER TABLE oracle_voting_contract_summaries
    ADD COLUMN secret_votes_count bigint;
ALTER TABLE oracle_voting_contract_summaries
    ADD COLUMN epoch_without_growth smallint;

ALTER TABLE oracle_voting_contract_summaries_changes
    ADD COLUMN secret_votes_count bigint;
ALTER TABLE oracle_voting_contract_summaries_changes
    ADD COLUMN epoch_without_growth smallint;

ALTER TABLE oracle_voting_contract_call_prolongations
    ADD COLUMN epoch_without_growth smallint;

ALTER TABLE oracle_voting_contract_call_prolongations
    ADD COLUMN prolong_vote_count bigint;