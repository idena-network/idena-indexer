CREATE TABLE IF NOT EXISTS pools_summary
(
    count bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS pool_sizes
(
    address_id bigint NOT NULL,
    size       bigint NOT NULL,
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