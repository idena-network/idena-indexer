DROP TYPE tp_oracle_voting_contract_call_vote_proof CASCADE;
DROP TYPE tp_oracle_voting_contract_call_vote CASCADE;

ALTER TABLE oracle_voting_contract_call_vote_proofs
    ADD COLUMN discriminated_newbie boolean;
ALTER TABLE oracle_voting_contract_call_vote_proofs
    ADD COLUMN discriminated_delegation boolean;
ALTER TABLE oracle_voting_contract_call_vote_proofs
    ADD COLUMN discriminated_stake boolean;
ALTER TABLE oracle_voting_contract_call_votes
    ADD COLUMN discriminated_newbie boolean;
ALTER TABLE oracle_voting_contract_call_votes
    ADD COLUMN discriminated_delegation boolean;
ALTER TABLE oracle_voting_contract_call_votes
    ADD COLUMN discriminated_stake boolean;

ALTER TABLE epochs
    ADD COLUMN discrimination_stake_threshold numeric(30, 18);