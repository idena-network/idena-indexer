DROP TYPE tp_oracle_voting_contract CASCADE;

ALTER TABLE oracle_voting_contracts
    ADD COLUMN owner_deposit numeric(48, 18);
ALTER TABLE oracle_voting_contracts
    ADD COLUMN oracle_reward_fund numeric(48, 18);
ALTER TABLE oracle_voting_contracts
    ADD COLUMN refund_recipient_address_id bigint;