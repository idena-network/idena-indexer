CREATE TABLE IF NOT EXISTS performance_logs
(
    timestamp timestamptz not null,
    message   character varying(100),
    duration  real
);

CREATE TABLE IF NOT EXISTS words_dictionary
(
    id          bigint                                              NOT NULL,
    name        character varying(50) COLLATE pg_catalog."default"  NOT NULL,
    description character varying(200) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT words_dictionary_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS dic_identity_states
(
    id   smallint                                           NOT NULL,
    name character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT dic_identity_states_pkey PRIMARY KEY (id),
    CONSTRAINT dic_identity_states_name_key UNIQUE (name)
);

INSERT INTO dic_identity_states
values (0, 'Undefined')
ON CONFLICT DO NOTHING;
INSERT INTO dic_identity_states
values (1, 'Invite')
ON CONFLICT DO NOTHING;
INSERT INTO dic_identity_states
values (2, 'Candidate')
ON CONFLICT DO NOTHING;
INSERT INTO dic_identity_states
values (3, 'Verified')
ON CONFLICT DO NOTHING;
INSERT INTO dic_identity_states
values (4, 'Suspended')
ON CONFLICT DO NOTHING;
INSERT INTO dic_identity_states
values (5, 'Killed')
ON CONFLICT DO NOTHING;
INSERT INTO dic_identity_states
values (6, 'Zombie')
ON CONFLICT DO NOTHING;
INSERT INTO dic_identity_states
values (7, 'Newbie')
ON CONFLICT DO NOTHING;
INSERT INTO dic_identity_states
values (8, 'Human')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS dic_tx_types
(
    id   smallint                                           NOT NULL,
    name character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT dic_tx_types_pkey PRIMARY KEY (id),
    CONSTRAINT dic_tx_types_name_key UNIQUE (name)
);

INSERT INTO dic_tx_types
VALUES (0, 'SendTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (1, 'ActivationTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (2, 'InviteTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (3, 'KillTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (4, 'SubmitFlipTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (5, 'SubmitAnswersHashTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (6, 'SubmitShortAnswersTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (7, 'SubmitLongAnswersTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (8, 'EvidenceTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (9, 'OnlineStatusTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (10, 'KillInviteeTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (11, 'ChangeGodAddressTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (12, 'BurnTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (13, 'ChangeProfileTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (14, 'DeleteFlipTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (15, 'DeployContract')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (16, 'CallContract')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (17, 'TerminateContract')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (18, 'DelegateTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (19, 'UndelegateTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (20, 'KillDelegatorTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (21, 'StoreToIpfsTx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_tx_types
VALUES (22, 'ReplenishStakeTx')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS dic_flip_statuses
(
    id   smallint                                           NOT NULL,
    name character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT dic_flip_statuses_pkey PRIMARY KEY (id),
    CONSTRAINT dic_flip_statuses_name_key UNIQUE (name)
);

INSERT INTO dic_flip_statuses
values (0, 'NotQualified')
ON CONFLICT DO NOTHING;
INSERT INTO dic_flip_statuses
values (1, 'Qualified')
ON CONFLICT DO NOTHING;
INSERT INTO dic_flip_statuses
values (2, 'WeaklyQualified')
ON CONFLICT DO NOTHING;
INSERT INTO dic_flip_statuses
values (3, 'QualifiedByNone')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS dic_answers
(
    id   smallint                                           NOT NULL,
    name character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT dic_answers_pkey PRIMARY KEY (id),
    CONSTRAINT dic_answers_name_key UNIQUE (name)
);

INSERT INTO dic_answers
values (0, 'None')
ON CONFLICT DO NOTHING;
INSERT INTO dic_answers
values (1, 'Left')
ON CONFLICT DO NOTHING;
INSERT INTO dic_answers
values (2, 'Right')
ON CONFLICT DO NOTHING;
INSERT INTO dic_answers
values (3, 'Inappropriate')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS dic_epoch_reward_types
(
    id   smallint                                           NOT NULL,
    name character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT dic_epoch_reward_types_pkey PRIMARY KEY (id),
    CONSTRAINT dic_epoch_reward_types_name_key UNIQUE (name)
);

INSERT INTO dic_epoch_reward_types
VALUES (0, 'Validation')
ON CONFLICT DO NOTHING;
INSERT INTO dic_epoch_reward_types
VALUES (1, 'Flips')
ON CONFLICT DO NOTHING;
INSERT INTO dic_epoch_reward_types
VALUES (2, 'Invitations')
ON CONFLICT DO NOTHING;
INSERT INTO dic_epoch_reward_types
VALUES (3, 'FoundationPayouts')
ON CONFLICT DO NOTHING;
INSERT INTO dic_epoch_reward_types
VALUES (4, 'ZeroWalletFund')
ON CONFLICT DO NOTHING;
INSERT INTO dic_epoch_reward_types
VALUES (5, 'Invitations2')
ON CONFLICT DO NOTHING;
INSERT INTO dic_epoch_reward_types
VALUES (6, 'Invitations3')
ON CONFLICT DO NOTHING;
INSERT INTO dic_epoch_reward_types
VALUES (7, 'SavedInvite')
ON CONFLICT DO NOTHING;
INSERT INTO dic_epoch_reward_types
VALUES (8, 'SavedInviteWin')
ON CONFLICT DO NOTHING;
INSERT INTO dic_epoch_reward_types
VALUES (9, 'Reports')
ON CONFLICT DO NOTHING;
INSERT INTO dic_epoch_reward_types
VALUES (10, 'Staking')
ON CONFLICT DO NOTHING;
INSERT INTO dic_epoch_reward_types
VALUES (11, 'Candidate')
ON CONFLICT DO NOTHING;
INSERT INTO dic_epoch_reward_types
VALUES (12, 'ExtraFlips')
ON CONFLICT DO NOTHING;
INSERT INTO dic_epoch_reward_types
VALUES (13, 'Invitee')
ON CONFLICT DO NOTHING;
INSERT INTO dic_epoch_reward_types
VALUES (14, 'Invitee2')
ON CONFLICT DO NOTHING;
INSERT INTO dic_epoch_reward_types
VALUES (15, 'Invitee3')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS dic_bad_author_reasons
(
    id   smallint                                           NOT NULL,
    name character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT dic_bad_author_reasons_pkey PRIMARY KEY (id),
    CONSTRAINT dic_bad_author_reasons_name_key UNIQUE (name)
);

INSERT INTO dic_bad_author_reasons
values (0, 'NoQualifiedFlips')
ON CONFLICT DO NOTHING;
INSERT INTO dic_bad_author_reasons
values (1, 'QualifiedByNone')
ON CONFLICT DO NOTHING;
INSERT INTO dic_bad_author_reasons
values (2, 'WrongWords')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS dic_balance_update_reasons
(
    id   smallint                                           NOT NULL,
    name character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT dic_balance_update_reasons_pkey PRIMARY KEY (id),
    CONSTRAINT dic_balance_update_reasons_name_key UNIQUE (name)
);

INSERT INTO dic_balance_update_reasons
VALUES (0, 'Tx')
ON CONFLICT DO NOTHING;
INSERT INTO dic_balance_update_reasons
VALUES (1, 'VerifiedStake')
ON CONFLICT DO NOTHING;
INSERT INTO dic_balance_update_reasons
VALUES (2, 'ProposerReward')
ON CONFLICT DO NOTHING;
INSERT INTO dic_balance_update_reasons
VALUES (3, 'CommitteeReward')
ON CONFLICT DO NOTHING;
INSERT INTO dic_balance_update_reasons
VALUES (4, 'EpochReward')
ON CONFLICT DO NOTHING;
INSERT INTO dic_balance_update_reasons
VALUES (5, 'FailedValidation')
ON CONFLICT DO NOTHING;
INSERT INTO dic_balance_update_reasons
VALUES (6, 'Penalty')
ON CONFLICT DO NOTHING;
INSERT INTO dic_balance_update_reasons
VALUES (7, 'EpochPenaltyReset')
ON CONFLICT DO NOTHING;
INSERT INTO dic_balance_update_reasons
VALUES (8, 'Initial')
ON CONFLICT DO NOTHING;
INSERT INTO dic_balance_update_reasons
VALUES (9, 'DustClearing')
ON CONFLICT DO NOTHING;
INSERT INTO dic_balance_update_reasons
VALUES (10, 'Contract')
ON CONFLICT DO NOTHING;
INSERT INTO dic_balance_update_reasons
VALUES (11, 'EmbeddedContractTerm')
ON CONFLICT DO NOTHING;
INSERT INTO dic_balance_update_reasons
VALUES (12, 'DelegatorEpochReward')
ON CONFLICT DO NOTHING;
INSERT INTO dic_balance_update_reasons
VALUES (13, 'DelegateeEpochReward')
ON CONFLICT DO NOTHING;
INSERT INTO dic_balance_update_reasons
VALUES (14, 'IdentityClearing')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS epochs
(
    epoch                          bigint NOT NULL,
    validation_time                bigint NOT NULL,
    root                           character(66),
    discrimination_stake_threshold numeric(30, 18),
    CONSTRAINT epochs_pkey PRIMARY KEY (epoch)
);

CREATE TABLE IF NOT EXISTS blocks
(
    height                 bigint                                     NOT NULL,
    hash                   character(66) COLLATE pg_catalog."default" NOT NULL,
    epoch                  bigint                                     NOT NULL,
    "timestamp"            bigint                                     NOT NULL,
    is_empty               boolean                                    NOT NULL,
    validators_count       integer                                    NOT NULL,
    body_size              integer                                    NOT NULL,
    vrf_proposer_threshold double precision                           NOT NULL,
    full_size              integer                                    NOT NULL,
    fee_rate               numeric(30, 18)                            NOT NULL,
    upgrade                integer,
    pool_validators_count  integer,
    offline_address_id     bigint,
    used_gas               integer,
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
CREATE INDEX IF NOT EXISTS blocks_epoch_heights_idx on blocks (epoch, height desc);
CREATE INDEX IF NOT EXISTS blocks_timestamp_idx on blocks ("timestamp" desc);
CREATE INDEX IF NOT EXISTS blocks_upgrades_history_api_idx ON blocks (height) WHERE coalesce(upgrade, 0) > 0;

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
    CONSTRAINT block_proposers_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
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

-- Table: block_proposer_vrf_scores

-- DROP TABLE block_proposer_vrf_scores;

CREATE TABLE IF NOT EXISTS block_proposer_vrf_scores
(
    block_height bigint           NOT NULL,
    vrf_score    double precision NOT NULL,
    CONSTRAINT block_proposer_vrf_scores_pkey PRIMARY KEY (block_height),
    CONSTRAINT block_proposer_vrf_scores_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE block_proposer_vrf_scores
    OWNER to postgres;

-- Table: mining_rewards

-- DROP TABLE mining_rewards;

CREATE TABLE IF NOT EXISTS mining_rewards
(
    address_id   bigint          NOT NULL,
    block_height bigint          NOT NULL,
    balance      numeric(30, 18) NOT NULL,
    stake        numeric(30, 18) NOT NULL,
    proposer     boolean         NOT NULL,
    stake_weight double precision,
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
    id           bigint                                     NOT NULL DEFAULT nextval('transactions_id_seq'::regclass),
    hash         character(66) COLLATE pg_catalog."default" NOT NULL,
    block_height bigint                                     NOT NULL,
    type         smallint                                   NOT NULL,
    "from"       bigint                                     NOT NULL,
    "to"         bigint,
    amount       numeric(30, 18)                            NOT NULL,
    tips         numeric(30, 18)                            NOT NULL,
    max_fee      numeric(30, 18)                            NOT NULL,
    fee          numeric(30, 18)                            NOT NULL,
    size         integer                                    NOT NULL,
    nonce        integer                                    NOT NULL,
    used_gas     integer,
    CONSTRAINT transactions_pkey PRIMARY KEY (id),
    CONSTRAINT transactions_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT transactions_type_fkey FOREIGN KEY (type)
        REFERENCES dic_tx_types (id) MATCH SIMPLE
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
CREATE INDEX IF NOT EXISTS transactions_from_idx on transactions ("from");
CREATE INDEX IF NOT EXISTS transactions_to_idx on transactions ("to") WHERE "to" is not null;
CREATE INDEX IF NOT EXISTS transactions_block_height_idx on transactions (block_height);
CREATE INDEX IF NOT EXISTS transactions_invite_id_idx on transactions (id) WHERE type = 2;

CREATE TABLE IF NOT EXISTS transaction_raws
(
    tx_id bigint NOT NULL,
    raw   bytea  NOT NULL,
    CONSTRAINT transaction_raws_pkey PRIMARY KEY (tx_id),
    CONSTRAINT transaction_raws_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

-- Table: activation_tx_transfers

CREATE TABLE IF NOT EXISTS activation_tx_transfers
(
    tx_id            bigint          NOT NULL,
    balance_transfer numeric(30, 18) NOT NULL,
    CONSTRAINT activation_tx_transfers_pkey PRIMARY KEY (tx_id),
    CONSTRAINT activation_tx_transfers_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

-- Table: kill_tx_transfers

CREATE TABLE IF NOT EXISTS kill_tx_transfers
(
    tx_id          bigint          NOT NULL,
    stake_transfer numeric(30, 18) NOT NULL,
    CONSTRAINT kill_tx_transfers_pkey PRIMARY KEY (tx_id),
    CONSTRAINT kill_tx_transfers_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

-- Table: kill_invitee_tx_transfers

CREATE TABLE IF NOT EXISTS kill_invitee_tx_transfers
(
    tx_id          bigint          NOT NULL,
    stake_transfer numeric(30, 18) NOT NULL,
    CONSTRAINT kill_invitee_tx_transfers_pkey PRIMARY KEY (tx_id),
    CONSTRAINT kill_invitee_tx_transfers_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS short_answer_txs
(
    tx_id       bigint   NOT NULL,
    client_type smallint NOT NULL,
    CONSTRAINT short_answer_txs_pkey PRIMARY KEY (tx_id)
);

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
    id           bigint   NOT NULL DEFAULT nextval('address_states_id_seq'::regclass),
    address_id   bigint   NOT NULL,
    state        smallint NOT NULL,
    is_actual    boolean  NOT NULL,
    block_height bigint   NOT NULL,
    tx_id        bigint,
    prev_id      bigint,
    CONSTRAINT address_states_pkey PRIMARY KEY (id),
    CONSTRAINT address_states_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT address_states_state_fkey FOREIGN KEY (state)
        REFERENCES dic_identity_states (id) MATCH SIMPLE
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

CREATE INDEX IF NOT EXISTS address_states_actual_idx on address_states (address_id) WHERE is_actual;
CREATE INDEX IF NOT EXISTS address_states_actual_terminated_idx on address_states (block_height)
    WHERE is_actual and state in (0, 5);
CREATE INDEX IF NOT EXISTS address_states_address_id_idx on address_states (address_id);

CREATE TABLE IF NOT EXISTS epoch_identities
(
    address_state_id        bigint          NOT NULL,
    epoch                   bigint          NOT NULL,
    short_point             real            NOT NULL,
    short_flips             integer         NOT NULL,
    total_short_point       real            NOT NULL,
    total_short_flips       integer         NOT NULL,
    long_point              real            NOT NULL,
    long_flips              integer         NOT NULL,
    approved                boolean         NOT NULL,
    missed                  boolean         NOT NULL,
    required_flips          smallint        NOT NULL,
    available_flips         smallint        NOT NULL,
    made_flips              smallint        NOT NULL,
    next_epoch_invites      smallint        NOT NULL,
    birth_epoch             bigint          NOT NULL,
    total_validation_reward numeric(30, 18) NOT NULL,
    short_answers           integer         NOT NULL,
    long_answers            integer         NOT NULL,
    wrong_words_flips       smallint        NOT NULL,
    delegatee_address_id    bigint,
    shard_id                integer,
    new_shard_id            integer,
    address_id              bigint          NOT NULL,
    wrong_grade_reason      smallint,
    CONSTRAINT epoch_identities_pkey PRIMARY KEY (address_state_id),
    CONSTRAINT epoch_identities_address_state_id_fkey FOREIGN KEY (address_state_id)
        REFERENCES address_states (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT epoch_identities_epoch_id_fkey FOREIGN KEY (epoch)
        REFERENCES epochs (epoch) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE UNIQUE INDEX IF NOT EXISTS epoch_identities_api_idx on epoch_identities (address_id, epoch desc);

CREATE TABLE IF NOT EXISTS epoch_identity_interim_states
(
    address_state_id bigint NOT NULL,
    block_height     bigint NOT NULL,
    CONSTRAINT epoch_identity_interim_states_pkey PRIMARY KEY (address_state_id, block_height),
    CONSTRAINT epoch_identity_interim_states_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS flips
(
    tx_id               bigint                                              NOT NULL,
    cid                 character varying(100) COLLATE pg_catalog."default" NOT NULL,
    size                integer                                             NOT NULL,
    pair                smallint                                            NOT NULL,
    status_block_height bigint,
    answer              smallint,
    status              smallint,
    delete_tx_id        bigint,
    grade               smallint,
    grade_score         real,
    CONSTRAINT flips_pkey PRIMARY KEY (tx_id),
    CONSTRAINT flips_status_block_height_fkey FOREIGN KEY (status_block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT flips_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT flips_status_fkey FOREIGN KEY (status)
        REFERENCES dic_flip_statuses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT flips_answer_fkey FOREIGN KEY (answer)
        REFERENCES dic_answers (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT flips_delete_tx_id_fkey FOREIGN KEY (delete_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE UNIQUE INDEX IF NOT EXISTS flips_cid_unique_idx on flips (LOWER(cid));
CREATE INDEX IF NOT EXISTS flips_zero_size_idx on flips (tx_id) WHERE size = 0 and delete_tx_id is NULL;
CREATE INDEX IF NOT EXISTS flips_actual_idx on flips (tx_id) WHERE delete_tx_id is NULL;

-- Table: flip_keys

-- DROP TABLE flip_keys;

CREATE TABLE IF NOT EXISTS flip_keys
(
    tx_id bigint                                              NOT NULL,
    key   character varying(100) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT flip_keys_pkey PRIMARY KEY (tx_id),
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

CREATE TABLE IF NOT EXISTS mem_pool_flip_keys
(
    ei_address_state_id bigint                                              NOT NULL,
    key                 character varying(100) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT mem_pool_flip_keys_pkey PRIMARY KEY (ei_address_state_id),
    CONSTRAINT mem_pool_flip_keys_ei_address_state_id_fkey FOREIGN KEY (ei_address_state_id)
        REFERENCES epoch_identities (address_state_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
) WITH (
      OIDS = FALSE
    )
  TABLESPACE pg_default;

ALTER TABLE mem_pool_flip_keys
    OWNER to postgres;

-- Table: answers

-- DROP TABLE answers;

CREATE TABLE IF NOT EXISTS answers
(
    flip_tx_id          bigint   NOT NULL,
    ei_address_state_id bigint   NOT NULL,
    is_short            boolean  NOT NULL,
    answer              smallint NOT NULL,
    point               real     NOT NULL,
    grade               smallint NOT NULL,
    index               smallint,
    considered          boolean,
    CONSTRAINT answers_ei_address_state_id_fkey FOREIGN KEY (ei_address_state_id)
        REFERENCES epoch_identities (address_state_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT answers_flip_tx_id_fkey FOREIGN KEY (flip_tx_id)
        REFERENCES flips (tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT answers_answer_fkey FOREIGN KEY (answer)
        REFERENCES dic_answers (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE answers
    OWNER to postgres;

CREATE INDEX IF NOT EXISTS answers_long_reported_idx on answers (flip_tx_id) WHERE not is_short and grade = 1;
CREATE INDEX IF NOT EXISTS answers_short_idx on answers (flip_tx_id) WHERE is_short;
CREATE INDEX IF NOT EXISTS answers_long_idx on answers (flip_tx_id) WHERE not is_short;
CREATE INDEX IF NOT EXISTS answers_short_respondent_idx on answers (ei_address_state_id) WHERE is_short;
CREATE INDEX IF NOT EXISTS answers_long_respondent_idx on answers (ei_address_state_id) WHERE not is_short;

-- Table: flips_to_solve

-- DROP TABLE flips_to_solve;

CREATE TABLE IF NOT EXISTS flips_to_solve
(
    ei_address_state_id bigint  NOT NULL,
    flip_tx_id          bigint  NOT NULL,
    is_short            boolean NOT NULL,
    index               smallint,
    CONSTRAINT flips_to_solve_ei_address_state_id_fkey FOREIGN KEY (ei_address_state_id)
        REFERENCES epoch_identities (address_state_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT flips_to_solve_flip_tx_id_fkey FOREIGN KEY (flip_tx_id)
        REFERENCES flips (tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE flips_to_solve
    OWNER to postgres;

CREATE TABLE IF NOT EXISTS balances
(
    address_id bigint NOT NULL,
    balance    numeric(30, 18),
    stake      numeric(30, 18),
    CONSTRAINT balances_pkey PRIMARY KEY (address_id),
    CONSTRAINT balances_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE INDEX IF NOT EXISTS balances_balance_address_id_idx on balances (balance desc, address_id);
CREATE INDEX IF NOT EXISTS balances_stake_address_id_api_idx on balances (stake desc, address_id);

-- Table: birthdays

-- DROP TABLE birthdays;

CREATE TABLE IF NOT EXISTS birthdays
(
    address_id  bigint  NOT NULL,
    birth_epoch integer NOT NULL,
    CONSTRAINT birthdays_pkey PRIMARY KEY (address_id),
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

-- Table: block_flags

-- DROP TABLE block_flags;

CREATE TABLE IF NOT EXISTS block_flags
(
    block_height bigint                                             NOT NULL,
    flag         character varying(50) COLLATE pg_catalog."default" NOT NULL,
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

-- Table: flips_data

-- DROP TABLE flips_data;

CREATE TABLE IF NOT EXISTS flips_data
(
    flip_tx_id bigint NOT NULL,
    CONSTRAINT flips_data_pkey PRIMARY KEY (flip_tx_id),
    CONSTRAINT flips_data_flip_tx_id_fkey FOREIGN KEY (flip_tx_id)
        REFERENCES flips (tx_id) MATCH SIMPLE
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
    fd_flip_tx_id bigint   NOT NULL,
    index         smallint NOT NULL,
    data          bytea    NOT NULL,
    CONSTRAINT flip_pics_fd_flip_tx_id_fkey FOREIGN KEY (fd_flip_tx_id)
        REFERENCES flips_data (flip_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE flip_pics
    OWNER to postgres;

CREATE INDEX IF NOT EXISTS flip_pics_fd_flip_tx_id_idx on flip_pics (fd_flip_tx_id);

-- Table: flip_icons

-- DROP TABLE flip_icons;

CREATE TABLE IF NOT EXISTS flip_icons
(
    fd_flip_tx_id bigint NOT NULL,
    data          bytea  NOT NULL,
    CONSTRAINT flip_icons_pkey PRIMARY KEY (fd_flip_tx_id),
    CONSTRAINT flip_icons_fd_flip_tx_id_fkey FOREIGN KEY (fd_flip_tx_id)
        REFERENCES flips_data (flip_tx_id) MATCH SIMPLE
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
    fd_flip_tx_id  bigint   NOT NULL,
    answer_index   smallint NOT NULL,
    pos_index      smallint NOT NULL,
    flip_pic_index smallint NOT NULL,
    CONSTRAINT flip_pic_orders_fd_flip_tx_id_fkey FOREIGN KEY (fd_flip_tx_id)
        REFERENCES flips_data (flip_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE flip_pic_orders
    OWNER to postgres;

CREATE INDEX IF NOT EXISTS flip_pic_orders_fd_flip_tx_id_idx on flip_pic_orders (fd_flip_tx_id);

CREATE SEQUENCE IF NOT EXISTS penalties_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE penalties_id_seq
    OWNER TO postgres;

CREATE TABLE IF NOT EXISTS penalties
(
    id                        bigint          NOT NULL DEFAULT nextval('penalties_id_seq'::regclass),
    address_id                bigint          NOT NULL,
    penalty                   numeric(30, 18) NOT NULL,
    block_height              bigint          NOT NULL,
    penalty_seconds           smallint,
    inherited_from_address_id bigint,
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

CREATE INDEX IF NOT EXISTS penalties_address_id_idx on penalties (address_id);
CREATE INDEX IF NOT EXISTS penalties_id_address_id_idx on penalties (id desc, address_id);

CREATE TABLE IF NOT EXISTS total_rewards
(
    epoch             bigint          NOT NULL,
    total             numeric(30, 18) NOT NULL,
    validation        numeric(30, 18) NOT NULL,
    flips             numeric(30, 18) NOT NULL,
    invitations       numeric(30, 18) NOT NULL,
    foundation        numeric(30, 18) NOT NULL,
    zero_wallet       numeric(30, 18) NOT NULL,
    validation_share  numeric(30, 18) NOT NULL,
    flips_share       numeric(30, 18) NOT NULL,
    invitations_share numeric(30, 18) NOT NULL,
    reports           numeric(30, 18),
    reports_share     numeric(30, 18),
    staking           numeric(30, 18),
    candidate         numeric(30, 18),
    staking_share     numeric(30, 18),
    candidate_share   numeric(30, 18),
    flips_extra       numeric(30, 18),
    flips_extra_share numeric(30, 18),
    CONSTRAINT total_rewards_pkey PRIMARY KEY (epoch),
    CONSTRAINT total_rewards_epoch_fkey FOREIGN KEY (epoch)
        REFERENCES epochs (epoch) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

-- Table: validation_rewards

-- DROP TABLE validation_rewards;

CREATE TABLE IF NOT EXISTS validation_rewards
(
    ei_address_state_id bigint          NOT NULL,
    balance             numeric(30, 18) NOT NULL,
    stake               numeric(30, 18) NOT NULL,
    type                smallint        NOT NULL,
    CONSTRAINT validation_rewards_ei_address_state_id_type_key UNIQUE (ei_address_state_id, type),
    CONSTRAINT validation_rewards_ei_address_state_id_fkey FOREIGN KEY (ei_address_state_id)
        REFERENCES epoch_identities (address_state_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT validation_rewards_type_fkey FOREIGN KEY (type)
        REFERENCES dic_epoch_reward_types (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
) WITH (
      OIDS = FALSE
    )
  TABLESPACE pg_default;

ALTER TABLE validation_rewards
    OWNER to postgres;

CREATE TABLE IF NOT EXISTS reward_ages
(
    ei_address_state_id bigint  NOT NULL,
    age                 integer NOT NULL,
    CONSTRAINT reward_ages_pkey PRIMARY KEY (ei_address_state_id),
    CONSTRAINT reward_ages_ei_address_state_id_fkey FOREIGN KEY (ei_address_state_id)
        REFERENCES epoch_identities (address_state_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
) WITH (
      OIDS = FALSE
    )
  TABLESPACE pg_default;

ALTER TABLE reward_ages
    OWNER to postgres;

CREATE TABLE IF NOT EXISTS reward_staked_amounts
(
    ei_address_state_id bigint          NOT NULL,
    amount              numeric(30, 18) NOT NULL,
    failed              boolean,
    CONSTRAINT reward_staked_amounts_pkey PRIMARY KEY (ei_address_state_id)
);

-- Table: fund_rewards

-- DROP TABLE fund_rewards;

CREATE TABLE IF NOT EXISTS fund_rewards
(
    address_id   bigint          NOT NULL,
    block_height bigint          NOT NULL,
    balance      numeric(30, 18) NOT NULL,
    type         smallint        NOT NULL,
    CONSTRAINT fund_rewards_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT fund_rewards_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT fund_rewards_type_fkey FOREIGN KEY (type)
        REFERENCES dic_epoch_reward_types (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
) WITH (
      OIDS = FALSE
    )
  TABLESPACE pg_default;

ALTER TABLE fund_rewards
    OWNER to postgres;

CREATE TABLE IF NOT EXISTS bad_authors
(
    ei_address_state_id bigint   NOT NULL,
    reason              smallint NOT NULL,
    CONSTRAINT bad_authors_pkey PRIMARY KEY (ei_address_state_id),
    CONSTRAINT bad_authors_ei_address_state_id_fkey FOREIGN KEY (ei_address_state_id)
        REFERENCES epoch_identities (address_state_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT bad_authors_reason_fkey FOREIGN KEY (reason)
        REFERENCES dic_bad_author_reasons (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

-- Table: good_authors

-- DROP TABLE good_authors;

CREATE TABLE IF NOT EXISTS good_authors
(
    ei_address_state_id bigint   NOT NULL,
    strong_flips        smallint NOT NULL,
    weak_flips          smallint NOT NULL,
    successful_invites  smallint NOT NULL,
    CONSTRAINT good_authors_pkey PRIMARY KEY (ei_address_state_id),
    CONSTRAINT good_authors_ei_address_state_id_fkey FOREIGN KEY (ei_address_state_id)
        REFERENCES epoch_identities (address_state_id) MATCH SIMPLE
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
    flip_tx_id bigint   NOT NULL,
    word_1     smallint NOT NULL,
    word_2     smallint NOT NULL,
    tx_id      bigint   NOT NULL,
    CONSTRAINT flip_words_pkey PRIMARY KEY (flip_tx_id),
    CONSTRAINT flip_words_flip_tx_id_fkey FOREIGN KEY (flip_tx_id)
        REFERENCES flips (tx_id) MATCH SIMPLE
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

CREATE TABLE IF NOT EXISTS activation_txs
(
    tx_id        bigint NOT NULL,
    invite_tx_id bigint NOT NULL,
    shard_id     integer,
    CONSTRAINT activation_txs_pkey PRIMARY KEY (tx_id),
    CONSTRAINT activation_txs_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT activation_txs_invite_tx_id_key UNIQUE (invite_tx_id),
    CONSTRAINT activation_txs_invite_tx_id_fkey FOREIGN KEY (invite_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
-- CREATE UNIQUE INDEX IF NOT EXISTS activation_txs_for_rewarded_invitee_summary_idx on activation_txs (invite_tx_id);

CREATE TABLE IF NOT EXISTS kill_invitee_txs
(
    tx_id        bigint NOT NULL,
    invite_tx_id bigint NOT NULL,
    CONSTRAINT kill_invitee_txs_pkey PRIMARY KEY (tx_id),
    CONSTRAINT kill_invitee_txs_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT kill_invitee_txs_invite_tx_id_key UNIQUE (invite_tx_id),
    CONSTRAINT kill_invitee_txs_invite_tx_id_fkey FOREIGN KEY (invite_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS become_online_txs
(
    tx_id bigint NOT NULL,
    CONSTRAINT become_online_txs_pkey PRIMARY KEY (tx_id),
    CONSTRAINT become_online_txs_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS become_offline_txs
(
    tx_id bigint NOT NULL,
    CONSTRAINT become_offline_txs_pkey PRIMARY KEY (tx_id),
    CONSTRAINT become_offline_txs_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS rewarded_flips
(
    flip_tx_id bigint NOT NULL,
    extra      boolean,
    CONSTRAINT rewarded_flips_pkey PRIMARY KEY (flip_tx_id),
    CONSTRAINT rewarded_flips_flip_tx_id_fkey FOREIGN KEY (flip_tx_id)
        REFERENCES flips (tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS rewarded_invitations
(
    invite_tx_id bigint   NOT NULL,
    block_height bigint   NOT NULL,
    reward_type  smallint NOT NULL,
    epoch_height integer,
    CONSTRAINT rewarded_invitations_pkey PRIMARY KEY (invite_tx_id, block_height),
    CONSTRAINT rewarded_invitations_invite_tx_id_fkey FOREIGN KEY (invite_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT rewarded_invitations_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT rewarded_invitations_reward_type_fkey FOREIGN KEY (reward_type)
        REFERENCES dic_epoch_reward_types (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS rewarded_invitees
(
    invite_tx_id bigint NOT NULL,
    block_height bigint NOT NULL,
    epoch_height integer,
    CONSTRAINT rewarded_invitees_pkey PRIMARY KEY (invite_tx_id, block_height)
);
-- CREATE TABLE IF NOT EXISTS rewarded_invitees
-- (
--     epoch        integer NOT NULL,
--     address_id   bigint  NOT NULL,
--     invite_tx_id bigint  NOT NULL,
--     block_height bigint  NOT NULL,
--     epoch_height integer,
--     CONSTRAINT rewarded_invitees_pkey PRIMARY KEY (invite_tx_id, block_height)
-- );
-- CREATE UNIQUE INDEX IF NOT EXISTS rewarded_invitees_api_idx on rewarded_invitees (epoch, address_id);

CREATE TABLE IF NOT EXISTS saved_invite_rewards
(
    ei_address_state_id bigint   NOT NULL,
    reward_type         smallint NOT NULL,
    count               smallint NOT NULL,
    CONSTRAINT saved_invite_rewards_pkey PRIMARY KEY (ei_address_state_id, reward_type),
    CONSTRAINT saved_invite_rewards_ei_address_state_id_fkey FOREIGN KEY (ei_address_state_id)
        REFERENCES epoch_identities (address_state_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT saved_invite_rewards_reward_type_fkey FOREIGN KEY (reward_type)
        REFERENCES dic_epoch_reward_types (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS reported_flip_rewards
(
    ei_address_state_id bigint          NOT NULL,
    address_id          bigint          NOT NULL,
    epoch               bigint          NOT NULL,
    flip_tx_id          bigint          NOT NULL,
    balance             numeric(30, 18) NOT NULL,
    stake               numeric(30, 18) NOT NULL,
    CONSTRAINT reported_flip_rewards_ei_address_state_id_fkey FOREIGN KEY (ei_address_state_id)
        REFERENCES epoch_identities (address_state_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT reported_flip_rewards_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT reported_flip_rewards_epoch_fkey FOREIGN KEY (epoch)
        REFERENCES epochs (epoch) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT reported_flip_rewards_flip_tx_id_fkey FOREIGN KEY (flip_tx_id)
        REFERENCES flips (tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE UNIQUE INDEX IF NOT EXISTS reported_flip_rewards_unique_idx1 on reported_flip_rewards (ei_address_state_id, flip_tx_id);
CREATE UNIQUE INDEX IF NOT EXISTS reported_flip_rewards_unique_idx2 on reported_flip_rewards (address_id, flip_tx_id);

CREATE SEQUENCE IF NOT EXISTS balance_updates_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE TABLE IF NOT EXISTS balance_updates
(
    id                     bigint          NOT NULL DEFAULT nextval('balance_updates_id_seq'::regclass),
    address_id             bigint          NOT NULL,
    balance_old            numeric(30, 18) NOT NULL,
    stake_old              numeric(30, 18) NOT NULL,
    penalty_old            numeric(30, 18),
    balance_new            numeric(30, 18) NOT NULL,
    stake_new              numeric(30, 18) NOT NULL,
    penalty_new            numeric(30, 18),
    reason                 smallint        NOT NULL,
    block_height           bigint          NOT NULL,
    tx_id                  bigint,
    last_block_height      bigint,
    committee_reward_share numeric(30, 18),
    blocks_count           integer,
    penalty_seconds_old    smallint,
    penalty_seconds_new    smallint,
    penalty_payment        numeric(30, 18),
    contract_address_id    bigint,
    CONSTRAINT balance_updates_pkey PRIMARY KEY (id),
    CONSTRAINT balance_updates_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT balance_updates_reason_fkey FOREIGN KEY (reason)
        REFERENCES dic_balance_update_reasons (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT balance_updates_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT balance_updates_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT balance_updates_last_block_height_fkey FOREIGN KEY (last_block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE INDEX IF NOT EXISTS balance_updates_id_address_id_idx on balance_updates (id desc, address_id);
CREATE INDEX IF NOT EXISTS balance_updates_address_id_idx on balance_updates (address_id);
CREATE INDEX IF NOT EXISTS balance_updates_block_height_idx on balance_updates (block_height);

CREATE TABLE IF NOT EXISTS latest_committee_reward_balance_updates
(
    block_height        bigint          NOT NULL,
    address_id          bigint          NOT NULL,
    balance_old         numeric(30, 18) NOT NULL,
    stake_old           numeric(30, 18) NOT NULL,
    penalty_old         numeric(30, 18),
    balance_new         numeric(30, 18) NOT NULL,
    stake_new           numeric(30, 18) NOT NULL,
    penalty_new         numeric(30, 18),
    balance_update_id   bigint          NOT NULL,
    penalty_seconds_old smallint,
    penalty_seconds_new smallint,
    penalty_payment     numeric(30, 18),
    CONSTRAINT latest_committee_reward_balance_updates_pkey PRIMARY KEY (block_height, address_id),
    CONSTRAINT latest_committee_reward_balance_updates_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT latest_committee_reward_balance_updates_bu_id_fkey FOREIGN KEY (balance_update_id)
        REFERENCES balance_updates (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT latest_committee_reward_balance_updates_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE INDEX IF NOT EXISTS latest_committee_reward_block_height_idx
    on latest_committee_reward_balance_updates (block_height);
CREATE INDEX IF NOT EXISTS latest_committee_reward_du_id_idx
    on latest_committee_reward_balance_updates (balance_update_id);

CREATE TABLE IF NOT EXISTS flips_queue
(
    cid                    character varying(100) COLLATE pg_catalog."default" NOT NULL,
    key                    character varying(100) COLLATE pg_catalog."default" NOT NULL,
    attempts               smallint                                            NOT NULL,
    next_attempt_timestamp bigint                                              NOT NULL
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;

ALTER TABLE flips_queue
    OWNER to postgres;

CREATE INDEX IF NOT EXISTS flips_queue_next_attempt_timestamp_idx on flips_queue (next_attempt_timestamp desc);
CREATE UNIQUE INDEX IF NOT EXISTS flips_queue_cid_unique_idx on flips_queue (LOWER(cid));

CREATE TABLE IF NOT EXISTS dic_change_types
(
    id   smallint               NOT NULL,
    name character varying(100) NOT NULL,
    CONSTRAINT dic_change_types_pkey PRIMARY KEY (id),
    CONSTRAINT dic_change_types_name_key UNIQUE (name)
);
INSERT INTO dic_change_types
VALUES (1, 'oracle_voting_contract_results')
ON CONFLICT DO NOTHING;
INSERT INTO dic_change_types
VALUES (2, 'oracle_voting_contract_summaries')
ON CONFLICT DO NOTHING;
INSERT INTO dic_change_types
VALUES (3, 'sorted_oracle_voting_contracts')
ON CONFLICT DO NOTHING;
INSERT INTO dic_change_types
VALUES (4, 'sorted_oracle_voting_contract_committees')
ON CONFLICT DO NOTHING;
INSERT INTO dic_change_types
VALUES (5, 'balance_update_summaries')
ON CONFLICT DO NOTHING;
INSERT INTO dic_change_types
VALUES (6, 'mining_reward_summaries')
ON CONFLICT DO NOTHING;
INSERT INTO dic_change_types
VALUES (7, 'token_balances')
ON CONFLICT DO NOTHING;
INSERT INTO dic_change_types
VALUES (8, 'delegation_history')
ON CONFLICT DO NOTHING;

CREATE SEQUENCE IF NOT EXISTS changes_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;
CREATE TABLE IF NOT EXISTS changes
(
    id           bigint   NOT NULL DEFAULT nextval('changes_id_seq'::regclass),
    block_height bigint   NOT NULL,
    type         smallint NOT NULL,
    CONSTRAINT changes_pkey PRIMARY KEY (id),
    CONSTRAINT changes_type_fkey FOREIGN KEY (type)
        REFERENCES dic_change_types (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE INDEX IF NOT EXISTS changes_block_height_idx on changes (block_height);

CREATE OR REPLACE VIEW epoch_identity_states AS
SELECT s.id AS address_state_id,
       s.address_id,
       s.prev_id,
       s.state,
       s.block_height,
       ei.epoch
FROM address_states s
         JOIN blocks b ON b.height = s.block_height
         JOIN epoch_identities ei ON s.id = ei.address_state_id;

-- Types
DO
$$
    BEGIN
        -- Type: tp_mining_reward
        CREATE TYPE tp_mining_reward AS
        (
            address      character(42),
            balance      numeric(30, 18),
            stake        numeric(30, 18),
            proposer     boolean,
            stake_weight double precision
        );

        ALTER TYPE tp_mining_reward
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_burnt_coins
        CREATE TYPE tp_burnt_coins AS
        (
            address character(42),
            amount  numeric(30, 18),
            reason  smallint,
            tx_id   bigint
        );

        ALTER TYPE tp_burnt_coins
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_balance
        CREATE TYPE tp_balance AS
        (
            address character(42),
            balance numeric(30, 18),
            stake   numeric(30, 18)
        );

        ALTER TYPE tp_balance
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_activation_tx_transfer
        CREATE TYPE tp_activation_tx_transfer AS
        (
            tx_hash          character(66),
            balance_transfer numeric(30, 18)
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_kill_tx_transfer
        CREATE TYPE tp_kill_tx_transfer AS
        (
            tx_hash        character(66),
            stake_transfer numeric(30, 18)
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_kill_invitee_tx_transfer
        CREATE TYPE tp_kill_invitee_tx_transfer AS
        (
            tx_hash        character(66),
            stake_transfer numeric(30, 18)
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_tx_hash_id
        CREATE TYPE tp_tx_hash_id AS
        (
            hash character(66),
            id   bigint
        );

        ALTER TYPE tp_tx_hash_id
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_address
        CREATE TYPE tp_address AS
        (
            address      character(42),
            is_temporary boolean
        );

        ALTER TYPE tp_address
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_address_state_change
        CREATE TYPE tp_address_state_change AS
        (
            address   character(42),
            new_state smallint,
            tx_hash   character(66)
        );

        ALTER TYPE tp_address_state_change
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_answer AS
        (
            flip_cid   character varying(100),
            address    character(42),
            is_short   boolean,
            answer     smallint,
            point      real,
            grade      smallint,
            index      smallint,
            considered boolean
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_flip_state AS
        (
            flip_cid    character varying(100),
            answer      smallint,
            status      smallint,
            grade       smallint,
            grade_score real
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_birthday
        CREATE TYPE tp_birthday AS
        (
            address     character(42),
            birth_epoch integer
        );

        ALTER TYPE tp_birthday
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_epoch_identity AS
        (
            address            character(42),
            state              smallint,
            short_point        real,
            short_flips        integer,
            total_short_point  real,
            total_short_flips  integer,
            long_point         real,
            long_flips         integer,
            approved           boolean,
            missed             boolean,
            required_flips     smallint,
            available_flips    smallint,
            made_flips         smallint,
            next_epoch_invites smallint,
            birth_epoch        bigint,
            short_answers      integer,
            long_answers       integer,
            wrong_words_flips  smallint,
            delegatee_address  character(42),
            shard_id           integer,
            new_shard_id       integer,
            wrong_grade_reason smallint
        );

        ALTER TYPE tp_epoch_identity
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_flip_to_solve AS
        (
            address  character(42),
            cid      character varying(100),
            is_short boolean,
            index    smallint
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_good_author
        CREATE TYPE tp_good_author AS
        (
            address            character(42),
            strong_flips       integer,
            weak_flips         integer,
            successful_invites integer
        );

        ALTER TYPE tp_good_author
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_bad_author AS
        (
            address character(42),
            reason  smallint
        );

        ALTER TYPE tp_bad_author
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_total_epoch_reward
        CREATE TYPE tp_total_epoch_reward AS
        (
            total             numeric(30, 18),
            validation        numeric(30, 18),
            staking           numeric(30, 18),
            candidate         numeric(30, 18),
            flips             numeric(30, 18),
            flips_extra       numeric(30, 18),
            invitations       numeric(30, 18),
            foundation        numeric(30, 18),
            zero_wallet       numeric(30, 18),
            validation_share  numeric(30, 18),
            staking_share     numeric(30, 18),
            candidate_share   numeric(30, 18),
            flips_share       numeric(30, 18),
            flips_extra_share numeric(30, 18),
            invitations_share numeric(30, 18),
            reports           numeric(30, 18),
            reports_share     numeric(30, 18)
        );

        ALTER TYPE tp_total_epoch_reward
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_epoch_reward
        CREATE TYPE tp_epoch_reward AS
        (
            address character(42),
            balance numeric(30, 18),
            stake   numeric(30, 18),
            type    smallint
        );

        ALTER TYPE tp_epoch_reward
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_reward_age
        CREATE TYPE tp_reward_age AS
        (
            address character(42),
            age     integer
        );

        ALTER TYPE tp_reward_age
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_reward_staked_amount AS
        (
            address character(42),
            amount  numeric(30, 18)
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_mem_pool_flip_key
        CREATE TYPE tp_mem_pool_flip_key AS
        (
            address character(42),
            key     character varying(100)
        );

        ALTER TYPE tp_mem_pool_flip_key
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_failed_flip_content
        CREATE TYPE tp_failed_flip_content AS
        (
            cid                    character varying(100),
            attempts_limit_reached boolean,
            next_attempt_timestamp bigint
        );

        ALTER TYPE tp_failed_flip_content
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_deleted_flip
        CREATE TYPE tp_deleted_flip AS
        (
            tx_hash character(66),
            cid     character varying(100)
        );

        ALTER TYPE tp_deleted_flip
            OWNER TO postgres;
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_activation_tx
        CREATE TYPE tp_activation_tx AS
        (
            tx_hash        character(66),
            invite_tx_hash character(66),
            shard_id       integer
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        -- Type: tp_kill_invitee_tx
        CREATE TYPE tp_kill_invitee_tx AS
        (
            tx_hash        character(66),
            invite_tx_hash character(66)
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_rewarded_invitation AS
        (
            tx_hash      character(66),
            reward_type  smallint,
            epoch_height integer
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_rewarded_invitee AS
        (
            tx_hash      character(66),
            epoch_height integer
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_saved_invite_rewards AS
        (
            address     text,
            reward_type smallint,
            count       smallint
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_reported_flip_reward AS
        (
            address text,
            cid     text,
            balance numeric,
            stake   numeric
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;


DO
$$
    BEGIN
        CREATE TYPE tp_balance_update AS
        (
            address             text,
            balance_old         numeric(30, 18),
            stake_old           numeric(30, 18),
            penalty_old         numeric(30, 18),
            penalty_seconds_old smallint,
            balance_new         numeric(30, 18),
            stake_new           numeric(30, 18),
            penalty_new         numeric(30, 18),
            penalty_seconds_new smallint,
            penalty_payment     numeric(30, 18),
            tx_hash             text,
            reason              smallint,
            contract_address    text
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_flip_words AS
        (
            cid     text,
            word_1  smallint,
            word_2  smallint,
            tx_hash text
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_oracle_voting_contract AS
        (
            tx_hash                text,
            contract_address       text,
            stake                  numeric,
            start_time             bigint,
            voting_duration        bigint,
            voting_min_payment     numeric,
            fact                   text,
            state                  smallint,
            public_voting_duration bigint,
            winner_threshold       smallint,
            quorum                 smallint,
            committee_size         bigint,
            owner_fee              smallint,
            owner_deposit          numeric,
            oracle_reward_fund     numeric,
            refund_recipient       text,
            hash                   text,
            network_size           bigint
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_oracle_voting_contract_call_vote_proof AS
        (
            tx_hash                  text,
            vote_hash                text,
            secret_votes_count       bigint,
            discriminated_newbie     boolean,
            discriminated_delegation boolean,
            discriminated_stake      boolean
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_oracle_voting_contract_call_vote AS
        (
            tx_hash                  text,
            vote                     smallint,
            salt                     text,
            option_votes             bigint,
            option_all_votes         bigint,
            secret_votes_count       bigint,
            delegatee                text,
            prev_pool_vote           smallint,
            prev_option_votes        bigint,
            discriminated_newbie     boolean,
            discriminated_delegation boolean,
            discriminated_stake      boolean
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_oracle_voting_contract_call_finish AS
        (
            tx_hash       text,
            state         smallint,
            result        smallint,
            fund          numeric,
            oracle_reward numeric,
            owner_reward  numeric
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_oracle_voting_contract_call_add_stake AS
        (
            tx_hash text
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_oracle_voting_contract_termination AS
        (
            tx_hash       text,
            fund          numeric,
            oracle_reward numeric,
            owner_reward  numeric
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_oracle_lock_contract AS
        (
            tx_hash               text,
            contract_address      text,
            stake                 numeric,
            oracle_voting_address text,
            "value"               smallint,
            success_address       text,
            fail_address          text
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_oracle_lock_contract_call_check_oracle_voting AS
        (
            tx_hash              text,
            oracle_voting_result smallint
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;


DO
$$
    BEGIN
        CREATE TYPE tp_oracle_lock_contract_call_push AS
        (
            tx_hash              text,
            success              boolean,
            oracle_voting_result smallint,
            transfer             numeric
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_oracle_lock_contract_termination AS
        (
            tx_hash text,
            dest    text
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_refundable_oracle_lock_contract AS
        (
            tx_hash               text,
            contract_address      text,
            stake                 numeric,
            oracle_voting_address text,
            "value"               smallint,
            success_address       text,
            fail_address          text,
            refund_delay          bigint,
            deposit_deadline      bigint,
            oracle_voting_fee     smallint,
            oracle_voting_fee_new integer
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_refundable_oracle_lock_contract_call_deposit AS
        (
            tx_hash text,
            own_sum numeric,
            sum     numeric,
            fee     numeric
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_refundable_oracle_lock_contract_call_push AS
        (
            tx_hash              text,
            state                smallint,
            oracle_voting_exists boolean,
            oracle_voting_result smallint,
            transfer             numeric,
            refund_block         bigint
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_refundable_oracle_lock_contract_call_refund AS
        (
            tx_hash text,
            balance numeric,
            coef    double precision
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_refundable_oracle_lock_contract_termination AS
        (
            tx_hash text,
            dest    text
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_multisig_contract AS
        (
            tx_hash          text,
            contract_address text,
            stake            numeric,
            min_votes        smallint,
            max_votes        smallint,
            state            smallint
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_multisig_contract_call_add AS
        (
            tx_hash   text,
            address   text,
            new_state smallint
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_multisig_contract_call_send AS
        (
            tx_hash text,
            dest    text,
            amount  numeric
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_multisig_contract_call_push AS
        (
            tx_hash          text,
            dest             text,
            amount           numeric,
            vote_address_cnt smallint,
            vote_amount_cnt  smallint
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_multisig_contract_termination AS
        (
            tx_hash text,
            dest    text
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_time_lock_contract AS
        (
            tx_hash          text,
            contract_address text,
            stake            numeric,
            "timestamp"      bigint
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_time_lock_contract_call_transfer AS
        (
            tx_hash text,
            dest    text,
            amount  numeric
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_time_lock_contract_termination AS
        (
            tx_hash text,
            dest    text
        );
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
            insert into mining_rewards (address_id, block_height, balance, stake, proposer, stake_weight)
            values ((select id from addresses where lower(address) = lower(mr_row.address)), height,
                    mr_row.balance, mr_row.stake, mr_row.proposer, mr_row.stake_weight);
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

CREATE OR REPLACE PROCEDURE save_balances(p_block_height bigint,
                                          b tp_balance[],
                                          p_updates tp_balance_update[],
                                          p_blocks_count integer,
                                          p_committee_reward_share numeric(30, 18))
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    b_row tp_balance;
BEGIN
    if b is not null then
        for i in 1..cardinality(b)
            loop
                b_row = b[i];
                insert into balances (address_id, balance, stake)
                values (get_address_id_or_insert(p_block_height, b_row.address), b_row.balance, b_row.stake)
                on conflict (address_id) do update set balance=b_row.balance, stake=b_row.stake;
            end loop;
    end if;

    if p_updates is not null then
        call save_balance_updates(p_block_height, p_updates, p_committee_reward_share);
        call apply_balance_updates_on_contracts(p_block_height, p_updates);
    end if;

    delete
    from latest_committee_reward_balance_updates
    where block_height <= p_block_height - p_blocks_count;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_balance_updates(p_block_height bigint,
                                                 p_updates tp_balance_update[],
                                                 p_committee_reward_share numeric(30, 18))
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    COMMITTEE_REASON CONSTANT smallint = 3;
    l_balance_update          tp_balance_update;
    l_tx_id                   bigint;
    l_address_id              bigint;
    l_contract_address_id     bigint;
BEGIN
    if p_updates is null then
        return;
    end if;

    for i in 1..cardinality(p_updates)
        loop
            l_balance_update = p_updates[i];
            if l_balance_update.reason = COMMITTEE_REASON then
                call save_committee_reward_balance_update(p_block_height, l_balance_update,
                                                          p_committee_reward_share);
            else
                if char_length(l_balance_update.tx_hash) > 0 then
                    select id into l_tx_id from transactions where lower(hash) = lower(l_balance_update.tx_hash);
                else
                    l_tx_id = null;
                end if;
                if char_length(l_balance_update.contract_address) > 0 then
                    l_contract_address_id = get_address_id_or_insert(p_block_height, l_balance_update.contract_address);
                else
                    l_contract_address_id = null;
                end if;
                l_address_id = get_address_id_or_insert(p_block_height, l_balance_update.address);
                insert into balance_updates (address_id, balance_old, stake_old, penalty_old, penalty_seconds_old,
                                             balance_new, stake_new, penalty_new, penalty_seconds_new, penalty_payment,
                                             reason, block_height, tx_id, last_block_height, committee_reward_share,
                                             blocks_count, contract_address_id)
                values (l_address_id,
                        l_balance_update.balance_old,
                        l_balance_update.stake_old,
                        null_if_zero(l_balance_update.penalty_old),
                        null_if_zero_smallint(l_balance_update.penalty_seconds_old),
                        l_balance_update.balance_new,
                        l_balance_update.stake_new,
                        null_if_zero(l_balance_update.penalty_new),
                        null_if_zero_smallint(l_balance_update.penalty_seconds_new),
                        null_if_zero(l_balance_update.penalty_payment),
                        l_balance_update.reason,
                        p_block_height,
                        l_tx_id,
                        null,
                        null,
                        null,
                        l_contract_address_id);
            end if;

            call update_balance_update_summary(p_block_height, l_balance_update);
            call update_mining_reward_summary(p_block_height, l_balance_update);
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_committee_reward_balance_update(p_block_height bigint,
                                                                 p_update tp_balance_update,
                                                                 p_committee_reward_share numeric(30, 18))
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    COMMITTEE_REASON CONSTANT smallint = 3;
    l_address_id              bigint;
    l_balance_update_id       bigint;
    l_reason                  smallint;
    l_committee_reward        numeric(30, 18);
BEGIN
    select id
    into l_address_id
    from addresses
    where lower(address) = lower(p_update.address);

    select id, reason, committee_reward_share
    into l_balance_update_id, l_reason, l_committee_reward
    from balance_updates
    where address_id = l_address_id
    order by id desc
    limit 1;

    if l_reason = COMMITTEE_REASON and l_committee_reward = p_committee_reward_share then
        update balance_updates
        set balance_new         = p_update.balance_new,
            stake_new           = p_update.stake_new,
            penalty_new         = null_if_zero(p_update.penalty_new),
            penalty_seconds_new = null_if_zero_smallint(p_update.penalty_seconds_new),
            penalty_payment     = null_if_zero(coalesce(penalty_payment, 0) + coalesce(p_update.penalty_payment, 0)),
            blocks_count        = blocks_count + 1,
            last_block_height   = p_block_height
        where id = l_balance_update_id;
    else
        insert into balance_updates (address_id, balance_old, stake_old, penalty_old, penalty_seconds_old, balance_new,
                                     stake_new, penalty_new, penalty_seconds_new, penalty_payment, reason, block_height,
                                     tx_id, last_block_height, committee_reward_share, blocks_count)
        values (l_address_id,
                p_update.balance_old,
                p_update.stake_old,
                null_if_zero(p_update.penalty_old),
                null_if_zero_smallint(p_update.penalty_seconds_old),
                p_update.balance_new,
                p_update.stake_new,
                null_if_zero(p_update.penalty_new),
                null_if_zero_smallint(p_update.penalty_seconds_new),
                null_if_zero(p_update.penalty_payment),
                COMMITTEE_REASON,
                p_block_height,
                null,
                p_block_height,
                p_committee_reward_share,
                1)
        returning id into l_balance_update_id;
    end if;

    insert into latest_committee_reward_balance_updates (block_height, address_id, balance_old, stake_old, penalty_old,
                                                         penalty_seconds_old, balance_new, stake_new, penalty_new,
                                                         penalty_seconds_new, penalty_payment, balance_update_id)
    values (p_block_height,
            l_address_id,
            p_update.balance_old,
            p_update.stake_old,
            null_if_zero(p_update.penalty_old),
            null_if_zero_smallint(p_update.penalty_seconds_old),
            p_update.balance_new,
            p_update.stake_new,
            null_if_zero(p_update.penalty_new),
            null_if_zero_smallint(p_update.penalty_seconds_new),
            null_if_zero(p_update.penalty_payment),
            l_balance_update_id);
END
$BODY$;

CREATE OR REPLACE FUNCTION null_if_zero(v numeric)
    RETURNS numeric
    LANGUAGE 'plpgsql'
AS
$BODY$
BEGIN
    if v = 0.0 then
        return null;
    end if;
    return v;
END
$BODY$;

CREATE OR REPLACE FUNCTION null_if_zero_smallint(v smallint)
    RETURNS numeric
    LANGUAGE 'plpgsql'
AS
$BODY$
BEGIN
    if v = 0 then
        return null;
    end if;
    return v;
END
$BODY$;

CREATE OR REPLACE FUNCTION null_if_negative_numeric(v numeric)
    RETURNS numeric
    LANGUAGE 'plpgsql'
AS
$BODY$
BEGIN
    if v < 0 then
        return null;
    end if;
    return v;
END
$BODY$;

CREATE OR REPLACE FUNCTION null_if_zero_bigint(v bigint)
    RETURNS numeric
    LANGUAGE 'plpgsql'
AS
$BODY$
BEGIN
    if v = 0 then
        return null;
    end if;
    return v;
END
$BODY$;

CREATE OR REPLACE FUNCTION null_if_negative_bigint(v bigint)
    RETURNS numeric
    LANGUAGE 'plpgsql'
AS
$BODY$
BEGIN
    if v < 0 then
        return null;
    end if;
    return v;
END
$BODY$;

CREATE OR REPLACE FUNCTION limited_text(v text, max_length integer)
    RETURNS text
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    if v is not null and char_length(v) > max_length then
        v = substring(v from 1 for max_length);
    end if;
    return v;
END
$$;

CREATE OR REPLACE FUNCTION save_addrs_and_txs(p_height bigint,
                                              p_changes_blocks_count smallint,
                                              addresses tp_address[],
                                              p_activation_tx_transfers tp_activation_tx_transfer[],
                                              p_kill_tx_transfers tp_kill_tx_transfer[],
                                              p_kill_invitee_tx_transfers tp_kill_invitee_tx_transfer[],
                                              address_state_changes tp_address_state_change[],
                                              deleted_flips tp_deleted_flip[],
                                              p_activation_txs tp_activation_tx[],
                                              p_kill_invitee_txs tp_kill_invitee_tx[],
                                              p_become_online_txs character(66)[],
                                              p_become_offline_txs character(66)[],
                                              p_oracle_voting_contracts tp_oracle_voting_contract[],
                                              p_oracle_voting_contract_call_starts jsonb[],
                                              p_oracle_voting_contract_call_vote_proofs tp_oracle_voting_contract_call_vote_proof[],
                                              p_oracle_voting_contract_call_votes tp_oracle_voting_contract_call_vote[],
                                              p_oracle_voting_contract_call_finishes tp_oracle_voting_contract_call_finish[],
                                              p_oracle_voting_contract_call_prolongations jsonb[],
                                              p_oracle_voting_contract_call_add_stakes tp_oracle_voting_contract_call_add_stake[],
                                              p_oracle_voting_contract_terminations tp_oracle_voting_contract_termination[],
                                              p_clear_old_ovc_committees boolean,
                                              p_oracle_lock_contracts tp_oracle_lock_contract[],
                                              p_oracle_lock_contract_call_check_oracle_votings tp_oracle_lock_contract_call_check_oracle_voting[],
                                              p_oracle_lock_contract_call_pushes tp_oracle_lock_contract_call_push[],
                                              p_oracle_lock_contract_terminations tp_oracle_lock_contract_termination[],
                                              p_refundable_oracle_lock_contracts tp_refundable_oracle_lock_contract[],
                                              p_refundable_oracle_lock_contract_call_deposits tp_refundable_oracle_lock_contract_call_deposit[],
                                              p_refundable_oracle_lock_contract_call_pushes tp_refundable_oracle_lock_contract_call_push[],
                                              p_refundable_oracle_lock_contract_call_refunds tp_refundable_oracle_lock_contract_call_refund[],
                                              p_refundable_oracle_lock_contract_terminations tp_refundable_oracle_lock_contract_termination[],
                                              p_time_lock_contracts tp_time_lock_contract[],
                                              p_time_lock_contract_call_transfers tp_time_lock_contract_call_transfer[],
                                              p_time_lock_contract_terminations tp_time_lock_contract_termination[],
                                              p_multisig_contracts tp_multisig_contract[],
                                              p_multisig_contract_call_adds tp_multisig_contract_call_add[],
                                              p_multisig_contract_call_sends tp_multisig_contract_call_send[],
                                              p_multisig_contract_call_pushes tp_multisig_contract_call_push[],
                                              p_multisig_contract_terminations tp_multisig_contract_termination[],
                                              p_contract_tx_balance_updates jsonb[],
                                              p_data jsonb)
    RETURNS tp_tx_hash_id[]
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_address_id             bigint;
    address_row              tp_address;
    address_state_change_row tp_address_state_change;
    l_prev_state_id          bigint;
    res                      tp_tx_hash_id[];
    deleted_flip             tp_deleted_flip;
BEGIN
    if addresses is not null then
        for i in 1..cardinality(addresses)
            loop
                address_row = addresses[i];

                l_address_id = get_address_id_or_insert(p_height, address_row.address);

                if address_row.is_temporary then
                    insert into temporary_identities (address_id, block_height)
                    values (l_address_id, p_height)
                    on conflict (address_id) do nothing;
                end if;
            end loop;
    end if;

    if p_data is not null then
        res = save_txs(p_height, p_data -> 'txs', p_data -> 'epochSummaryUpdate');
    end if;

    if p_activation_tx_transfers is not null then
        call save_activation_tx_transfers(p_activation_tx_transfers);
    end if;

    if p_kill_tx_transfers is not null then
        call save_kill_tx_transfers(p_kill_tx_transfers);
    end if;

    if p_kill_invitee_tx_transfers is not null then
        call save_kill_invitee_tx_transfers(p_kill_invitee_tx_transfers);
    end if;

    if p_activation_txs is not null then
        call save_activation_txs(p_activation_txs);
    end if;

    if p_kill_invitee_txs is not null then
        call save_kill_invitee_txs(p_kill_invitee_txs);
    end if;

    if p_become_online_txs is not null then
        call save_become_online_txs(p_become_online_txs);
    end if;

    if p_become_offline_txs is not null then
        call save_become_offline_txs(p_become_offline_txs);
    end if;

    if address_state_changes is not null then
        for i in 1..cardinality(address_state_changes)
            loop
                address_state_change_row := address_state_changes[i];

                select id
                into l_address_id
                from addresses
                where lower(address) = lower(address_state_change_row.address);

                update address_states
                set is_actual = false
                where address_id = l_address_id
                  and is_actual
                returning id into l_prev_state_id;

                insert into address_states (address_id, state, is_actual, block_height, tx_id, prev_id)
                values (l_address_id, address_state_change_row.new_state, true, p_height,
                        (select id from transactions where lower(hash) = lower(address_state_change_row.tx_hash)),
                        l_prev_state_id);
            end loop;
    end if;

    if deleted_flips is not null then
        for i in 1..cardinality(deleted_flips)
            loop
                deleted_flip = deleted_flips[i];
                update flips
                set delete_tx_id=(SELECT ID FROM TRANSACTIONS WHERE LOWER(HASH) = lower(deleted_flip.tx_hash))
                where lower(cid) = lower(deleted_flip.cid);
            end loop;
    end if;

    if p_oracle_voting_contracts is not null then
        call save_oracle_voting_contracts(p_height, p_oracle_voting_contracts);
    end if;

    if p_oracle_voting_contract_call_starts is not null then
        call save_oracle_voting_contract_call_starts(p_height, p_oracle_voting_contract_call_starts);
    end if;

    if p_oracle_voting_contract_call_vote_proofs is not null then
        call save_oracle_voting_contract_call_vote_proofs(p_height, p_oracle_voting_contract_call_vote_proofs);
    end if;

    if p_oracle_voting_contract_call_votes is not null then
        call save_oracle_voting_contract_call_votes(p_height, p_oracle_voting_contract_call_votes);
    end if;

    if p_oracle_voting_contract_call_finishes is not null then
        call save_oracle_voting_contract_call_finishes(p_height, p_oracle_voting_contract_call_finishes);
    end if;

    if p_oracle_voting_contract_call_prolongations is not null then
        call save_oracle_voting_contract_call_prolongations(p_height, p_oracle_voting_contract_call_prolongations);
    end if;

    if p_oracle_voting_contract_call_add_stakes is not null then
        call save_oracle_voting_contract_call_add_stakes(p_height, p_oracle_voting_contract_call_add_stakes);
    end if;

    if p_oracle_voting_contract_terminations is not null then
        call save_oracle_voting_contract_terminations(p_height, p_oracle_voting_contract_terminations);
    end if;

    if p_oracle_lock_contracts is not null then
        call save_oracle_lock_contracts(p_height, p_oracle_lock_contracts);
    end if;

    if p_oracle_lock_contract_call_check_oracle_votings is not null then
        call save_oracle_lock_contract_call_check_oracle_votings(p_oracle_lock_contract_call_check_oracle_votings);
    end if;

    if p_oracle_lock_contract_call_pushes is not null then
        call save_oracle_lock_contract_call_pushes(p_oracle_lock_contract_call_pushes);
    end if;

    if p_oracle_lock_contract_terminations is not null then
        call save_oracle_lock_contract_terminations(p_height, p_oracle_lock_contract_terminations);
    end if;

    if p_refundable_oracle_lock_contracts is not null then
        call save_refundable_oracle_lock_contracts(p_height, p_refundable_oracle_lock_contracts);
    end if;

    if p_refundable_oracle_lock_contract_call_deposits is not null then
        call save_refundable_oracle_lock_contract_call_deposits(p_refundable_oracle_lock_contract_call_deposits);
    end if;

    if p_refundable_oracle_lock_contract_call_pushes is not null then
        call save_refundable_oracle_lock_contract_call_pushes(p_refundable_oracle_lock_contract_call_pushes);
    end if;

    if p_refundable_oracle_lock_contract_call_refunds is not null then
        call save_refundable_oracle_lock_contract_call_refunds(p_refundable_oracle_lock_contract_call_refunds);
    end if;

    if p_refundable_oracle_lock_contract_terminations is not null then
        call save_refundable_oracle_lock_contract_terminations(p_height,
                                                               p_refundable_oracle_lock_contract_terminations);
    end if;

    if p_time_lock_contracts is not null then
        call save_time_lock_contracts(p_height, p_time_lock_contracts);
    end if;

    if p_time_lock_contract_call_transfers is not null then
        call save_time_lock_contract_call_transfers(p_height, p_time_lock_contract_call_transfers);
    end if;

    if p_time_lock_contract_terminations is not null then
        call save_time_lock_contract_terminations(p_height, p_time_lock_contract_terminations);
    end if;

    if p_multisig_contracts is not null then
        call save_multisig_contracts(p_height, p_multisig_contracts);
    end if;

    if p_multisig_contract_call_adds is not null then
        call save_multisig_contract_call_adds(p_height, p_multisig_contract_call_adds);
    end if;

    if p_multisig_contract_call_sends is not null then
        call save_multisig_contract_call_sends(p_height, p_multisig_contract_call_sends);
    end if;

    if p_multisig_contract_call_pushes is not null then
        call save_multisig_contract_call_pushes(p_height, p_multisig_contract_call_pushes);
    end if;

    if p_multisig_contract_terminations is not null then
        call save_multisig_contract_terminations(p_height, p_multisig_contract_terminations);
    end if;

    if p_data is not null then
        call save_contracts(p_height, p_data -> 'contracts');
    end if;

    if p_contract_tx_balance_updates is not null then
        call save_contract_tx_balance_updates(p_height, p_contract_tx_balance_updates);
    end if;

    if p_data is not null then
        call save_tx_receipts(p_height, p_data -> 'txReceipts');
        call save_delegation_switches(p_height, p_data -> 'delegationSwitches');
        call update_pool_sizes(p_height, p_data -> 'poolSizes');
        call save_upgrades_votes(p_height, p_data -> 'upgradesVotes');
        call save_miners_history_item(p_height, p_data -> 'minersHistoryItem');
        call save_removed_transitive_delegations(p_height, p_data -> 'removedTransitiveDelegations');
        call save_oracle_voting_contracts_to_prolong(p_height, p_data -> 'oracleVotingContractsToProlong');
        call save_tokens(p_height, p_data -> 'tokens');
        call save_token_balance_updates(p_height, p_data -> 'tokenBalanceUpdates');
        call save_delegation_history_updates(p_height, p_data -> 'delegationHistoryUpdates');
    end if;

    call apply_block_on_sorted_contracts(p_height, p_clear_old_ovc_committees);

    DELETE FROM changes WHERE block_height <= p_height - p_changes_blocks_count;

    return res;
END
$BODY$;

CREATE OR REPLACE FUNCTION save_txs(p_block_height bigint,
                                    p_items jsonb,
                                    p_epoch_summary_update jsonb)
    RETURNS tp_tx_hash_id[]
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                 jsonb;
    l_invites_count        integer = 0;
    l_flips_count_diff     integer = 0;
    l_cur_epoch_min_tx_id  bigint;
    l_to                   bigint;
    l_address_id           bigint;
    l_tx_id                bigint;
    l_res                  tp_tx_hash_id[];
    l_type                 smallint;
    l_epoch_min_tx_id      bigint;
    l_epoch_max_tx_id      bigint;
    l_data                 jsonb;
    l_candidate_count_diff integer;
    l_tx_count_diff        integer = 0;
BEGIN

    if p_items is not null then

        l_tx_count_diff = jsonb_array_length(p_items);

        SELECT min_tx_id
        INTO l_cur_epoch_min_tx_id
        FROM epoch_summaries
        WHERE epoch = (SELECT epoch FROM blocks WHERE height = p_block_height);

        for i in 0..jsonb_array_length(p_items) - 1
            loop
                l_item = (p_items ->> i)::jsonb;
                l_type = (l_item ->> 'type')::smallint;

                l_to = null;
                IF char_length((l_item ->> 'to')::text) > 0 THEN
                    select id into l_to from addresses where lower(address) = lower((l_item ->> 'to')::text);
                end if;
                SELECT id INTO l_address_id FROM addresses WHERE lower(address) = lower((l_item ->> 'from')::text);
                INSERT INTO transactions (HASH, BLOCK_HEIGHT, type, "from", "to", AMOUNT, TIPS, MAX_FEE, FEE, SIZE,
                                          nonce, used_gas)
                VALUES ((l_item ->> 'hash')::text,
                        p_block_height,
                        l_type,
                        l_address_id,
                        l_to,
                        (l_item ->> 'amount')::numeric,
                        (l_item ->> 'tips')::numeric,
                        (l_item ->> 'maxFee')::numeric,
                        (l_item ->> 'fee')::numeric,
                        (l_item ->> 'size')::integer,
                        (l_item ->> 'nonce')::integer,
                        (l_item ->> 'usedGas')::integer)
                RETURNING id into l_tx_id;

                INSERT INTO transaction_raws (tx_id, raw) VALUES (l_tx_id, decode((l_item ->> 'raw')::text, 'hex'));

                l_res = array_append(l_res, ((l_item ->> 'hash')::text, l_tx_id)::tp_tx_hash_id);

                l_data = l_item -> 'data';

                if l_type = 6 then
                    -- SubmitShortAnswersTx
                    if l_data is not null then
                        INSERT INTO short_answer_txs (tx_id, client_type)
                        VALUES (l_tx_id, (l_data ->> 'clientType')::smallint);
                    end if;
                end if;

                if l_type = 2 then
                    -- InviteTx
                    l_invites_count = l_invites_count + 1;
                end if;
                if l_type = 4 then
                    -- SubmitFlipTx
                    l_flips_count_diff = l_flips_count_diff + 1;
                    CALL update_address_summary(p_address_id => l_address_id, p_flips_diff => 1);
                end if;
                if l_type = 14 then
                    -- DeleteFlipTx
                    l_flips_count_diff = l_flips_count_diff - 1;
                    CALL update_address_summary(p_address_id => l_address_id, p_flips_diff => -1);
                end if;

                if l_cur_epoch_min_tx_id is null and l_epoch_min_tx_id is null then
                    l_epoch_min_tx_id = l_tx_id;
                end if;

                l_epoch_max_tx_id = l_tx_id;

            end loop;
    end if;

    if p_epoch_summary_update is not null then
        l_candidate_count_diff = (p_epoch_summary_update ->> 'candidateCountDiff')::integer;
    end if;

    call update_epoch_summary(p_block_height => p_block_height,
                              p_tx_count_diff => l_tx_count_diff,
                              p_invite_count_diff => l_invites_count,
                              p_flip_count_diff => l_flips_count_diff,
                              p_min_tx_id => l_epoch_min_tx_id,
                              p_max_tx_id => l_epoch_max_tx_id,
                              p_candidate_count_diff => l_candidate_count_diff);

    return l_res;
END
$$;

-- PROCEDURE: save_activation_tx_transfers
CREATE OR REPLACE PROCEDURE save_activation_tx_transfers(p_activation_tx_transfers tp_activation_tx_transfer[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_activation_tx_transfer tp_activation_tx_transfer;
BEGIN
    for i in 1..cardinality(p_activation_tx_transfers)
        loop
            l_activation_tx_transfer := p_activation_tx_transfers[i];
            insert into activation_tx_transfers (tx_id, balance_transfer)
            values ((select id from transactions where lower(hash) = lower(l_activation_tx_transfer.tx_hash)),
                    l_activation_tx_transfer.balance_transfer);
        end loop;
END
$BODY$;

-- PROCEDURE: save_kill_tx_transfers
CREATE OR REPLACE PROCEDURE save_kill_tx_transfers(p_kill_tx_transfers tp_kill_tx_transfer[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_kill_tx_transfer tp_kill_tx_transfer;
BEGIN
    for i in 1..cardinality(p_kill_tx_transfers)
        loop
            l_kill_tx_transfer := p_kill_tx_transfers[i];
            insert into kill_tx_transfers (tx_id, stake_transfer)
            values ((select id from transactions where lower(hash) = lower(l_kill_tx_transfer.tx_hash)),
                    l_kill_tx_transfer.stake_transfer);
        end loop;
END
$BODY$;

-- PROCEDURE: save_kill_invitee_tx_transfers
CREATE OR REPLACE PROCEDURE save_kill_invitee_tx_transfers(p_kill_invitee_tx_transfers tp_kill_invitee_tx_transfer[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_kill_invitee_tx_transfer tp_kill_invitee_tx_transfer;
BEGIN
    for i in 1..cardinality(p_kill_invitee_tx_transfers)
        loop
            l_kill_invitee_tx_transfer := p_kill_invitee_tx_transfers[i];
            insert into kill_invitee_tx_transfers (tx_id, stake_transfer)
            values ((select id from transactions where lower(hash) = lower(l_kill_invitee_tx_transfer.tx_hash)),
                    l_kill_invitee_tx_transfer.stake_transfer);
        end loop;
END
$BODY$;

-- PROCEDURE: save_activation_txs
CREATE OR REPLACE PROCEDURE save_activation_txs(p_activation_txs tp_activation_tx[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_activation_tx               tp_activation_tx;
    l_invite_tx_id                bigint;
    l_activation_tx_id            bigint;
    l_activation_tx_to_address_id bigint;
    l_activation_tx_epoch         bigint;
BEGIN
    for i in 1..cardinality(p_activation_txs)
        loop
            l_activation_tx := p_activation_txs[i];
            select id into l_invite_tx_id from transactions where lower(hash) = lower(l_activation_tx.invite_tx_hash);
            if l_invite_tx_id is null then
                continue;
            end if;

            SELECT t.id, t."to", b.epoch
            INTO l_activation_tx_id, l_activation_tx_to_address_id, l_activation_tx_epoch
            FROM transactions t
                     LEFT JOIN blocks b ON b.height = t.block_height
            WHERE lower(t.hash) = lower(l_activation_tx.tx_hash);

            insert into activation_txs (tx_id, invite_tx_id, shard_id)
            values (l_activation_tx_id, l_invite_tx_id,
                    l_activation_tx.shard_id);

            INSERT INTO latest_activation_txs (activation_tx_id, address_id, epoch)
            VALUES (l_activation_tx_id, l_activation_tx_to_address_id, l_activation_tx_epoch);
        end loop;
END
$BODY$;

-- PROCEDURE: save_kill_invitee_txs
CREATE OR REPLACE PROCEDURE save_kill_invitee_txs(p_kill_invitee_txs tp_kill_invitee_tx[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_kill_invitee_tx tp_kill_invitee_tx;
    l_invite_tx_id    bigint;
BEGIN
    for i in 1..cardinality(p_kill_invitee_txs)
        loop
            l_kill_invitee_tx := p_kill_invitee_txs[i];
            select id into l_invite_tx_id from transactions where lower(hash) = lower(l_kill_invitee_tx.invite_tx_hash);
            if l_invite_tx_id is null then
                continue;
            end if;
            insert into kill_invitee_txs (tx_id, invite_tx_id)
            values ((select id from transactions where lower(hash) = lower(l_kill_invitee_tx.tx_hash)), l_invite_tx_id);
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_become_online_txs(p_become_online_txs character(66)[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_become_online_tx character(66);
BEGIN
    for i in 1..cardinality(p_become_online_txs)
        loop
            l_become_online_tx := p_become_online_txs[i];
            insert into become_online_txs (tx_id)
            values ((select id from transactions where lower(hash) = lower(l_become_online_tx)));
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_become_offline_txs(p_become_offline_txs character(66)[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_become_offline_tx character(66);
BEGIN
    for i in 1..cardinality(p_become_offline_txs)
        loop
            l_become_offline_tx := p_become_offline_txs[i];
            insert into become_offline_txs (tx_id)
            values ((select id from transactions where lower(hash) = lower(l_become_offline_tx)));
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_flips_words(p_flips_words tp_flip_words[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_flip_words tp_flip_words;
    l_flip_tx_id bigint;
BEGIN
    for i in 1..cardinality(p_flips_words)
        loop
            l_flip_words := p_flips_words[i];
            SELECT tx_id into l_flip_tx_id FROM flips WHERE lower(cid) = lower(l_flip_words.cid);
            if l_flip_tx_id is null then
                continue;
            end if;
            INSERT INTO flip_words (flip_tx_id, word_1, word_2, tx_id)
            VALUES (l_flip_tx_id,
                    l_flip_words.word_1,
                    l_flip_words.word_2,
                    (SELECT id FROM transactions WHERE lower(hash) = lower(l_flip_words.tx_hash)));
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE update_flips_queue()
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_epoch                      bigint;
    l_max_height                 bigint;
    l_last_block_timestamp       bigint;
    l_epoch_min_tx_id            bigint;
    l_max_tx_id                  bigint;
    l_epoch_min_address_state_id bigint;
    l_epoch_max_address_state_id bigint;
BEGIN
    select max(epoch) into l_epoch from epochs;

    SELECT max(id) INTO l_max_tx_id FROM transactions;

    SELECT min(id)
    INTO l_epoch_min_tx_id
    FROM transactions
    WHERE block_height = (SELECT min(block_height)
                          FROM transactions
                          WHERE block_height >=
                                (SELECT min(height) FROM blocks WHERE epoch = l_epoch));

    SELECT max(height) INTO l_max_height FROM blocks WHERE epoch = l_epoch;
    SELECT "timestamp" INTO l_last_block_timestamp FROM blocks WHERE height = l_max_height;

    SELECT min(address_state_id),
           max(address_state_id)
    INTO l_epoch_min_address_state_id, l_epoch_max_address_state_id
    FROM cur_epoch_identities;

    insert into flips_queue
        (select f.cid, coalesce(fk.key, mpfk.key) "key", 0, l_last_block_timestamp + 30 * 60
         from flips f

                  join transactions t on f.tx_id = t.id

                  left join (select distinct on (t.from) fk.key, t.from
                             from flip_keys fk
                                      join transactions t on t.id = fk.tx_id
                             WHERE fk.tx_id >= l_epoch_min_tx_id
                               AND fk.tx_id <= l_max_tx_id) fk on t.from = fk.from

                  left join (select mpfk.key, s.address_id
                             from mem_pool_flip_keys mpfk
                                      join address_states s on s.id = mpfk.ei_address_state_id
                             WHERE mpfk.ei_address_state_id >= l_epoch_min_address_state_id
                               AND mpfk.ei_address_state_id <= l_epoch_max_address_state_id) mpfk
                            on t.from = mpfk.address_id

         where f.tx_id >= l_epoch_min_tx_id
           AND f.tx_id <= l_max_tx_id
           AND (fk.key is not null or mpfk.key is not null));
END
$BODY$;

CREATE OR REPLACE PROCEDURE reset_to(p_block_height bigint)
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_epoch               bigint;
    l_is_epoch_last_block boolean;
    l_timestamp           bigint;
    l_min_tx_id_to_delete bigint;
BEGIN

    SET session_replication_role = replica;

    call reset_balance_updates_to(p_block_height);

    if p_block_height < (select min(height) - 1 from blocks) then
        raise exception 'wrong block height % to reset', p_block_height;
    end if;

    call reset_changes_to(p_block_height);
    call reset_contracts_to(p_block_height);
    call reset_upgrade_voting_history_to(p_block_height);

    select epoch, "timestamp" into l_epoch, l_timestamp from blocks where height = greatest(2, p_block_height);

    l_epoch = coalesce(l_epoch, 0);
    l_timestamp = coalesce(l_timestamp, 0);

    SELECT min(t.id) INTO l_min_tx_id_to_delete FROM transactions t WHERE t.block_height > p_block_height;

    l_is_epoch_last_block = (SELECT exists(SELECT 1
                                           FROM block_flags
                                           WHERE block_height = p_block_height
                                             AND flag = 'ValidationFinished'));

    DELETE FROM oracle_voting_contract_authors_and_voters WHERE deploy_or_vote_tx_id >= l_min_tx_id_to_delete;

    delete
    from flips_queue
    where lower(cid) in (select f.cid
                         from flips f,
                              transactions t,
                              blocks b
                         where f.tx_id = t.id
                           and t.block_height = b.height
                           and b.epoch + 1 > l_epoch);

    delete
    from flip_pics
    where fd_flip_tx_id in
          (select t.id
           from transactions t
                    join blocks b on b.height = t.block_height and
                                     b.epoch + 1 > l_epoch);

    delete
    from flip_icons
    where fd_flip_tx_id in
          (select t.id
           from transactions t
                    join blocks b on b.height = t.block_height and
                                     b.epoch + 1 > l_epoch);

    delete
    from flip_pic_orders
    where fd_flip_tx_id in
          (select t.id
           from transactions t
                    join blocks b on b.height = t.block_height and
                                     b.epoch + 1 > l_epoch);

    delete
    from flips_data
    where flip_tx_id in
          (select t.id
           from transactions t
                    join blocks b on b.height = t.block_height and
                                     b.epoch + 1 > l_epoch);

    delete
    from rewarded_flips
    where flip_tx_id in
          (select t.id
           from transactions t
                    join blocks b on b.height = t.block_height and
                                     b.epoch + 1 > l_epoch);

    delete
    from reported_flip_rewards
    where flip_tx_id in
          (select t.id
           from transactions t
                    join blocks b on b.height = t.block_height and
                                     b.epoch + 1 > l_epoch);

    delete
    from flip_summaries
    where flip_tx_id in
          (select t.id
           from transactions t
                    join blocks b on b.height = t.block_height and
                                     b.epoch + 1 > l_epoch);

    DELETE FROM miners_history WHERE block_timestamp > l_timestamp;

    delete
    from epoch_identity_interim_states
    where block_height > p_block_height;

    delete
    from block_proposer_vrf_scores
    where block_height > p_block_height;

    delete
    from burnt_coins
    where block_height > p_block_height;

    delete
    from failed_validations
    where block_height > p_block_height;

    delete
    from fund_rewards
    where block_height > p_block_height;

    delete
    from penalties
    where block_height > p_block_height;

    delete
    from epoch_summaries
    where block_height > p_block_height;

    DELETE FROM epoch_reward_bounds WHERE epoch > l_epoch OR NOT l_is_epoch_last_block AND epoch = l_epoch;
    DELETE FROM total_rewards WHERE epoch > l_epoch OR NOT l_is_epoch_last_block AND epoch = l_epoch;
    DELETE FROM epoch_flip_statuses WHERE epoch > l_epoch OR NOT l_is_epoch_last_block AND epoch = l_epoch;
    DELETE FROM removed_transitive_delegations WHERE epoch > l_epoch OR NOT l_is_epoch_last_block AND epoch = l_epoch;
    DELETE FROM delegatee_validation_rewards WHERE epoch > l_epoch OR NOT l_is_epoch_last_block AND epoch = l_epoch;
    DELETE
    FROM delegatee_total_validation_rewards
    WHERE epoch > l_epoch
       OR NOT l_is_epoch_last_block AND epoch = l_epoch;

    DELETE
    FROM validation_reward_summaries
    WHERE epoch > l_epoch
       OR NOT l_is_epoch_last_block AND epoch = l_epoch;

    DELETE FROM pool_size_history WHERE epoch > l_epoch OR NOT l_is_epoch_last_block AND epoch = l_epoch;

    DELETE FROM address_summaries;
    DELETE FROM balances;
    DELETE FROM birthdays;
    DELETE FROM pool_sizes;
    DELETE FROM delegations;
    DELETE FROM pools_summary;

    delete
    from coins
    where block_height > p_block_height;

    delete
    from mem_pool_flip_keys
    where ei_address_state_id in
          (select id
           from address_states
           where block_height > p_block_height);

    delete
    from flips_to_solve
    where ei_address_state_id in
          (select id
           from address_states
           where block_height > p_block_height);

    delete
    from answers
    where ei_address_state_id in
          (select id
           from address_states
           where block_height > p_block_height);

    delete
    from validation_rewards
    where ei_address_state_id in
          (select id
           from address_states
           where block_height > p_block_height);

    delete
    from reward_ages
    where ei_address_state_id in
          (select id
           from address_states
           where block_height > p_block_height);

    delete
    from reward_staked_amounts
    where ei_address_state_id in
          (select id
           from address_states
           where block_height > p_block_height);

    delete
    from bad_authors
    where ei_address_state_id in
          (select id
           from address_states
           where block_height > p_block_height);

    delete
    from good_authors
    where ei_address_state_id in
          (select id
           from address_states
           where block_height > p_block_height);

    delete
    from saved_invite_rewards
    where ei_address_state_id in
          (select id
           from address_states
           where block_height > p_block_height);

    delete
    from epoch_identities
    where address_state_id in
          (select id
           from address_states
           where block_height > p_block_height);

    delete
    from address_states
    where block_height > p_block_height;
    update address_states
    set is_actual = true
    where id in
          (select s.id
           from address_states s
           where (s.address_id, s.block_height) in
                 (select s.address_id, max(s.block_height)
                  from address_states s
                  group by address_id)
             and not s.is_actual);

    delete
    from flip_words
    where tx_id in
          (select t.id
           from transactions t
           where t.block_height > p_block_height);

    delete
    from flips
    where tx_id in
          (select id
           from transactions
           where block_height > p_block_height);
    update flips
    set status_block_height=null,
        status=null,
        answer=null,
        grade=null,
        grade_score=null
    where status_block_height > p_block_height;
    update flips
    set delete_tx_id=null
    where delete_tx_id in (select t.id
                           from transactions t
                           where t.block_height > p_block_height);

    delete
    from flip_keys
    where tx_id in
          (select t.id
           from transactions t
           where t.block_height > p_block_height);

    delete
    from transaction_raws
    where tx_id in
          (select t.id
           from transactions t
           where t.block_height > p_block_height);

    delete
    from activation_txs
    where tx_id in
          (select t.id
           from transactions t
           where t.block_height > p_block_height);

    DELETE FROM latest_activation_txs WHERE activation_tx_id >= l_min_tx_id_to_delete;

    delete
    from kill_invitee_txs
    where tx_id in
          (select t.id
           from transactions t
           where t.block_height > p_block_height);

    delete
    from become_online_txs
    where tx_id in
          (select t.id
           from transactions t
           where t.block_height > p_block_height);

    delete
    from become_offline_txs
    where tx_id in
          (select t.id
           from transactions t
           where t.block_height > p_block_height);

    delete
    from activation_tx_transfers
    where tx_id in
          (select t.id
           from transactions t
           where t.block_height > p_block_height);

    delete
    from kill_tx_transfers
    where tx_id in
          (select t.id
           from transactions t
           where t.block_height > p_block_height);

    delete
    from kill_invitee_tx_transfers
    where tx_id in
          (select t.id
           from transactions t
           where t.block_height > p_block_height);

    DELETE FROM short_answer_txs WHERE tx_id >= l_min_tx_id_to_delete;

    delete
    from rewarded_invitations
    where block_height > p_block_height;

    delete
    from rewarded_invitees
    where block_height > p_block_height;

    delete
    from transactions
    where block_height > p_block_height;

    delete
    from block_proposers
    where block_height > p_block_height;

    delete
    from mining_rewards
    where block_height > p_block_height;

    delete
    from temporary_identities
    where block_height > p_block_height;

    delete
    from addresses
    where block_height > p_block_height;

    delete
    from block_flags
    where block_height > p_block_height;

    delete
    from blocks
    where height > p_block_height;

    delete
    from epochs
    where epoch not in (select distinct epoch from blocks);

    call restore_coins_summary();
    call restore_epoch_summary(p_block_height);
    call restore_address_summaries();

    SET session_replication_role = DEFAULT;
END
$BODY$;

CREATE OR REPLACE PROCEDURE reset_balance_updates_to(p_block_height bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    rec                  record;
    l_history_min_height bigint;
    l_last_block_height  bigint;
BEGIN
    select min(block_height) into l_history_min_height from latest_committee_reward_balance_updates;

    if l_history_min_height is null then
        if (select exists(select 1
                          from balance_updates
                          where block_height <= p_block_height
                            and last_block_height is not null
                            and last_block_height > p_block_height)) then
            raise exception 'there is no committee rewards history to restore balance updates';
        end if;
    end if;

    if p_block_height < l_history_min_height then
        if (select exists(select 1
                          from balance_updates
                          where block_height <= p_block_height
                            and last_block_height is not null
                            and last_block_height > p_block_height)) then
            raise exception 'height to reset is lower than committee rewards history min height';
        end if;
    end if;

    delete
    from latest_committee_reward_balance_updates
    where balance_update_id in (select id from balance_updates where block_height > p_block_height);

    delete from balance_updates where block_height > p_block_height;

    for rec in select block_height,
                      address_id,
                      balance_old,
                      stake_old,
                      penalty_old,
                      penalty_seconds_old,
                      penalty_payment,
                      balance_update_id
               from latest_committee_reward_balance_updates
               where block_height > p_block_height
               order by block_height desc
        loop
            select max(block_height)
            into l_last_block_height
            from latest_committee_reward_balance_updates
            where address_id = rec.address_id
              and block_height < rec.block_height;

            update balance_updates
            set balance_new         = rec.balance_old,
                stake_new           = rec.stake_old,
                penalty_new         = rec.penalty_old,
                penalty_seconds_new = rec.penalty_seconds_old,
                penalty_payment     = null_if_zero(coalesce(penalty_payment, 0) - coalesce(rec.penalty_payment, 0)),
                blocks_count        = blocks_count - 1,
                last_block_height   = l_last_block_height
            where id = rec.balance_update_id;
        end loop;

    delete
    from latest_committee_reward_balance_updates
    where block_height > p_block_height;
END
$$;

CREATE OR REPLACE PROCEDURE log_performance(p_message text,
                                            p_start timestamp,
                                            p_end timestamp)
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    insert into performance_logs (timestamp, message, duration)
    values ((select current_timestamp), p_message, (SELECT EXTRACT(EPOCH FROM (age(p_end, p_start)))));
END
$$;

CREATE OR REPLACE FUNCTION get_address_id_or_insert(p_block_height bigint, p_address text)
    RETURNS bigint
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_address_id bigint;
BEGIN
    if char_length(p_address) = 0 then
        raise exception 'empty address to insert';
    end if;
    SELECT id INTO l_address_id FROM addresses WHERE lower(address) = lower(p_address);
    if l_address_id is null then
        INSERT INTO addresses (address, block_height)
        VALUES (p_address, p_block_height)
        RETURNING id INTO l_address_id;
    end if;
    return l_address_id;
END
$$;

CREATE OR REPLACE PROCEDURE save_restored_data(p_data jsonb)
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    if p_data is null then
        return;
    end if;
    call save_delegations(p_data -> 'delegations');
    call save_pool_sizes(p_data -> 'poolSizes');

    call generate_pools_summary();
END
$$;
