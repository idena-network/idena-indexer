ALTER TABLE delegatee_total_validation_rewards
    ADD COLUMN penalized_delegators integer;

DO
$$
    DECLARE
        l_record record;
    BEGIN
        CREATE INDEX tmp_index ON epoch_identities (epoch, delegatee_address_id) WHERE delegatee_address_id is NOT NULL;
        for l_record in SELECT epoch, delegatee_address_id FROM delegatee_total_validation_rewards
            loop
                UPDATE delegatee_total_validation_rewards
                SET penalized_delegators = (SELECT count(*)
                                            FROM epoch_identities ei
                                                     JOIN bad_authors ba ON ba.ei_address_state_id = ei.address_state_id
                                            WHERE ei.delegatee_address_id IS NOT NULL
                                              AND ei.epoch = l_record.epoch
                                              AND ei.delegatee_address_id = l_record.delegatee_address_id)
                WHERE epoch = l_record.epoch
                  AND delegatee_address_id = l_record.delegatee_address_id;
            end loop;

        INSERT INTO delegatee_total_validation_rewards (epoch,
                                                        delegatee_address_id,
                                                        total_balance,
                                                        delegators,
                                                        penalized_delegators)
        SELECT DISTINCT ei.epoch, ei.delegatee_address_id, 0, 0, count(*) cnt
        FROM epoch_identities ei
                 JOIN bad_authors ba ON ba.ei_address_state_id = ei.address_state_id
                 LEFT JOIN delegatee_total_validation_rewards vr
                           ON vr.epoch = ei.epoch AND vr.delegatee_Address_id = ei.delegatee_address_id
        WHERE vr.delegatee_address_id is null
          AND ei.delegatee_address_id is not null
        GROUP BY ei.epoch, ei.delegatee_address_id;

        DROP INDEX tmp_index;
    END;
$$;

ALTER TABLE delegatee_total_validation_rewards
    ALTER COLUMN penalized_delegators SET NOT NULL;