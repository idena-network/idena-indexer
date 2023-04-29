do
$$
    declare
        l_record                                 record;
        l_start_tx_id                            bigint;
        l_finish_tx_id                           bigint;
        l_height                                 bigint;
        l_identity_update_height                 bigint;
        l_record_2                               record;
        l_address_id                             bigint;
        l_validation_height                      bigint;
        l_epoch_min_tx_id                        bigint;
        l_epoch_max_tx_id                        bigint;
        L_INACTIVE_IDENTITY_FIRST_EPOCH constant integer = 91;
    begin
        l_start_tx_id = (SELECT min(id) FROM transactions WHERE "type" = 18);
        l_finish_tx_id = (SELECT max(id) FROM transactions);

        CREATE TABLE tmp_delegation_switches
        (
            idx       integer NOT NULL,
            delegator bigint  NOT NULL,
            delegatee bigint,
            CONSTRAINT tmp_delegation_switches_pkey PRIMARY KEY (idx)
        );
        CREATE UNIQUE INDEX tmp_delegation_switches_idx ON tmp_delegation_switches (delegator);

        CREATE TABLE delegation_history
        (
            delegator_address_id      bigint NOT NULL,
            delegation_tx_id          bigint NOT NULL,
            delegation_block_height   bigint,
            undelegation_reason       smallint,
            undelegation_tx_id        bigint,
            undelegation_block_height integer,
            is_actual                 boolean
        );
        CREATE INDEX delegation_history_api_idx ON delegation_history (delegator_address_id, delegation_tx_id desc);
        CREATE UNIQUE INDEX delegation_history_api_actual_idx ON delegation_history (delegator_address_id) WHERE is_actual;

        for l_record in SELECT * FROM transactions WHERE id >= l_start_tx_id and id <= l_finish_tx_id ORDER BY id
            loop
                if l_height is null then
                    l_height = l_record.block_height;
                end if;

                if l_height <> l_record.block_height then
                    l_identity_update_height =
                            (SELECT min(block_height)
                             FROM block_flags
                             WHERE block_height >= l_height
                               AND block_height < l_record.block_height
                               AND flag = 'IdentityUpdate');

                    if l_identity_update_height is not null then

                        for l_record_2 in SELECT *
                                          FROM transactions
                                          WHERE block_height = l_identity_update_height
                                            AND "type" in (1, 3, 10, 20)
                                          ORDER BY id
                            loop

                                if l_record_2."type" in (1, 3) then
                                    l_address_id = l_record_2."from";
                                end if;

                                if l_record_2."type" in (10, 20) then
                                    l_address_id = l_record_2."to";
                                end if;

                                UPDATE delegation_history
                                SET undelegation_block_height = l_identity_update_height,
                                    undelegation_tx_id        = l_record_2.id,
                                    undelegation_reason       = 2, -- Termination
                                    is_actual                 = null
                                WHERE delegator_address_id = l_address_id
                                  AND is_actual;

                            end loop;

                        for l_record_2 in SELECT * FROM tmp_delegation_switches
                            loop
                                if l_record_2.delegatee is null then
                                    UPDATE delegation_history
                                    SET undelegation_block_height = l_identity_update_height,
                                        is_actual                 = null
                                    WHERE delegator_address_id = l_record_2.delegator
                                      AND is_actual;
                                else
                                    if exists(SELECT 1
                                              FROM tmp_delegation_switches
                                              where delegator = l_record_2.delegatee
                                                AND idx < l_record_2.idx
                                                AND delegatee is not null) then

                                        UPDATE delegation_history
                                        SET delegation_block_height   = l_identity_update_height,
                                            undelegation_block_height = l_identity_update_height,
                                            undelegation_reason       = 6, -- ApplyingFailure
                                            is_actual                 = null
                                        WHERE delegator_address_id = l_record_2.delegator
                                          AND is_actual;
                                    else
                                        UPDATE delegation_history
                                        SET delegation_block_height = l_identity_update_height
                                        WHERE delegator_address_id = l_record_2.delegator
                                          AND is_actual;
                                    end if;


                                end if;
                            end loop;
                        DELETE FROM tmp_delegation_switches;
                    end if;

                    l_validation_height = (SELECT min(block_height)
                                           FROM block_flags
                                           WHERE block_height >= l_height
                                             AND block_height < l_record.block_height
                                             AND flag = 'ValidationFinished');

                    if l_validation_height is not null then
                        for l_record_2 in SELECT *
                                          FROM removed_transitive_delegations
                                          WHERE epoch = (SELECT epoch FROM blocks WHERE height = l_validation_height)
                            loop
                                UPDATE delegation_history
                                SET undelegation_block_height = l_identity_update_height,
                                    undelegation_reason       = 4, -- TransitiveDelegationRemove
                                    is_actual                 = null
                                WHERE delegator_address_id = l_record_2.delegator_address_id
                                  AND is_actual;
                            end loop;

                        for l_record_2 in SELECT ei.address_id
                                          FROM epoch_identities ei
                                                   JOIN address_states s ON s.id = ei.address_state_id AND s.state = 0
                                                   JOIN address_states prevs ON prevs.id = s.prev_id AND prevs.state = 0
                                          WHERE ei.epoch >= L_INACTIVE_IDENTITY_FIRST_EPOCH
                                            AND ei.epoch = (SELECT epoch FROM blocks WHERE height = l_validation_height)
                            loop
                                if not exists(SELECT 1
                                              FROM delegation_history
                                              WHERE delegator_address_id = l_record_2.address_id
                                                AND is_actual) then
                                    continue;
                                end if;

                                SELECT min_tx_id, max_tx_id
                                INTO l_epoch_min_tx_id, l_epoch_max_tx_id
                                FROM epoch_summaries
                                WHERE epoch = (SELECT epoch FROM blocks WHERE height = l_validation_height);

                                if not exists(SELECT 1
                                              FROM transactions
                                              WHERE id >= l_epoch_min_tx_id
                                                AND id <= l_epoch_max_tx_id
                                                AND "from" = l_record_2.address_id) then
                                    UPDATE delegation_history
                                    SET undelegation_block_height = l_identity_update_height,
                                        undelegation_reason       = 5, -- InactiveIdentity
                                        is_actual                 = null
                                    WHERE delegator_address_id = l_record_2.address_id
                                      AND is_actual;
                                end if;

                            end loop;

                        for l_record_2 in SELECT ei.address_id
                                          FROM epoch_identities ei
                                                   JOIN address_states s ON s.id = ei.address_state_id AND s.state = 0
                                                   JOIN address_states prevs ON prevs.id = s.prev_id AND prevs.state <> 0
                                          WHERE ei.epoch = (SELECT epoch FROM blocks WHERE height = l_validation_height)
                            loop
                                if not exists(SELECT 1
                                              FROM delegation_history
                                              WHERE delegator_address_id = l_record_2.address_id
                                                AND is_actual) then
                                    continue;
                                end if;

                                UPDATE delegation_history
                                SET undelegation_block_height = l_identity_update_height,
                                    undelegation_reason       = 3, -- ValidationFailure
                                    is_actual                 = null
                                WHERE delegator_address_id = l_record_2.address_id
                                  AND is_actual;

                            end loop;
                    end if;

                    l_height = l_record.block_height;
                end if;

                if l_record.type = 18 then
                    -- DelegateTx

                    if exists(SELECT 1 FROM tmp_delegation_switches WHERE delegator = l_record."from") then
                        DELETE FROM tmp_delegation_switches WHERE delegator = l_record."from";
                    else
                        INSERT INTO tmp_delegation_switches
                        VALUES (coalesce((SELECT max(idx) FROM tmp_delegation_switches), 0) + 1, l_record."from",
                                l_record."to");
                    end if;

                    UPDATE delegation_history
                    SET is_actual = null
                    WHERE delegator_address_id = l_record."from"
                      AND is_actual;

                    INSERT INTO delegation_history (delegator_address_id, delegation_tx_id, is_actual)
                    VALUES (l_record."from", l_record.id, true);

                end if;

                if l_record.type = 19 then
                    -- UndelegateTx

                    if exists(SELECT 1 FROM tmp_delegation_switches WHERE delegator = l_record."from") then
                        DELETE FROM tmp_delegation_switches WHERE delegator = l_record."from";
                    else
                        INSERT INTO tmp_delegation_switches
                        VALUES (coalesce((SELECT max(idx) FROM tmp_delegation_switches), 0) + 1, l_record."from", null);
                    end if;

                    UPDATE delegation_history
                    SET undelegation_tx_id  = l_record.id,
                        undelegation_reason = 1, -- UndelegationReasonUndelegation
                        is_actual           = (case when delegation_block_height is null then null else true end)
                    WHERE delegator_address_id = l_record."from"
                      AND is_actual;
                end if;


            end loop;
    end;
$$;

