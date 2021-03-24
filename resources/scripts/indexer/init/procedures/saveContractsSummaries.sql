CREATE OR REPLACE FUNCTION calculate_oracle_voting_contract_sort_key(p_estimated_oracle_reward numeric, p_id bigint)
    RETURNS character(68)
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    return to_char(p_estimated_oracle_reward, 'FM000000000000000000000000000000.000000000000000000') ||
           to_char(p_id, 'FM0000000000000000000');
END
$$;

CREATE OR REPLACE FUNCTION calculate_estimated_oracle_reward(p_balance numeric, p_id bigint)
    RETURNS numeric(30, 18)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_owner_fee               smallint;
    l_fee_rate                double precision;
    l_committee_size          bigint;
    l_voting_min_payment      numeric;
    l_estimated_oracle_reward numeric(30, 18);
BEGIN
    SELECT committee_size, owner_fee, voting_min_payment
    INTO l_committee_size, l_owner_fee, l_voting_min_payment
    FROM oracle_voting_contracts
    WHERE contract_tx_id = p_id;

    if l_committee_size = 0 then
        l_committee_size = 1;
    end if;

    if l_voting_min_payment is null then
        SELECT voting_min_payment
        INTO l_voting_min_payment
        FROM oracle_voting_contract_call_starts
        WHERE ov_contract_tx_id = p_id;
    end if;

    if l_voting_min_payment is null then
        l_voting_min_payment = 0;
    end if;

    l_fee_rate = (l_owner_fee::double precision) / 100;

    l_estimated_oracle_reward = p_balance * (1.0 - l_fee_rate) / l_committee_size + l_fee_rate * l_voting_min_payment;
    return l_estimated_oracle_reward;
END
$$;

CREATE OR REPLACE PROCEDURE update_sorted_oracle_voting_contract_state(p_block_height bigint,
                                                                       p_contract_tx_id bigint,
                                                                       p_state smallint,
                                                                       p_state_tx_id bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    SOVC_STATE_TERMINATED CONSTANT smallint = 4;
BEGIN
    call update_sorted_oracle_voting_contracts(p_block_height, p_contract_tx_id, null, p_state, p_state_tx_id, null,
                                               null);
    if p_state = SOVC_STATE_TERMINATED then
        call delete_not_voted_committee_from_sovcc(p_block_height, p_contract_tx_id);
    end if;
    call update_sorted_oracle_voting_contract_committees(p_block_height, p_contract_tx_id, null, null, p_state,
                                                         p_state_tx_id, null);
END
$$;

CREATE OR REPLACE PROCEDURE apply_prolongation_on_sorted_contracts(p_block_height bigint,
                                                                   p_prolongation_tx_id bigint,
                                                                   p_contract_tx_id bigint,
                                                                   p_start_block bigint,
                                                                   p_epoch bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    SOVC_STATE_VOTING CONSTANT smallint = 1;
    l_voting_duration          bigint;
    l_balance                  numeric;
    l_estimated_oracle_reward  numeric;
    l_sort_key                 text;
BEGIN

    if p_start_block is not null then
        SELECT voting_duration
        INTO l_voting_duration
        FROM oracle_voting_contracts
        WHERE contract_tx_id = p_contract_tx_id;

        SELECT balance
        INTO l_balance
        FROM balances
        WHERE address_id = (SELECT contract_address_id FROM contracts WHERE tx_id = p_contract_tx_id);
        l_balance = coalesce(l_balance, 0);
        l_estimated_oracle_reward = calculate_estimated_oracle_reward(l_balance, p_contract_tx_id);
        l_sort_key = calculate_oracle_voting_contract_sort_key(l_estimated_oracle_reward, p_contract_tx_id);

        call update_sorted_oracle_voting_contract_committees(p_block_height, p_contract_tx_id, null, l_sort_key,
                                                             SOVC_STATE_VOTING, p_prolongation_tx_id, null);

        call update_sorted_oracle_voting_contracts(p_block_height, p_contract_tx_id, l_sort_key, SOVC_STATE_VOTING,
                                                   p_prolongation_tx_id, p_start_block + l_voting_duration, p_epoch);
    else

        call update_sorted_oracle_voting_contracts(p_block_height, p_contract_tx_id, null, null, null, null, p_epoch);

    end if;

END
$$;

CREATE OR REPLACE PROCEDURE update_sorted_oracle_voting_contracts(p_block_height bigint,
                                                                  p_contract_tx_id bigint,
                                                                  p_sort_key text,
                                                                  p_state smallint,
                                                                  p_state_tx_id bigint,
                                                                  p_counting_block bigint,
                                                                  p_epoch bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CHANGE_TYPE_SORTED_ORACLE_VOTING CONSTANT smallint = 3;
    SOVC_STATE_PENDING               CONSTANT smallint = 0;
    SOVC_STATE_VOTING                CONSTANT smallint = 1;
    l_change_id                               bigint;
    l_reset_sort_key                          boolean;
BEGIN

    if not (SELECT exists(SELECT 1
                          FROM sorted_oracle_voting_contracts
                          WHERE contract_tx_id = p_contract_tx_id)) then
        return;
    end if;

    INSERT INTO changes (block_height, type)
    VALUES (p_block_height, CHANGE_TYPE_SORTED_ORACLE_VOTING)
    RETURNING id INTO l_change_id;

    INSERT INTO sorted_oracle_voting_contracts_changes (change_id, contract_tx_id, sort_key, state, state_tx_id,
                                                        counting_block, epoch)
        (SELECT l_change_id, p_contract_tx_id, sort_key, state, state_tx_id, counting_block, epoch
         FROM sorted_oracle_voting_contracts
         WHERE contract_tx_id = p_contract_tx_id);

    l_reset_sort_key =
                p_sort_key IS NULL AND p_state IS NOT NULL AND p_state NOT IN (SOVC_STATE_PENDING, SOVC_STATE_VOTING);

    UPDATE sorted_oracle_voting_contracts
    SET sort_key       = CASE WHEN l_reset_sort_key THEN NULL ELSE coalesce(p_sort_key, sort_key) END,
        state          = coalesce(p_state, state),
        state_tx_id    = coalesce(p_state_tx_id, state_tx_id),
        counting_block = coalesce(p_counting_block, counting_block),
        epoch          = coalesce(p_epoch, epoch)
    WHERE contract_tx_id = p_contract_tx_id;
END
$$;

CREATE OR REPLACE PROCEDURE save_oracle_voting_committee(p_block_height bigint,
                                                         p_ov_contract_tx_id bigint,
                                                         p_committee jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_address_id        bigint;
    l_sort_key          text;
    l_state             smallint;
    l_state_tx_id       bigint;
    l_author_address_id bigint;
BEGIN

    call delete_cur_not_voted_committee_from_sovcc(p_block_height, p_ov_contract_tx_id);

    if p_committee is null then
        return;
    end if;

    for j in 0..jsonb_array_length(p_committee) - 1
        loop
            SELECT id INTO l_address_id FROM addresses WHERE lower(address) = lower((p_committee ->> j)::text);

            SELECT sort_key, state, state_tx_id, author_address_id
            INTO l_sort_key, l_state, l_state_tx_id, l_author_address_id
            FROM sorted_oracle_voting_contracts
            WHERE contract_tx_id = p_ov_contract_tx_id;

            call insert_sorted_oracle_voting_contract_committees(p_block_height, p_ov_contract_tx_id,
                                                                 l_author_address_id, l_address_id, l_sort_key, l_state,
                                                                 l_state_tx_id, false);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE insert_sorted_oracle_voting_contract_committees(p_block_height bigint,
                                                                            p_contract_tx_id bigint,
                                                                            p_author_address_id bigint,
                                                                            p_address_id bigint,
                                                                            p_sort_key text,
                                                                            p_state smallint,
                                                                            p_state_tx_id bigint,
                                                                            p_voted boolean)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CHANGE_TYPE_SORTED_ORACLE_VOTING_COMMITTEE CONSTANT smallint = 4;
    l_change_id                                         bigint;
BEGIN
    if (SELECT exists(SELECT 1
                      FROM sorted_oracle_voting_contract_committees
                      WHERE contract_tx_id = p_contract_tx_id
                        AND address_id = p_address_id)) then
        return;
    end if;

    INSERT INTO changes (block_height, type)
    VALUES (p_block_height, CHANGE_TYPE_SORTED_ORACLE_VOTING_COMMITTEE)
    RETURNING id INTO l_change_id;

    INSERT INTO sorted_oracle_voting_contract_committees_changes (change_id, contract_tx_id, author_address_id,
                                                                  address_id, sort_key, state, state_tx_id, voted,
                                                                  deleted)
    VALUES (l_change_id, p_contract_tx_id, p_author_address_id, p_address_id, null, null, null, null, null);

    INSERT INTO sorted_oracle_voting_contract_committees (contract_tx_id, author_address_id, sort_key, state,
                                                          state_tx_id, address_id, voted)
    VALUES (p_contract_tx_id, p_author_address_id, p_sort_key, p_state, p_state_tx_id, p_address_id, p_voted);
END
$$;

CREATE OR REPLACE PROCEDURE update_sorted_oracle_voting_contract_committees(p_block_height bigint,
                                                                            p_contract_tx_id bigint,
                                                                            p_address_id bigint,
                                                                            p_sort_key text,
                                                                            p_state smallint,
                                                                            p_state_tx_id bigint,
                                                                            p_voted boolean)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    SOVC_STATE_PENDING                         CONSTANT smallint = 0;
    SOVC_STATE_VOTING                          CONSTANT smallint = 1;
    SOVCC_STATE_VOTED                          CONSTANT smallint = 5;
    CHANGE_TYPE_SORTED_ORACLE_VOTING_COMMITTEE CONSTANT smallint = 4;
    l_change_id                                         bigint;
    l_reset_sort_key                                    boolean;
BEGIN

    if not (SELECT exists(SELECT 1
                          FROM sorted_oracle_voting_contract_committees
                          WHERE contract_tx_id = p_contract_tx_id
                            AND (p_address_id is null or address_id = p_address_id))) then
        return;
    end if;

    INSERT INTO changes (block_height, type)
    VALUES (p_block_height, CHANGE_TYPE_SORTED_ORACLE_VOTING_COMMITTEE)
    RETURNING id INTO l_change_id;

    INSERT INTO sorted_oracle_voting_contract_committees_changes (change_id, contract_tx_id, author_address_id,
                                                                  address_id, sort_key, state, state_tx_id, voted,
                                                                  deleted)
        (SELECT l_change_id,
                p_contract_tx_id,
                author_address_id,
                address_id,
                sort_key,
                state,
                state_tx_id,
                voted,
                false
         FROM sorted_oracle_voting_contract_committees
         WHERE contract_tx_id = p_contract_tx_id
           AND (p_address_id is null or address_id = p_address_id));

    l_reset_sort_key =
                p_sort_key IS NULL AND p_state IS NOT NULL AND p_state NOT IN (SOVC_STATE_PENDING, SOVC_STATE_VOTING);

    UPDATE sorted_oracle_voting_contract_committees
    SET sort_key    = CASE
                          WHEN l_reset_sort_key OR state = SOVCC_STATE_VOTED THEN NULL
                          ELSE coalesce(p_sort_key, sort_key) END,
        state       = CASE
                          WHEN p_state = SOVC_STATE_VOTING AND voted THEN SOVCC_STATE_VOTED
                          ELSE coalesce(p_state, state) END,
        state_tx_id = coalesce(p_state_tx_id, state_tx_id),
        voted       = coalesce(p_voted, voted)
    WHERE contract_tx_id = p_contract_tx_id
      AND (p_address_id is null or address_id = p_address_id);
END
$$;

CREATE OR REPLACE PROCEDURE delete_cur_not_voted_committee_from_sovcc(p_block_height bigint,
                                                                      p_contract_tx_id bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CHANGE_TYPE_SORTED_ORACLE_VOTING_COMMITTEE CONSTANT smallint = 4;
    l_change_id                                         bigint;
BEGIN

    if not (SELECT exists(SELECT 1
                          FROM sorted_oracle_voting_contract_committees
                          WHERE contract_tx_id = p_contract_tx_id
                            AND state = 1)) then
        return;
    end if;

    INSERT INTO changes (block_height, type)
    VALUES (p_block_height, CHANGE_TYPE_SORTED_ORACLE_VOTING_COMMITTEE)
    RETURNING id INTO l_change_id;

    INSERT INTO sorted_oracle_voting_contract_committees_changes (change_id, contract_tx_id, author_address_id,
                                                                  address_id, sort_key, state, state_tx_id, voted,
                                                                  deleted)
        (SELECT l_change_id,
                p_contract_tx_id,
                author_address_id,
                address_id,
                sort_key,
                state,
                state_tx_id,
                voted,
                true
         FROM sorted_oracle_voting_contract_committees
         WHERE contract_tx_id = p_contract_tx_id
           AND state = 1);

    DELETE
    FROM sorted_oracle_voting_contract_committees
    WHERE contract_tx_id = p_contract_tx_id
      AND state = 1;
END
$$;

CREATE OR REPLACE PROCEDURE delete_old_epoch_not_voted_committee_from_sovcc(p_block_height bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CHANGE_TYPE_SORTED_ORACLE_VOTING_COMMITTEE CONSTANT smallint = 4;
    SOVC_STATE_VOTING                          CONSTANT smallint = 1;
    l_change_id                                         bigint;
    l_cur_epoch                                         bigint;
BEGIN
    SELECT max(epoch) INTO l_cur_epoch FROM epochs;

    if NOT (SELECT exists(SELECT 1
                          FROM sorted_oracle_voting_contract_committees
                          WHERE contract_tx_id IN (SELECT contract_tx_id
                                                   FROM sorted_oracle_voting_contracts
                                                   WHERE state = SOVC_STATE_VOTING
                                                     AND epoch < l_cur_epoch)
                            AND NOT voted)) then
        return;
    end if;

    INSERT INTO changes (block_height, type)
    VALUES (p_block_height, CHANGE_TYPE_SORTED_ORACLE_VOTING_COMMITTEE)
    RETURNING id INTO l_change_id;

    INSERT INTO sorted_oracle_voting_contract_committees_changes (change_id, contract_tx_id, author_address_id,
                                                                  address_id, sort_key, state, state_tx_id, voted,
                                                                  deleted)
        (SELECT l_change_id,
                contract_tx_id,
                author_address_id,
                address_id,
                sort_key,
                state,
                state_tx_id,
                voted,
                true
         FROM sorted_oracle_voting_contract_committees
         WHERE contract_tx_id IN (SELECT contract_tx_id
                                  FROM sorted_oracle_voting_contracts
                                  WHERE state = SOVC_STATE_VOTING
                                    AND epoch < l_cur_epoch)
           AND NOT voted);

    DELETE
    FROM sorted_oracle_voting_contract_committees
    WHERE contract_tx_id IN (SELECT contract_tx_id
                             FROM sorted_oracle_voting_contracts
                             WHERE state = SOVC_STATE_VOTING
                               AND epoch < l_cur_epoch)
      AND NOT voted;
END
$$;

CREATE OR REPLACE PROCEDURE delete_not_voted_committee_from_sovcc(p_block_height bigint,
                                                                  p_contract_tx_id bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CHANGE_TYPE_SORTED_ORACLE_VOTING_COMMITTEE CONSTANT smallint = 4;
    l_change_id                                         bigint;
BEGIN
    if NOT (SELECT exists(SELECT 1
                          FROM sorted_oracle_voting_contract_committees
                          WHERE contract_tx_id = p_contract_tx_id
                            AND NOT voted)) then
        return;
    end if;

    INSERT INTO changes (block_height, type)
    VALUES (p_block_height, CHANGE_TYPE_SORTED_ORACLE_VOTING_COMMITTEE)
    RETURNING id INTO l_change_id;

    INSERT INTO sorted_oracle_voting_contract_committees_changes (change_id, contract_tx_id, author_address_id,
                                                                  address_id, sort_key, state, state_tx_id, voted,
                                                                  deleted)
        (SELECT l_change_id,
                contract_tx_id,
                author_address_id,
                address_id,
                sort_key,
                state,
                state_tx_id,
                voted,
                true
         FROM sorted_oracle_voting_contract_committees
         WHERE contract_tx_id = p_contract_tx_id
           AND NOT voted);

    DELETE
    FROM sorted_oracle_voting_contract_committees
    WHERE contract_tx_id = p_contract_tx_id
      AND NOT voted;
END
$$;

CREATE OR REPLACE PROCEDURE update_oracle_voting_contract_summaries(p_block_height bigint,
                                                                    p_contract_tx_id bigint,
                                                                    p_vote_proofs_diff bigint,
                                                                    p_votes_diff bigint,
                                                                    p_finish_timestamp bigint,
                                                                    p_termination_timestamp bigint,
                                                                    p_total_reward_diff numeric,
                                                                    p_stake_diff numeric,
                                                                    p_secret_votes_count bigint,
                                                                    p_epoch_without_growth smallint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CHANGE_TYPE_ORACLE_VOTING_SUMMARIES CONSTANT smallint = 2;
    l_change_id                                  bigint;
    l_prev_vote_proofs                           bigint;
    l_prev_votes                                 bigint;
    l_prev_finish_timestamp                      bigint;
    l_prev_termination_timestamp                 bigint;
    l_prev_total_reward                          numeric;
    l_new_total_reward                           numeric;
    l_prev_stake                                 numeric;
    l_prev_secret_votes_count                    bigint;
    l_prev_epoch_without_growth                  smallint;
BEGIN

    SELECT vote_proofs,
           votes,
           finish_timestamp,
           termination_timestamp,
           total_reward,
           stake,
           secret_votes_count,
           epoch_without_growth
    INTO l_prev_vote_proofs, l_prev_votes, l_prev_finish_timestamp, l_prev_termination_timestamp, l_prev_total_reward,
        l_prev_stake, l_prev_secret_votes_count, l_prev_epoch_without_growth
    FROM oracle_voting_contract_summaries
    WHERE contract_tx_id = p_contract_tx_id;

    if p_total_reward_diff is not null then
        l_new_total_reward = coalesce(l_prev_total_reward, 0) + p_total_reward_diff;
    else
        l_new_total_reward = l_prev_total_reward;
    end if;

    UPDATE oracle_voting_contract_summaries
    SET vote_proofs           = vote_proofs + p_vote_proofs_diff,
        votes                 = votes + p_votes_diff,
        finish_timestamp      = coalesce(p_finish_timestamp, finish_timestamp),
        termination_timestamp = coalesce(p_termination_timestamp, termination_timestamp),
        total_reward          = l_new_total_reward,
        stake                 = stake + p_stake_diff,
        secret_votes_count    = coalesce(p_secret_votes_count, secret_votes_count),
        epoch_without_growth  = coalesce(p_epoch_without_growth, epoch_without_growth)
    WHERE contract_tx_id = p_contract_tx_id;

    INSERT INTO changes (block_height, type)
    VALUES (p_block_height, CHANGE_TYPE_ORACLE_VOTING_SUMMARIES)
    RETURNING id INTO l_change_id;

    INSERT INTO oracle_voting_contract_summaries_changes (change_id, contract_tx_id, vote_proofs, votes,
                                                          finish_timestamp, termination_timestamp, total_reward, stake,
                                                          secret_votes_count, epoch_without_growth)
    VALUES (l_change_id, p_contract_tx_id, l_prev_vote_proofs, l_prev_votes, l_prev_finish_timestamp,
            l_prev_termination_timestamp, l_prev_total_reward, l_prev_stake, l_prev_secret_votes_count,
            l_prev_epoch_without_growth);
END
$$;

CREATE OR REPLACE PROCEDURE apply_balance_updates_on_contracts(p_block_height bigint,
                                                               p_items tp_balance_update[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    SOVC_STATE_PENDING CONSTANT smallint = 0;
    SOVC_STATE_VOTING  CONSTANT smallint = 1;
    l_item                      tp_balance_update;
    l_contract_tx_id            bigint;
    l_sort_key                  text;
    l_estimated_oracle_reward   numeric;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT c.tx_id
            INTO l_contract_tx_id
            FROM contracts c
                     JOIN sorted_oracle_voting_contracts sovc
                          on sovc.contract_tx_id = c.tx_id AND sovc.state in (SOVC_STATE_PENDING, SOVC_STATE_VOTING)
            WHERE contract_address_id = (SELECT id FROM addresses WHERE lower(address) = lower(l_item.address));

            if l_contract_tx_id is null then
                continue;
            end if;

            l_estimated_oracle_reward = calculate_estimated_oracle_reward(l_item.balance_new, l_contract_tx_id);
            l_sort_key = calculate_oracle_voting_contract_sort_key(l_estimated_oracle_reward, l_contract_tx_id);

            call update_sorted_oracle_voting_contracts(p_block_height, l_contract_tx_id, l_sort_key, null, null, null,
                                                       null);

            call update_sorted_oracle_voting_contract_committees(p_block_height, l_contract_tx_id, null, l_sort_key,
                                                                 null, null, null);

        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE apply_block_on_sorted_contracts(p_height bigint,
                                                            p_clear_old_ovc_committees boolean)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    SOVC_STATE_VOTING   CONSTANT smallint = 1;
    SOVC_STATE_COUNTING CONSTANT smallint = 3;
    l_rec                        record;
BEGIN
    for l_rec in SELECT contract_tx_id
                 FROM sorted_oracle_voting_contracts
                 WHERE state = SOVC_STATE_VOTING
                   AND p_height = counting_block
        loop
            call update_sorted_oracle_voting_contract_state(p_height, l_rec.contract_tx_id,
                                                            SOVC_STATE_COUNTING, null);
        end loop;

    if p_clear_old_ovc_committees then
        call delete_old_epoch_not_voted_committee_from_sovcc(p_height);
    end if;
END
$$;

CREATE OR REPLACE PROCEDURE add_vote_to_oracle_voting_contract_result(p_block_height bigint,
                                                                      p_contract_tx_id bigint,
                                                                      p_option smallint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CHANGE_TYPE_ORACLE_VOTING_RESULTS CONSTANT smallint = 1;
    l_votes_count                              bigint;
    l_change_id                                bigint;
BEGIN
    SELECT votes_count
    INTO l_votes_count
    FROM oracle_voting_contract_results
    WHERE contract_tx_id = p_contract_tx_id
      AND option = p_option;

    if l_votes_count is null then
        l_votes_count = 0;
        INSERT INTO oracle_voting_contract_results (contract_tx_id, "option", votes_count)
        VALUES (p_contract_tx_id, p_option, 1);
    else
        UPDATE oracle_voting_contract_results
        SET votes_count = l_votes_count + 1
        WHERE contract_tx_id = p_contract_tx_id
          AND option = p_option;
    end if;

    INSERT INTO changes (block_height, type)
    VALUES (p_block_height, CHANGE_TYPE_ORACLE_VOTING_RESULTS)
    RETURNING id INTO l_change_id;

    INSERT INTO oracle_voting_contract_results_changes (change_id, contract_tx_id, option, votes_count)
    VALUES (l_change_id, p_contract_tx_id, p_option, l_votes_count);
END
$$;

CREATE OR REPLACE PROCEDURE update_oracle_voting_contract_result(p_block_height bigint,
                                                                 p_contract_tx_id bigint,
                                                                 p_option smallint,
                                                                 p_votes bigint,
                                                                 p_all_votes bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CHANGE_TYPE_ORACLE_VOTING_RESULTS CONSTANT smallint = 1;
    l_votes_count                              bigint;
    l_all_votes_cnt                            bigint;
    l_change_id                                bigint;
BEGIN
    SELECT votes_count, all_votes_count
    INTO l_votes_count, l_all_votes_cnt
    FROM oracle_voting_contract_results
    WHERE contract_tx_id = p_contract_tx_id
      AND option = p_option;

    if l_votes_count is null then
        l_votes_count = 0;
        l_all_votes_cnt = 0;
        INSERT INTO oracle_voting_contract_results (contract_tx_id, "option", votes_count, all_votes_count)
        VALUES (p_contract_tx_id, p_option, coalesce(p_votes, 0), coalesce(p_all_votes, 0));
    else
        UPDATE oracle_voting_contract_results
        SET votes_count     = coalesce(p_votes, votes_count),
            all_votes_count = coalesce(p_all_votes, all_votes_count)
        WHERE contract_tx_id = p_contract_tx_id
          AND option = p_option;
    end if;

    INSERT INTO changes (block_height, type)
    VALUES (p_block_height, CHANGE_TYPE_ORACLE_VOTING_RESULTS)
    RETURNING id INTO l_change_id;

    INSERT INTO oracle_voting_contract_results_changes (change_id, contract_tx_id, option, votes_count, all_votes_count)
    VALUES (l_change_id, p_contract_tx_id, p_option, l_votes_count, l_all_votes_cnt);
END
$$;

CREATE OR REPLACE PROCEDURE save_contract_tx_balance_updates(p_block_height bigint,
                                                             p_items jsonb[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item           jsonb;
    l_update         jsonb;
    l_contract_tx_id bigint;
    l_contract_type  bigint;
    l_tx_id          bigint;
    l_call_method    smallint;
    l_address_id     bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];
            if l_item -> 'updates' is null then
                continue;
            end if;

            SELECT tx_id, "type"
            INTO l_contract_tx_id, l_contract_type
            FROM contracts
            WHERE contract_address_id =
                  (SELECT id FROM addresses WHERE lower(address) = lower((l_item ->> 'contractAddress')::text));

            SELECT id
            INTO l_tx_id
            FROM transactions
            WHERE lower(hash) = lower((l_item ->> 'txHash')::text);

            l_call_method = (l_item ->> 'contractCallMethod')::smallint;

            for j in 0..jsonb_array_length(l_item -> 'updates') - 1
                loop
                    l_update = l_item -> 'updates' ->> j;

                    l_address_id = get_address_id_or_insert(p_block_height, (l_update ->> 'address')::text);

                    INSERT INTO contract_tx_balance_updates (contract_tx_id, address_id, contract_type, tx_id,
                                                             call_method, balance_old, balance_new)
                    VALUES (l_contract_tx_id, l_address_id, l_contract_type, l_tx_id, l_call_method,
                            (l_update ->> 'balanceOld')::numeric,
                            (l_update ->> 'balanceNew')::numeric);
                end loop;
        end loop;
END
$$;
