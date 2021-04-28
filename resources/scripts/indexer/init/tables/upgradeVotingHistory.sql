CREATE TABLE IF NOT EXISTS upgrades
(
    upgrade               smallint NOT NULL,
    start_activation_date bigint   NOT NULL,
    end_activation_date   bigint   NOT NULL,
    CONSTRAINT upgrades_pkey PRIMARY KEY (upgrade)
);

CREATE TABLE IF NOT EXISTS upgrade_voting_history
(
    block_height bigint   NOT NULL,
    upgrade      smallint NOT NULL,
    votes        bigint   NOT NULL,
    "timestamp"  bigint   NOT NULL,
    CONSTRAINT upgrade_voting_history_pkey PRIMARY KEY (upgrade, block_height)
);

CREATE TABLE IF NOT EXISTS upgrade_voting_history_summary
(
    upgrade smallint NOT NULL,
    items   bigint   NOT NULL,
    CONSTRAINT upgrade_voting_history_summary_pkey PRIMARY KEY (upgrade)
);

CREATE TABLE IF NOT EXISTS upgrade_voting_short_history
(
    block_height bigint   NOT NULL,
    upgrade      smallint NOT NULL,
    votes        bigint   NOT NULL,
    CONSTRAINT upgrade_voting_short_history_pkey PRIMARY KEY (upgrade, block_height)
);

CREATE TABLE IF NOT EXISTS upgrade_voting_short_history_summary
(
    upgrade     smallint NOT NULL,
    last_height bigint   NOT NULL,
    last_step   integer  NOT NULL,
    CONSTRAINT upgrade_voting_short_history_summary_pkey PRIMARY KEY (upgrade)
);