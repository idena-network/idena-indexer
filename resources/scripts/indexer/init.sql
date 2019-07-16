-- SEQUENCE: public.epochs_id_seq

-- DROP SEQUENCE public.epochs_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.epochs_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 2147483647
    CACHE 1;

ALTER SEQUENCE public.epochs_id_seq
    OWNER TO postgres;

-- Table: public.epochs

-- DROP TABLE public.epochs;

CREATE TABLE IF NOT EXISTS public.epochs
(
    id    integer NOT NULL DEFAULT nextval('epochs_id_seq'::regclass),
    epoch integer,
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
    MAXVALUE 2147483647
    CACHE 1;

ALTER SEQUENCE public.blocks_id_seq
    OWNER TO postgres;

-- Table: public.blocks

-- DROP TABLE public.blocks;

CREATE TABLE IF NOT EXISTS public.blocks
(
    id          integer                                    NOT NULL DEFAULT nextval('blocks_id_seq'::regclass),
    height      integer                                    NOT NULL,
    hash        character(66) COLLATE pg_catalog."default" NOT NULL,
    epoch_id    integer                                    NOT NULL,
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

-- SEQUENCE: public.transactions_id_seq

-- DROP SEQUENCE public.transactions_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.transactions_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 2147483647
    CACHE 1;

ALTER SEQUENCE public.transactions_id_seq
    OWNER TO postgres;

-- Table: public.transactions

-- DROP TABLE public.transactions;

CREATE TABLE IF NOT EXISTS public.transactions
(
    id       integer                                            NOT NULL DEFAULT nextval('transactions_id_seq'::regclass),
    hash     character(66) COLLATE pg_catalog."default"         NOT NULL,
    block_id integer                                            NOT NULL,
    type     character varying(20) COLLATE pg_catalog."default" NOT NULL,
    "from"   character(42) COLLATE pg_catalog."default"         NOT NULL,
    "to"     character(42) COLLATE pg_catalog."default",
    amount   bigint                                             NOT NULL,
    fee      bigint                                             NOT NULL,
    CONSTRAINT transactions_pkey PRIMARY KEY (id),
    CONSTRAINT transactions_hash_key UNIQUE (hash)

)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.transactions
    OWNER to postgres;

-- SEQUENCE: public.identities_id_seq

-- DROP SEQUENCE public.identities_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.identities_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 2147483647
    CACHE 1;

ALTER SEQUENCE public.identities_id_seq
    OWNER TO postgres;

-- Table: public.identities

-- DROP TABLE public.identities;

CREATE TABLE IF NOT EXISTS public.identities
(
    id      integer                                    NOT NULL DEFAULT nextval('identities_id_seq'::regclass),
    address character(42) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT identities_pkey PRIMARY KEY (id),
    CONSTRAINT identities_address_key UNIQUE (address)

)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.identities
    OWNER to postgres;

-- SEQUENCE: public.epoch_identities_id_seq

-- DROP SEQUENCE public.epoch_identities_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.epoch_identities_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 2147483647
    CACHE 1;

ALTER SEQUENCE public.epoch_identities_id_seq
    OWNER TO postgres;

-- Table: public.epoch_identities

-- DROP TABLE public.epoch_identities;

CREATE TABLE IF NOT EXISTS public.epoch_identities
(
    id          integer NOT NULL DEFAULT nextval('epoch_identities_id_seq'::regclass),
    epoch_id    integer,
    identity_id integer,
    state       character varying(20) COLLATE pg_catalog."default",
    short_point real,
    short_flips integer,
    long_point  real,
    long_flips  integer,
    CONSTRAINT epoch_identities_pkey PRIMARY KEY (id),
    CONSTRAINT epoch_identities_epoch_id_identity_id_key UNIQUE (epoch_id, identity_id)

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
    MAXVALUE 2147483647
    CACHE 1;

ALTER SEQUENCE public.flips_id_seq
    OWNER TO postgres;

-- Table: public.flips

-- DROP TABLE public.flips;

CREATE TABLE IF NOT EXISTS public.flips
(
    id     integer                                             NOT NULL DEFAULT nextval('flips_id_seq'::regclass),
    tx_id  integer                                             NOT NULL,
    cid    character varying(100) COLLATE pg_catalog."default" NOT NULL,
    answer character varying(20) COLLATE pg_catalog."default",
    status character varying(20) COLLATE pg_catalog."default",
    CONSTRAINT flips_pkey PRIMARY KEY (id),
    CONSTRAINT flips_cid_key UNIQUE (cid)

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
    MAXVALUE 2147483647
    CACHE 1;

ALTER SEQUENCE public.flip_keys_id_seq
    OWNER TO postgres;

-- Table: public.flip_keys

-- DROP TABLE public.flip_keys;

CREATE TABLE IF NOT EXISTS public.flip_keys
(
    id    integer                                             NOT NULL DEFAULT nextval('flip_keys_id_seq'::regclass),
    tx_id integer                                             NOT NULL,
    key   character varying(100) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT flip_keys_pkey PRIMARY KEY (id)
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
    MAXVALUE 2147483647
    CACHE 1;

ALTER SEQUENCE public.answers_id_seq
    OWNER TO postgres;

-- Table: public.answers

-- DROP TABLE public.answers;

CREATE TABLE IF NOT EXISTS public.answers
(
    id          integer                                            NOT NULL DEFAULT nextval('answers_id_seq'::regclass),
    flip_id     integer                                            NOT NULL,
    identity_id integer                                            NOT NULL,
    is_short    boolean                                            NOT NULL,
    answer      character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT answers_pkey PRIMARY KEY (id)
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.answers
    OWNER to postgres;