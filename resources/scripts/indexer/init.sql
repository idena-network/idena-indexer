-- SEQUENCE: public.epochs_id_seq

-- DROP SEQUENCE public.epochs_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.epochs_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.epochs_id_seq
    OWNER TO postgres;

-- Table: public.epochs

-- DROP TABLE public.epochs;

CREATE TABLE IF NOT EXISTS public.epochs
(
    id              bigint  NOT NULL DEFAULT nextval('epochs_id_seq'::regclass),
    epoch           integer NOT NULL,
    validation_time bigint  NOT NULL,
    CONSTRAINT epochs_pkey PRIMARY KEY (id),
    CONSTRAINT epochs_epoch_key UNIQUE (epoch)

)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.epochs
    OWNER to postgres;

-- SEQUENCE: public.blocks_id_seq

-- DROP SEQUENCE public.blocks_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.blocks_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.blocks_id_seq
    OWNER TO postgres;

-- Table: public.blocks

-- DROP TABLE public.blocks;

CREATE TABLE IF NOT EXISTS public.blocks
(
    id          bigint                                     NOT NULL DEFAULT nextval('blocks_id_seq'::regclass),
    height      integer                                    NOT NULL,
    hash        character(66) COLLATE pg_catalog."default" NOT NULL,
    epoch_id    bigint                                     NOT NULL,
    "timestamp" bigint                                     NOT NULL,
    CONSTRAINT blocks_pkey PRIMARY KEY (id),
    CONSTRAINT blocks_hash_key UNIQUE (hash),
    CONSTRAINT blocks_height_key UNIQUE (height),
    CONSTRAINT blocks_epoch_id_fkey FOREIGN KEY (epoch_id)
        REFERENCES public.epochs (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.blocks
    OWNER to postgres;

-- SEQUENCE: public.addresses_id_seq

-- DROP SEQUENCE public.addresses_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.addresses_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.addresses_id_seq
    OWNER TO postgres;

-- Table: public.addresses

-- DROP TABLE public.addresses;

CREATE TABLE IF NOT EXISTS public.addresses
(
    id       bigint                                     NOT NULL DEFAULT nextval('addresses_id_seq'::regclass),
    address  character(42) COLLATE pg_catalog."default" NOT NULL,
    block_id bigint                                     NOT NULL,
    CONSTRAINT addresses_pkey PRIMARY KEY (id),
    CONSTRAINT addresses_address_key UNIQUE (address),
    CONSTRAINT addresses_block_id_fkey FOREIGN KEY (block_id)
        REFERENCES public.blocks (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.addresses
    OWNER to postgres;

-- Table: public.proposers

-- DROP TABLE public.proposers;

CREATE TABLE IF NOT EXISTS public.proposers
(
    address_id bigint NOT NULL,
    block_id   bigint NOT NULL,
    CONSTRAINT proposers_pkey PRIMARY KEY (block_id)
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.proposers
    OWNER to postgres;

-- SEQUENCE: public.transactions_id_seq

-- DROP SEQUENCE public.transactions_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.transactions_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.transactions_id_seq
    OWNER TO postgres;

-- Table: public.transactions

-- DROP TABLE public.transactions;

CREATE TABLE IF NOT EXISTS public.transactions
(
    id       bigint                                             NOT NULL DEFAULT nextval('transactions_id_seq'::regclass),
    hash     character(66) COLLATE pg_catalog."default"         NOT NULL,
    block_id bigint                                             NOT NULL,
    type     character varying(20) COLLATE pg_catalog."default" NOT NULL,
    "from"   bigint                                             NOT NULL,
    "to"     bigint,
    amount   character varying(20) COLLATE pg_catalog."default" NOT NULL,
    fee      character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT transactions_pkey PRIMARY KEY (id),
    CONSTRAINT transactions_hash_key UNIQUE (hash),
    CONSTRAINT transactions_block_id_fkey FOREIGN KEY (block_id)
        REFERENCES public.blocks (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT transactions_from_fkey FOREIGN KEY ("from")
        REFERENCES public.addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT transactions_to_fkey FOREIGN KEY ("to")
        REFERENCES public.addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.transactions
    OWNER to postgres;

-- SEQUENCE: public.address_states_id_seq

-- DROP SEQUENCE public.address_states_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.address_states_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.address_states_id_seq
    OWNER TO postgres;

-- Table: public.address_states

-- DROP TABLE public.address_states;

CREATE TABLE IF NOT EXISTS public.address_states
(
    id         bigint                                             NOT NULL DEFAULT nextval('address_states_id_seq'::regclass),
    address_id bigint                                             NOT NULL,
    state      character varying(20) COLLATE pg_catalog."default" NOT NULL,
    is_actual  boolean                                            NOT NULL,
    block_id   bigint                                             NOT NULL,
    tx_id      bigint,
    CONSTRAINT address_states_pkey PRIMARY KEY (id),
    CONSTRAINT address_states_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES public.addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT address_states_block_id_fkey FOREIGN KEY (block_id)
        REFERENCES public.blocks (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT address_states_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES public.transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.address_states
    OWNER to postgres;

-- SEQUENCE: public.epoch_identities_id_seq

-- DROP SEQUENCE public.epoch_identities_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.epoch_identities_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.epoch_identities_id_seq
    OWNER TO postgres;

-- Table: public.epoch_identities

-- DROP TABLE public.epoch_identities;

CREATE TABLE IF NOT EXISTS public.epoch_identities
(
    id                bigint  NOT NULL DEFAULT nextval('epoch_identities_id_seq'::regclass),
    epoch_id          bigint  NOT NULL,
    address_state_id  bigint  NOT NULL,
    short_point       real    NOT NULL,
    short_flips       integer NOT NULL,
    total_short_point real    NOT NULL,
    total_short_flips integer NOT NULL,
    long_point        real    NOT NULL,
    long_flips        integer NOT NULL,
    approved          boolean NOT NULL,
    missed            boolean NOT NULL,
    CONSTRAINT epoch_identities_pkey PRIMARY KEY (id),
    CONSTRAINT epoch_identities_epoch_id_identity_id_key UNIQUE (epoch_id, address_state_id),
    CONSTRAINT epoch_identities_address_state_id_fkey FOREIGN KEY (address_state_id)
        REFERENCES public.address_states (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT epoch_identities_epoch_id_fkey FOREIGN KEY (epoch_id)
        REFERENCES public.epochs (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.epoch_identities
    OWNER to postgres;

-- SEQUENCE: public.flips_id_seq

-- DROP SEQUENCE public.flips_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.flips_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.flips_id_seq
    OWNER TO postgres;

-- Table: public.flips

-- DROP TABLE public.flips;

CREATE TABLE IF NOT EXISTS public.flips
(
    id              bigint                                              NOT NULL DEFAULT nextval('flips_id_seq'::regclass),
    tx_id           bigint                                              NOT NULL,
    cid             character varying(100) COLLATE pg_catalog."default" NOT NULL,
    status_block_id bigint,
    answer          character varying(20) COLLATE pg_catalog."default",
    status          character varying(20) COLLATE pg_catalog."default",
    data_tx_id      bigint,
    data            bytea,
    CONSTRAINT flips_pkey PRIMARY KEY (id),
    CONSTRAINT flips_cid_key UNIQUE (cid),
    CONSTRAINT flips_data_tx_id_fkey FOREIGN KEY (data_tx_id)
        REFERENCES public.transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT flips_status_block_id_fkey FOREIGN KEY (status_block_id)
        REFERENCES public.blocks (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT flips_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES public.transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.flips
    OWNER to postgres;

-- SEQUENCE: public.flip_keys_id_seq

-- DROP SEQUENCE public.flip_keys_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.flip_keys_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.flip_keys_id_seq
    OWNER TO postgres;

-- Table: public.flip_keys

-- DROP TABLE public.flip_keys;

CREATE TABLE IF NOT EXISTS public.flip_keys
(
    id    bigint                                              NOT NULL DEFAULT nextval('flip_keys_id_seq'::regclass),
    tx_id bigint                                              NOT NULL,
    key   character varying(100) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT flip_keys_pkey PRIMARY KEY (id),
    CONSTRAINT flip_keys_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES public.transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.flip_keys
    OWNER to postgres;

-- SEQUENCE: public.answers_id_seq

-- DROP SEQUENCE public.answers_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.answers_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.answers_id_seq
    OWNER TO postgres;

-- Table: public.answers

-- DROP TABLE public.answers;

CREATE TABLE IF NOT EXISTS public.answers
(
    id                bigint                                             NOT NULL DEFAULT nextval('answers_id_seq'::regclass),
    flip_id           bigint                                             NOT NULL,
    epoch_identity_id bigint                                             NOT NULL,
    is_short          boolean                                            NOT NULL,
    answer            character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT answers_pkey PRIMARY KEY (id),
    CONSTRAINT answers_epoch_identity_id_fkey FOREIGN KEY (epoch_identity_id)
        REFERENCES public.epoch_identities (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT answers_flip_id_fkey FOREIGN KEY (flip_id)
        REFERENCES public.flips (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.answers
    OWNER to postgres;

-- SEQUENCE: public.flips_to_solve_id_seq

-- DROP SEQUENCE public.flips_to_solve_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.flips_to_solve_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.flips_to_solve_id_seq
    OWNER TO postgres;

-- Table: public.flips_to_solve

-- DROP TABLE public.flips_to_solve;

CREATE TABLE IF NOT EXISTS public.flips_to_solve
(
    id                bigint  NOT NULL DEFAULT nextval('flips_to_solve_id_seq'::regclass),
    epoch_identity_id bigint  NOT NULL,
    flip_id           bigint  NOT NULL,
    is_short          boolean NOT NULL,
    CONSTRAINT flips_to_solve_pkey PRIMARY KEY (id),
    CONSTRAINT flips_to_solve_epoch_identity_id_fkey FOREIGN KEY (epoch_identity_id)
        REFERENCES public.epoch_identities (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT flips_to_solve_flip_id_fkey FOREIGN KEY (flip_id)
        REFERENCES public.flips (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.flips_to_solve
    OWNER to postgres;

-- SEQUENCE: public.balances_id_seq

-- DROP SEQUENCE public.balances_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.balances_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.balances_id_seq
    OWNER TO postgres;

-- Table: public.balances

-- DROP TABLE public.balances;

CREATE TABLE IF NOT EXISTS public.balances
(
    id         bigint                                             NOT NULL DEFAULT nextval('balances_id_seq'::regclass),
    address_id bigint                                             NOT NULL,
    balance    character varying(50) COLLATE pg_catalog."default" NOT NULL,
    stake      character varying(50) COLLATE pg_catalog."default" NOT NULL,
    block_id   bigint                                             NOT NULL,
    CONSTRAINT balances_pkey PRIMARY KEY (id),
    CONSTRAINT balances_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES public.addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT balances_block_id_fkey FOREIGN KEY (block_id)
        REFERENCES public.blocks (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.balances
    OWNER to postgres;

-- SEQUENCE: public.block_flags_id_seq

-- DROP SEQUENCE public.block_flags_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.block_flags_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.block_flags_id_seq
    OWNER TO postgres;

-- Table: public.block_flags

-- DROP TABLE public.block_flags;

CREATE TABLE IF NOT EXISTS public.block_flags
(
    id       bigint                                             NOT NULL DEFAULT nextval('block_flags_id_seq'::regclass),
    block_id bigint                                             NOT NULL,
    flag     character varying(50) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT block_flags_pkey PRIMARY KEY (id),
    CONSTRAINT block_flags_block_id_flag_key UNIQUE (block_id, flag),
    CONSTRAINT block_flags_block_id_fkey FOREIGN KEY (block_id)
        REFERENCES public.blocks (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.block_flags
    OWNER to postgres;

-- Table: public.temporary_identities

-- DROP TABLE public.temporary_identities;

CREATE TABLE IF NOT EXISTS public.temporary_identities
(
    address_id bigint NOT NULL,
    block_id   bigint NOT NULL,
    CONSTRAINT temporary_identities_pkey PRIMARY KEY (address_id),
    CONSTRAINT temporary_identities_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES public.addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT temporary_identities_block_id_fkey FOREIGN KEY (block_id)
        REFERENCES public.blocks (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.temporary_identities
    OWNER to postgres;

-- View: public.current_balances

-- DROP VIEW public.current_balances;

CREATE OR REPLACE VIEW public.current_balances AS
SELECT a.id address_id,
       a.address,
       ab.balance,
       ab.stake
FROM balances ab
         JOIN blocks b ON b.id = ab.block_id
         JOIN addresses a ON a.id = ab.address_id
WHERE ((ab.address_id, b.height) IN (SELECT bh.address_id,
                                            max(bh.height) AS max
                                     FROM (SELECT ab_1.address_id,
                                                  b_1.height
                                           FROM balances ab_1
                                                    JOIN blocks b_1 ON b_1.id = ab_1.block_id) bh
                                     GROUP BY bh.address_id));

-- View: public.epoch_identity_states

-- DROP VIEW public.epoch_identity_states;

CREATE OR REPLACE VIEW public.epoch_identity_states AS
SELECT s.id AS address_state_id,
       s.address_id,
       s.state,
       s.block_id,
       e.id AS epoch_id,
       e.epoch
FROM address_states s
         JOIN blocks b ON b.id = s.block_id
         LEFT JOIN epoch_identities ei ON s.id = ei.address_state_id
         LEFT JOIN temporary_identities ti ON ti.address_id = s.address_id,
     epochs e
WHERE ti.address_id IS NULL
  AND (e.id = b.epoch_id AND s.is_actual OR e.id = b.epoch_id AND ei.address_state_id IS NOT NULL OR
       e.epoch = ((SELECT max(epochs.epoch) AS max_epoch
                   FROM epochs)) AND s.is_actual AND NOT (b.epoch_id <> e.id AND (s.state::text = ANY
                                                                                  (ARRAY ['Undefined'::character varying, 'Killed'::character varying]::text[]))));

ALTER TABLE public.epoch_identity_states
    OWNER TO postgres;