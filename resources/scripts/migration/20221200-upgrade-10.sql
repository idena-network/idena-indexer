DROP INDEX sovc_voting_idx;

DROP TYPE tp_refundable_oracle_lock_contract CASCADE;

ALTER TABLE refundable_oracle_lock_contracts
    ADD COLUMN oracle_voting_fee_new integer;

DROP TYPE tp_oracle_voting_contract CASCADE;
ALTER TABLE oracle_voting_contracts
    ADD COLUMN network_size bigint;
ALTER TABLE oracle_voting_contract_call_starts
    ADD COLUMN committee_size bigint;
ALTER TABLE oracle_voting_contract_call_prolongations
    ADD COLUMN committee_size bigint;

DROP TYPE tp_total_epoch_reward CASCADE;

ALTER TABLE total_rewards
    ADD COLUMN flips_extra numeric(30, 18);
ALTER TABLE total_rewards
    ADD COLUMN flips_extra_share numeric(30, 18);

DROP PROCEDURE IF EXISTS save_epoch_result;
DROP PROCEDURE IF EXISTS save_epoch_rewards;
DROP PROCEDURE IF EXISTS save_rewarded_flips;
DROP PROCEDURE IF EXISTS save_reward_staked_amounts;
DROP FUNCTION IF EXISTS calculate_invitations_missed_reward;

ALTER TABLE rewarded_flips
    ADD COLUMN extra boolean;

ALTER TABLE reward_staked_amounts
    ADD COLUMN failed boolean;

ALTER TABLE validation_reward_summaries
    ADD COLUMN extra_flips numeric(30, 18);
ALTER TABLE validation_reward_summaries
    ADD COLUMN extra_flips_missed numeric(30, 18);
ALTER TABLE validation_reward_summaries
    ADD COLUMN extra_flips_missed_reason smallint;

ALTER TABLE penalties
    ADD COLUMN inherited_from_address_id bigint;