CREATE TABLE IF NOT EXISTS tokens
(
    contract_address_id bigint NOT NULL,
    "name"              character varying(50),
    symbol              character varying(10),
    decimals            smallint
);
CREATE UNIQUE INDEX IF NOT EXISTS tokens_pkey ON tokens (contract_address_id);

CREATE TABLE IF NOT EXISTS token_balances
(
    contract_address_id bigint        NOT NULL,
    address             character(42) NOT NULL,
    balance             numeric       NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS token_balances_pkey ON token_balances (contract_address_id, lower(address));
CREATE INDEX IF NOT EXISTS token_balances_api_idx1 ON token_balances (lower(address), contract_address_id desc);

CREATE TABLE IF NOT EXISTS token_balances_changes
(
    change_id           bigint        NOT NULL,
    contract_address_id bigint        NOT NULL,
    address             character(42) NOT NULL,
    balance             numeric,
    CONSTRAINT token_balances_changes_change_id_fkey FOREIGN KEY (change_id)
        REFERENCES changes (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS token_balances_changes_pkey ON token_balances_changes (change_id);