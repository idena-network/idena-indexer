CREATE OR REPLACE PROCEDURE save_delegatee_epoch_rewards(p_epoch bigint,
                                                         p_block_height bigint,
                                                         p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                 jsonb;
    l_delegation_reward    jsonb;
    l_delegator_rewqard    jsonb;
    l_delegator_address_id bigint;
    l_delegatee_address_id bigint;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            l_delegatee_address_id = get_address_id_or_insert(p_block_height, (l_item ->> 'address')::text);
            l_delegation_reward = l_item -> 'totalReward';

            INSERT INTO delegatee_total_validation_rewards (epoch, delegatee_address_id, total_balance,
                                                            validation_balance, flips_balance, invitations_balance,
                                                            invitations2_balance, invitations3_balance,
                                                            saved_invites_balance, saved_invites_win_balance,
                                                            reports_balance, candidate_balance, staking_balance,
                                                            delegators, penalized_delegators, extra_flips_balance,
                                                            invitee1_balance, invitee2_balance, invitee3_balance)
            VALUES (p_epoch, l_delegatee_address_id, (l_delegation_reward ->> 'total')::numeric,
                    (l_delegation_reward ->> 'validation')::numeric, (l_delegation_reward ->> 'flips')::numeric,
                    (l_delegation_reward ->> 'invitations')::numeric, (l_delegation_reward ->> 'invitations2')::numeric,
                    (l_delegation_reward ->> 'invitations3')::numeric,
                    (l_delegation_reward ->> 'savedInvites')::numeric,
                    (l_delegation_reward ->> 'savedInvitesWin')::numeric,
                    (l_delegation_reward ->> 'reports')::numeric,
                    (l_delegation_reward ->> 'candidate')::numeric,
                    (l_delegation_reward ->> 'staking')::numeric,
                    (case
                         when (l_item -> 'delegatorRewards') is null then 0
                         else jsonb_array_length(l_item -> 'delegatorRewards') end),
                    (l_item ->> 'penalizedDelegators')::integer,
                    (l_delegation_reward ->> 'extraFlips')::numeric,
                    (l_delegation_reward ->> 'invitee1')::numeric,
                    (l_delegation_reward ->> 'invitee2')::numeric,
                    (l_delegation_reward ->> 'invitee3')::numeric);

            if (l_item -> 'delegatorRewards') is not null then
                for j in 0..jsonb_array_length(l_item -> 'delegatorRewards') - 1
                    loop
                        l_delegator_rewqard = l_item -> 'delegatorRewards' ->> j;
                        l_delegator_address_id =
                                get_address_id_or_insert(p_block_height, (l_delegator_rewqard ->> 'address')::text);
                        l_delegation_reward = l_delegator_rewqard -> 'totalReward';

                        INSERT INTO delegatee_validation_rewards (epoch, delegatee_address_id, delegator_address_id,
                                                                  total_balance,
                                                                  validation_balance, flips_balance,
                                                                  invitations_balance,
                                                                  invitations2_balance, invitations3_balance,
                                                                  saved_invites_balance, saved_invites_win_balance,
                                                                  candidate_balance, staking_balance,
                                                                  reports_balance, extra_flips_balance,
                                                                  invitee1_balance, invitee2_balance, invitee3_balance)
                        VALUES (p_epoch, l_delegatee_address_id, l_delegator_address_id,
                                (l_delegation_reward ->> 'total')::numeric,
                                (l_delegation_reward ->> 'validation')::numeric,
                                (l_delegation_reward ->> 'flips')::numeric,
                                (l_delegation_reward ->> 'invitations')::numeric,
                                (l_delegation_reward ->> 'invitations2')::numeric,
                                (l_delegation_reward ->> 'invitations3')::numeric,
                                (l_delegation_reward ->> 'savedInvites')::numeric,
                                (l_delegation_reward ->> 'savedInvitesWin')::numeric,
                                (l_delegation_reward ->> 'candidate')::numeric,
                                (l_delegation_reward ->> 'staking')::numeric,
                                (l_delegation_reward ->> 'reports')::numeric,
                                (l_delegation_reward ->> 'extraFlips')::numeric,
                                (l_delegation_reward ->> 'invitee1')::numeric,
                                (l_delegation_reward ->> 'invitee2')::numeric,
                                (l_delegation_reward ->> 'invitee3')::numeric);
                    end loop;
            end if;
        end loop;
END
$$;