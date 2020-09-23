CREATE OR REPLACE PROCEDURE generate_epoch_summaries(p_epoch bigint,
                                                     p_block_height bigint,
                                                     p_min_score_for_invite real)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_validated_count integer;
    l_start           timestamp;
    l_end             timestamp;
BEGIN
    select clock_timestamp() into l_start;
    select count(*)
    into l_validated_count
    from cur_epoch_identities ei
             join address_states s on s.id = ei.address_state_id
         -- 'Verified', 'Newbie', 'Human'
    where s.state in (3, 7, 8);
    call update_epoch_summary(p_block_height => p_block_height,
                              p_validated_count => l_validated_count,
                              p_min_score_for_invite => p_min_score_for_invite);
    select clock_timestamp() into l_end;
    call log_performance('update_epoch_summary', l_start, l_end);

    select clock_timestamp() into l_start;
    call generate_epoch_flips_summary(p_epoch);
    select clock_timestamp() into l_end;
    call log_performance('generate_epoch_flips_summary', l_start, l_end);

    select clock_timestamp() into l_start;
    call update_epoch_identities_rewards(p_epoch);
    select clock_timestamp() into l_end;
    call log_performance('update_epoch_identities_rewards', l_start, l_end);
END
$$;

CREATE OR REPLACE PROCEDURE generate_epoch_flips_summary(p_epoch bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
BEGIN
    insert into flip_summaries (
        select f.tx_id,
               COALESCE(ww.cnt, 0),
               COALESCE(short.answers, 0),
               COALESCE(long.answers, 0),
               false
        from flips f
                 join transactions t on t.id = f.tx_id
                 join blocks b on b.height = t.block_height and b.epoch = p_epoch
                 left join (select a.flip_tx_id, count(*) answers
                            from answers a
                            where a.is_short
                            group by a.flip_tx_id) short
                           on short.flip_tx_id = f.tx_id
                 left join (select a.flip_tx_id, count(*) answers
                            from answers a
                            where not a.is_short
                            group by a.flip_tx_id) long
                           on long.flip_tx_id = f.tx_id
                 left join (select a.flip_tx_id, count(*) cnt
                            from answers a
                            where not a.is_short
                              and a.grade = 1 -- reported
                            group by a.flip_tx_id) ww
                           on ww.flip_tx_id = f.tx_id
        where f.delete_tx_id is null
    );
END
$$;

CREATE OR REPLACE PROCEDURE update_epoch_identities_rewards(p_epoch bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
BEGIN
    update epoch_identities
    set total_validation_reward = total_rewards.reward
    from (
             select vr.ei_address_state_id, sum(vr.balance + vr.stake) reward
             from validation_rewards vr
                      join epoch_identities ei on vr.ei_address_state_id = ei.address_state_id and ei.epoch = p_epoch
             group by ei_address_state_id
         ) total_rewards
    where total_rewards.ei_address_state_id = epoch_identities.address_state_id;
END
$$;

CREATE OR REPLACE PROCEDURE update_epoch_summary(p_block_height bigint,
                                                 p_validated_count integer DEFAULT null,
                                                 p_block_count_diff bigint DEFAULT null,
                                                 p_empty_block_count_diff bigint DEFAULT null,
                                                 p_tx_count_diff bigint DEFAULT null,
                                                 p_invite_count_diff bigint DEFAULT null,
                                                 p_flip_count_diff integer DEFAULT null,
                                                 p_burnt_diff numeric DEFAULT null,
                                                 p_minted_diff numeric DEFAULT null,
                                                 p_total_balance numeric DEFAULT null,
                                                 p_total_stake numeric DEFAULT null,
                                                 p_min_score_for_invite real DEFAULT null)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_epoch       bigint;
    l_check_epoch bigint;
BEGIN
    select epoch into l_epoch from blocks where height = p_block_height;
    update epoch_summaries
    set validated_count      = coalesce(p_validated_count, validated_count),
        block_count          = block_count + coalesce(p_block_count_diff, 0),
        empty_block_count    = empty_block_count + coalesce(p_empty_block_count_diff, 0),
        tx_count             = tx_count + coalesce(p_tx_count_diff, 0),
        invite_count         = invite_count + coalesce(p_invite_count_diff, 0),
        flip_count           = flip_count + coalesce(p_flip_count_diff, 0),
        burnt                = burnt + coalesce(p_burnt_diff, 0),
        minted               = minted + coalesce(p_minted_diff, 0),
        total_balance        = coalesce(p_total_balance, total_balance),
        total_stake          = coalesce(p_total_stake, total_stake),
        block_height         = p_block_height,
        min_score_for_invite = coalesce(p_min_score_for_invite, min_score_for_invite)
    where epoch = l_epoch
    RETURNING epoch into l_check_epoch;

    if l_check_epoch is null then
        insert into epoch_summaries (epoch,
                                     validated_count,
                                     block_count,
                                     empty_block_count,
                                     tx_count,
                                     invite_count,
                                     flip_count,
                                     burnt,
                                     minted,
                                     total_balance,
                                     total_stake,
                                     block_height,
                                     min_score_for_invite)
        values (l_epoch,
                coalesce(p_validated_count, 0),
                coalesce(p_block_count_diff, 0),
                coalesce(p_empty_block_count_diff, 0),
                coalesce(p_tx_count_diff, 0),
                coalesce(p_invite_count_diff, 0),
                coalesce(p_flip_count_diff, 0),
                coalesce(p_burnt_diff, 0),
                coalesce(p_minted_diff, 0),
                coalesce(p_total_balance, 0),
                coalesce(p_total_stake, 0),
                p_block_height,
                coalesce(p_min_score_for_invite, 0));
    end if;
END
$$;

CREATE OR REPLACE PROCEDURE restore_epoch_summary(p_block_height bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_epoch         bigint;
    l_total_balance numeric;
    l_total_stake   numeric;
    l_exists        boolean;
BEGIN

    select epoch into l_epoch from blocks where height = p_block_height;
    if l_epoch is null then
        return;
    end if;

    l_exists = (select exists(select 1 from epoch_summaries where epoch = l_epoch));
    if l_exists then
        return;
    end if;

    select c.total_balance, c.total_stake
    into l_total_balance, l_total_stake
    from coins c
             join blocks b on b.height = c.block_height and b.epoch = l_epoch;

    insert into epoch_summaries
    (epoch,
     validated_count,
     block_count,
     empty_block_count,
     tx_count,
     invite_count,
     flip_count,
     burnt,
     minted,
     total_balance,
     total_stake,
     block_height,
     min_score_for_invite)
    values (l_epoch,
            0,
            (select count(*)
             from blocks b
             where b.epoch = l_epoch),
            (select count(*)
             from blocks b
             where b.epoch = l_epoch
               and b.is_empty),
            (select count(*)
             from transactions t
                      join blocks b on t.block_height = b.height and b.epoch = l_epoch),
            (select count(*)
             from transactions t
                      join blocks b on t.block_height = b.height and b.epoch = l_epoch
                  -- 'InviteTx'
             where t.type = 2),
            (select count(*)
             from flips f
                      join transactions t on t.id = f.tx_id
                      join blocks b on b.height = t.block_height and b.epoch = l_epoch
             where f.delete_tx_id is null),
            (select coalesce(sum(burnt), 0)
             from coins c
                      join blocks b on b.height = c.block_height
             where b.epoch = l_epoch),
            (select coalesce(sum(minted), 0)
             from coins c
                      join blocks b on b.height = c.block_height
             where b.epoch = l_epoch),
            l_total_balance,
            l_total_stake,
            p_block_height,
            0);
END
$$;