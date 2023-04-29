do
$$
    declare
        l_record         record;
        l_prev_size      integer;
        l_new_size       integer;
        l_new_delegators integer;
    begin
        CREATE TABLE tmp_delegation_history
        (
            delegator_address_id bigint NOT NULL,
            delegatee_address_id bigint NOT NULL,
            delegation_epoch     bigint NOT NULL,
            undelegation_epoch   bigint
        );

        INSERT INTO tmp_delegation_history
        SELECT t.from, t.to, b.epoch, coalesce(bu.epoch, buu.epoch)
        FROM delegation_history dh
                 LEFT JOIN transactions t ON t.id = dh.delegation_tx_id
                 LEFT JOIN blocks b ON b.height = t.block_height
                 LEFT JOIN blocks bu ON bu.height = dh.undelegation_block_height
                 LEFT JOIN transactions tu ON tu.id = dh.undelegation_tx_id
                 LEFT JOIN blocks buu ON buu.height = tu.block_height;

        CREATE INDEX tmp_delegation_history_idx ON tmp_delegation_history (delegatee_address_id, delegation_epoch,
                                                                           coalesce(undelegation_epoch, 999));
        CREATE INDEX tmp_epoch_identities_idx ON epoch_identities (delegatee_address_id, epoch) WHERE delegatee_address_id is not null;

        CREATE TABLE pool_size_history
        (
            address_id            bigint   NOT NULL,
            epoch                 smallint NOT NULL,
            validation_size       integer  NOT NULL,
            validation_delegators integer  NOT NULL,
            end_size              integer  NOT NULL,
            end_delegators        integer  NOT NULL
        );
        CREATE INDEX pool_size_history_pkey ON pool_size_history (address_id, epoch desc);

        for l_record in (SELECT delegatee_address_id, epoch, count(*) cnt
                         FROM epoch_identities
                         WHERE delegatee_address_id is not null
                         group by delegatee_address_id, epoch
                         order by epoch)
            loop

                SELECT count(*)
                INTO l_prev_size
                FROM epoch_identities ei
                         JOIN address_states s ON s.id = ei.address_state_id
                         JOIN address_states prevs ON prevs.id = s.prev_id AND prevs.state in (3, 7, 8)
                WHERE (ei.delegatee_address_id is not null
                           AND ei.delegatee_address_id = l_record.delegatee_address_id OR
                       ei.address_id = l_record.delegatee_address_id)
                  AND ei.epoch = l_record.epoch;

                SELECT count(*)
                INTO l_new_size
                FROM epoch_identities ei
                         JOIN address_states s ON s.id = ei.address_state_id AND s.state in (3, 7, 8)
                         LEFT JOIN removed_transitive_delegations rtd
                                   ON rtd.epoch = l_record.epoch AND rtd.delegator_address_id = ei.address_id
                WHERE (ei.delegatee_address_id is not null
                           AND ei.delegatee_address_id = l_record.delegatee_address_id OR
                       ei.address_id = l_record.delegatee_address_id)
                  AND ei.epoch = l_record.epoch
                  AND rtd.delegator_address_id is null;

                SELECT count(*)
                INTO l_new_delegators
                FROM tmp_delegation_history
                WHERE delegatee_address_id = l_record.delegatee_address_id
                  AND delegation_epoch <= l_record.epoch
                  AND coalesce(undelegation_epoch, 999) > l_record.epoch;

                INSERT INTO pool_size_history
                VALUES (l_record.delegatee_address_id, l_record.epoch, l_prev_size, l_record.cnt, l_new_size,
                        l_new_delegators);
            end loop;

        DROP INDEX tmp_epoch_identities_idx;
        DROP TABLE tmp_delegation_history;

    end;
$$;