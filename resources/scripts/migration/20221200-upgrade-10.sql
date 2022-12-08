DROP TYPE tp_refundable_oracle_lock_contract CASCADE;

ALTER TABLE refundable_oracle_lock_contracts
    ADD COLUMN oracle_voting_fee_new integer;