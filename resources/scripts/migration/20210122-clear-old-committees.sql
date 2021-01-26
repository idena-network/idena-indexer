DROP FUNCTION save_addrs_and_txs;
DROP PROCEDURE apply_block_on_sorted_contracts;

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

DO
$$
    DECLARE
        l_height bigint;
    BEGIN
        SELECT max(height) INTO l_height FROM blocks;
        CALL delete_old_epoch_not_voted_committee_from_sovcc(l_height);
    END
$$