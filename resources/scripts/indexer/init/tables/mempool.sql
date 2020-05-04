CREATE TABLE IF NOT EXISTS flip_key_timestamps
(
    address     character(42) COLLATE pg_catalog."default" NOT NULL,
    epoch       bigint                                     NOT NULL,
    "timestamp" bigint                                     NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS flip_key_timestamps_address_epoch_unique_idx on
    flip_key_timestamps (LOWER(address), epoch);

CREATE TABLE IF NOT EXISTS answers_hash_tx_timestamps
(
    address     character(42) COLLATE pg_catalog."default" NOT NULL,
    epoch       bigint                                     NOT NULL,
    "timestamp" bigint                                     NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS answers_hash_tx_timestamps_address_epoch_unique_idx on answers_hash_tx_timestamps
    (LOWER(address), epoch);