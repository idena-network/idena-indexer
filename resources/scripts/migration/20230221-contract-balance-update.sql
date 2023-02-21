DROP TYPE tp_balance_update CASCADE;

UPDATE dic_balance_update_reasons
SET name = 'Contract'
WHERE name = 'EmbeddedContract';

ALTER TABLE balance_updates
    ADD COLUMN contract_address_id bigint;