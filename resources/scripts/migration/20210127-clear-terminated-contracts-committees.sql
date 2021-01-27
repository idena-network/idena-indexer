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

DO
$$
    DECLARE
        SOVC_STATE_TERMINATED CONSTANT smallint = 4;
        l_height                       bigint;
        l_record                       record;
    BEGIN
        SELECT max(height) INTO l_height FROM blocks;
        for l_record in SELECT contract_tx_id FROM sorted_oracle_voting_contracts WHERE state = SOVC_STATE_TERMINATED
            loop
                call delete_not_voted_committee_from_sovcc(l_height, l_record.contract_tx_id);
            end loop;
    END
$$