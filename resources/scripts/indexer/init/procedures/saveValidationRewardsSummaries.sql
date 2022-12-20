CREATE OR REPLACE PROCEDURE save_validation_rewards_summaries(p_epoch bigint,
                                                              p_block_height bigint,
                                                              p_items jsonb,
                                                              p_invitations_share numeric)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    MISSED_REWARD_REASON_NOT_ALL_INVITES CONSTANT smallint = 6;
    l_item                                        jsonb;
    l_address_id                                  bigint;
    l_validation_reward_summary                   jsonb;
    l_candidate_reward_summary                    jsonb;
    l_staking_reward_summary                      jsonb;
    l_flips_reward_summary                        jsonb;
    l_extra_flips_reward_summary                  jsonb;
    l_invitations_reward_summary                  jsonb;
    l_reports_reward_summary                      jsonb;
    l_invitations                                 numeric;
    l_invitations_missed                          numeric;
    l_invitations_missed_reason                   smallint;
    l_enable_upgrade_10                           boolean;
BEGIN
    if p_items is null then
        return;
    end if;
    l_enable_upgrade_10 = (select exists(select * from blocks where coalesce(upgrade, 0) > 0 and upgrade = 10));
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;

            l_address_id = get_address_id_or_insert(p_block_height, (l_item ->> 'address')::text);

            l_validation_reward_summary = l_item -> 'validation';
            l_candidate_reward_summary = l_item -> 'candidate';
            l_staking_reward_summary = l_item -> 'staking';
            l_flips_reward_summary = l_item -> 'flips';
            l_extra_flips_reward_summary = l_item -> 'extraFlips';
            l_invitations_reward_summary = l_item -> 'invitations';
            l_reports_reward_summary = l_item -> 'reports';

            l_invitations = (l_invitations_reward_summary ->> 'earned')::numeric;
            l_invitations_missed_reason = (l_invitations_reward_summary ->> 'missedReason')::smallint;
            if l_enable_upgrade_10 then
                l_invitations_missed =
                        calculate_invitations_missed_reward(p_epoch, l_address_id, l_invitations, p_invitations_share,
                                                            (l_item ->> 'prevStake')::numeric);
            else
                l_invitations_missed =
                        calculate_invitations_missed_reward_old(p_epoch, l_address_id, l_invitations, p_invitations_share);
            end if;
            if l_invitations_missed_reason is not null and l_invitations_missed is null or l_invitations_missed = 0 then
                l_invitations_missed_reason = null;
            end if;
            if l_invitations_missed_reason is null and l_invitations_missed > 0 then
                l_invitations_missed_reason = MISSED_REWARD_REASON_NOT_ALL_INVITES;
            end if;

            INSERT INTO validation_reward_summaries (epoch, address_id, validation, validation_missed,
                                                     validation_missed_reason, flips, flips_missed, flips_missed_reason,
                                                     invitations, invitations_missed, invitations_missed_reason,
                                                     reports, reports_missed, reports_missed_reason,
                                                     candidate, candidate_missed, candidate_missed_reason,
                                                     staking, staking_missed, staking_missed_reason,
                                                     extra_flips, extra_flips_missed, extra_flips_missed_reason)
            VALUES (p_epoch, l_address_id,
                    null_if_zero((l_validation_reward_summary ->> 'earned')::numeric),
                    null_if_zero((l_validation_reward_summary ->> 'missed')::numeric),
                    (l_validation_reward_summary ->> 'missedReason')::smallint,
                    null_if_zero((l_flips_reward_summary ->> 'earned')::numeric),
                    null_if_zero((l_flips_reward_summary ->> 'missed')::numeric),
                    (l_flips_reward_summary ->> 'missedReason')::smallint,
                    null_if_zero(l_invitations),
                    null_if_zero(l_invitations_missed),
                    l_invitations_missed_reason,
                    null_if_zero((l_reports_reward_summary ->> 'earned')::numeric),
                    null_if_zero((l_reports_reward_summary ->> 'missed')::numeric),
                    (l_reports_reward_summary ->> 'missedReason')::smallint,
                    null_if_zero((l_candidate_reward_summary ->> 'earned')::numeric),
                    null_if_zero((l_candidate_reward_summary ->> 'missed')::numeric),
                    (l_candidate_reward_summary ->> 'missedReason')::smallint,
                    null_if_zero((l_staking_reward_summary ->> 'earned')::numeric),
                    null_if_zero((l_staking_reward_summary ->> 'missed')::numeric),
                    (l_staking_reward_summary ->> 'missedReason')::smallint,
                    null_if_zero((l_extra_flips_reward_summary ->> 'earned')::numeric),
                    null_if_zero((l_extra_flips_reward_summary ->> 'missed')::numeric),
                    (l_extra_flips_reward_summary ->> 'missedReason')::smallint);

        end loop;
END
$$;

CREATE OR REPLACE FUNCTION calculate_invitations_missed_reward_old(p_epoch bigint,
                                                                   p_address_id bigint,
                                                                   p_reward numeric,
                                                                   p_reward_share numeric)
    RETURNS numeric
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_start_epoch             bigint;
    l_record                  record;
    l_epoch_available_invites smallint;
    l_max_reward              numeric;
    l_missed_reward           numeric;
    l_reward_coef             smallint;
BEGIN
    l_start_epoch = p_epoch - 3;
    if l_start_epoch < 0 then
        l_start_epoch = 0;
    end if;
    l_epoch_available_invites = 0;
    l_max_reward = 0;
    for l_record in SELECT ei.address_state_id, ei.epoch, ei.next_epoch_invites
                    FROM epoch_identities ei
                             JOIN address_states s ON s.id = ei.address_state_id AND s.address_id = p_address_id
                    WHERE ei.epoch >= l_start_epoch
                      AND ei.epoch <= p_epoch
                    ORDER BY ei.epoch
        loop
            if l_record.epoch > p_epoch - 3 then
                l_reward_coef = 3;
                if l_record.epoch = p_epoch - 1 then
                    l_reward_coef = l_reward_coef * 3;
                end if;
                if l_record.epoch = p_epoch - 2 then
                    l_reward_coef = l_reward_coef * 6;
                end if;
                l_max_reward = l_max_reward + l_reward_coef * p_reward_share * l_epoch_available_invites;
            end if;
            l_epoch_available_invites = l_record.next_epoch_invites;
        end loop;
    l_missed_reward = l_max_reward - coalesce(p_reward, 0);
    if l_missed_reward < 0 then
        l_missed_reward = 0;
    end if;
    return l_missed_reward;
END
$$;

CREATE OR REPLACE FUNCTION calculate_invitations_missed_reward(p_epoch bigint,
                                                               p_address_id bigint,
                                                               p_reward numeric,
                                                               p_reward_share numeric,
                                                               p_stake numeric)
    RETURNS numeric
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_start_epoch             bigint;
    l_record                  record;
    l_epoch_available_invites smallint;
    l_max_reward              numeric;
    l_missed_reward           numeric;
    l_reward_coef             double precision;
    l_stake_weight            double precision;
BEGIN
    l_start_epoch = p_epoch - 3;
    if l_start_epoch < 0 then
        l_start_epoch = 0;
    end if;
    l_epoch_available_invites = 0;
    l_max_reward = 0;
    l_stake_weight = coalesce(p_stake, 0) ^ 0.9;
    for l_record in SELECT ei.address_state_id, ei.epoch, ei.next_epoch_invites
                    FROM epoch_identities ei
                             JOIN address_states s ON s.id = ei.address_state_id AND s.address_id = p_address_id
                    WHERE ei.epoch >= l_start_epoch
                      AND ei.epoch <= p_epoch
                    ORDER BY ei.epoch
        loop
            if l_record.epoch > p_epoch - 3 then
                l_reward_coef = 0.2;
                if l_record.epoch = p_epoch - 1 then
                    l_reward_coef = 0.5;
                end if;
                if l_record.epoch = p_epoch - 2 then
                    l_reward_coef = 0.8;
                end if;
                l_max_reward = l_max_reward + l_stake_weight * l_reward_coef * p_reward_share * l_epoch_available_invites;
            end if;
            l_epoch_available_invites = l_record.next_epoch_invites;
        end loop;
    l_missed_reward = l_max_reward - coalesce(p_reward, 0);
    if l_missed_reward < 0 then
        l_missed_reward = 0;
    end if;
    return l_missed_reward;
END
$$;