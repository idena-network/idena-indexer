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
                                              p_fund_rewards tp_epoch_reward[],
                                              p_rewarded_flip_cids text[],
                                              p_rewarded_invitations tp_rewarded_invitation[],
                                              p_saved_invite_rewards tp_saved_invite_rewards[],
                                              p_reported_flip_rewards tp_reported_flip_reward[],
                                              p_failed_validation boolean,
                                              p_min_score_for_invite real)
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_start timestamp;
    l_end   timestamp;
BEGIN
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
    call save_epoch_rewards(p_epoch, p_height, p_bad_authors, p_total, p_validation_rewards, p_ages, p_fund_rewards,
                            p_rewarded_flip_cids,
                            p_rewarded_invitations,
                            p_saved_invite_rewards,
                            p_reported_flip_rewards);
    select clock_timestamp() into l_end;
    call log_performance('save_epoch_rewards', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_failed_validation then
        insert into failed_validations (block_height) values (p_height);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('failed_validations', l_start, l_end);

    select clock_timestamp() into l_start;
    call generate_epoch_summaries(p_epoch, p_height, p_min_score_for_invite);
    select clock_timestamp() into l_end;
    call log_performance('generate_epoch_summaries', l_start, l_end);

    select clock_timestamp() into l_start;
    call update_flips_queue();
    select clock_timestamp() into l_end;
    call log_performance('update_flips_queue', l_start, l_end);
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_birthdays(p_birthdays tp_birthday[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    birthday tp_birthday;
BEGIN
    for i in 1..cardinality(p_birthdays)
        loop
            birthday := p_birthdays[i];
            insert into birthdays (address_id, birth_epoch)
            values ((select id from addresses where lower(address) = lower(birthday.address)), birthday.birth_epoch)
            on conflict (address_id) do update set birth_epoch=birthday.birth_epoch;
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
    identity        tp_epoch_identity;
    l_address_id    bigint;
    l_prev_state_id bigint;
    l_state_id      bigint;
BEGIN

    CREATE TEMP TABLE cur_epoch_identities
    (
        address          character(42),
        address_state_id bigint
    ) ON COMMIT DROP;
    CREATE UNIQUE INDEX ON cur_epoch_identities (lower(address));

    for i in 1..cardinality(p_identities)
        loop
            identity := p_identities[i];

            select id into l_address_id from addresses where lower(address) = lower(identity.address);

            update address_states
            set is_actual = false
            where address_id = l_address_id
              and is_actual
            returning id into l_prev_state_id;

            insert into address_states (address_id, state, is_actual, block_height, prev_id)
            values (l_address_id, identity.state, true, p_height, l_prev_state_id)
            returning id into l_state_id;

            insert into epoch_identities (epoch, address_state_id, short_point, short_flips, total_short_point,
                                          total_short_flips, long_point, long_flips, approved, missed,
                                          required_flips, available_flips, made_flips, next_epoch_invites, birth_epoch,
                                          total_validation_reward, short_answers, long_answers, wrong_words_flips)
            values (p_epoch, l_state_id, identity.short_point, identity.short_flips, identity.total_short_point,
                    identity.total_short_flips, identity.long_point, identity.long_flips, identity.approved,
                    identity.missed, identity.required_flips, identity.available_flips, identity.made_flips,
                    identity.next_epoch_invites, identity.birth_epoch, 0, identity.short_answers, identity.long_answers,
                    identity.wrong_words_flips);

            if identity.wrong_words_flips > 0 then
                CALL update_address_summary(p_address_id => l_address_id,
                                            p_wrong_words_flips_diff => identity.wrong_words_flips);
            end if;

            insert into cur_epoch_identities values (identity.address, l_state_id);

            if l_prev_state_id is not null then
                insert into epoch_identity_interim_states (address_state_id, block_height)
                values (l_prev_state_id, p_height);
            end if;

        end loop;

    insert into epoch_identity_interim_states (
        select s.id, p_height
        from address_states s
                 join blocks b on b.height = s.block_height and b.epoch = p_epoch
                 left join temporary_identities ti on ti.address_id = s.address_id
                 left join cur_epoch_identities ei on ei.address_state_id = s.id
        where s.is_actual
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
DECLARE
    l_flip_to_solve    tp_flip_to_solve;
    l_address_state_id bigint;
    l_flip_tx_id       bigint;
BEGIN
    for i in 1..cardinality(p_flips_to_solve)
        loop
            l_flip_to_solve := p_flips_to_solve[i];

            if char_length(l_flip_to_solve.address) > 0 then
                select address_state_id
                into l_address_state_id
                from cur_epoch_identities
                where lower(address) = lower(l_flip_to_solve.address);
            end if;

            select tx_id into l_flip_tx_id from flips where lower(cid) = lower(l_flip_to_solve.cid);
            if l_flip_tx_id is null then
                continue;
            end if;

            insert into flips_to_solve (ei_address_state_id, flip_tx_id, is_short)
            values (l_address_state_id,
                    l_flip_tx_id,
                    l_flip_to_solve.is_short);
        end loop;
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
    answer       tp_answer;
    state        tp_flip_state;
    l_flip_tx_id bigint;
BEGIN
    if answers is not null then
        for i in 1..cardinality(answers)
            loop
                answer := answers[i];
                IF char_length(answer.flip_cid) > 0 THEN
                    select tx_id into l_flip_tx_id from flips where lower(cid) = lower(answer.flip_cid);
                end if;
                if l_flip_tx_id is null then
                    continue;
                end if;
                INSERT INTO ANSWERS (FLIP_TX_ID, ei_address_state_id, IS_SHORT, ANSWER, POINT, GRADE)
                VALUES (l_flip_tx_id,
                        (select address_state_id
                         from cur_epoch_identities
                         where lower(address) = lower(answer.address)),
                        answer.is_short, answer.answer, answer.point, answer.grade);
            end loop;
    end if;

    if states is not null then
        for i in 1..cardinality(states)
            loop
                state := states[i];
                UPDATE FLIPS
                SET STATUS=state.status,
                    ANSWER=state.answer,
                    GRADE=state.grade,
                    STATUS_BLOCK_HEIGHT=block_height
                WHERE lower(CID) = lower(state.flip_cid);
            end loop;
    end if;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_epoch_rewards(p_epoch bigint,
                                               p_block_height bigint,
                                               p_bad_authors tp_bad_author[],
                                               p_total tp_total_epoch_reward,
                                               p_validation_rewards tp_epoch_reward[],
                                               p_ages tp_reward_age[],
                                               p_fund_rewards tp_epoch_reward[],
                                               p_rewarded_flip_cids text[],
                                               p_rewarded_invitations tp_rewarded_invitation[],
                                               p_saved_invite_rewards tp_saved_invite_rewards[],
                                               p_reported_flip_rewards tp_reported_flip_reward[])
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
        call save_total_reward(p_block_height, p_total);
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
    if p_fund_rewards is not null then
        call save_fund_rewards(p_block_height, p_fund_rewards);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_fund_rewards', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_rewarded_flip_cids is not null then
        call save_rewarded_flips(p_rewarded_flip_cids);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_rewarded_flips', l_start, l_end);

    select clock_timestamp() into l_start;
    if p_rewarded_invitations is not null then
        call save_rewarded_invitations(p_block_height, p_rewarded_invitations);
    end if;
    select clock_timestamp() into l_end;
    call log_performance('save_rewarded_invitations', l_start, l_end);

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

CREATE OR REPLACE PROCEDURE save_total_reward(p_block_height bigint,
                                              p_total tp_total_epoch_reward)
    LANGUAGE 'plpgsql'
AS
$BODY$
BEGIN
    insert into total_rewards (block_height,
                               total, validation,
                               flips,
                               invitations,
                               foundation,
                               zero_wallet,
                               validation_share,
                               flips_share,
                               invitations_share)
    values (p_block_height,
            p_total.total,
            p_total.validation,
            p_total.flips,
            p_total.invitations,
            p_total.foundation,
            p_total.zero_wallet,
            p_total.validation_share,
            p_total.flips_share,
            p_total.invitations_share);
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

CREATE OR REPLACE PROCEDURE save_rewarded_flips(p_rewarded_flip_cids text[])
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
            insert into rewarded_flips (flip_tx_id)
            values (l_flip_tx_id);
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
            insert into rewarded_invitations (invite_tx_id, block_height, reward_type)
            values (l_invite_tx_id,
                    p_block_height,
                    l_rewarded_invitation.reward_type);
        end loop;
END
$BODY$;

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