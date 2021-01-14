ALTER TABLE contract_tx_balance_updates
    RENAME TO contract_tx_balance_updates_old;

ALTER TABLE contract_tx_balance_updates_old
    RENAME CONSTRAINT contract_tx_balance_updates_pkey TO contract_tx_balance_updates_old_pkey;

ALTER TABLE contract_tx_balance_updates_old
    RENAME CONSTRAINT contract_tx_balance_updates_contract_tx_id_fkey TO contract_tx_balance_updates_old_contract_tx_id_fkey;

ALTER TABLE contract_tx_balance_updates_old
    RENAME CONSTRAINT contract_tx_balance_updates_address_id_fkey TO contract_tx_balance_updates_old_address_id_fkey;

ALTER TABLE contract_tx_balance_updates_old
    RENAME CONSTRAINT contract_tx_balance_updates_contract_type_fkey TO contract_tx_balance_updates_old_contract_type_fkey;

ALTER TABLE contract_tx_balance_updates_old
    RENAME CONSTRAINT contract_tx_balance_updates_tx_id_fkey TO contract_tx_balance_updates_old_tx_id_fkey;

CREATE SEQUENCE IF NOT EXISTS contract_tx_balance_updates_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE TABLE IF NOT EXISTS contract_tx_balance_updates
(
    id             bigint NOT NULL DEFAULT nextval('contract_tx_balance_updates_id_seq'::regclass),
    contract_tx_id bigint NOT NULL,
    address_id     bigint NOT NULL,
    contract_type  bigint NOT NULL,
    tx_id          bigint NOT NULL,
    call_method    smallint,
    balance_old    numeric(30, 18),
    balance_new    numeric(30, 18),
    CONSTRAINT contract_tx_balance_updates_pkey PRIMARY KEY (tx_id, address_id),
    CONSTRAINT contract_tx_balance_updates_contract_tx_id_fkey FOREIGN KEY (contract_tx_id)
        REFERENCES contracts (tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT contract_tx_balance_updates_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT contract_tx_balance_updates_contract_type_fkey FOREIGN KEY (contract_type)
        REFERENCES dic_contract_types (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT contract_tx_balance_updates_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE INDEX IF NOT EXISTS contract_tx_balance_updates_api_idx_1 on contract_tx_balance_updates (contract_tx_id, address_id, tx_id desc);
CREATE INDEX IF NOT EXISTS contract_tx_balance_updates_api_idx_2 on contract_tx_balance_updates (contract_tx_id, id desc);

INSERT INTO contract_tx_balance_updates (contract_tx_id,
                                         address_id,
                                         contract_type,
                                         tx_id,
                                         call_method,
                                         balance_old,
                                         balance_new)
SELECT contract_tx_id,
       address_id,
       contract_type,
       tx_id,
       call_method,
       balance_old,
       balance_new
FROM contract_tx_balance_updates_old
ORDER BY tx_id;


DROP TABLE contract_tx_balance_updates_old;