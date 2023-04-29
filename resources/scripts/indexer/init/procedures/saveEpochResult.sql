CREATE OR REPLACE PROCEDURE save_epoch_result(p_epoch bigint,
                                              p_height bigint,
                                              p_birthdays tp_birthday[],
                                              p_identities tp_epoch_identity[],
                                              p_flips_to_solve tp_flip_to_solve[],
                                              p_keys tp_mem_pool_flip_key[],
                                              p_answers tp_answer[],
                                              p_states tp_flip_state[],
                                              p_bad_authors tp_bad_author[],
                                              p_total tp_total_epoch_reward,
                                              p_validation_rewards tp_epoch_reward[],
                                              p_ages tp_reward_age[],
                                              p_staked_amounts tp_reward_staked_amount[],
                                              p_failed_staked_amounts tp_reward_staked_amount[],
                                              p_fund_rewards tp_epoch_reward[],
                                              p_rewarded_flip_cids text[],
                                              p_rewarded_extra_flip_cids text[],
                                              p_rewarded_invitations tp_rewarded_invitation[],
                                              p_saved_invite_rewards tp_saved_invite_rewards[],
                                              p_reported_flip_rewards tp_reported_flip_reward[],
                                              p_failed_validation boolean,
                                              p_min_score_for_invite real,
                                              p_rewarded_invitees tp_rewarded_invitee[],
                                              p_data jsonb)
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_start timestamp;
    l_end   timestamp;
BEGIN
    SET session_replication_role = replica;

    select clock_timestamp() into l_start;
    if p_birthdays is not null then
        call save_birthdays(p_birthdays);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_birthdays', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_identities is not null then
        call save_epoch_identities(p_epoch, p_height, p_identities);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_epoch_identities', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_flips_to_solve is not null then
        call save_flips_to_solve(p_flips_to_solve);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_flips_to_solve', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_keys is not null then
        call save_mem_pool_flip_keys(p_keys);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_mem_pool_flip_keys', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_states is not null or p_states is not null then
        call save_flip_stats(p_height, p_answers, p_states);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_flip_stats', l_start, l_end);

    select clock_timestamp() into l_start;
    call save_epoch_rewards(p_epoch, p_height, p_bad_authors, p_total, p_validation_rewards, p_ages,
                            p_staked_amounts,
                            p_failed_staked_amounts,
                            p_fund_rewards,
                            p_rewarded_flip_cids,
                            p_rewarded_extra_flip_cids,
                            p_rewarded_invitations,
                            p_saved_invite_rewards,
                            p_reported_flip_rewards,
                            p_rewarded_invitees);
    select clock_timestamp() into l_end;
    call log_performance('save_epoch_rewards', l_start, l_end);

    if p_data is not null then
        select clock_timestamp() into l_start;
        call save_epoch_rewards_bounds(p_epoch, p_height, p_data -> 'rewardsBounds');
        select clock_timestamp() into l_end;
        call log_performance('save_epoch_rewards_bounds', l_start, l_end);

        select clock_timestamp() into l_start;
        call save_epoch_flip_statuses(p_epoch, p_data -> 'flipStatuses');
        select clock_timestamp() into l_end;
        call log_performance('save_epoch_flip_statuses', l_start, l_end);

        select clock_timestamp() into l_start;
        call save_delegatee_epoch_rewards(p_epoch, p_height, p_data -> 'delegateeEpochRewards');
        select clock_timestamp() into l_end;
        call log_performance('save_delegatee_epoch_rewards', l_start, l_end);

        select clock_timestamp() into l_start;
        call save_validation_rewards_summaries(p_epoch, p_height, p_data -> 'validationRewardsSummaries',
                                               p_total.invitations_share);
        select clock_timestamp() into l_end;
        call log_performance('save_validation_rewards_summaries', l_start, l_end);

        select clock_timestamp() into l_start;
        call save_pool_size_history(p_epoch, p_height, p_data -> 'poolSizeChanges');
        select clock_timestamp() into l_end;
        call log_performance('save_pool_size_history', l_start, l_end);
    end if;


    select clock_timestamp() into l_start;
    if p_failed_validation then
        insert into failed_validations (block_height) values (p_height);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('failed_validations', l_start, l_end);

    select clock_timestamp() into l_start;
    call generate_epoch_summaries(p_epoch, p_height, p_min_score_for_invite, (p_data ->> 'reportedFlips')::integer);
    select clock_timestamp() into l_end;
    call log_performance('generate_epoch_summaries', l_start, l_end);

    select clock_timestamp() into l_start;
    call update_flips_queue();
    select clock_timestamp() into l_end;
    call log_performance('update_flips_queue', l_start, l_end);

    --     select clock_timestamp() into l_start;
--     DELETE FROM latest_activation_txs WHERE epoch < p_epoch - 2;
--     select clock_timestamp() into l_end;
--     call log_performance('delete_activation_txs', l_start, l_end);

    SET session_replication_role = DEFAULT;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_birthdays(p_birthdays tp_birthday[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    birthday     tp_birthday;
    l_address_id bigint;
BEGIN
    for i in 1..cardinality(p_birthdays)
        loop
            birthday := p_birthdays[i];
            select id into l_address_id from addresses where lower(address) = lower(birthday.address);
            insert into birthdays (address_id, birth_epoch)
            values (l_address_id, birthday.birth_epoch)
            on conflict (address_id) do update set birth_epoch=birthday.birth_epoch;

            call apply_birthday_on_delegations(l_address_id, birthday.birth_epoch);
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_epoch_identities(p_epoch bigint,
                                                  p_height bigint,
                                                  p_identities tp_epoch_identity[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_identity             tp_epoch_identity;
    l_address_id           bigint;
    l_prev_state_id        bigint;
    l_state_id             bigint;
    l_min_epoch_height     bigint;
    l_max_epoch_height     bigint;
    l_delegatee_address_id bigint;
BEGIN

    CREATE TEMP TABLE cur_epoch_identities
    (
        address          character(42),
        address_state_id bigint
    ) ON COMMIT DROP;

    for i in 1..cardinality(p_identities)
        loop
            l_identity := p_identities[i];

            select id into l_address_id from addresses where lower(address) = lower(l_identity.address);

            update address_states
            set is_actual = false
            where address_id = l_address_id
              and is_actual
            returning id into l_prev_state_id;

            insert into address_states (address_id, state, is_actual, block_height, prev_id)
            values (l_address_id, l_identity.state, true, p_height, l_prev_state_id)
            returning id into l_state_id;

            if char_length(l_identity.delegatee_address) > 0 then
                SELECT id
                INTO l_delegatee_address_id
                FROM addresses
                WHERE lower(address) = lower(l_identity.delegatee_address);
            else
                l_delegatee_address_id = null;
            end if;

            insert into epoch_identities (epoch, address_id, address_state_id, short_point, short_flips,
                                          total_short_point,
                                          total_short_flips, long_point, long_flips, approved, missed, required_flips,
                                          available_flips, made_flips, next_epoch_invites, birth_epoch,
                                          total_validation_reward, short_answers, long_answers, wrong_words_flips,
                                          delegatee_address_id, shard_id, new_shard_id, wrong_grade_reason)
            values (p_epoch, l_address_id, l_state_id, l_identity.short_point, l_identity.short_flips,
                    l_identity.total_short_point,
                    l_identity.total_short_flips, l_identity.long_point, l_identity.long_flips, l_identity.approved,
                    l_identity.missed, l_identity.required_flips, l_identity.available_flips, l_identity.made_flips,
                    l_identity.next_epoch_invites, l_identity.birth_epoch, 0, l_identity.short_answers,
                    l_identity.long_answers, l_identity.wrong_words_flips, l_delegatee_address_id, l_identity.shard_id,
                    l_identity.new_shard_id, null_if_zero_smallint(l_identity.wrong_grade_reason));

            if l_identity.wrong_words_flips > 0 then
                CALL update_address_summary(p_address_id => l_address_id,
                                            p_wrong_words_flips_diff => l_identity.wrong_words_flips);
            end if;

            insert into cur_epoch_identities values (l_identity.address, l_state_id);

            if l_prev_state_id is not null then
                insert into epoch_identity_interim_states (address_state_id, block_height)
                values (l_prev_state_id, p_height);
            end if;

        end loop;

    CREATE UNIQUE INDEX ON cur_epoch_identities (lower(address));
    CREATE UNIQUE INDEX ON cur_epoch_identities (address_state_id);

    SELECT min(height), max(height) INTO l_min_epoch_height, l_max_epoch_height FROM blocks WHERE epoch = p_epoch;

    insert into epoch_identity_interim_states (select s.id, p_height
                                               from address_states s
                                                        left join temporary_identities ti on ti.address_id = s.address_id
                                                        left join cur_epoch_identities ei on ei.address_state_id = s.id
                                               where s.block_height >= l_min_epoch_height
                                                 and s.block_height <= l_max_epoch_height
                                                 and s.is_actual
                                                 and s.state in (0, 5)
                                                 and ti.address_id is null -- exclude temporary identities
                                                 and ei.address_state_id is null -- exclude epoch identities (at least god node if it was killed before validation)
    );
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_flips_to_solve(p_flips_to_solve tp_flip_to_solve[])
    LANGUAGE 'plpgsql'
AS
$BODY$
BEGIN
    INSERT INTO flips_to_solve (ei_address_state_id, flip_tx_id, is_short, index)
    SELECT cei.address_state_id,
           f.tx_id,
           flips_to_solve_arr.is_short,
           flips_to_solve_arr.index
    FROM unnest(p_flips_to_solve) AS flips_to_solve_arr
             JOIN flips f ON lower(f.cid) = lower(flips_to_solve_arr.cid)
             JOIN cur_epoch_identities cei ON lower(cei.address) = lower(flips_to_solve_arr.address);
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_mem_pool_flip_keys(p_keys tp_mem_pool_flip_key[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_key tp_mem_pool_flip_key;
BEGIN
    for i in 1..cardinality(p_keys)
        loop
            l_key := p_keys[i];
            insert into mem_pool_flip_keys (ei_address_state_id, key)
            values ((select address_state_id
                     from cur_epoch_identities
                     where lower(address) = lower(l_key.address)),
                    l_key.key);
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_flip_stats(block_height bigint,
                                            answers tp_answer[],
                                            states tp_flip_state[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    state   tp_flip_state;
    l_start timestamp;
    l_end   timestamp;
BEGIN

    select clock_timestamp() into l_start;
    if answers is not null then
        INSERT INTO answers (flip_tx_id, ei_address_state_id, is_short, answer, point, grade, index, considered)
        SELECT f.tx_id,
               cei.address_state_id,
               answers_arr.is_short,
               answers_arr.answer,
               answers_arr.point,
               answers_arr.grade,
               answers_arr.index,
               answers_arr.considered
        FROM unnest(answers) AS answers_arr
                 JOIN flips f ON lower(f.cid) = lower(answers_arr.flip_cid)
                 JOIN cur_epoch_identities cei ON lower(cei.address) = lower(answers_arr.address);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_flip_answers', l_start, l_end);

    select clock_timestamp() into l_start;
    if states is not null then
        for i in 1..cardinality(states)
            loop
                state := states[i];
                UPDATE FLIPS
                SET STATUS=state.status,
                    ANSWER=state.answer,
                    GRADE=state.grade,
                    GRADE_SCORE=state.grade_score,
                    STATUS_BLOCK_HEIGHT=block_height
                WHERE lower(CID) = lower(state.flip_cid);
            end loop;
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_flip_states', l_start, l_end);
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_epoch_rewards(p_epoch bigint,
                                               p_block_height bigint,
                                               p_bad_authors tp_bad_author[],
                                               p_total tp_total_epoch_reward,
                                               p_validation_rewards tp_epoch_reward[],
                                               p_ages tp_reward_age[],
                                               p_staked_amounts tp_reward_staked_amount[],
                                               p_failed_staked_amounts tp_reward_staked_amount[],
                                               p_fund_rewards tp_epoch_reward[],
                                               p_rewarded_flip_cids text[],
                                               p_rewarded_extra_flip_cids text[],
                                               p_rewarded_invitations tp_rewarded_invitation[],
                                               p_saved_invite_rewards tp_saved_invite_rewards[],
                                               p_reported_flip_rewards tp_reported_flip_reward[],
                                               p_rewarded_invitees tp_rewarded_invitee[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_start timestamp;
    l_end   timestamp;
BEGIN
    select clock_timestamp() into l_start;
    if p_bad_authors is not null then
        call save_bad_authors(p_bad_authors);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_bad_authors', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_total is not null then
        call save_total_reward(p_epoch, p_total);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_total_reward', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_validation_rewards is not null then
        call save_validation_rewards(p_validation_rewards);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_validation_rewards', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_ages is not null then
        call save_reward_ages(p_ages);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_reward_ages', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_staked_amounts is not null then
        call save_reward_staked_amounts(p_staked_amounts, false);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_reward_staked_amounts', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_failed_staked_amounts is not null then
        call save_reward_staked_amounts(p_failed_staked_amounts, true);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_reward_failed_staked_amounts', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_fund_rewards is not null then
        call save_fund_rewards(p_block_height, p_fund_rewards);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_fund_rewards', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_rewarded_flip_cids is not null then
        call save_rewarded_flips(p_rewarded_flip_cids, false);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_rewarded_flips', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_rewarded_extra_flip_cids is not null then
        call save_rewarded_flips(p_rewarded_extra_flip_cids, true);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_rewarded_extra_flips', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_rewarded_invitations is not null then
        call save_rewarded_invitations(p_block_height, p_rewarded_invitations);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_rewarded_invitations', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_rewarded_invitees is not null then
        call save_rewarded_invitees(p_block_height, p_rewarded_invitees);
--         call save_rewarded_invitees(p_epoch, p_block_height, p_rewarded_invitees);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_rewarded_invitees', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_saved_invite_rewards is not null then
        call save_saved_invite_rewards(p_saved_invite_rewards);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_saved_invite_rewards', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_reported_flip_rewards is not null then
        call save_reported_flip_rewards(p_epoch, p_reported_flip_rewards);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_reported_flip_rewards', l_start, l_end);
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_bad_authors(p_bad_authors tp_bad_author[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_bad_author tp_bad_author;
BEGIN
    for i in 1..cardinality(p_bad_authors)
        loop
            l_bad_author := p_bad_authors[i];
            insert into bad_authors (ei_address_state_id, reason)
            values ((select address_state_id
                     from cur_epoch_identities
                     where lower(address) = lower(l_bad_author.address)), l_bad_author.reason);
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_total_reward(p_epoch bigint,
                                              p_total tp_total_epoch_reward)
    LANGUAGE 'plpgsql'
AS
$BODY$
BEGIN
    insert into total_rewards (epoch,
                               total, validation,
                               flips,
                               flips_extra,
                               invitations,
                               foundation,
                               zero_wallet,
                               validation_share,
                               flips_share,
                               flips_extra_share,
                               invitations_share,
                               reports,
                               reports_share,
                               staking,
                               candidate,
                               staking_share,
                               candidate_share)
    values (p_epoch,
            p_total.total,
            p_total.validation,
            p_total.flips,
            p_total.flips_extra,
            p_total.invitations,
            p_total.foundation,
            p_total.zero_wallet,
            p_total.validation_share,
            p_total.flips_share,
            p_total.flips_extra_share,
            p_total.invitations_share,
            p_total.reports,
            p_total.reports_share,
            p_total.staking,
            p_total.candidate,
            p_total.staking_share,
            p_total.candidate_share);
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_validation_rewards(p_validation_rewards tp_epoch_reward[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_validation_reward tp_epoch_reward;
BEGIN
    for i in 1..cardinality(p_validation_rewards)
        loop
            l_validation_reward := p_validation_rewards[i];
            insert into validation_rewards (ei_address_state_id, balance, stake, type)
            values ((select address_state_id
                     from cur_epoch_identities
                     where lower(address) = lower(l_validation_reward.address)),
                    l_validation_reward.balance,
                    l_validation_reward.stake,
                    l_validation_reward.type);
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_reward_ages(p_ages tp_reward_age[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_age tp_reward_age;
BEGIN
    for i in 1..cardinality(p_ages)
        loop
            l_age := p_ages[i];
            insert into reward_ages (ei_address_state_id, age)
            values ((select address_state_id
                     from cur_epoch_identities
                     where lower(address) = lower(l_age.address)),
                    l_age.age);
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_reward_staked_amounts(p_items tp_reward_staked_amount[], p_failed boolean)
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_item tp_reward_staked_amount;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item := p_items[i];
            INSERT INTO reward_staked_amounts (ei_address_state_id, amount, failed)
            VALUES ((SELECT address_state_id
                     FROM cur_epoch_identities
                     WHERE lower(address) = lower(l_item.address)),
                    l_item.amount, case when p_failed then true end);
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_fund_rewards(p_block_height bigint,
                                              p_fund_rewards tp_epoch_reward[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_fund_reward tp_epoch_reward;
BEGIN
    for i in 1..cardinality(p_fund_rewards)
        loop
            l_fund_reward := p_fund_rewards[i];
            insert into fund_rewards (address_id, block_height, balance, type)
            values ((select id from addresses where lower(address) = lower(l_fund_reward.address)),
                    p_block_height,
                    l_fund_reward.balance,
                    l_fund_reward.type);
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_rewarded_flips(p_rewarded_flip_cids text[], p_extra boolean)
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_flip_tx_id bigint;
BEGIN
    for i in 1..cardinality(p_rewarded_flip_cids)
        loop
            select tx_id into l_flip_tx_id from flips where lower(cid) = lower(p_rewarded_flip_cids[i]);
            if l_flip_tx_id is null then
                continue;
            end if;
            insert into rewarded_flips (flip_tx_id, extra)
            values (l_flip_tx_id, case when p_extra then true end);
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_rewarded_invitations(p_block_height bigint,
                                                      p_rewarded_invitations tp_rewarded_invitation[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_rewarded_invitation tp_rewarded_invitation;
    l_invite_tx_id        bigint;
BEGIN
    for i in 1..cardinality(p_rewarded_invitations)
        loop
            l_rewarded_invitation = p_rewarded_invitations[i];
            select id into l_invite_tx_id from transactions where lower(hash) = lower(l_rewarded_invitation.tx_hash);
            if l_invite_tx_id is null then
                continue;
            end if;
            insert into rewarded_invitations (invite_tx_id, block_height, reward_type, epoch_height)
            values (l_invite_tx_id,
                    p_block_height,
                    l_rewarded_invitation.reward_type,
                    l_rewarded_invitation.epoch_height);
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_rewarded_invitees(p_block_height bigint,
                                                   p_rewarded_invitees tp_rewarded_invitee[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_rewarded_invitee tp_rewarded_invitee;
    l_invite_tx_id     bigint;
BEGIN
    for i in 1..cardinality(p_rewarded_invitees)
        loop
            l_rewarded_invitee = p_rewarded_invitees[i];
            select id into l_invite_tx_id from transactions where lower(hash) = lower(l_rewarded_invitee.tx_hash);
            if l_invite_tx_id is null then
                continue;
            end if;
            insert into rewarded_invitees (invite_tx_id, block_height, epoch_height)
            values (l_invite_tx_id,
                    p_block_height,
                    l_rewarded_invitee.epoch_height);
        end loop;
END
$$;
-- CREATE OR REPLACE PROCEDURE save_rewarded_invitees(p_epoch bigint,
--                                                    p_block_height bigint,
--                                                    p_rewarded_invitees tp_rewarded_invitee[])
--     LANGUAGE 'plpgsql'
-- AS
-- $$
-- DECLARE
--     l_rewarded_invitee   tp_rewarded_invitee;
--     l_invite_tx_id       bigint;
--     l_invitee_address_id bigint;
-- BEGIN
--     for i in 1..cardinality(p_rewarded_invitees)
--         loop
--             l_rewarded_invitee = p_rewarded_invitees[i];
--             select id into l_invite_tx_id from transactions where lower(hash) = lower(l_rewarded_invitee.tx_hash);
--             if l_invite_tx_id is null then
--                 continue;
--             end if;
--
--             SELECT "to"
--             INTO l_invitee_address_id
--             FROM activation_txs act
--                      LEFT JOIN transactions t ON t.id = act.tx_id
--             WHERE act.invite_tx_id = l_invite_tx_id;
--
--             insert into rewarded_invitees (epoch, address_id, invite_tx_id, block_height, epoch_height)
--             values (p_epoch,
--                     l_invitee_address_id,
--                     l_invite_tx_id,
--                     p_block_height,
--                     l_rewarded_invitee.epoch_height);
--         end loop;
-- END
-- $$;

CREATE OR REPLACE PROCEDURE save_saved_invite_rewards(p_saved_invite_rewards tp_saved_invite_rewards[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_saved_invite_rewards tp_saved_invite_rewards;
BEGIN
    for i in 1..cardinality(p_saved_invite_rewards)
        loop
            l_saved_invite_rewards = p_saved_invite_rewards[i];
            insert into saved_invite_rewards (ei_address_state_id, reward_type, count)
            values ((select address_state_id
                     from cur_epoch_identities
                     where lower(address) = lower(l_saved_invite_rewards.address)),
                    l_saved_invite_rewards.reward_type,
                    l_saved_invite_rewards.count);
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_reported_flip_rewards(p_epoch bigint,
                                                       p_reported_flip_rewards tp_reported_flip_reward[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_reported_flip_reward tp_reported_flip_reward;
    l_flip_tx_id           bigint;
BEGIN
    for i in 1..cardinality(p_reported_flip_rewards)
        loop
            l_reported_flip_reward = p_reported_flip_rewards[i];
            select tx_id into l_flip_tx_id from flips where lower(cid) = lower(l_reported_flip_reward.cid);
            if l_flip_tx_id is null then
                continue;
            end if;
            insert into reported_flip_rewards (ei_address_state_id, address_id, epoch, flip_tx_id, balance, stake)
            values ((select address_state_id
                     from cur_epoch_identities
                     where lower(address) = lower(l_reported_flip_reward.address)),
                    (select id
                     from addresses
                     where lower(address) = lower(l_reported_flip_reward.address)),
                    p_epoch,
                    l_flip_tx_id,
                    l_reported_flip_reward.balance,
                    l_reported_flip_reward.stake);
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_epoch_rewards_bounds(p_epoch bigint,
                                                      p_block_height bigint,
                                                      p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item           jsonb;
    l_min_address_id bigint;
    l_max_address_id bigint;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            l_min_address_id = get_address_id_or_insert(p_block_height, (l_item ->> 'minAddress')::text);
            l_max_address_id = get_address_id_or_insert(p_block_height, (l_item ->> 'maxAddress')::text);
            INSERT INTO epoch_reward_bounds (epoch, bound_type, min_amount, min_address_id, max_amount, max_address_id)
            VALUES (p_epoch, (l_item ->> 'boundType')::smallint, (l_item ->> 'minAmount')::numeric, l_min_address_id,
                    (l_item ->> 'maxAmount')::numeric,
                    l_max_address_id);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_epoch_flip_statuses(p_epoch bigint,
                                                     p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item jsonb;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            INSERT INTO epoch_flip_statuses (epoch, flip_status, count)
            VALUES (p_epoch, (l_item ->> 'status')::smallint, (l_item ->> 'count')::integer);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_pool_size_history(p_epoch bigint,
                                                   p_block_height bigint,
                                                   p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item jsonb;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            INSERT INTO pool_size_history (address_id, epoch, validation_size, validation_delegators, end_size,
                                           end_delegators)
            VALUES (get_address_id_or_insert(p_block_height, (l_item ->> 'address')::text), p_epoch,
                    (l_item -> 'old' ->> 'size')::integer, (l_item -> 'old' ->> 'totalDelegated')::integer,
                    (l_item -> 'new' ->> 'size')::integer, (l_item -> 'new' ->> 'totalDelegated')::integer);
        end loop;
END
$$;