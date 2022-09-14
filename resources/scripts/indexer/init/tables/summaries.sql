CREATE TABLE IF NOT EXISTS epoch_summaries
(
    epoch                     bigint          NOT NULL,
    validated_count           integer         NOT NULL,
    block_count               bigint          NOT NULL,
    empty_block_count         bigint          NOT NULL,
    tx_count                  bigint          NOT NULL,
    invite_count              bigint          NOT NULL,
    flip_count                integer         NOT NULL,
    burnt                     numeric(30, 18) NOT NULL,
    minted                    numeric(30, 18) NOT NULL,
    total_balance             numeric(30, 18) NOT NULL,
    total_stake               numeric(30, 18) NOT NULL,
    block_height              bigint          NOT NULL,
    min_score_for_invite      real            NOT NULL,
    flip_lottery_block_height bigint,
    min_tx_id                 bigint,
    max_tx_id                 bigint,
    reported_flips            integer,
    candidate_count           bigint          NOT NULL,
    CONSTRAINT epoch_summaries_pkey PRIMARY KEY (epoch),
    CONSTRAINT epoch_summaries_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT epoch_summaries_epoch_fkey FOREIGN KEY (epoch)
        REFERENCES epochs (epoch) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS flip_summaries
(
    flip_tx_id        bigint   NOT NULL,
    wrong_words_votes smallint NOT NULL,
    short_answers     smallint NOT NULL,
    long_answers      smallint NOT NULL,
    encrypted         boolean  NOT NULL,
    CONSTRAINT flip_summaries_pkey PRIMARY KEY (flip_tx_id),
    CONSTRAINT flip_summaries_flip_tx_id_fkey FOREIGN KEY (flip_tx_id)
        REFERENCES flips (tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS coins_summary
(
    total_burnt  numeric(30, 18),
    total_minted numeric(30, 18)
);

CREATE TABLE IF NOT EXISTS address_summaries
(
    address_id        bigint  NOT NULL,
    flips             integer NOT NULL,
    wrong_words_flips integer NOT NULL,
    CONSTRAINT address_summaries_pkey PRIMARY KEY (address_id),
    CONSTRAINT address_summaries_address_id_fkey FOREIGN KEY (address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS epoch_reward_bounds
(
    epoch          bigint          NOT NULL,
    bound_type     smallint        NOT NULL,
    min_amount     numeric(30, 18) NOT NULL,
    min_address_id bigint          NOT NULL,
    max_amount     numeric(30, 18) NOT NULL,
    max_address_id bigint          NOT NULL,
    CONSTRAINT epoch_reward_bounds_pkey PRIMARY KEY (epoch, bound_type)
);

CREATE TABLE IF NOT EXISTS epoch_flip_statuses
(
    epoch       bigint   NOT NULL,
    flip_status smallint NOT NULL,
    count       integer  NOT NULL,
    CONSTRAINT epoch_flip_statuses_pkey PRIMARY KEY (epoch, flip_status)
);

CREATE TABLE IF NOT EXISTS balance_update_summaries
(
    address_id  bigint          NOT NULL,
    balance_in  numeric(30, 18) NOT NULL,
    balance_out numeric(30, 18) NOT NULL,
    stake_in    numeric(30, 18) NOT NULL,
    stake_out   numeric(30, 18) NOT NULL,
    penalty_in  numeric(30, 18) NOT NULL,
    penalty_out numeric(30, 18) NOT NULL,
    CONSTRAINT balance_update_summaries_pkey PRIMARY KEY (address_id)
);

CREATE TABLE IF NOT EXISTS balance_update_summaries_changes
(
    change_id   bigint NOT NULL,
    address_id  bigint NOT NULL,
    balance_in  numeric(30, 18),
    balance_out numeric(30, 18),
    stake_in    numeric(30, 18),
    stake_out   numeric(30, 18),
    penalty_in  numeric(30, 18),
    penalty_out numeric(30, 18),
    CONSTRAINT balance_update_summaries_changes_pkey PRIMARY KEY (change_id),
    CONSTRAINT balance_update_summaries_changes_change_id_fkey FOREIGN KEY (change_id)
        REFERENCES changes (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS delegatee_total_validation_rewards
(
    epoch                     bigint          NOT NULL,
    delegatee_address_id      bigint          NOT NULL,
    total_balance             numeric(30, 18) NOT NULL,
    validation_balance        numeric(30, 18),
    flips_balance             numeric(30, 18),
    invitations_balance       numeric(30, 18),
    invitations2_balance      numeric(30, 18),
    invitations3_balance      numeric(30, 18),
    saved_invites_balance     numeric(30, 18),
    saved_invites_win_balance numeric(30, 18),
    reports_balance           numeric(30, 18),
    candidate_balance         numeric(30, 18),
    staking_balance           numeric(30, 18),
    delegators                integer         NOT NULL,
    CONSTRAINT delegatee_total_validation_rewards_pkey PRIMARY KEY (epoch, delegatee_address_id)
);
CREATE INDEX IF NOT EXISTS delegatee_total_validation_rewards_api_idx1 on delegatee_total_validation_rewards (epoch, total_balance desc, delegatee_address_id);
CREATE INDEX IF NOT EXISTS delegatee_total_validation_rewards_api_idx2 on delegatee_total_validation_rewards (delegatee_address_id, epoch desc);

CREATE TABLE IF NOT EXISTS delegatee_validation_rewards
(
    epoch                     bigint          NOT NULL,
    delegatee_address_id      bigint          NOT NULL,
    delegator_address_id      bigint          NOT NULL,
    total_balance             numeric(30, 18) NOT NULL,
    validation_balance        numeric(30, 18),
    flips_balance             numeric(30, 18),
    invitations_balance       numeric(30, 18),
    invitations2_balance      numeric(30, 18),
    invitations3_balance      numeric(30, 18),
    saved_invites_balance     numeric(30, 18),
    saved_invites_win_balance numeric(30, 18),
    reports_balance           numeric(30, 18),
    candidate_balance         numeric(30, 18),
    staking_balance           numeric(30, 18),
    CONSTRAINT delegatee_validation_rewards_pkey PRIMARY KEY (epoch, delegatee_address_id, delegator_address_id)
);
CREATE INDEX IF NOT EXISTS delegatee_validation_rewards_api_idx1 on delegatee_validation_rewards (epoch, delegatee_address_id, total_balance desc, delegator_address_id);
CREATE UNIQUE INDEX IF NOT EXISTS delegatee_validation_rewards_api_idx2 on delegatee_validation_rewards (epoch, delegator_address_id);

CREATE TABLE IF NOT EXISTS validation_reward_summaries
(
    epoch                     bigint NOT NULL,
    address_id                bigint NOT NULL,
    validation                numeric(30, 18),
    validation_missed         numeric(30, 18),
    validation_missed_reason  smallint,
    flips                     numeric(30, 18),
    flips_missed              numeric(30, 18),
    flips_missed_reason       smallint,
    invitations               numeric(30, 18),
    invitations_missed        numeric(30, 18),
    invitations_missed_reason smallint,
    reports                   numeric(30, 18),
    reports_missed            numeric(30, 18),
    reports_missed_reason     smallint,
    candidate                 numeric(30, 18),
    candidate_missed          numeric(30, 18),
    candidate_missed_reason   smallint,
    staking                   numeric(30, 18),
    staking_missed            numeric(30, 18),
    staking_missed_reason     smallint,
    CONSTRAINT validation_reward_summaries_pkey PRIMARY KEY (epoch, address_id)
);

CREATE TABLE IF NOT EXISTS mining_reward_summaries
(
    address_id bigint          NOT NULL,
    epoch      smallint        NOT NULL,
    amount     numeric(30, 18) NOT NULL,
    burnt      numeric(30, 18) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS mining_reward_summaries_pkey ON mining_reward_summaries (address_id, epoch desc);

CREATE TABLE IF NOT EXISTS mining_reward_summaries_changes
(
    change_id  bigint   NOT NULL,
    address_id bigint   NOT NULL,
    epoch      smallint NOT NULL,
    amount     numeric(30, 18),
    burnt      numeric(30, 18),
    CONSTRAINT mining_reward_summaries_changes_pkey PRIMARY KEY (change_id),
    CONSTRAINT mining_reward_summaries_changes_change_id_fkey FOREIGN KEY (change_id)
        REFERENCES changes (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);