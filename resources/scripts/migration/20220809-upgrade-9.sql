ALTER TABLE penalties
    ADD COLUMN penalty_seconds smallint;

DROP TYPE tp_mining_reward CASCADE;
DROP TYPE tp_balance_update CASCADE;

ALTER TABLE mining_rewards
    ADD COLUMN stake_weight double precision;

ALTER TABLE balance_updates
    ADD COLUMN penalty_seconds_old smallint;
ALTER TABLE balance_updates
    ADD COLUMN penalty_seconds_new smallint;
ALTER TABLE balance_updates
    ADD COLUMN penalty_payment numeric(30, 18);

ALTER TABLE latest_committee_reward_balance_updates
    ADD COLUMN penalty_seconds_old smallint;
ALTER TABLE latest_committee_reward_balance_updates
    ADD COLUMN penalty_seconds_new smallint;
ALTER TABLE latest_committee_reward_balance_updates
    ADD COLUMN penalty_payment numeric(30, 18);