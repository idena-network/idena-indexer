CREATE TABLE IF NOT EXISTS flip_private_keys_package_timestamps
(
    address     character(42) COLLATE pg_catalog."default" NOT NULL,
    epoch       smallint                                   NOT NULL,
    "timestamp" bigint                                     NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS flip_private_keys_package_timestamps_idx ON
    flip_private_keys_package_timestamps (LOWER(address), epoch);

CREATE TABLE IF NOT EXISTS flip_key_timestamps
(
    address     character(42) COLLATE pg_catalog."default" NOT NULL,
    epoch       bigint                                     NOT NULL,
    "timestamp" bigint                                     NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS flip_key_timestamps_address_epoch_unique_idx ON
    flip_key_timestamps (LOWER(address), epoch);

CREATE TABLE IF NOT EXISTS answers_hash_tx_timestamps
(
    address     character(42) COLLATE pg_catalog."default" NOT NULL,
    epoch       bigint                                     NOT NULL,
    "timestamp" bigint                                     NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS answers_hash_tx_timestamps_address_epoch_unique_idx ON answers_hash_tx_timestamps
    (LOWER(address), epoch);

CREATE TABLE IF NOT EXISTS short_answers_tx_timestamps
(
    address     character(42) COLLATE pg_catalog."default" NOT NULL,
    epoch       smallint                                   NOT NULL,
    "timestamp" bigint                                     NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS short_answers_tx_timestamps_idx ON
    short_answers_tx_timestamps (LOWER(address), epoch);