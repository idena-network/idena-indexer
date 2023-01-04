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
    l_invitee_reward_summary                      jsonb;
    l_invitations                                 numeric;
    l_invitations_missed                          numeric;
    l_invitations_missed_reason                   smallint;
    l_invitee                                     numeric;
    l_invitee_missed                              numeric;
    l_invitee_missed_reason                       smallint;
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
            l_invitee_reward_summary = l_item -> 'invitee';

            l_invitations = (l_invitations_reward_summary ->> 'earned')::numeric;
            l_invitations_missed_reason = (l_invitations_reward_summary ->> 'missedReason')::smallint;
            if l_enable_upgrade_10 then
                l_invitations_missed =
                        calculate_invitations_missed_reward(p_epoch, l_address_id, l_invitations, p_invitations_share,
                                                            (l_item ->> 'prevStake')::numeric);
            else
                l_invitations_missed =
                        calculate_invitations_missed_reward_old(p_epoch, l_address_id, l_invitations,
                                                                p_invitations_share);
            end if;
            if l_invitations_missed_reason is not null and l_invitations_missed is null or l_invitations_missed = 0 then
                l_invitations_missed_reason = null;
            end if;
            if l_invitations_missed_reason is null and l_invitations_missed > 0 then
                l_invitations_missed_reason = MISSED_REWARD_REASON_NOT_ALL_INVITES;
            end if;

            l_invitee = (l_invitee_reward_summary ->> 'earned')::numeric;
            l_invitee_missed_reason = (l_invitee_reward_summary ->> 'missedReason')::smallint;
            l_invitee_missed = null;
            if coalesce(l_invitee, 0) = 0 then
                SELECT p_missed, p_missed_reason
                INTO l_invitee_missed, l_invitee_missed_reason
                FROM
                    calculate_invitee_missed_reward(p_epoch, l_address_id, p_invitations_share,
                                                    l_invitee_missed_reason);
            end if;

            INSERT INTO validation_reward_summaries (epoch, address_id, validation, validation_missed,
                                                     validation_missed_reason, flips, flips_missed, flips_missed_reason,
                                                     invitations, invitations_missed, invitations_missed_reason,
                                                     reports, reports_missed, reports_missed_reason,
                                                     candidate, candidate_missed, candidate_missed_reason,
                                                     staking, staking_missed, staking_missed_reason,
                                                     extra_flips, extra_flips_missed, extra_flips_missed_reason,
                                                     invitee, invitee_missed, invitee_missed_reason)
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
                    (l_extra_flips_reward_summary ->> 'missedReason')::smallint,
                    null_if_zero(l_invitee),
                    null_if_zero(l_invitee_missed),
                    l_invitee_missed_reason);

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
                l_max_reward =
                            l_max_reward + l_stake_weight * l_reward_coef * p_reward_share * l_epoch_available_invites;
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

CREATE OR REPLACE FUNCTION calculate_invitee_missed_reward(p_epoch bigint,
                                                           p_address_id bigint,
                                                           p_reward_share numeric,
                                                           inout p_missed_reason smallint,
                                                           out p_missed numeric)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    MISSED_REWARD_REASON_INVITER_REPORTED      CONSTANT smallint = 7;
    MISSED_REWARD_REASON_INVITER_NOT_VALIDATED CONSTANT smallint = 8;
    MISSED_REWARD_REASON_INVITER_RESET         CONSTANT smallint = 9;
    l_activation_tx_id                                  bigint;
    l_activation_epoch                                  integer;
    l_age                                               smallint;
    l_inviter_stake_weight                              numeric;
    l_reward_weight                                     numeric;
    l_is_inviter_reported                               boolean;
    l_is_inviter_not_validated                          boolean;
BEGIN

    SELECT lat.activation_tx_id,
           lat.epoch,
           pow(rsa.amount, 0.9),
           ba.reason is not null    as is_inviter_reported,
           s.state not in (3, 7, 8) as is_inviter_not_validated
    INTO l_activation_tx_id, l_activation_epoch, l_inviter_stake_weight, l_is_inviter_reported, l_is_inviter_not_validated
    FROM latest_activation_txs lat
             LEFT JOIN activation_txs act ON act.tx_id = lat.activation_tx_id
             LEFT JOIN transactions invt ON invt.id = act.invite_tx_id
             LEFT JOIN addresses inva ON inva.id = invt.from
             LEFT JOIN cur_epoch_identities ei ON lower(ei.address) = lower(inva.address)
             LEFT JOIN bad_authors ba ON ba.ei_address_state_id = ei.address_state_id
             LEFT JOIN address_states s ON s.id = ei.address_state_id
             LEFT JOIN reward_staked_amounts rsa ON rsa.ei_address_state_id = ei.address_state_id
    WHERE lat.address_id = p_address_id
    ORDER BY lat.activation_tx_id DESC
    LIMIT 1;

    p_missed = 0;

    l_age = p_epoch - l_activation_epoch + 1;
    if l_activation_tx_id is null or l_age > 3 then
        return;
    end if;

    if coalesce(l_inviter_stake_weight, 0) = 0 then
        return;
    end if;

    l_reward_weight = 0.8;
    if l_age = 2 then
        l_reward_weight = 0.5;
    end if;
    if l_age = 3 then
        l_reward_weight = 0.2;
    end if;

    l_reward_weight = l_reward_weight * l_inviter_stake_weight;
    p_missed = p_reward_share * l_reward_weight;

    if p_missed > 0 and coalesce(p_missed_reason, 0) = 0 then
        if l_is_inviter_reported then
            p_missed_reason = MISSED_REWARD_REASON_INVITER_REPORTED;
            return;
        end if;

        if l_is_inviter_not_validated then
            p_missed_reason = MISSED_REWARD_REASON_INVITER_NOT_VALIDATED;
            return;
        end if;
        p_missed_reason = MISSED_REWARD_REASON_INVITER_RESET;
    end if;
END
$$;