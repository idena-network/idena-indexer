-- Table: words_dictionary

-- DROP TABLE words_dictionary;

CREATE TABLE IF NOT EXISTS words_dictionary
(
    id          bigint                                              NOT NULL,
    name        character varying(20) COLLATE pg_catalog."default"  NOT NULL,
    description character varying(100) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT words_dictionary_pkey PRIMARY KEY (id)
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE words_dictionary
    OWNER to postgres;

-- Table: epochs

-- DROP TABLE epochs;

CREATE TABLE IF NOT EXISTS epochs
(
    epoch           bigint NOT NULL,
    validation_time bigint NOT NULL,
    CONSTRAINT epochs_pkey PRIMARY KEY (epoch)
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE epochs
    OWNER to postgres;

-- Table: blocks

-- DROP TABLE blocks;

CREATE TABLE IF NOT EXISTS blocks
(
    height                 bigint                                     NOT NULL,
    hash                   character(66) COLLATE pg_catalog."default" NOT NULL,
    epoch                  bigint                                     NOT NULL,
    "timestamp"            bigint                                     NOT NULL,
    is_empty               boolean                                    NOT NULL,
    validators_count       integer                                    NOT NULL,
    size                   integer                                    NOT NULL,
    vrf_proposer_threshold double precision                           NOT NULL,
    CONSTRAINT blocks_pkey PRIMARY KEY (height),
    CONSTRAINT blocks_epoch_fkey FOREIGN KEY (epoch)
        REFERENCES epochs (epoch) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE blocks
    OWNER to postgres;

CREATE UNIQUE INDEX IF NOT EXISTS blocks_hash_unique_idx on blocks (LOWER(hash));
CREATE INDEX IF NOT EXISTS blocks_epoch_idx on blocks (epoch);
CREATE INDEX IF NOT EXISTS blocks_timestamp_idx on blocks ("timestamp" desc);

-- Table: epoch_summaries

-- DROP TABLE epoch_summaries;

CREATE TABLE IF NOT EXISTS epoch_summaries
(
    epoch             bigint          NOT NULL,
    validated_count   integer         NOT NULL,
    block_count       bigint          NOT NULL,
    empty_block_count bigint          NOT NULL,
    tx_count          bigint          NOT NULL,
    invite_count      bigint          NOT NULL,
    flip_count        integer         NOT NULL,
    burnt             numeric(30, 18) NOT NULL,
    minted            numeric(30, 18) NOT NULL,
    total_balance     numeric(30, 18) NOT NULL,
    total_stake       numeric(30, 18) NOT NULL,
    block_height      bigint          NOT NULL,
    CONSTRAINT epoch_summaries_pkey PRIMARY KEY (epoch),
    CONSTRAINT epoch_summaries_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT epoch_summaries_epoch_fkey FOREIGN KEY (epoch)
        REFERENCES epochs (epoch) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE epoch_summaries
    OWNER to postgres;

-- Table: failed_validations

-- DROP TABLE failed_validations;

CREATE TABLE IF NOT EXISTS failed_validations
(
    block_height bigint NOT NULL,
    CONSTRAINT failed_validations_block_height_key UNIQUE (block_height),
    CONSTRAINT failed_validations_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE failed_validations
    OWNER to postgres;

-- SEQUENCE: addresses_id_seq

-- DROP SEQUENCE addresses_id_seq;

CREATE SEQUENCE IF NOT EXISTS addresses_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE addresses_id_seq
    OWNER TO postgres;

-- Table: addresses

-- DROP TABLE addresses;

CREATE TABLE IF NOT EXISTS addresses
(
    id           bigint                                     NOT NULL DEFAULT nextval('addresses_id_seq'::regclass),
    address      character(42) COLLATE pg_catalog."default" NOT NULL,
    block_height bigint                                     NOT NULL,
    CONSTRAINT addresses_pkey PRIMARY KEY (id),
    CONSTRAINT addresses_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE addresses
    OWNER to postgres;

CREATE UNIQUE INDEX IF NOT EXISTS addresses_address_unique_idx on addresses (LOWER(address));

-- Table: block_proposers

-- DROP TABLE block_proposers;

CREATE TABLE IF NOT EXISTS block_proposers
(
    address_id   bigint NOT NULL,
    block_height bigint NOT NULL,
    CONSTRAINT block_proposers_pkey PRIMARY KEY (block_height),
    CONSTRAINT block_proposers_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE block_proposers
    OWNER to postgres;

-- Table: mining_rewards

-- DROP TABLE mining_rewards;

CREATE TABLE IF NOT EXISTS mining_rewards
(
    address_id   bigint                                             NOT NULL,
    block_height bigint                                             NOT NULL,
    balance      numeric(30, 18)                                    NOT NULL,
    stake        numeric(30, 18)                                    NOT NULL,
    type         character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT mining_rewards_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT mining_rewards_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
) WITH (
      OIDS = FALSE
    )
  TABLESPACE pg_default;

ALTER TABLE mining_rewards
    OWNER to postgres;

CREATE INDEX IF NOT EXISTS mining_rewards_block_height_desc_idx on mining_rewards (block_height desc);

-- SEQUENCE: transactions_id_seq

-- DROP SEQUENCE transactions_id_seq;

CREATE SEQUENCE IF NOT EXISTS transactions_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE transactions_id_seq
    OWNER TO postgres;

-- Table: transactions

-- DROP TABLE transactions;

CREATE TABLE IF NOT EXISTS transactions
(
    id           bigint                                             NOT NULL DEFAULT nextval('transactions_id_seq'::regclass),
    hash         character(66) COLLATE pg_catalog."default"         NOT NULL,
    block_height bigint                                             NOT NULL,
    type         character varying(20) COLLATE pg_catalog."default" NOT NULL,
    "from"       bigint                                             NOT NULL,
    "to"         bigint,
    amount       numeric(30, 18)                                    NOT NULL,
    tips         numeric(30, 18)                                    NOT NULL,
    max_fee      numeric(30, 18)                                    NOT NULL,
    fee          numeric(30, 18)                                    NOT NULL,
    size         integer                                            NOT NULL,
    CONSTRAINT transactions_pkey PRIMARY KEY (id),
    CONSTRAINT transactions_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT transactions_from_fkey FOREIGN KEY ("from")
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT transactions_to_fkey FOREIGN KEY ("to")
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE transactions
    OWNER to postgres;

CREATE UNIQUE INDEX IF NOT EXISTS transactions_hash_unique_idx on transactions (LOWER(hash));

-- SEQUENCE: address_states_id_seq

-- DROP SEQUENCE address_states_id_seq;

CREATE SEQUENCE IF NOT EXISTS address_states_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE address_states_id_seq
    OWNER TO postgres;

-- Table: address_states

-- DROP TABLE address_states;

CREATE TABLE IF NOT EXISTS address_states
(
    id           bigint                                             NOT NULL DEFAULT nextval('address_states_id_seq'::regclass),
    address_id   bigint                                             NOT NULL,
    state        character varying(20) COLLATE pg_catalog."default" NOT NULL,
    is_actual    boolean                                            NOT NULL,
    block_height bigint                                             NOT NULL,
    tx_id        bigint,
    prev_id      bigint,
    CONSTRAINT address_states_pkey PRIMARY KEY (id),
    CONSTRAINT address_states_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT address_states_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT address_states_prev_id_fkey FOREIGN KEY (prev_id)
        REFERENCES address_states (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT address_states_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE address_states
    OWNER to postgres;

-- SEQUENCE: epoch_identities_id_seq

-- DROP SEQUENCE epoch_identities_id_seq;

CREATE SEQUENCE IF NOT EXISTS epoch_identities_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE epoch_identities_id_seq
    OWNER TO postgres;

-- Table: epoch_identities

-- DROP TABLE epoch_identities;

CREATE TABLE IF NOT EXISTS epoch_identities
(
    id                bigint   NOT NULL DEFAULT nextval('epoch_identities_id_seq'::regclass),
    epoch             bigint   NOT NULL,
    address_state_id  bigint   NOT NULL,
    short_point       real     NOT NULL,
    short_flips       integer  NOT NULL,
    total_short_point real     NOT NULL,
    total_short_flips integer  NOT NULL,
    long_point        real     NOT NULL,
    long_flips        integer  NOT NULL,
    approved          boolean  NOT NULL,
    missed            boolean  NOT NULL,
    required_flips    smallint NOT NULL,
    made_flips        smallint NOT NULL,
    CONSTRAINT epoch_identities_pkey PRIMARY KEY (id),
    CONSTRAINT epoch_identities_epoch_identity_id_key UNIQUE (epoch, address_state_id),
    CONSTRAINT epoch_identities_address_state_id_fkey FOREIGN KEY (address_state_id)
        REFERENCES address_states (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT epoch_identities_epoch_id_fkey FOREIGN KEY (epoch)
        REFERENCES epochs (epoch) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE epoch_identities
    OWNER to postgres;

-- SEQUENCE: flips_id_seq

-- DROP SEQUENCE flips_id_seq;

CREATE SEQUENCE IF NOT EXISTS flips_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE flips_id_seq
    OWNER TO postgres;

-- Table: flips

-- DROP TABLE flips;

CREATE TABLE IF NOT EXISTS flips
(
    id                  bigint                                              NOT NULL DEFAULT nextval('flips_id_seq'::regclass),
    tx_id               bigint                                              NOT NULL,
    cid                 character varying(100) COLLATE pg_catalog."default" NOT NULL,
    size                integer                                             NOT NULL,
    pair                smallint                                            NOT NULL,
    status_block_height bigint,
    answer              character varying(20) COLLATE pg_catalog."default",
    wrong_words         boolean,
    status              character varying(20) COLLATE pg_catalog."default",
    CONSTRAINT flips_pkey PRIMARY KEY (id),
    CONSTRAINT flips_status_block_height_fkey FOREIGN KEY (status_block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT flips_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE flips
    OWNER to postgres;

CREATE UNIQUE INDEX IF NOT EXISTS flips_cid_unique_idx on flips (LOWER(cid));

-- SEQUENCE: flip_keys_id_seq

-- DROP SEQUENCE flip_keys_id_seq;

CREATE SEQUENCE IF NOT EXISTS flip_keys_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE flip_keys_id_seq
    OWNER TO postgres;

-- Table: flip_keys

-- DROP TABLE flip_keys;

CREATE TABLE IF NOT EXISTS flip_keys
(
    id    bigint                                              NOT NULL DEFAULT nextval('flip_keys_id_seq'::regclass),
    tx_id bigint                                              NOT NULL,
    key   character varying(100) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT flip_keys_pkey PRIMARY KEY (id),
    CONSTRAINT flip_keys_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE flip_keys
    OWNER to postgres;

-- SEQUENCE: answers_id_seq

-- DROP SEQUENCE answers_id_seq;

CREATE SEQUENCE IF NOT EXISTS answers_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE answers_id_seq
    OWNER TO postgres;

-- Table: answers

-- DROP TABLE answers;

CREATE TABLE IF NOT EXISTS answers
(
    id                bigint                                             NOT NULL DEFAULT nextval('answers_id_seq'::regclass),
    flip_id           bigint                                             NOT NULL,
    epoch_identity_id bigint                                             NOT NULL,
    is_short          boolean                                            NOT NULL,
    answer            character varying(20) COLLATE pg_catalog."default" NOT NULL,
    wrong_words       boolean                                            NOT NULL,
    point             real                                               NOT NULL,
    CONSTRAINT answers_pkey PRIMARY KEY (id),
    CONSTRAINT answers_epoch_identity_id_fkey FOREIGN KEY (epoch_identity_id)
        REFERENCES epoch_identities (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT answers_flip_id_fkey FOREIGN KEY (flip_id)
        REFERENCES flips (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE answers
    OWNER to postgres;

-- SEQUENCE: flips_to_solve_id_seq

-- DROP SEQUENCE flips_to_solve_id_seq;

CREATE SEQUENCE IF NOT EXISTS flips_to_solve_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE flips_to_solve_id_seq
    OWNER TO postgres;

-- Table: flips_to_solve

-- DROP TABLE flips_to_solve;

CREATE TABLE IF NOT EXISTS flips_to_solve
(
    id                bigint  NOT NULL DEFAULT nextval('flips_to_solve_id_seq'::regclass),
    epoch_identity_id bigint  NOT NULL,
    flip_id           bigint  NOT NULL,
    is_short          boolean NOT NULL,
    CONSTRAINT flips_to_solve_pkey PRIMARY KEY (id),
    CONSTRAINT flips_to_solve_epoch_identity_id_fkey FOREIGN KEY (epoch_identity_id)
        REFERENCES epoch_identities (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT flips_to_solve_flip_id_fkey FOREIGN KEY (flip_id)
        REFERENCES flips (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE flips_to_solve
    OWNER to postgres;

-- Table: balances

-- DROP TABLE balances;

CREATE TABLE IF NOT EXISTS balances
(
    address_id bigint NOT NULL,
    balance    numeric(30, 18),
    stake      numeric(30, 18),
    CONSTRAINT balances_address_id_key UNIQUE (address_id),
    CONSTRAINT balances_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE balances
    OWNER to postgres;

-- Table: birthdays

-- DROP TABLE birthdays;

CREATE TABLE IF NOT EXISTS birthdays
(
    address_id  bigint  NOT NULL,
    birth_epoch integer NOT NULL,
    CONSTRAINT birthdays_address_id_key UNIQUE (address_id),
    CONSTRAINT birthdays_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE birthdays
    OWNER to postgres;

-- Table: coins

-- DROP TABLE coins;

CREATE TABLE IF NOT EXISTS coins
(
    block_height  bigint NOT NULL,
    burnt         numeric(30, 18),
    minted        numeric(30, 18),
    total_balance numeric(30, 18),
    total_stake   numeric(30, 18),
    CONSTRAINT coins_pkey PRIMARY KEY (block_height),
    CONSTRAINT coins_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE coins
    OWNER to postgres;

-- SEQUENCE: block_flags_id_seq

-- DROP SEQUENCE block_flags_id_seq;

CREATE SEQUENCE IF NOT EXISTS block_flags_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE block_flags_id_seq
    OWNER TO postgres;

-- Table: block_flags

-- DROP TABLE block_flags;

CREATE TABLE IF NOT EXISTS block_flags
(
    id           bigint                                             NOT NULL DEFAULT nextval('block_flags_id_seq'::regclass),
    block_height bigint                                             NOT NULL,
    flag         character varying(50) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT block_flags_pkey PRIMARY KEY (id),
    CONSTRAINT block_flags_block_height_flag_key UNIQUE (block_height, flag),
    CONSTRAINT block_flags_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE block_flags
    OWNER to postgres;

-- Table: temporary_identities

-- DROP TABLE temporary_identities;

CREATE TABLE IF NOT EXISTS temporary_identities
(
    address_id   bigint NOT NULL,
    block_height bigint NOT NULL,
    CONSTRAINT temporary_identities_pkey PRIMARY KEY (address_id),
    CONSTRAINT temporary_identities_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT temporary_identities_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE temporary_identities
    OWNER to postgres;

-- SEQUENCE: flips_data_id_seq

-- DROP SEQUENCE flips_data_id_seq;

CREATE SEQUENCE IF NOT EXISTS flips_data_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE flips_data_id_seq
    OWNER TO postgres;

-- Table: flips_data

-- DROP TABLE flips_data;

CREATE TABLE IF NOT EXISTS flips_data
(
    id           bigint NOT NULL DEFAULT nextval('flips_data_id_seq'::regclass),
    flip_id      bigint NOT NULL,
    block_height bigint,
    tx_id        bigint,
    CONSTRAINT flips_data_pkey PRIMARY KEY (id),
    CONSTRAINT flips_data_flip_id_key UNIQUE (flip_id),
    CONSTRAINT flips_data_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT flips_data_flip_id_fkey FOREIGN KEY (flip_id)
        REFERENCES flips (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT flips_data_tx_id_fkey1 FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE flips_data
    OWNER to postgres;

-- Table: flip_pics

-- DROP TABLE flip_pics;

CREATE TABLE IF NOT EXISTS flip_pics
(
    flip_data_id bigint   NOT NULL,
    index        smallint NOT NULL,
    data         bytea    NOT NULL,
    CONSTRAINT flip_pics_flip_data_id_fkey FOREIGN KEY (flip_data_id)
        REFERENCES flips_data (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE flip_pics
    OWNER to postgres;

-- Table: flip_icons

-- DROP TABLE flip_icons;

CREATE TABLE IF NOT EXISTS flip_icons
(
    flip_data_id bigint NOT NULL,
    data         bytea  NOT NULL,
    CONSTRAINT flip_icons_flip_data_id_fkey FOREIGN KEY (flip_data_id)
        REFERENCES flips_data (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE flip_icons
    OWNER to postgres;

-- Table: flip_pic_orders

-- DROP TABLE flip_pic_orders;

CREATE TABLE IF NOT EXISTS flip_pic_orders
(
    flip_data_id   bigint   NOT NULL,
    answer_index   smallint NOT NULL,
    pos_index      smallint NOT NULL,
    flip_pic_index smallint NOT NULL,
    CONSTRAINT flip_pic_orders_flip_data_id_fkey FOREIGN KEY (flip_data_id)
        REFERENCES flips_data (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE flip_pic_orders
    OWNER to postgres;

-- SEQUENCE: penalties_id_seq

-- DROP SEQUENCE penalties_id_seq;

CREATE SEQUENCE IF NOT EXISTS penalties_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE penalties_id_seq
    OWNER TO postgres;

-- Table: penalties

-- DROP TABLE penalties;

CREATE TABLE IF NOT EXISTS penalties
(
    id           bigint          NOT NULL DEFAULT nextval('penalties_id_seq'::regclass),
    address_id   bigint          NOT NULL,
    penalty      numeric(30, 18) NOT NULL,
    block_height bigint          NOT NULL,
    CONSTRAINT penalties_pkey PRIMARY KEY (id),
    CONSTRAINT penalties_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT penalties_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE penalties
    OWNER to postgres;

-- Table: paid_penalties

-- DROP TABLE penalties;

CREATE TABLE IF NOT EXISTS paid_penalties
(
    penalty_id   bigint          NOT NULL,
    penalty      numeric(30, 18) NOT NULL,
    block_height bigint          NOT NULL,
    CONSTRAINT paid_penalties_pkey PRIMARY KEY (penalty_id),
    CONSTRAINT paid_penalties_penalty_id_fkey FOREIGN KEY (penalty_id)
        REFERENCES penalties (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT paid_penalties_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE penalties
    OWNER to postgres;

-- Table: total_rewards

-- DROP TABLE total_rewards;

CREATE TABLE IF NOT EXISTS total_rewards
(
    block_height bigint          NOT NULL,
    total        numeric(30, 18) NOT NULL,
    validation   numeric(30, 18) NOT NULL,
    flips        numeric(30, 18) NOT NULL,
    invitations  numeric(30, 18) NOT NULL,
    foundation   numeric(30, 18) NOT NULL,
    zero_wallet  numeric(30, 18) NOT NULL,
    CONSTRAINT total_validation_rewards_pkey UNIQUE (block_height),
    CONSTRAINT total_validation_rewards_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
) WITH (
      OIDS = FALSE
    )
  TABLESPACE pg_default;

ALTER TABLE total_rewards
    OWNER to postgres;

-- Table: validation_rewards

-- DROP TABLE validation_rewards;

CREATE TABLE IF NOT EXISTS validation_rewards
(
    epoch_identity_id bigint                                             NOT NULL,
    balance           numeric(30, 18)                                    NOT NULL,
    stake             numeric(30, 18)                                    NOT NULL,
    type              character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT validation_rewards_epoch_identity_id_type_key UNIQUE (epoch_identity_id, type),
    CONSTRAINT validation_rewards_epoch_identity_id_fkey FOREIGN KEY (epoch_identity_id)
        REFERENCES epoch_identities (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
) WITH (
      OIDS = FALSE
    )
  TABLESPACE pg_default;

ALTER TABLE validation_rewards
    OWNER to postgres;

-- Table: reward_ages

-- DROP TABLE reward_ages;

CREATE TABLE IF NOT EXISTS reward_ages
(
    epoch_identity_id bigint  NOT NULL,
    age               integer NOT NULL,
    CONSTRAINT reward_ages_epoch_identity_id_type_key UNIQUE (epoch_identity_id),
    CONSTRAINT reward_ages_epoch_identity_id_fkey FOREIGN KEY (epoch_identity_id)
        REFERENCES epoch_identities (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
) WITH (
      OIDS = FALSE
    )
  TABLESPACE pg_default;

ALTER TABLE reward_ages
    OWNER to postgres;

-- Table: fund_rewards

-- DROP TABLE fund_rewards;

CREATE TABLE IF NOT EXISTS fund_rewards
(
    address_id   bigint                                             NOT NULL,
    block_height bigint                                             NOT NULL,
    balance      numeric(30, 18)                                    NOT NULL,
    type         character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT fund_rewards_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT fund_rewards_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
) WITH (
      OIDS = FALSE
    )
  TABLESPACE pg_default;

ALTER TABLE fund_rewards
    OWNER to postgres;

-- Table: bad_authors

-- DROP TABLE bad_authors;

CREATE TABLE IF NOT EXISTS bad_authors
(
    epoch_identity_id bigint NOT NULL,
    CONSTRAINT bad_authors_pkey UNIQUE (epoch_identity_id),
    CONSTRAINT bad_authors_epoch_identity_id_fkey FOREIGN KEY (epoch_identity_id)
        REFERENCES epoch_identities (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE bad_authors
    OWNER to postgres;

-- Table: good_authors

-- DROP TABLE good_authors;

CREATE TABLE IF NOT EXISTS good_authors
(
    epoch_identity_id  bigint   NOT NULL,
    strong_flips       smallint NOT NULL,
    weak_flips         smallint NOT NULL,
    successful_invites smallint NOT NULL,
    CONSTRAINT good_authors_pkey UNIQUE (epoch_identity_id),
    CONSTRAINT good_authors_epoch_identity_id_fkey FOREIGN KEY (epoch_identity_id)
        REFERENCES epoch_identities (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE good_authors
    OWNER to postgres;

-- Table: flip_words

-- DROP TABLE flip_words;

CREATE TABLE IF NOT EXISTS flip_words
(
    flip_id bigint   NOT NULL,
    word_1  smallint NOT NULL,
    word_2  smallint NOT NULL,
    tx_id   bigint   NOT NULL,
    CONSTRAINT flip_words_flip_id_key UNIQUE (flip_id),
    CONSTRAINT flip_words_flip_id_fkey FOREIGN KEY (flip_id)
        REFERENCES flips (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT flip_words_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE flip_words
    OWNER to postgres;

-- Table: burnt_coins

-- DROP TABLE burnt_coins;

CREATE TABLE IF NOT EXISTS burnt_coins
(
    address_id   bigint          NOT NULL,
    block_height bigint          NOT NULL,
    amount       numeric(30, 18) NOT NULL,
    reason       smallint        NOT NULL,
    tx_id        bigint,
    CONSTRAINT burnt_coins_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT burnt_coins_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT burnt_coins_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
) WITH (
      OIDS = FALSE
    )
  TABLESPACE pg_default;

ALTER TABLE burnt_coins
    OWNER to postgres;

CREATE INDEX IF NOT EXISTS burnt_coins_block_height_desc_idx on burnt_coins (block_height desc);

-- Table: flip_key_timestamps

-- DROP TABLE flip_key_timestamps;

CREATE TABLE IF NOT EXISTS flip_key_timestamps
(
    address     character(42) COLLATE pg_catalog."default" NOT NULL,
    epoch       bigint                                     NOT NULL,
    "timestamp" bigint                                     NOT NULL
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE flip_key_timestamps
    OWNER to postgres;

CREATE UNIQUE INDEX IF NOT EXISTS flip_key_timestamps_address_epoch_unique_idx on flip_key_timestamps
    (LOWER(address), epoch);

-- Table: answers_hash_tx_timestamps

-- DROP TABLE answers_hash_tx_timestamps;

CREATE TABLE IF NOT EXISTS answers_hash_tx_timestamps
(
    address     character(42) COLLATE pg_catalog."default" NOT NULL,
    epoch       bigint                                     NOT NULL,
    "timestamp" bigint                                     NOT NULL
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE answers_hash_tx_timestamps
    OWNER to postgres;

CREATE UNIQUE INDEX IF NOT EXISTS answers_hash_tx_timestamps_address_epoch_unique_idx on answers_hash_tx_timestamps
    (LOWER(address), epoch);

-- View: epoch_identity_states

-- DROP VIEW epoch_identity_states;

CREATE OR REPLACE VIEW epoch_identity_states AS
    SELECT s.id AS address_state_id,
           s.address_id,
           s.prev_id,
           s.state,
           s.block_height,
           ei.epoch
    FROM address_states s
             JOIN blocks b ON b.height = s.block_height
             JOIN epoch_identities ei ON s.id = ei.address_state_id
    UNION
    SELECT s.id AS address_state_id,
           s.address_id,
           s.prev_id,
           s.state,
           s.block_height,
           max_epoch.epoch
    FROM address_states s
             JOIN blocks b ON b.height = s.block_height
             LEFT JOIN temporary_identities ti ON ti.address_id = s.address_id,
         (SELECT max(epochs.epoch) AS epoch FROM epochs) max_epoch
             LEFT JOIN epoch_identities ei ON ei.epoch = max_epoch.epoch
    WHERE s.is_actual
      AND ti.address_id IS NULL
      AND ei.address_state_id IS NULL
      AND NOT (b.epoch <> max_epoch.epoch AND (s.state::text = ANY
                                               (ARRAY ['Undefined'::character varying, 'Killed'::character varying]::text[])));

ALTER TABLE epoch_identity_states
    OWNER TO postgres;

-- View: used_invites

-- DROP VIEW used_invites;

CREATE OR REPLACE VIEW used_invites AS
SELECT DISTINCT ON (b.epoch, it."to") it.id AS invite_tx_id,
                                      t.id  AS activation_tx_id
FROM transactions t
         JOIN blocks b ON b.height = t.block_height
         JOIN blocks ib ON ib.epoch = b.epoch AND ib.height < b.height
         JOIN transactions it ON it.block_height = ib.height AND it.type::text = 'InviteTx'::text AND it."to" = t."from"
WHERE t.type::text = 'ActivationTx'::text
ORDER BY b.epoch, it."to", ib.height DESC;

ALTER TABLE used_invites
    OWNER TO postgres;

-- View: epochs_detail

-- DROP VIEW epochs_detail;

CREATE OR REPLACE VIEW epochs_detail AS
SELECT e.epoch,
       COALESCE(es.validated_count::bigint, (SELECT count(*) AS count
                                             FROM epoch_identities ei
                                                      JOIN address_states s ON s.id = ei.address_state_id
                                             WHERE ei.epoch = e.epoch
                                               AND (s.state::text = ANY
                                                    (ARRAY ['Verified'::character varying, 'Newbie'::character varying]::text[])))) AS validated_count,
       COALESCE(es.block_count, (SELECT count(*) AS count
                                 FROM blocks b
                                 WHERE b.epoch = e.epoch))                                                                          AS block_count,
       COALESCE(es.empty_block_count, (SELECT count(*) AS count
                                       FROM blocks b
                                       WHERE b.epoch = e.epoch
                                         and b.is_empty))                                                                           AS empty_block_count,
       COALESCE(es.tx_count, (SELECT count(*) AS count
                              FROM transactions t,
                                   blocks b
                              WHERE t.block_height = b.height
                                AND b.epoch = e.epoch))                                                                             AS tx_count,
       COALESCE(es.invite_count, (SELECT count(*) AS count
                                  FROM transactions t,
                                       blocks b
                                  WHERE t.block_height = b.height
                                    AND b.epoch = e.epoch
                                    AND t.type::text = 'InviteTx'::text))                                                           AS invite_count,
       COALESCE(es.flip_count::bigint, (SELECT count(*) AS count
                                        FROM flips f,
                                             transactions t,
                                             blocks b
                                        WHERE f.tx_id = t.id
                                          AND t.block_height = b.height
                                          AND b.epoch = e.epoch))                                                                   AS flip_count,
       COALESCE(es.burnt, (SELECT COALESCE(sum(c.burnt), 0::numeric) AS "coalesce"
                           FROM coins c
                                    JOIN blocks b ON b.height = c.block_height
                           WHERE b.epoch = e.epoch))                                                                                AS burnt,
       COALESCE(es.minted, (SELECT COALESCE(sum(c.minted), 0::numeric) AS "coalesce"
                            FROM coins c
                                     JOIN blocks b ON b.height = c.block_height
                            WHERE b.epoch = e.epoch))                                                                               AS minted,
       COALESCE(es.total_balance, (SELECT c.total_balance
                                   FROM coins c
                                            JOIN blocks b ON b.height = c.block_height
                                   WHERE b.epoch = e.epoch
                                   ORDER BY c.block_height DESC
                                   LIMIT 1))                                                                                        AS total_balance,
       COALESCE(es.total_stake, (SELECT c.total_stake
                                 FROM coins c
                                          JOIN blocks b ON b.height = c.block_height
                                 WHERE b.epoch = e.epoch
                                 ORDER BY c.block_height DESC
                                 LIMIT 1))                                                                                          AS total_stake,
       e.validation_time                                                                                                               validation_time,
       coalesce(trew.total, 0)                                                                                                         total_reward,
       coalesce(trew.validation, 0)                                                                                                    validation_reward,
       coalesce(trew.flips, 0)                                                                                                         flips_reward,
       coalesce(trew.invitations, 0)                                                                                                   invitations_reward,
       coalesce(trew.foundation, 0)                                                                                                    foundation_payout,
       coalesce(trew.zero_wallet, 0)                                                                                                   zero_wallet_payout
FROM epochs e
         LEFT JOIN epoch_summaries es ON es.epoch = e.epoch
         left join (select b.epoch,
                           trew.total,
                           trew.validation,
                           trew.flips,
                           trew.invitations,
                           trew.foundation,
                           trew.zero_wallet
                    from total_rewards trew
                             join blocks b on b.height = trew.block_height) trew on trew.epoch = e.epoch
ORDER BY e.epoch DESC;

ALTER TABLE epochs_detail
    OWNER TO postgres;

-- Types
DO
$$
    BEGIN
        -- Type: tp_mining_reward
        CREATE TYPE tp_mining_reward AS
            (
            address character(42),
            balance numeric(30, 18),
            stake numeric(30, 18),
            type character varying(20)
            );

        ALTER TYPE tp_mining_reward
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

-- Type: tp_burnt_coins
DO
$$
    BEGIN
        CREATE TYPE tp_burnt_coins AS
            (
            address character(42),
            amount numeric(30, 18),
            reason smallint,
            tx_id bigint
            );

        ALTER TYPE tp_burnt_coins
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

-- PROCEDURE: save_mining_rewards

CREATE OR REPLACE PROCEDURE save_mining_rewards(height bigint, mr tp_mining_reward[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    mr_row tp_mining_reward;
BEGIN
    for i in 1..cardinality(mr)
        loop
            mr_row = mr[i];
            insert into mining_rewards (address_id, block_height, balance, stake, type)
            values ((select id from addresses where lower(address) = lower(mr_row.address)), height,
                    mr_row.balance, mr_row.stake, mr_row.type);
        end loop;
END
$BODY$;

-- PROCEDURE: save_burnt_coins

CREATE OR REPLACE PROCEDURE save_burnt_coins(height bigint, bc tp_burnt_coins[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    bc_row     tp_burnt_coins;
    address_id bigint;
    tx_id      bigint;
BEGIN
    for i in 1..cardinality(bc)
        loop
            bc_row = bc[i];
            IF char_length(bc_row.address) > 0 THEN
                select id into address_id from addresses where lower(address) = lower(bc_row.address);
            end if;
            if bc_row.tx_id > 0 then
                tx_id = bc_row.tx_id;
            else
                tx_id = null;
            end if;
            insert into burnt_coins (address_id, block_height, amount, reason, tx_id)
            values (address_id, height, bc_row.amount, bc_row.reason, tx_id);
        end loop;
END
$BODY$;