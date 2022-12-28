ALTER TABLE delegatee_validation_rewards
    ADD COLUMN extra_flips_balance numeric(30, 18);
ALTER TABLE delegatee_validation_rewards
    ADD COLUMN invitee1_balance numeric(30, 18);
ALTER TABLE delegatee_validation_rewards
    ADD COLUMN invitee2_balance numeric(30, 18);
ALTER TABLE delegatee_validation_rewards
    ADD COLUMN invitee3_balance numeric(30, 18);

ALTER TABLE delegatee_total_validation_rewards
    ADD COLUMN extra_flips_balance numeric(30, 18);
ALTER TABLE delegatee_total_validation_rewards
    ADD COLUMN invitee1_balance numeric(30, 18);
ALTER TABLE delegatee_total_validation_rewards
    ADD COLUMN invitee2_balance numeric(30, 18);
ALTER TABLE delegatee_total_validation_rewards
    ADD COLUMN invitee3_balance numeric(30, 18);

ALTER TABLE validation_reward_summaries
    ADD COLUMN invitee numeric(30, 18);
ALTER TABLE validation_reward_summaries
    ADD COLUMN invitee_missed numeric(30, 18);
ALTER TABLE validation_reward_summaries
    ADD COLUMN invitee_missed_reason smallint;