CREATE OR REPLACE FUNCTION tmp_calculate_invitations_missed_reward(p_epoch bigint,
                                                                   p_address_id bigint,
                                                                   p_reward numeric,
                                                                   p_reward_share numeric,
                                                                   p_new_invitations_reward_coeffs_epoch bigint)
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
                if p_new_invitations_reward_coeffs_epoch <= l_record.epoch then
                    l_reward_coef = 3;
                else
                    l_reward_coef = 1;
                end if;
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

DO
$$
    DECLARE
        l_ei_record                                    record;
        l_vr_record                                    record;
        l_is_newbie                                    boolean;
        l_amount                                       numeric;
        l_delegatee_address_state_id                   bigint;
        l_validation                                   numeric(30, 18);
        l_validation_missed                            numeric(30, 18);
        l_validation_missed_reason                     smallint;
        l_flips                                        numeric(30, 18);
        l_flips_missed                                 numeric(30, 18);
        l_flips_missed_reason                          smallint;
        l_invitations                                  numeric(30, 18);
        l_invitations_missed                           numeric(30, 18);
        l_invitations_missed_reason                    smallint;
        l_reports                                      numeric(30, 18);
        l_reports_missed                               numeric(30, 18);
        l_reports_missed_reason                        smallint;
        l_potential_age                                smallint;
        l_rewarded_flips_cnt                           smallint;
        l_missed_reports_cnt                           smallint;
        l_new_invitations_reward_coeffs_epoch constant smallint = 37;
        l_report_rewards_epoch                constant smallint = 54;
        l_delegation_epoch                    constant smallint = 65;
        l_last_epoch_to_generate_data         constant smallint = 74;
    BEGIN

        for l_ei_record in SELECT ei.*, s.address_id, s.state
                           FROM epoch_identities ei
                                    LEFT JOIN address_states s on s.id = ei.address_state_id
                           WHERE ei.epoch >= l_delegation_epoch
                             AND ei.epoch <= l_last_epoch_to_generate_data
                             AND ei.delegatee_address_id is not null
                           ORDER BY ei.address_state_id
            loop

                SELECT ei.address_state_id
                INTO l_delegatee_address_state_id
                FROM epoch_identities ei
                         JOIN address_states s
                              ON s.id = ei.address_state_id AND s.address_id = l_ei_record.delegatee_address_id
                WHERE ei.epoch = l_ei_record.epoch;

                l_is_newbie = (l_ei_record.state = 7);

                for l_vr_record in SELECT *
                                   FROM validation_rewards
                                   WHERE ei_address_state_id = l_ei_record.address_state_id
                    loop
                        if l_vr_record.type = 9 then
                            l_amount = coalesce((SELECT sum(balance)
                                                 FROM reported_flip_rewards
                                                 WHERE ei_address_State_id = l_ei_record.address_state_id), 0);
                        else
                            l_amount = l_vr_record.stake;
                            if l_is_newbie then
                                l_amount = l_amount / 4.0;
                            else
                                l_amount = l_amount * 4.0;
                            end if;
                        end if;

                        if l_ei_record.epoch <= l_last_epoch_to_generate_data then
                            UPDATE validation_rewards
                            SET balance = balance - l_amount
                            WHERE ei_address_state_id = l_delegatee_address_state_id
                              AND "type" = l_vr_record.type;

                            UPDATE epoch_identities
                            SET total_validation_reward = total_validation_reward - l_amount
                            WHERE address_state_id = l_delegatee_address_state_id;

                            UPDATE validation_rewards
                            SET balance = balance + l_amount
                            WHERE ei_address_state_id = l_ei_record.address_state_id
                              AND "type" = l_vr_record.type;

                            UPDATE epoch_identities
                            SET total_validation_reward = total_validation_reward + l_amount
                            WHERE address_State_id = l_ei_record.address_state_id;
                        end if;

                        if not exists(SELECT 1
                                      FROM delegatee_validation_rewards
                                      WHERE epoch = l_ei_record.epoch
                                        AND delegatee_address_id = l_ei_record.delegatee_address_id
                                        AND delegator_address_id = l_ei_record.address_id) then
                            INSERT INTO delegatee_validation_rewards (epoch, delegatee_address_id, delegator_address_id, total_balance)
                            VALUES (l_ei_record.epoch, l_ei_record.delegatee_address_id, l_ei_record.address_id, 0);
                        end if;

                        UPDATE delegatee_validation_rewards
                        SET total_balance             = total_balance + l_amount,
                            validation_balance        = (case
                                                             when l_vr_record.type = 0
                                                                 then coalesce(validation_balance, 0) + l_amount
                                                             else validation_balance end),
                            flips_balance             = (case
                                                             when l_vr_record.type = 1
                                                                 then coalesce(flips_balance, 0) + l_amount
                                                             else flips_balance end),
                            invitations_balance       = (case
                                                             when l_vr_record.type = 2
                                                                 then coalesce(invitations_balance, 0) + l_amount
                                                             else invitations_balance end),
                            invitations2_balance      = (case
                                                             when l_vr_record.type = 5
                                                                 then coalesce(invitations2_balance, 0) + l_amount
                                                             else invitations2_balance end),
                            invitations3_balance      = (case
                                                             when l_vr_record.type = 6
                                                                 then coalesce(invitations3_balance, 0) + l_amount
                                                             else invitations3_balance end),
                            saved_invites_balance     = (case
                                                             when l_vr_record.type = 7
                                                                 then coalesce(saved_invites_balance, 0) + l_amount
                                                             else saved_invites_balance end),
                            saved_invites_win_balance = (case
                                                             when l_vr_record.type = 8
                                                                 then coalesce(saved_invites_win_balance, 0) + l_amount
                                                             else saved_invites_win_balance end),
                            reports_balance           = (case
                                                             when l_vr_record.type = 9
                                                                 then coalesce(reports_balance, 0) + l_amount
                                                             else reports_balance end)
                        WHERE epoch = l_ei_record.epoch
                          AND delegatee_address_id = l_ei_record.delegatee_address_id
                          AND delegator_address_id = l_ei_record.address_id;

                    end loop;
            end loop;

        INSERT INTO delegatee_total_validation_rewards
        SELECT epoch,
               delegatee_address_id,
               sum(total_balance),
               case
                   when sum(coalesce(validation_balance, 0)) = 0 then null
                   else sum(coalesce(validation_balance, 0)) end,
               case
                   when sum(coalesce(flips_balance, 0)) = 0 then null
                   else sum(coalesce(flips_balance, 0)) end,
               case
                   when sum(coalesce(invitations_balance, 0)) = 0 then null
                   else sum(coalesce(invitations_balance, 0)) end,
               case
                   when sum(coalesce(invitations2_balance, 0)) = 0 then null
                   else sum(coalesce(invitations2_balance, 0)) end,
               case
                   when sum(coalesce(invitations3_balance, 0)) = 0 then null
                   else sum(coalesce(invitations3_balance, 0)) end,
               case
                   when sum(coalesce(saved_invites_balance, 0)) = 0 then null
                   else sum(coalesce(saved_invites_balance, 0)) end,
               case
                   when sum(coalesce(saved_invites_win_balance, 0)) = 0 then null
                   else sum(coalesce(saved_invites_win_balance, 0)) end,
               case
                   when sum(coalesce(reports_balance, 0)) = 0 then null
                   else sum(coalesce(reports_balance, 0)) end,
               count(*) delegators
        FROM delegatee_validation_rewards
        WHERE epoch <= l_last_epoch_to_generate_data
        GROUP BY epoch, delegatee_address_id
        ORDER BY epoch, delegatee_address_id;

        UPDATE epoch_identities
        SET total_validation_reward = 0
        WHERE total_validation_reward < 0.00000001
          and address_state_id IN
              (SELECT ei_address_state_id FROM validation_rewards WHERE abs(balance) + abs(stake) < 0.00000001);
        DELETE FROM validation_rewards WHERE abs(balance) + abs(stake) < 0.00000001;


        -- !!!!!!!!!!!!!!!!!!!!!! REWARD SUMMARIES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
        for l_ei_record in SELECT ei.*,
                                  s.address_id,
                                  s.state,
                                  prevs.state prev_state,
                                  (ba.ei_address_state_id is not null) penalized,
                                  tr.validation_share,
                                  tr.flips_share,
                                  tr.invitations_share
                           FROM epoch_identities ei
                                    LEFT JOIN address_states s ON s.id = ei.address_state_id
                                    LEFT JOIN address_states prevs ON prevs.id = s.prev_id
                                    LEFT JOIN total_rewards tr ON tr.epoch = ei.epoch
                                    LEFT JOIN bad_authors ba ON ba.ei_address_state_id = ei.address_state_id
                           WHERE ei.epoch <= l_last_epoch_to_generate_data
                           ORDER BY address_state_id
            loop
                SELECT balance + stake
                INTO l_validation
                FROM validation_rewards
                WHERE ei_address_state_id = l_ei_record.address_state_id
                  AND "type" = 0;
                if l_validation = 0 then
                    l_validation = null;
                end if;
                l_validation_missed = null;
                l_validation_missed_reason = null;

                SELECT balance + stake
                INTO l_flips
                FROM validation_rewards
                WHERE ei_address_state_id = l_ei_record.address_state_id
                  AND "type" = 1;
                if l_flips = 0 then
                    l_flips = null;
                end if;
                l_flips_missed = null;
                l_flips_missed_reason = null;

                SELECT sum(balance + stake)
                INTO l_invitations
                FROM validation_rewards
                WHERE ei_address_state_id = l_ei_record.address_state_id
                  AND "type" in (2, 5, 6, 7, 8);
                if l_invitations = 0 then
                    l_invitations = null;
                end if;
                l_invitations_missed = null;
                l_invitations_missed_reason = null;

                l_reports = null;
                l_reports_missed = null;
                l_reports_missed_reason = null;

                -- VALIDATION
                if l_ei_record.state not in (3, 7, 8) or l_ei_record.penalized then
                    if l_ei_record.prev_state in (0, 1, 2, 5) then
                        l_potential_age = 1;
                    else
                        l_potential_age = 1 + l_ei_record.epoch - l_ei_record.birth_epoch;
                    end if;
                    l_validation_missed = l_ei_record.validation_share * (l_potential_age ^ (1.0 / 3.0));
                    if l_validation_missed > 0 then
                        if l_ei_record.penalized then
                            l_validation_missed_reason = 1;
                        else
                            l_validation_missed_reason = 2;
                        end if;
                    end if;
                end if;

                -- FLIPS
                SELECT count(*)
                INTO l_rewarded_flips_cnt
                FROM rewarded_flips rf
                         JOIN transactions t ON t.id = rf.flip_tx_id AND t.from = l_ei_record.address_id
                         JOIN blocks b on b.height = t.block_height AND b.epoch = l_ei_record.epoch;

                if l_ei_record.available_flips > 0 then
                    l_flips_missed = l_ei_record.flips_share * (l_ei_record.available_flips - l_rewarded_flips_cnt);
                end if;

                if l_flips_missed is not null and l_flips_missed > 0 then
                    if l_ei_record.penalized then
                        l_flips_missed_reason = 1;
                    else
                        if l_ei_record.missed then
                            l_flips_missed_reason = 3;
                        else
                            l_flips_missed_reason = 4;
                        end if;
                    end if;
                end if;

                -- INVITATIONS
                l_invitations_missed =
                        tmp_calculate_invitations_missed_reward(l_ei_record.epoch,
                                                                l_ei_record.address_id,
                                                                l_invitations,
                                                                l_ei_record.invitations_share,
                                                                l_new_invitations_reward_coeffs_epoch);
                if l_invitations_missed_reason is not null and l_invitations_missed is null or
                   l_invitations_missed = 0 then
                    l_invitations_missed = null;
                    l_invitations_missed_reason = null;
                end if;
                if l_invitations_missed_reason is null and l_invitations_missed > 0 then
                    l_invitations_missed_reason = 6;
                end if;

                -- REPORTS
                if l_ei_record.epoch >= l_report_rewards_epoch then

                    SELECT balance + stake
                    INTO l_reports
                    FROM validation_rewards
                    WHERE ei_address_state_id = l_ei_record.address_state_id
                      AND "type" = 9;
                    if l_reports = 0 then
                        l_reports = null;
                    end if;

                    l_missed_reports_cnt = (SELECT count(*)
                                            FROM answers a
                                                     JOIN flips f ON f.tx_id = a.flip_tx_id AND f.grade = 1
                                            WHERE a.ei_address_state_id = l_ei_record.address_state_id
                                              AND NOT a.is_short) -
                                           (SELECT count(*)
                                            FROM reported_flip_rewards
                                            WHERE epoch = l_ei_record.epoch
                                              AND address_id = l_ei_record.address_id);
                    if l_missed_reports_cnt > 0 then
                        l_reports_missed = l_missed_reports_cnt * l_ei_record.flips_share / 5.0;
                        if l_ei_record.penalized then
                            l_reports_missed_reason = 1;
                        else
                            l_reports_missed_reason = 5;
                        end if;
                    end if;
                end if;


                INSERT INTO validation_reward_summaries (epoch, address_id, validation, validation_missed,
                                                         validation_missed_reason, flips, flips_missed,
                                                         flips_missed_reason,
                                                         invitations, invitations_missed, invitations_missed_reason,
                                                         reports, reports_missed, reports_missed_reason)
                VALUES (l_ei_record.epoch,
                        l_ei_record.address_id,
                        l_validation,
                        l_validation_missed,
                        l_validation_missed_reason,
                        l_flips,
                        l_flips_missed,
                        l_flips_missed_reason,
                        l_invitations,
                        l_invitations_missed,
                        l_invitations_missed_reason,
                        l_reports,
                        l_reports_missed,
                        l_reports_missed_reason);

            end loop;

        DROP function tmp_calculate_invitations_missed_reward;
    END
$$