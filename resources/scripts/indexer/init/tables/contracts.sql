CREATE TABLE IF NOT EXISTS dic_contract_types
(
    id   smallint                                           NOT NULL,
    name character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT dic_contract_types_pkey PRIMARY KEY (id),
    CONSTRAINT dic_contract_types_name_key UNIQUE (name)
);
INSERT INTO dic_contract_types
VALUES (1, 'TimeLock')
ON CONFLICT DO NOTHING;
INSERT INTO dic_contract_types
VALUES (2, 'OracleVoting')
ON CONFLICT DO NOTHING;
INSERT INTO dic_contract_types
VALUES (3, 'OracleLock')
ON CONFLICT DO NOTHING;
INSERT INTO dic_contract_types
VALUES (4, 'Multisig')
ON CONFLICT DO NOTHING;
INSERT INTO dic_contract_types
VALUES (5, 'RefundableOracleLock')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS dic_oracle_voting_contract_states
(
    id   smallint                                           NOT NULL,
    name character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT dic_oracle_voting_contract_states_pkey PRIMARY KEY (id),
    CONSTRAINT dic_oracle_voting_contract_states_name_key UNIQUE (name)
);
INSERT INTO dic_oracle_voting_contract_states
VALUES (0, 'Pending')
ON CONFLICT DO NOTHING;
INSERT INTO dic_oracle_voting_contract_states
VALUES (1, 'Voting')
ON CONFLICT DO NOTHING;
INSERT INTO dic_oracle_voting_contract_states
VALUES (3, 'Counting')
ON CONFLICT DO NOTHING;
INSERT INTO dic_oracle_voting_contract_states
VALUES (2, 'Completed')
ON CONFLICT DO NOTHING;
INSERT INTO dic_oracle_voting_contract_states
VALUES (4, 'Terminated')
ON CONFLICT DO NOTHING;
INSERT INTO dic_oracle_voting_contract_states
VALUES (5, 'Voted')
ON CONFLICT DO NOTHING;
INSERT INTO dic_oracle_voting_contract_states
VALUES (6, 'CanBeProlonged')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS tx_receipts
(
    tx_id     bigint          NOT NULL,
    success   boolean         NOT NULL,
    gas_used  bigint          NOT NULL,
    gas_cost  numeric(30, 18) NOT NULL,
    method    character varying(100),
    error_msg character varying(50),
    CONSTRAINT tx_receipts_pkey PRIMARY KEY (tx_id),
    CONSTRAINT tx_receipts_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS contracts
(
    tx_id               bigint          NOT NULL,
    contract_address_id bigint          NOT NULL,
    type                smallint        NOT NULL,
    stake               numeric(30, 18) NOT NULL,
    CONSTRAINT contracts_pkey PRIMARY KEY (tx_id),
    CONSTRAINT contracts_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT contracts_type_fkey FOREIGN KEY (type)
        REFERENCES dic_contract_types (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT contracts_contract_address_id_fkey FOREIGN KEY (contract_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE UNIQUE INDEX IF NOT EXISTS contracts_oracle_voting_address_id_idx ON contracts (contract_address_id);

CREATE TABLE IF NOT EXISTS tx_receipts
(
    tx_id     bigint          NOT NULL,
    success   boolean         NOT NULL,
    gas_used  bigint          NOT NULL,
    gas_cost  numeric(30, 18) NOT NULL,
    error_msg character varying(50),
    CONSTRAINT tx_receipts_pkey PRIMARY KEY (tx_id),
    CONSTRAINT tx_receipts_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS oracle_voting_contracts
(
    contract_tx_id              bigint   NOT NULL,
    start_time                  bigint   NOT NULL,
    voting_duration             bigint   NOT NULL,
    voting_min_payment          numeric(48, 18),
    fact                        bytea,
    public_voting_duration      bigint   NOT NULL,
    winner_threshold            smallint NOT NULL,
    quorum                      smallint NOT NULL,
    committee_size              bigint   NOT NULL,
    owner_fee                   smallint NOT NULL,
    state                       smallint NOT NULL,
    owner_deposit               numeric(48, 18),
    oracle_reward_fund          numeric(48, 18),
    refund_recipient_address_id bigint,
    CONSTRAINT fec_pkey PRIMARY KEY (contract_tx_id),
    CONSTRAINT fec_contract_tx_id_fkey FOREIGN KEY (contract_tx_id)
        REFERENCES contracts (tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS oracle_voting_contract_call_starts
(
    call_tx_id         bigint   NOT NULL,
    ov_contract_tx_id  bigint   NOT NULL,
    start_block_height bigint   NOT NULL,
    epoch              bigint   NOT NULL,
    voting_min_payment numeric(48, 18),
    vrf_seed           bytea,
    state              smallint NOT NULL,
    CONSTRAINT oracle_voting_contract_call_starts_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT oracle_voting_contract_call_starts_call_tx_id_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT oracle_voting_contract_call_starts_contract_tx_id_fkey FOREIGN KEY (ov_contract_tx_id)
        REFERENCES oracle_voting_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE INDEX IF NOT EXISTS oracle_voting_contract_call_starts_ov_contract_tx_id_idx ON oracle_voting_contract_call_starts (ov_contract_tx_id);

CREATE TABLE IF NOT EXISTS oracle_voting_contract_call_vote_proofs
(
    call_tx_id         bigint NOT NULL,
    ov_contract_tx_id  bigint NOT NULL,
    address_id         bigint NOT NULL,
    vote_hash          bytea,
    secret_votes_count bigint,
    discriminated      boolean,
    CONSTRAINT oracle_voting_contract_call_vote_proofs_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT oracle_voting_contract_call_vote_proofs_call_tx_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT oracle_voting_contract_call_vote_proofs_contract_tx_id_fkey FOREIGN KEY (ov_contract_tx_id)
        REFERENCES oracle_voting_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT oracle_voting_contract_call_vote_proofs_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE INDEX IF NOT EXISTS oracle_voting_contract_call_vote_proofs_contract_address_idx ON oracle_voting_contract_call_vote_proofs (ov_contract_tx_id, address_id);

CREATE TABLE IF NOT EXISTS oracle_voting_contract_call_votes
(
    call_tx_id           bigint   NOT NULL,
    ov_contract_tx_id    bigint   NOT NULL,
    vote                 smallint NOT NULL,
    salt                 bytea,
    option_votes         bigint,
    option_all_votes     bigint,
    secret_votes_count   bigint,
    delegatee_address_id bigint,
    prev_pool_vote       smallint,
    prev_option_votes    bigint,
    discriminated        boolean,
    CONSTRAINT oracle_voting_contract_call_votes_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT oracle_voting_contract_call_votes_call_tx_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT oracle_voting_contract_call_votes_contract_tx_id_fkey FOREIGN KEY (ov_contract_tx_id)
        REFERENCES oracle_voting_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS oracle_voting_contract_call_finishes
(
    call_tx_id        bigint          NOT NULL,
    ov_contract_tx_id bigint          NOT NULL,
    result            smallint,
    fund              numeric(30, 18) NOT NULL,
    oracle_reward     numeric(30, 18) NOT NULL,
    owner_reward      numeric(30, 18) NOT NULL,
    state             smallint        NOT NULL,
    CONSTRAINT oracle_voting_contract_call_finishes_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT oracle_voting_contract_call_finishes_call_tx_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT oracle_voting_contract_call_finishes_contract_tx_id_fkey FOREIGN KEY (ov_contract_tx_id)
        REFERENCES oracle_voting_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS oracle_voting_contract_call_prolongations
(
    call_tx_id           bigint NOT NULL,
    ov_contract_tx_id    bigint NOT NULL,
    epoch                bigint,
    start_block          bigint,
    vrf_seed             bytea,
    epoch_without_growth smallint,
    prolong_vote_count   bigint,
    CONSTRAINT oracle_voting_contract_call_prolongations_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT oracle_voting_contract_call_prolongations_call_tx_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT oracle_voting_contract_call_prolongations_contract_tx_id_fkey FOREIGN KEY (ov_contract_tx_id)
        REFERENCES oracle_voting_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS oracle_voting_contract_call_add_stakes
(
    call_tx_id        bigint NOT NULL,
    ov_contract_tx_id bigint NOT NULL,
    CONSTRAINT oracle_voting_contract_call_add_stakes_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT oracle_voting_contract_call_add_stakes_call_tx_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT oracle_voting_contract_call_add_stakes_contract_tx_id_fkey FOREIGN KEY (ov_contract_tx_id)
        REFERENCES oracle_voting_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS oracle_voting_contract_terminations
(
    termination_tx_id bigint NOT NULL,
    ov_contract_tx_id bigint NOT NULL,
    fund              numeric(30, 18),
    oracle_reward     numeric(30, 18),
    owner_reward      numeric(30, 18),
    CONSTRAINT oracle_voting_contract_terminations_pkey PRIMARY KEY (termination_tx_id),
    CONSTRAINT oracle_voting_contract_terminations_termination_tx_fkey FOREIGN KEY (termination_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT oracle_voting_contract_terminations_contract_tx_id_fkey FOREIGN KEY (ov_contract_tx_id)
        REFERENCES oracle_voting_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE INDEX IF NOT EXISTS oracle_voting_contract_terminations_api_idx ON oracle_voting_contract_terminations (ov_contract_tx_id);

CREATE TABLE IF NOT EXISTS oracle_lock_contracts
(
    contract_tx_id           bigint   NOT NULL,
    oracle_voting_address_id bigint   NOT NULL,
    value                    smallint NOT NULL,
    success_address_id       bigint   NOT NULL,
    fail_address_id          bigint   NOT NULL,
    CONSTRAINT oracle_lock_contracts_pkey PRIMARY KEY (contract_tx_id),
    CONSTRAINT oracle_lock_contracts_contract_tx_id_fkey FOREIGN KEY (contract_tx_id)
        REFERENCES contracts (tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT oracle_lock_contracts_oracle_voting_address_id_fkey FOREIGN KEY (oracle_voting_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT oracle_lock_contracts_success_address_id_fkey FOREIGN KEY (success_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT oracle_lock_contracts_fail_address_id_fkey FOREIGN KEY (fail_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS oracle_lock_contract_call_check_oracle_votings
(
    call_tx_id           bigint NOT NULL,
    ol_contract_tx_id    bigint NOT NULL,
    oracle_voting_result smallint,
    CONSTRAINT olcccov_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT olcccov_call_tx_id_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT olcccov_ol_contract_tx_id_fkey FOREIGN KEY (ol_contract_tx_id)
        REFERENCES oracle_lock_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS oracle_lock_contract_call_pushes
(
    call_tx_id           bigint          NOT NULL,
    ol_contract_tx_id    bigint          NOT NULL,
    success              boolean         NOT NULL,
    oracle_voting_result smallint        NOT NULL,
    transfer             numeric(30, 18) NOT NULL,
    CONSTRAINT oracle_lock_contract_call_pushes_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT oracle_lock_contract_call_pushes_call_tx_id_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT oracle_lock_contract_call_pushes_ol_contract_tx_id_fkey FOREIGN KEY (ol_contract_tx_id)
        REFERENCES oracle_lock_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS oracle_lock_contract_terminations
(
    termination_tx_id bigint NOT NULL,
    ol_contract_tx_id bigint NOT NULL,
    dest_address_id   bigint NOT NULL,
    CONSTRAINT oracle_lock_contract_terminations_terminations_pkey PRIMARY KEY (termination_tx_id),
    CONSTRAINT oracle_lock_contract_terminations_termination_tx_fkey FOREIGN KEY (termination_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT oracle_lock_contract_terminations_ol_contract_tx_id_fkey FOREIGN KEY (ol_contract_tx_id)
        REFERENCES oracle_lock_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT oracle_lock_contract_terminations_dest_address_id_fkey FOREIGN KEY (dest_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE INDEX IF NOT EXISTS oracle_lock_contract_terminations_api_idx ON oracle_lock_contract_terminations (ol_contract_tx_id);

CREATE TABLE IF NOT EXISTS refundable_oracle_lock_contracts
(
    contract_tx_id           bigint   NOT NULL,
    oracle_voting_address_id bigint   NOT NULL,
    value                    smallint NOT NULL,
    success_address_id       bigint,
    fail_address_id          bigint,
    refund_delay             bigint   NOT NULL,
    deposit_deadline         bigint   NOT NULL,
    oracle_voting_fee        smallint NOT NULL,
    CONSTRAINT refundable_oracle_lock_contracts_pkey PRIMARY KEY (contract_tx_id),
    CONSTRAINT refundable_oracle_lock_contracts_contract_tx_id_fkey FOREIGN KEY (contract_tx_id)
        REFERENCES contracts (tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT refundable_oracle_lock_contracts_oracle_voting_address_id_fkey FOREIGN KEY (oracle_voting_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT refundable_oracle_lock_contracts_success_address_id_fkey FOREIGN KEY (success_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT refundable_oracle_lock_contracts_fail_address_id_fkey FOREIGN KEY (fail_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS refundable_oracle_lock_contract_call_deposits
(
    call_tx_id        bigint          NOT NULL,
    ol_contract_tx_id bigint          NOT NULL,
    own_sum           numeric(30, 18) NOT NULL,
    sum               numeric(30, 18) NOT NULL,
    fee               numeric(30, 18) NOT NULL,
    CONSTRAINT rol_contract_call_deposits_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT rol_contract_call_deposits_call_tx_id_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT rol_contract_call_deposits_ol_contract_tx_id_fkey FOREIGN KEY (ol_contract_tx_id)
        REFERENCES refundable_oracle_lock_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS refundable_oracle_lock_contract_call_pushes
(
    call_tx_id           bigint          NOT NULL,
    ol_contract_tx_id    bigint          NOT NULL,
    oracle_voting_exists boolean         NOT NULL,
    oracle_voting_result smallint,
    transfer             numeric(30, 18) NOT NULL,
    refund_block         bigint,
    CONSTRAINT rol_contract_call_pushes_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT rol_contract_call_pushes_call_tx_id_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT rol_contract_call_pushes_ol_contract_tx_id_fkey FOREIGN KEY (ol_contract_tx_id)
        REFERENCES refundable_oracle_lock_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS refundable_oracle_lock_contract_call_refunds
(
    call_tx_id        bigint           NOT NULL,
    ol_contract_tx_id bigint           NOT NULL,
    balance           numeric(30, 18)  NOT NULL,
    coef              double precision NOT NULL,
    CONSTRAINT rol_contract_call_refunds_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT rol_contract_call_refunds_call_tx_id_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT rol_contract_call_refunds_ol_contract_tx_id_fkey FOREIGN KEY (ol_contract_tx_id)
        REFERENCES refundable_oracle_lock_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS refundable_oracle_lock_contract_terminations
(
    termination_tx_id bigint NOT NULL,
    ol_contract_tx_id bigint NOT NULL,
    dest_address_id   bigint NOT NULL,
    CONSTRAINT rol_contract_terminations_pkey PRIMARY KEY (termination_tx_id),
    CONSTRAINT rol_contract_terminations_termination_tx_fkey FOREIGN KEY (termination_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT rol_contract_terminations_ol_contract_tx_id_fkey FOREIGN KEY (ol_contract_tx_id)
        REFERENCES refundable_oracle_lock_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT rol_contract_terminations_dest_address_id_fkey FOREIGN KEY (dest_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE INDEX IF NOT EXISTS rol_contract_terminations_api_idx ON refundable_oracle_lock_contract_terminations (ol_contract_tx_id);

CREATE TABLE IF NOT EXISTS time_lock_contracts
(
    contract_tx_id bigint NOT NULL,
    "timestamp"    bigint NOT NULL,
    CONSTRAINT time_lock_contracts_pkey PRIMARY KEY (contract_tx_id),
    CONSTRAINT time_lock_contracts_contract_tx_id_fkey FOREIGN KEY (contract_tx_id)
        REFERENCES contracts (tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS time_lock_contract_call_transfers
(
    call_tx_id        bigint          NOT NULL,
    tl_contract_tx_id bigint          NOT NULL,
    dest_address_id   bigint          NOT NULL,
    amount            numeric(30, 18) NOT NULL,
    CONSTRAINT time_lock_contract_call_transfers_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT time_lock_contract_call_transfers_call_tx_id_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT time_lock_contract_call_transfers_tl_contract_tx_id_fkey FOREIGN KEY (tl_contract_tx_id)
        REFERENCES time_lock_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT time_lock_contract_call_transfers_dest_address_id_fkey FOREIGN KEY (dest_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS time_lock_contract_terminations
(
    termination_tx_id bigint NOT NULL,
    tl_contract_tx_id bigint NOT NULL,
    dest_address_id   bigint NOT NULL,
    CONSTRAINT time_lock_contract_terminations_pkey PRIMARY KEY (termination_tx_id),
    CONSTRAINT time_lock_contract_terminations_termination_tx_fkey FOREIGN KEY (termination_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT time_lock_contract_terminations_tl_contract_tx_id_fkey FOREIGN KEY (tl_contract_tx_id)
        REFERENCES time_lock_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT time_lock_contract_terminations_dest_address_id_fkey FOREIGN KEY (dest_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE INDEX IF NOT EXISTS time_lock_contract_terminations_api_idx ON time_lock_contract_terminations (tl_contract_tx_id);

CREATE TABLE IF NOT EXISTS multisig_contracts
(
    contract_tx_id bigint   NOT NULL,
    min_votes      smallint NOT NULL,
    max_votes      smallint NOT NULL,
    state          smallint NOT NULL,
    CONSTRAINT multisig_contracts_pkey PRIMARY KEY (contract_tx_id),
    CONSTRAINT multisig_contracts_contract_tx_id_fkey FOREIGN KEY (contract_tx_id)
        REFERENCES contracts (tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS multisig_contract_call_adds
(
    call_tx_id        bigint NOT NULL,
    ms_contract_tx_id bigint NOT NULL,
    address_id        bigint NOT NULL,
    new_state         smallint,
    CONSTRAINT multisig_contract_call_adds_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT multisig_contract_call_adds_call_tx_id_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT multisig_contract_call_adds_ms_contract_tx_id_fkey FOREIGN KEY (ms_contract_tx_id)
        REFERENCES multisig_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT multisig_contract_call_adds_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS multisig_contract_call_sends
(
    call_tx_id        bigint          NOT NULL,
    ms_contract_tx_id bigint          NOT NULL,
    dest_address_id   bigint          NOT NULL,
    amount            numeric(30, 18) NOT NULL,
    CONSTRAINT multisig_contract_call_sends_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT multisig_contract_call_sends_call_tx_id_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT multisig_contract_call_sends_ms_contract_tx_id_fkey FOREIGN KEY (ms_contract_tx_id)
        REFERENCES multisig_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT multisig_contract_call_sends_dest_address_id_fkey FOREIGN KEY (dest_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS multisig_contract_call_pushes
(
    call_tx_id        bigint          NOT NULL,
    ms_contract_tx_id bigint          NOT NULL,
    dest_address_id   bigint          NOT NULL,
    amount            numeric(30, 18) NOT NULL,
    vote_address_cnt  smallint        NOT NULL,
    vote_amount_cnt   smallint        NOT NULL,
    CONSTRAINT multisig_contract_call_pushes_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT multisig_contract_call_pushes_call_tx_id_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT multisig_contract_call_pushes_ms_contract_tx_id_fkey FOREIGN KEY (ms_contract_tx_id)
        REFERENCES multisig_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT multisig_contract_call_pushes_dest_address_id_fkey FOREIGN KEY (dest_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS multisig_contract_terminations
(
    termination_tx_id bigint NOT NULL,
    ms_contract_tx_id bigint NOT NULL,
    dest_address_id   bigint NOT NULL,
    CONSTRAINT multisig_contract_terminations_pkey PRIMARY KEY (termination_tx_id),
    CONSTRAINT multisig_contract_terminations_termination_tx_fkey FOREIGN KEY (termination_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT multisig_contract_terminations_ms_contract_tx_id_fkey FOREIGN KEY (ms_contract_tx_id)
        REFERENCES multisig_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT multisig_contract_terminations_dest_address_id_fkey FOREIGN KEY (dest_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE INDEX IF NOT EXISTS multisig_contract_terminations_api_idx ON multisig_contract_terminations (ms_contract_tx_id);

--------------------- CONTRACT SUMMARIES -------------------------

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

CREATE TABLE IF NOT EXISTS sorted_oracle_voting_contracts
(
    contract_tx_id    bigint   NOT NULL,
    author_address_id bigint   NOT NULL,
    sort_key          character(68),
    state             smallint NOT NULL,
    state_tx_id       bigint   NOT NULL,
    counting_block    bigint,
    epoch             bigint,
    CONSTRAINT sovc_pkey PRIMARY KEY (contract_tx_id),
    CONSTRAINT sovc_contract_tx_id_fkey FOREIGN KEY (contract_tx_id)
        REFERENCES oracle_voting_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT sovc_author_address_id_fkey FOREIGN KEY (author_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT sovc_state_fkey FOREIGN KEY (state)
        REFERENCES dic_oracle_voting_contract_states (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE INDEX IF NOT EXISTS sovc_voting_idx ON sorted_oracle_voting_contracts (counting_block) WHERE state = 1;
CREATE INDEX IF NOT EXISTS ovc_api_idx_1 ON sorted_oracle_voting_contracts (author_address_id, sort_key desc, state) where state in (0, 1);
CREATE INDEX IF NOT EXISTS ovc_api_idx_2 ON sorted_oracle_voting_contracts (author_address_id, state_tx_id desc, state);
CREATE INDEX IF NOT EXISTS ovc_api_idx_3 ON sorted_oracle_voting_contracts (sort_key desc, state) where state in (0, 1);
CREATE INDEX IF NOT EXISTS ovc_api_idx_4 ON sorted_oracle_voting_contracts (state_tx_id desc, state);

CREATE TABLE IF NOT EXISTS sorted_oracle_voting_contracts_changes
(
    change_id      bigint NOT NULL,
    contract_tx_id bigint NOT NULL,
    sort_key       character(68),
    state          smallint,
    state_tx_id    bigint,
    counting_block bigint,
    epoch          bigint,
    CONSTRAINT sorted_oracle_voting_contracts_changes_pkey PRIMARY KEY (change_id),
    CONSTRAINT sorted_oracle_voting_contracts_changes_change_id_fkey FOREIGN KEY (change_id)
        REFERENCES changes (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sorted_oracle_voting_contract_committees
(
    contract_tx_id    bigint   NOT NULL,
    author_address_id bigint   NOT NULL,
    sort_key          character(68),
    state             smallint NOT NULL,
    state_tx_id       bigint   NOT NULL,
    address_id        bigint   NOT NULL,
    voted             boolean  NOT NULL,
    CONSTRAINT sovcc_pkey PRIMARY KEY (contract_tx_id, address_id),
    CONSTRAINT sovcc_contract_tx_id_fkey FOREIGN KEY (contract_tx_id)
        REFERENCES sorted_oracle_voting_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT sovcc_author_address_id_fkey FOREIGN KEY (author_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT sovcc_state_fkey FOREIGN KEY (state)
        REFERENCES dic_oracle_voting_contract_states (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT sovcc_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE INDEX IF NOT EXISTS sovcc_not_voted_oracles_idx ON sorted_oracle_voting_contract_committees (contract_tx_id) WHERE state = 1;
CREATE INDEX IF NOT EXISTS ovcc_api_idx_1 ON sorted_oracle_voting_contract_committees (address_id, author_address_id, sort_key desc, state) where state in (0, 1);
CREATE INDEX IF NOT EXISTS ovcc_api_idx_2 ON sorted_oracle_voting_contract_committees (address_id, author_address_id, state_tx_id desc, state);
CREATE INDEX IF NOT EXISTS ovcc_api_idx_3 ON sorted_oracle_voting_contract_committees (address_id, sort_key desc, state) where state in (0, 1);
CREATE INDEX IF NOT EXISTS ovcc_api_idx_4 ON sorted_oracle_voting_contract_committees (address_id, state_tx_id desc, state);

CREATE TABLE IF NOT EXISTS sorted_oracle_voting_contract_committees_changes
(
    change_id         bigint NOT NULL,
    contract_tx_id    bigint NOT NULL,
    author_address_id bigint NOT NULL,
    address_id        bigint NOT NULL,
    sort_key          character(68),
    state             smallint,
    state_tx_id       bigint,
    voted             boolean,
    deleted           boolean,
    CONSTRAINT sorted_oracle_voting_contract_committees_changes_change_id_fkey FOREIGN KEY (change_id)
        REFERENCES changes (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS sorted_oracle_voting_contract_committees_changes_change_id_idx ON sorted_oracle_voting_contract_committees_changes (change_id);

CREATE TABLE IF NOT EXISTS oracle_voting_contract_summaries
(
    contract_tx_id        bigint          NOT NULL,
    vote_proofs           bigint          NOT NULL,
    votes                 bigint          NOT NULL,
    finish_timestamp      bigint,
    termination_timestamp bigint,
    total_reward          numeric(30, 18),
    stake                 numeric(30, 18) NOT NULL,
    secret_votes_count    bigint,
    epoch_without_growth  smallint,
    CONSTRAINT oracle_voting_contract_summaries_pkey PRIMARY KEY (contract_tx_id),
    CONSTRAINT oracle_voting_contract_summaries_contract_tx_id_fkey FOREIGN KEY (contract_tx_id)
        REFERENCES oracle_voting_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS oracle_voting_contract_summaries_changes
(
    change_id             bigint NOT NULL,
    contract_tx_id        bigint NOT NULL,
    vote_proofs           bigint NOT NULL,
    votes                 bigint NOT NULL,
    finish_timestamp      bigint,
    termination_timestamp bigint,
    total_reward          numeric(30, 18),
    stake                 numeric(30, 18),
    secret_votes_count    bigint,
    epoch_without_growth  smallint,
    CONSTRAINT oracle_voting_contract_summaries_changes_pkey PRIMARY KEY (change_id),
    CONSTRAINT oracle_voting_contract_summaries_changes_change_id_fkey FOREIGN KEY (change_id)
        REFERENCES changes (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS oracle_voting_contract_results
(
    contract_tx_id  bigint   NOT NULL,
    option          smallint NOT NULL,
    votes_count     bigint   NOT NULL,
    all_votes_count bigint,
    CONSTRAINT oracle_voting_contract_results_pkey PRIMARY KEY (contract_tx_id, option),
    CONSTRAINT oracle_voting_contract_results_contract_tx_id_fkey FOREIGN KEY (contract_tx_id)
        REFERENCES oracle_voting_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS oracle_voting_contract_results_changes
(
    change_id       bigint   NOT NULL,
    contract_tx_id  bigint   NOT NULL,
    option          smallint NOT NULL,
    votes_count     bigint   NOT NULL,
    all_votes_count bigint,
    CONSTRAINT oracle_voting_contract_results_changes_pkey PRIMARY KEY (change_id),
    CONSTRAINT oracle_voting_contract_results_changes_change_id_fkey FOREIGN KEY (change_id)
        REFERENCES changes (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS oracle_voting_contract_authors_and_open_voters
(
    deploy_or_vote_tx_id bigint NOT NULL,
    address_id           bigint NOT NULL,
    contract_tx_id       bigint NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS ovc_authors_and_open_voters_pkey ON oracle_voting_contract_authors_and_open_voters (deploy_or_vote_tx_id);
CREATE UNIQUE INDEX IF NOT EXISTS ovc_authors_and_open_voters_ukey ON oracle_voting_contract_authors_and_open_voters (address_id, contract_tx_id);
CREATE INDEX IF NOT EXISTS ovc_authors_and_open_voters_api ON oracle_voting_contract_authors_and_open_voters (address_id, contract_tx_id, deploy_or_vote_tx_id desc);