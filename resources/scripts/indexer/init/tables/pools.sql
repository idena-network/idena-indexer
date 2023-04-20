CREATE TABLE IF NOT EXISTS pools_summary
(
    count bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS pool_sizes
(
    address_id      bigint NOT NULL,
    size            bigint NOT NULL,
    total_delegated bigint NOT NULL,
    CONSTRAINT pool_sizes_pkey PRIMARY KEY (address_id)
);
CREATE INDEX IF NOT EXISTS pool_sizes_pools_api_idx on pool_sizes (size desc, address_id);

CREATE TABLE IF NOT EXISTS delegations
(
    delegator_address_id bigint NOT NULL,
    delegatee_address_id bigint NOT NULL,
    birth_epoch          integer
);
CREATE INDEX IF NOT EXISTS delegations_pool_delegators_api_idx on delegations (delegatee_address_id,
                                                                               coalesce(birth_epoch, 9999),
                                                                               delegator_address_id);

CREATE TABLE IF NOT EXISTS removed_transitive_delegations
(
    epoch                integer NOT NULL,
    delegator_address_id bigint  NOT NULL,
    delegatee_address_id bigint  NOT NULL
);

CREATE TABLE IF NOT EXISTS dic_undelegation_reasons
(
    id   smallint              NOT NULL,
    name character varying(30) NOT NULL,
    CONSTRAINT dic_uindelegation_reasons_pkey PRIMARY KEY (id),
    CONSTRAINT dic_uindelegation_reasons_name_key UNIQUE (name)
);

INSERT INTO dic_undelegation_reasons
VALUES (1, 'Undelegation')
ON CONFLICT DO NOTHING;
INSERT INTO dic_undelegation_reasons
VALUES (2, 'Termination')
ON CONFLICT DO NOTHING;
INSERT INTO dic_undelegation_reasons
VALUES (3, 'ValidationFailure')
ON CONFLICT DO NOTHING;
INSERT INTO dic_undelegation_reasons
VALUES (4, 'TransitiveDelegationRemove')
ON CONFLICT DO NOTHING;
INSERT INTO dic_undelegation_reasons
VALUES (5, 'InactiveIdentity')
ON CONFLICT DO NOTHING;
INSERT INTO dic_undelegation_reasons
VALUES (6, 'ApplyingFailure')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS delegation_history
(
    delegator_address_id      bigint NOT NULL,
    delegation_tx_id          bigint NOT NULL,
    delegation_block_height   bigint,
    undelegation_reason       smallint,
    undelegation_tx_id        bigint,
    undelegation_block_height integer,
    is_actual                 boolean
);
CREATE INDEX IF NOT EXISTS delegation_history_api_idx ON delegation_history (delegator_address_id, delegation_tx_id desc);
CREATE UNIQUE INDEX IF NOT EXISTS delegation_history_api_actual_idx ON delegation_history (delegator_address_id) WHERE is_actual;

CREATE TABLE IF NOT EXISTS delegation_history_changes
(
    change_id                 bigint NOT NULL,
    delegator_address_id      bigint NOT NULL,
    delegation_tx_id          bigint NOT NULL,
    delegation_block_height   bigint,
    undelegation_reason       smallint,
    undelegation_tx_id        bigint,
    undelegation_block_height integer,
    is_actual                 boolean,
    CONSTRAINT delegation_history_changes_change_id_fkey FOREIGN KEY (change_id)
        REFERENCES changes (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS delegation_history_changes_pkey ON delegation_history_changes (change_id);