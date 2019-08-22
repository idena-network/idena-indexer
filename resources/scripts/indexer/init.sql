-- Table: public.epochs

-- DROP TABLE public.epochs;

CREATE TABLE IF NOT EXISTS public.epochs
(
    epoch           bigint NOT NULL,
    validation_time bigint NOT NULL,
    CONSTRAINT epochs_pkey PRIMARY KEY (epoch)
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.epochs
    OWNER to postgres;

-- Table: public.blocks

-- DROP TABLE public.blocks;

CREATE TABLE IF NOT EXISTS public.blocks
(
    height           bigint                                     NOT NULL,
    hash             character(66) COLLATE pg_catalog."default" NOT NULL,
    epoch            bigint                                     NOT NULL,
    "timestamp"      bigint                                     NOT NULL,
    is_empty         boolean                                    NOT NULL,
    validators_count integer                                    NOT NULL,
    CONSTRAINT blocks_pkey PRIMARY KEY (height),
    CONSTRAINT blocks_epoch_fkey FOREIGN KEY (epoch)
        REFERENCES public.epochs (epoch) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.blocks
    OWNER to postgres;

CREATE UNIQUE INDEX IF NOT EXISTS blocks_hash_unique_idx on blocks (LOWER(hash));

-- Table: public.epoch_summaries

-- DROP TABLE public.epoch_summaries;

CREATE TABLE IF NOT EXISTS public.epoch_summaries
(
    epoch             bigint          NOT NULL,
    validated_count   integer         NOT NULL,
    block_count       bigint          NOT NULL,
    empty_block_count bigint          NOT NULL,
    tx_count          bigint          NOT NULL,
    invite_count      bigint          NOT NULL,
    flip_count        integer         NOT NULL,
    burnt_balance     numeric(30, 18) NOT NULL,
    minted_balance    numeric(30, 18) NOT NULL,
    total_balance     numeric(30, 18) NOT NULL,
    burnt_stake       numeric(30, 18) NOT NULL,
    minted_stake      numeric(30, 18) NOT NULL,
    total_stake       numeric(30, 18) NOT NULL,
    block_height      bigint          NOT NULL,
    CONSTRAINT epoch_summaries_pkey PRIMARY KEY (epoch),
    CONSTRAINT epoch_summaries_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES public.blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT epoch_summaries_epoch_fkey FOREIGN KEY (epoch)
        REFERENCES public.epochs (epoch) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.epoch_summaries
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
    id           bigint                                     NOT NULL DEFAULT nextval('addresses_id_seq'::regclass),
    address      character(42) COLLATE pg_catalog."default" NOT NULL,
    block_height bigint                                     NOT NULL,
    CONSTRAINT addresses_pkey PRIMARY KEY (id),
    CONSTRAINT addresses_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES public.blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.addresses
    OWNER to postgres;

CREATE UNIQUE INDEX IF NOT EXISTS addresses_address_unique_idx on addresses (LOWER(address));

-- Table: public.block_proposers

-- DROP TABLE public.block_proposers;

CREATE TABLE IF NOT EXISTS public.block_proposers
(
    address_id   bigint NOT NULL,
    block_height bigint NOT NULL,
    CONSTRAINT block_proposers_pkey PRIMARY KEY (block_height),
    CONSTRAINT block_proposers_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES public.addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.block_proposers
    OWNER to postgres;

-- Table: public.block_validators

-- DROP TABLE public.block_validators;

CREATE TABLE IF NOT EXISTS public.block_validators
(
    block_height bigint NOT NULL,
    address_id   bigint NOT NULL,
    CONSTRAINT block_validators_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES public.addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT block_validators_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES public.blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.block_validators
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
    id           bigint                                             NOT NULL DEFAULT nextval('transactions_id_seq'::regclass),
    hash         character(66) COLLATE pg_catalog."default"         NOT NULL,
    block_height bigint                                             NOT NULL,
    type         character varying(20) COLLATE pg_catalog."default" NOT NULL,
    "from"       bigint                                             NOT NULL,
    "to"         bigint,
    amount       numeric(30, 18),
    fee          numeric(30, 18),
    CONSTRAINT transactions_pkey PRIMARY KEY (id),
    CONSTRAINT transactions_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES public.blocks (height) MATCH SIMPLE
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

CREATE UNIQUE INDEX IF NOT EXISTS transactions_hash_unique_idx on transactions (LOWER(hash));

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
    id           bigint                                             NOT NULL DEFAULT nextval('address_states_id_seq'::regclass),
    address_id   bigint                                             NOT NULL,
    state        character varying(20) COLLATE pg_catalog."default" NOT NULL,
    is_actual    boolean                                            NOT NULL,
    block_height bigint                                             NOT NULL,
    tx_id        bigint,
    prev_id      bigint,
    CONSTRAINT address_states_pkey PRIMARY KEY (id),
    CONSTRAINT address_states_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES public.addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT address_states_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES public.blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT address_states_prev_id_fkey FOREIGN KEY (prev_id)
        REFERENCES public.address_states (id) MATCH SIMPLE
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
    epoch             bigint  NOT NULL,
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
    CONSTRAINT epoch_identities_epoch_identity_id_key UNIQUE (epoch, address_state_id),
    CONSTRAINT epoch_identities_address_state_id_fkey FOREIGN KEY (address_state_id)
        REFERENCES public.address_states (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT epoch_identities_epoch_id_fkey FOREIGN KEY (epoch)
        REFERENCES public.epochs (epoch) MATCH SIMPLE
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
    id                  bigint                                              NOT NULL DEFAULT nextval('flips_id_seq'::regclass),
    tx_id               bigint                                              NOT NULL,
    cid                 character varying(100) COLLATE pg_catalog."default" NOT NULL,
    size                integer                                             NOT NULL,
    status_block_height bigint,
    answer              character varying(20) COLLATE pg_catalog."default",
    status              character varying(20) COLLATE pg_catalog."default",
    CONSTRAINT flips_pkey PRIMARY KEY (id),
    CONSTRAINT flips_status_block_height_fkey FOREIGN KEY (status_block_height)
        REFERENCES public.blocks (height) MATCH SIMPLE
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

CREATE UNIQUE INDEX IF NOT EXISTS flips_cid_unique_idx on flips (LOWER(cid));

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
    id           bigint NOT NULL DEFAULT nextval('balances_id_seq'::regclass),
    address_id   bigint NOT NULL,
    balance      numeric(30, 18),
    stake        numeric(30, 18),
    tx_id        bigint,
    block_height bigint NOT NULL,
    CONSTRAINT balances_pkey PRIMARY KEY (id),
    CONSTRAINT balances_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES public.addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT balances_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES public.blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT balances_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES public.transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.balances
    OWNER to postgres;

-- Table: public.coins

-- DROP TABLE public.coins;

CREATE TABLE IF NOT EXISTS public.coins
(
    block_height   bigint NOT NULL,
    burnt_balance  numeric(30, 18),
    minted_balance numeric(30, 18),
    total_balance  numeric(30, 18),
    burnt_stake    numeric(30, 18),
    minted_stake   numeric(30, 18),
    total_stake    numeric(30, 18),
    CONSTRAINT coins_pkey PRIMARY KEY (block_height),
    CONSTRAINT coins_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES public.blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.coins
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
    id           bigint                                             NOT NULL DEFAULT nextval('block_flags_id_seq'::regclass),
    block_height bigint                                             NOT NULL,
    flag         character varying(50) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT block_flags_pkey PRIMARY KEY (id),
    CONSTRAINT block_flags_block_height_flag_key UNIQUE (block_height, flag),
    CONSTRAINT block_flags_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES public.blocks (height) MATCH SIMPLE
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
    address_id   bigint NOT NULL,
    block_height bigint NOT NULL,
    CONSTRAINT temporary_identities_pkey PRIMARY KEY (address_id),
    CONSTRAINT temporary_identities_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES public.addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT temporary_identities_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES public.blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.temporary_identities
    OWNER to postgres;

-- SEQUENCE: public.flips_data_id_seq

-- DROP SEQUENCE public.flips_data_id_seq;

CREATE SEQUENCE IF NOT EXISTS public.flips_data_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.flips_data_id_seq
    OWNER TO postgres;

-- Table: public.flips_data

-- DROP TABLE public.flips_data;

CREATE TABLE IF NOT EXISTS public.flips_data
(
    id           bigint NOT NULL DEFAULT nextval('flips_data_id_seq'::regclass),
    flip_id      bigint NOT NULL,
    block_height bigint,
    tx_id        bigint,
    CONSTRAINT flips_data_pkey PRIMARY KEY (id),
    CONSTRAINT flips_data_flip_id_key UNIQUE (flip_id),
    CONSTRAINT flips_data_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES public.blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT flips_data_flip_id_fkey FOREIGN KEY (flip_id)
        REFERENCES public.flips (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT flips_data_tx_id_fkey1 FOREIGN KEY (tx_id)
        REFERENCES public.transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.flips_data
    OWNER to postgres;

-- Table: public.flip_pics

-- DROP TABLE public.flip_pics;

CREATE TABLE IF NOT EXISTS public.flip_pics
(
    flip_data_id bigint   NOT NULL,
    index        smallint NOT NULL,
    data         bytea    NOT NULL,
    CONSTRAINT flip_pics_flip_data_id_fkey FOREIGN KEY (flip_data_id)
        REFERENCES public.flips_data (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.flip_pics
    OWNER to postgres;

-- Table: public.flip_icons

-- DROP TABLE public.flip_icons;

CREATE TABLE IF NOT EXISTS public.flip_icons
(
    flip_data_id bigint NOT NULL,
    data         bytea  NOT NULL,
    CONSTRAINT flip_icons_flip_data_id_fkey FOREIGN KEY (flip_data_id)
        REFERENCES public.flips_data (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.flip_icons
    OWNER to postgres;

-- Table: public.flip_pic_orders

-- DROP TABLE public.flip_pic_orders;

CREATE TABLE IF NOT EXISTS public.flip_pic_orders
(
    flip_data_id   bigint   NOT NULL,
    answer_index   smallint NOT NULL,
    pos_index      smallint NOT NULL,
    flip_pic_index smallint NOT NULL,
    CONSTRAINT flip_pic_orders_flip_data_id_fkey FOREIGN KEY (flip_data_id)
        REFERENCES public.flips_data (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE public.flip_pic_orders
    OWNER to postgres;

-- View: public.current_balances

-- DROP VIEW public.current_balances;

CREATE OR REPLACE VIEW public.current_balances AS
SELECT DISTINCT ON (ab.address_id) ab.address_id,
                                   a.address,
                                   ab.balance,
                                   ab.stake
FROM balances ab
         JOIN addresses a ON a.id = ab.address_id
ORDER BY ab.address_id, ab.block_height DESC;

ALTER TABLE public.current_balances
    OWNER TO postgres;

-- View: public.epoch_identity_states

-- DROP VIEW public.epoch_identity_states;

CREATE OR REPLACE VIEW public.epoch_identity_states AS
SELECT s.id AS address_state_id,
       s.address_id,
       s.prev_id,
       s.state,
       s.block_height,
       e.epoch
FROM address_states s
         JOIN blocks b ON b.height = s.block_height
         LEFT JOIN epoch_identities ei ON s.id = ei.address_state_id
         LEFT JOIN temporary_identities ti ON ti.address_id = s.address_id,
     epochs e
WHERE ti.address_id IS NULL
  AND (e.epoch = b.epoch AND s.is_actual OR e.epoch = b.epoch AND ei.address_state_id IS NOT NULL OR
       e.epoch = ((SELECT max(epochs.epoch) AS max_epoch
                   FROM epochs)) AND s.is_actual AND NOT (b.epoch <> e.epoch AND (s.state::text = ANY
                                                                                  (ARRAY ['Undefined'::character varying, 'Killed'::character varying]::text[]))));

ALTER TABLE public.epoch_identity_states
    OWNER TO postgres;

-- View: public.used_invites

-- DROP VIEW public.used_invites;

CREATE OR REPLACE VIEW public.used_invites AS
SELECT DISTINCT ON (b.epoch, it."to") it.id AS invite_tx_id,
                                      t.id  AS activation_tx_id
FROM transactions t
         JOIN blocks b ON b.height = t.block_height
         JOIN blocks ib ON ib.epoch = b.epoch AND ib.height < b.height
         JOIN transactions it ON it.block_height = ib.height AND it.type::text = 'InviteTx'::text AND it."to" = t."from"
WHERE t.type::text = 'ActivationTx'::text
ORDER BY b.epoch, it."to", ib.height DESC;

ALTER TABLE public.used_invites
    OWNER TO postgres;

-- View: public.epochs_detail

-- DROP VIEW public.epochs_detail;

CREATE OR REPLACE VIEW public.epochs_detail AS
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
       COALESCE(es.burnt_balance, (SELECT COALESCE(sum(c.burnt_balance), 0::numeric) AS "coalesce"
                                   FROM coins c
                                            JOIN blocks b ON b.height = c.block_height
                                   WHERE b.epoch = e.epoch))                                                                        AS burnt_balance,
       COALESCE(es.minted_balance, (SELECT COALESCE(sum(c.minted_balance), 0::numeric) AS "coalesce"
                                    FROM coins c
                                             JOIN blocks b ON b.height = c.block_height
                                    WHERE b.epoch = e.epoch))                                                                       AS minted_balance,
       COALESCE(es.total_balance, (SELECT c.total_balance
                                   FROM coins c
                                            JOIN blocks b ON b.height = c.block_height
                                   WHERE b.epoch = e.epoch
                                   ORDER BY c.block_height DESC
                                   LIMIT 1))                                                                                        AS total_balance,
       COALESCE(es.burnt_stake, (SELECT COALESCE(sum(c.burnt_stake), 0::numeric) AS "coalesce"
                                 FROM coins c
                                          JOIN blocks b ON b.height = c.block_height
                                 WHERE b.epoch = e.epoch))                                                                          AS burnt_stake,
       COALESCE(es.minted_stake, (SELECT COALESCE(sum(c.minted_stake), 0::numeric) AS minted_stake
                                  FROM coins c
                                           JOIN blocks b ON b.height = c.block_height
                                  WHERE b.epoch = e.epoch))                                                                         AS minted_stake,
       COALESCE(es.total_stake, (SELECT c.total_stake
                                 FROM coins c
                                          JOIN blocks b ON b.height = c.block_height
                                 WHERE b.epoch = e.epoch
                                 ORDER BY c.block_height DESC
                                 LIMIT 1))                                                                                          AS total_stake
FROM epochs e
         LEFT JOIN epoch_summaries es ON es.epoch = e.epoch
ORDER BY e.epoch DESC;

ALTER TABLE public.epochs_detail
    OWNER TO postgres;

