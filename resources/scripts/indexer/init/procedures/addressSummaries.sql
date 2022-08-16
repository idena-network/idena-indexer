CREATE OR REPLACE PROCEDURE update_address_summary(p_address_id bigint,
                                                   p_flips_diff integer DEFAULT null,
                                                   p_wrong_words_flips_diff integer DEFAULT null)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_check_address_id bigint;
BEGIN
    UPDATE address_summaries
    SET flips             = flips + coalesce(p_flips_diff, 0),
        wrong_words_flips = wrong_words_flips + coalesce(p_wrong_words_flips_diff, 0)
    WHERE address_id = p_address_id
    RETURNING address_id INTO l_check_address_id;

    if l_check_address_id is null then
        INSERT INTO address_summaries (address_id, flips, wrong_words_flips)
        VALUES (p_address_id, coalesce(p_flips_diff, 0), coalesce(p_wrong_words_flips_diff, 0));
    end if;
END
$$;

CREATE OR REPLACE PROCEDURE restore_address_summaries()
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_epoch bigint;
    rec     record;
BEGIN
    SELECT max(epoch) INTO l_epoch FROM epoch_identities;
    INSERT INTO address_summaries (SELECT address_id, made_flips, wrong_words_flips
                                   FROM (SELECT s.address_id,
                                                coalesce(sum(ei.made_flips), 0)        made_flips,
                                                coalesce(sum(ei.wrong_words_flips), 0) wrong_words_flips
                                         FROM epoch_identities ei
                                                  JOIN address_states s ON s.id = ei.address_State_id
                                         GROUP BY s.address_id) t
                                   WHERE t.made_flips > 0
                                      OR t.wrong_words_flips > 0);

    for rec in (SELECT t.from, count(*) cnt
                FROM flips f
                         JOIN transactions t ON t.id = f.tx_id
                         JOIN blocks b ON b.height = t.block_height AND b.epoch > l_epoch
                WHERE f.delete_tx_id is null
                GROUP BY t.from)
        loop
            CALL update_address_summary(p_address_id => rec."from", p_flips_diff => rec.cnt::integer);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE update_balance_update_summary(p_block_height bigint,
                                                          p_balance_update tp_balance_update)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CHANGE_TYPE_BALANCE_UPDATE_SUMMARIES CONSTANT smallint = 5;
    l_balance_diff                                numeric;
    l_prev_balance_in                             numeric;
    l_balance_in                                  numeric;
    l_prev_balance_out                            numeric;
    l_balance_out                                 numeric;
    l_stake_diff                                  numeric;
    l_prev_stake_in                               numeric;
    l_stake_in                                    numeric;
    l_prev_stake_out                              numeric;
    l_stake_out                                   numeric;
    l_penalty_diff                                numeric;
    l_prev_penalty_in                             numeric;
    l_penalty_in                                  numeric;
    l_prev_penalty_out                            numeric;
    l_penalty_out                                 numeric;
    l_address_id                                  bigint;
    l_change_id                                   bigint;
BEGIN
    l_balance_diff = coalesce(p_balance_update.balance_new, 0) - coalesce(p_balance_update.balance_old, 0);
    l_stake_diff = coalesce(p_balance_update.stake_new, 0) - coalesce(p_balance_update.stake_old, 0);
    l_penalty_diff = coalesce(p_balance_update.penalty_new, 0) - coalesce(p_balance_update.penalty_old, 0);

    if l_balance_diff > 0 then
        l_balance_in = l_balance_diff;
    end if;

    if l_balance_diff < 0 then
        l_balance_out = -l_balance_diff;
    end if;

    if l_stake_diff > 0 then
        l_stake_in = l_stake_diff;
    end if;

    if l_stake_diff < 0 then
        l_stake_out = -l_stake_diff;
    end if;

    if l_penalty_diff > 0 then
        l_penalty_in = l_penalty_diff;
    end if;

    if l_penalty_diff < 0 then
        l_penalty_out = -l_penalty_diff;
    end if;

    if l_balance_in is null and l_balance_out is null and l_stake_in is null and l_stake_out is null and
       l_penalty_in is null and l_penalty_out is null then
        return;
    end if;

    l_address_id = get_address_id_or_insert(p_block_height, p_balance_update.address);

    SELECT balance_in, balance_out, stake_in, stake_out, penalty_in, penalty_out
    INTO l_prev_balance_in, l_prev_balance_out, l_prev_stake_in, l_prev_stake_out, l_prev_penalty_in, l_prev_penalty_out
    FROM balance_update_summaries
    WHERE address_id = l_address_id;

    if l_prev_balance_in is null then
        INSERT INTO balance_update_summaries (address_id, balance_in, balance_out, stake_in, stake_out, penalty_in,
                                              penalty_out)
        VALUES (l_address_id, coalesce(l_balance_in, 0), coalesce(l_balance_out, 0), coalesce(l_stake_in, 0),
                coalesce(l_stake_out, 0), coalesce(l_penalty_in, 0), coalesce(l_penalty_out, 0));
    else
        UPDATE balance_update_summaries
        SET balance_in  = balance_in + coalesce(l_balance_in, 0),
            balance_out = balance_out + coalesce(l_balance_out, 0),
            stake_in    = stake_in + coalesce(l_stake_in, 0),
            stake_out   = stake_out + coalesce(l_stake_out, 0),
            penalty_in  = penalty_in + coalesce(l_penalty_in, 0),
            penalty_out = penalty_out + coalesce(l_penalty_out, 0)
        WHERE address_id = l_address_id;
    end if;

    INSERT INTO changes (block_height, "type")
    VALUES (p_block_height, CHANGE_TYPE_BALANCE_UPDATE_SUMMARIES)
    RETURNING id INTO l_change_id;

    INSERT INTO balance_update_summaries_changes (change_id, address_id, balance_in, balance_out, stake_in, stake_out,
                                                  penalty_in, penalty_out)
    VALUES (l_change_id, l_address_id, l_prev_balance_in, l_prev_balance_out, l_prev_stake_in, l_prev_stake_out,
            l_prev_penalty_in, l_prev_penalty_out);
END
$$;

CREATE OR REPLACE PROCEDURE reset_balance_update_summaries_changes(p_change_id bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_address_id  bigint;
    l_balance_in  bigint;
    l_balance_out bigint;
    l_stake_in    bigint;
    l_stake_out   bigint;
    l_penalty_in  bigint;
    l_penalty_out bigint;
BEGIN
    SELECT address_id, balance_in, balance_out, stake_in, stake_out, penalty_in, penalty_out
    INTO l_address_id, l_balance_in , l_balance_out, l_stake_in , l_stake_out , l_penalty_in , l_penalty_out
    FROM balance_update_summaries_changes
    WHERE change_id = p_change_id;

    if l_balance_in is null then
        DELETE
        FROM balance_update_summaries
        WHERE address_id = l_address_id;
    else
        UPDATE balance_update_summaries
        SET balance_in  = l_balance_in,
            balance_out = l_balance_out,
            stake_in    = l_stake_in,
            stake_out   = l_stake_out,
            penalty_in  = l_penalty_in,
            penalty_out = l_penalty_out
        WHERE address_id = l_address_id;
    end if;

    DELETE FROM balance_update_summaries_changes WHERE change_id = p_change_id;
END
$$;