DROP TYPE tp_refundable_oracle_lock_contract CASCADE;

ALTER TABLE refundable_oracle_lock_contracts
    ADD COLUMN oracle_voting_fee_new integer;

DROP TYPE tp_oracle_voting_contract CASCADE;
ALTER TABLE oracle_voting_contracts
    ADD COLUMN network_size bigint;
ALTER TABLE oracle_voting_contract_call_starts
    ADD COLUMN committee_size bigint;
ALTER TABLE oracle_voting_contract_call_prolongations
    ADD COLUMN committee_size bigint;