DROP FUNCTION save_addrs_and_txs;
DROP PROCEDURE save_tx_receipts;
DROP PROCEDURE save_block;
DROP TYPE tp_tx_receipt CASCADE;

ALTER TABLE tx_receipts
    ADD COLUMN contract_address_id bigint;
ALTER TABLE tx_receipts
    ADD COLUMN "from" bigint;
ALTER TABLE tx_receipts
    ADD COLUMN action_result bytea;
ALTER TABLE tx_receipts
    ALTER COLUMN error_msg TYPE character varying(500);
ALTER TABLE blocks
    ADD COLUMN used_gas integer;
ALTER TABLE transactions
    ADD COLUMN used_gas integer;
ALTER TABLE contract_tx_balance_updates
    ADD COLUMN base_contract_tx_id bigint;

ALTER TABLE contract_tx_balance_updates
    DROP CONSTRAINT contract_tx_balance_updates_pkey;
ALTER TABLE contract_tx_balance_updates
    DROP CONSTRAINT contract_tx_balance_updates_contract_tx_id_fkey;

DO
$$
    BEGIN
        UPDATE contract_tx_balance_updates SET base_contract_tx_id = contract_tx_id;

        UPDATE tx_receipts
        SET contract_address_id = t.contract_address_id,
            "from"              = t."from"
        FROM (SELECT r.tx_id, t."from", coalesce(t."to", c.contract_address_id) contract_address_id
              FROM tx_receipts r
                       LEFT JOIN transactions t ON t.id = r.tx_id
                       LEFT JOIN contracts c ON c.tx_id = r.tx_id) t
        WHERE tx_receipts.tx_id = t.tx_id;
    END
$$;