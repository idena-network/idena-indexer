CREATE OR REPLACE PROCEDURE reset_contracts_to(p_block_height bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_tx_id bigint;
BEGIN
    SELECT min(id) INTO l_tx_id FROM transactions WHERE block_height > p_block_height;

    if l_tx_id is null then
        return;
    end if;

    DELETE FROM multisig_contract_terminations WHERE termination_tx_id >= l_tx_id;
    DELETE FROM multisig_contract_call_pushes WHERE call_tx_id >= l_tx_id;
    DELETE FROM multisig_contract_call_sends WHERE call_tx_id >= l_tx_id;
    DELETE FROM multisig_contract_call_adds WHERE call_tx_id >= l_tx_id;
    DELETE FROM multisig_contracts WHERE contract_tx_id >= l_tx_id;

    DELETE FROM time_lock_contract_terminations WHERE termination_tx_id >= l_tx_id;
    DELETE FROM time_lock_contract_call_transfers WHERE call_tx_id >= l_tx_id;
    DELETE FROM time_lock_contracts WHERE contract_tx_id >= l_tx_id;

    DELETE FROM refundable_oracle_lock_contract_terminations WHERE termination_tx_id >= l_tx_id;
    DELETE FROM refundable_oracle_lock_contract_call_refunds WHERE call_tx_id >= l_tx_id;
    DELETE FROM refundable_oracle_lock_contract_call_pushes WHERE call_tx_id >= l_tx_id;
    DELETE FROM refundable_oracle_lock_contract_call_deposits WHERE call_tx_id >= l_tx_id;
    DELETE FROM refundable_oracle_lock_contracts WHERE contract_tx_id >= l_tx_id;

    DELETE FROM oracle_lock_contract_terminations WHERE termination_tx_id >= l_tx_id;
    DELETE FROM oracle_lock_contract_call_check_oracle_votings WHERE call_tx_id >= l_tx_id;
    DELETE FROM oracle_lock_contract_call_pushes WHERE call_tx_id >= l_tx_id;
    DELETE FROM oracle_lock_contracts WHERE contract_tx_id >= l_tx_id;

    DELETE FROM sorted_oracle_voting_contracts WHERE contract_tx_id >= l_tx_id;
    DELETE FROM sorted_oracle_voting_contract_committees WHERE contract_tx_id >= l_tx_id;
    DELETE FROM oracle_voting_contract_summaries WHERE contract_tx_id >= l_tx_id;
    DELETE FROM oracle_voting_contract_results WHERE contract_tx_id >= l_tx_id;

    DELETE FROM oracle_voting_contracts WHERE contract_tx_id >= l_tx_id;
    DELETE FROM oracle_voting_contract_call_starts WHERE call_tx_id >= l_tx_id;
    DELETE FROM oracle_voting_contract_call_vote_proofs WHERE call_tx_id >= l_tx_id;
    DELETE FROM oracle_voting_contract_call_votes WHERE call_tx_id >= l_tx_id;
    DELETE FROM oracle_voting_contract_call_finishes WHERE call_tx_id >= l_tx_id;
    DELETE FROM oracle_voting_contract_call_prolongations WHERE call_tx_id >= l_tx_id;
    DELETE FROM oracle_voting_contract_call_add_stakes WHERE call_tx_id >= l_tx_id;
    DELETE FROM oracle_voting_contract_terminations WHERE termination_tx_id >= l_tx_id;

    DELETE FROM contract_tx_balance_updates WHERE tx_id >= l_tx_id;
    DELETE FROM contracts WHERE tx_id >= l_tx_id;
    DELETE FROM tx_receipts WHERE tx_id >= l_tx_id;
    DELETE FROM tx_events WHERE tx_id >= l_tx_id;
END
$$;

CREATE OR REPLACE PROCEDURE reset_oracle_voting_contract_results_changes(p_change_id bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_contract_tx_id  bigint;
    l_option          smallint;
    l_votes_count     bigint;
    l_all_votes_count bigint;
BEGIN
    SELECT contract_tx_id, option, votes_count, all_votes_count
    INTO l_contract_tx_id, l_option, l_votes_count, l_all_votes_count
    FROM oracle_voting_contract_results_changes
    WHERE change_id = p_change_id;

    if l_votes_count = 0 then
        DELETE
        FROM oracle_voting_contract_results
        WHERE contract_tx_id = l_contract_tx_id
          AND option = l_option;
    else
        UPDATE oracle_voting_contract_results
        SET votes_count     = l_votes_count,
            all_votes_count = l_all_votes_count
        WHERE contract_tx_id = l_contract_tx_id
          AND option = l_option;
    end if;

    DELETE FROM oracle_voting_contract_results_changes WHERE change_id = p_change_id;
END
$$;

CREATE OR REPLACE PROCEDURE reset_oracle_voting_contract_summaries_changes(p_change_id bigint)
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    UPDATE oracle_voting_contract_summaries ovcs
    SET vote_proofs           = t.vote_proofs,
        votes                 = t.votes,
        finish_timestamp      = t.finish_timestamp,
        termination_timestamp = t.termination_timestamp,
        total_reward          = t.total_reward,
        stake                 = t.stake,
        secret_votes_count    = t.secret_votes_count,
        epoch_without_growth  = t.epoch_without_growth
    FROM (SELECT contract_tx_id,
                 vote_proofs,
                 votes,
                 finish_timestamp,
                 termination_timestamp,
                 total_reward,
                 stake,
                 secret_votes_count,
                 epoch_without_growth
          FROM oracle_voting_contract_summaries_changes
          WHERE change_id = p_change_id) t
    WHERE ovcs.contract_tx_id = t.contract_tx_id;

    DELETE FROM oracle_voting_contract_summaries_changes WHERE change_id = p_change_id;
END
$$;

CREATE OR REPLACE PROCEDURE reset_sorted_oracle_voting_contracts_changes(p_change_id bigint)
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    UPDATE sorted_oracle_voting_contracts sovc
    SET sort_key       = t.sort_key,
        state          = t.state,
        state_tx_id    = t.state_tx_id,
        counting_block = t.counting_block,
        epoch          = t.epoch
    FROM (SELECT contract_tx_id, sort_key, state, state_tx_id, counting_block, epoch
          FROM sorted_oracle_voting_contracts_changes
          WHERE change_id = p_change_id) t
    WHERE sovc.contract_tx_id = t.contract_tx_id;

    DELETE FROM sorted_oracle_voting_contracts_changes WHERE change_id = p_change_id;
END
$$;

CREATE OR REPLACE PROCEDURE reset_sorted_oracle_voting_contract_committees_changes(p_change_id bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_rec record;
BEGIN
    for l_rec in SELECT contract_tx_id,
                        author_address_id,
                        address_id,
                        sort_key,
                        state,
                        state_tx_id,
                        voted,
                        deleted
                 FROM sorted_oracle_voting_contract_committees_changes
                 WHERE change_id = p_change_id
        loop

            if l_rec.deleted is null then
                DELETE
                FROM sorted_oracle_voting_contract_committees
                WHERE contract_tx_id = l_rec.contract_tx_id
                  AND address_id = l_rec.address_id;
                continue;
            end if;

            if l_rec.deleted then
                INSERT INTO sorted_oracle_voting_contract_committees (contract_tx_id,
                                                                      author_address_id,
                                                                      address_id,
                                                                      sort_key,
                                                                      state,
                                                                      state_tx_id,
                                                                      voted)
                VALUES (l_rec.contract_tx_id,
                        l_rec.author_address_id,
                        l_rec.address_id,
                        l_rec.sort_key,
                        l_rec.state,
                        l_rec.state_tx_id,
                        l_rec.voted);
                continue;
            end if;

            UPDATE sorted_oracle_voting_contract_committees sovc
            SET sort_key    = t.sort_key,
                state       = t.state,
                state_tx_id = t.state_tx_id,
                voted       = t.voted
            FROM (SELECT contract_tx_id, address_id, sort_key, state, state_tx_id, voted
                  FROM sorted_oracle_voting_contract_committees_changes
                  WHERE change_id = p_change_id) t
            WHERE sovc.contract_tx_id = t.contract_tx_id
              AND sovc.address_id = t.address_id;
        end loop;

    DELETE FROM sorted_oracle_voting_contract_committees_changes WHERE change_id = p_change_id;
END
$$;