ALTER TABLE oracle_voting_contracts
    DROP CONSTRAINT fec_contract_tx_id_fkey;
ALTER TABLE oracle_lock_contracts
    DROP CONSTRAINT oracle_lock_contracts_contract_tx_id_fkey;
ALTER TABLE refundable_oracle_lock_contracts
    DROP CONSTRAINT refundable_oracle_lock_contracts_contract_tx_id_fkey;
ALTER TABLE time_lock_contracts
    DROP CONSTRAINT time_lock_contracts_contract_tx_id_fkey;
ALTER TABLE multisig_contracts
    DROP CONSTRAINT multisig_contracts_contract_tx_id_fkey;

DROP INDEX contract_tx_balance_updates_api_idx_1;
DROP INDEX contract_tx_balance_updates_api_idx_2;

ALTER TABLE contract_tx_balance_updates
    ADD COLUMN contract_address_id BIGINT;
ALTER TABLE contract_tx_balance_updates
    ADD COLUMN base_contract_address_id BIGINT;

UPDATE contract_tx_balance_updates
SET contract_address_id = c.contract_address_id
FROM contracts c
WHERE contract_tx_balance_updates.contract_tx_id = c.tx_id;

UPDATE contract_tx_balance_updates
SET base_contract_address_id = c.contract_address_id
FROM contracts c
WHERE contract_tx_balance_updates.base_contract_tx_id = c.tx_id;

ALTER TABLE contracts
    DROP CONSTRAINT contracts_pkey;

ALTER TABLE contract_tx_balance_updates
    ALTER COLUMN contract_address_id SET NOT NULL;

ALTER TABLE contract_tx_balance_updates
    DROP COLUMN contract_tx_id;
ALTER TABLE contract_tx_balance_updates
    DROP COLUMN base_contract_tx_id;

