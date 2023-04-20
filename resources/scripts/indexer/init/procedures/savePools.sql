CREATE OR REPLACE PROCEDURE save_delegation_switches(p_block_height bigint,
                                                     p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                      jsonb;
    l_delegator_address_id      bigint;
    l_delegatee_address_id      bigint;
    l_prev_delegatee_address_id bigint;
    l_total_delegated           bigint;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            l_delegator_address_id = get_address_id_or_insert(p_block_height, (l_item ->> 'delegator')::text);
            l_delegatee_address_id = null;
            if l_item ->> 'delegatee' is not null then
                l_delegatee_address_id =
                        get_address_id_or_insert(p_block_height, (l_item ->> 'delegatee')::text);
            end if;

            DELETE
            FROM delegations
            WHERE delegator_address_id = l_delegator_address_id
            RETURNING delegatee_address_id INTO l_prev_delegatee_address_id;

            if l_prev_delegatee_address_id is not null then
                UPDATE pool_sizes
                SET total_delegated = total_delegated - 1
                WHERE address_id = l_prev_delegatee_address_id
                RETURNING total_delegated INTO l_total_delegated;

                if l_total_delegated = 0 then
                    DELETE FROM pool_sizes WHERE address_id = l_prev_delegatee_address_id;
                    UPDATE pools_summary SET count = count - 1;
                end if;
            end if;

            if l_delegatee_address_id is not null then
                INSERT INTO delegations (delegator_address_id, delegatee_address_id, birth_epoch)
                VALUES (l_delegator_address_id, l_delegatee_address_id, (l_item ->> 'birthEpoch')::integer);

                INSERT INTO pool_sizes (address_id, total_delegated, size)
                VALUES (l_delegatee_address_id, 1, 0)
                ON CONFLICT (address_id) DO UPDATE SET total_delegated = pool_sizes.total_delegated + 1
                RETURNING total_delegated INTO l_total_delegated;

                if l_total_delegated = 1 then
                    if EXISTS(SELECT 1 FROM pools_summary) then
                        UPDATE pools_summary SET count = count + 1;
                    else
                        INSERT INTO pools_summary (count) VALUES (1);
                    end if;
                end if;
            end if;

        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE apply_birthday_on_delegations(p_address_id bigint,
                                                          p_birth_epoch integer)
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    UPDATE delegations SET birth_epoch = p_birth_epoch WHERE delegator_address_id = p_address_id;
END
$$;

CREATE OR REPLACE PROCEDURE save_delegations(p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                 jsonb;
    l_delegator_address_id bigint;
    l_delegatee_address_id bigint;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;

            SELECT id
            INTO l_delegator_address_id
            FROM addresses
            WHERE lower(address) = lower((l_item ->> 'delegator')::text);

            SELECT id
            INTO l_delegatee_address_id
            FROM addresses
            WHERE lower(address) = lower((l_item ->> 'delegatee')::text);

            INSERT INTO delegations (delegator_address_id, delegatee_address_id, birth_epoch)
            VALUES (l_delegator_address_id, l_delegatee_address_id, (l_item ->> 'birthEpoch')::integer);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE update_pool_sizes(p_block_height bigint,
                                              p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item       jsonb;
    l_address_id bigint;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            l_address_id = get_address_id_or_insert(p_block_height, (l_item ->> 'address')::text);
            UPDATE pool_sizes SET size = (l_item ->> 'size')::bigint WHERE address_id = l_address_id;
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_pool_sizes(p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item       jsonb;
    l_address_id bigint;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            SELECT id INTO l_address_id FROM addresses WHERE lower(address) = lower((l_item ->> 'address')::text);
            INSERT INTO pool_sizes (address_id, total_delegated, size)
            VALUES (l_address_id, (l_item ->> 'totalDelegated')::bigint, (l_item ->> 'size')::bigint);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE generate_pools_summary()
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    INSERT INTO pools_summary (count) VALUES ((SELECT count(*) FROM pool_sizes));
END
$$;

CREATE OR REPLACE PROCEDURE save_removed_transitive_delegations(p_block_height bigint,
                                                                p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                 jsonb;
    l_epoch                bigint;
    l_delegator_address_id bigint;
    l_delegatee_address_id bigint;
BEGIN
    if p_items is null then
        return;
    end if;
    SELECT epoch INTO l_epoch FROM blocks WHERE height = p_block_height;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            l_delegator_address_id = get_address_id_or_insert(p_block_height, (l_item ->> 'delegator')::text);
            l_delegatee_address_id = get_address_id_or_insert(p_block_height, (l_item ->> 'delegatee')::text);

            INSERT INTO removed_transitive_delegations (epoch, delegator_address_id, delegatee_address_id)
            VALUES (l_epoch, l_delegator_address_id, l_delegatee_address_id);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_delegation_history_updates(p_block_height bigint,
                                                            p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                 jsonb;
    l_delegator_address_id bigint;
    l_delegation_tx_id     bigint;
    l_undelegation_tx_id   bigint;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;

            l_delegator_address_id = get_address_id_or_insert(p_block_height, (l_item ->> 'delegatorAddress')::text);

            l_delegation_tx_id = null;
            if l_item ->> 'delegationTx' is not null then
                SELECT id
                INTO l_delegation_tx_id
                FROM transactions
                WHERE lower(hash) = lower((l_item ->> 'delegationTx')::text);
            end if;

            l_undelegation_tx_id = null;
            if l_item ->> 'undelegationTx' is not null then
                SELECT id
                INTO l_undelegation_tx_id
                FROM transactions
                WHERE lower(hash) = lower((l_item ->> 'undelegationTx')::text);
            end if;

            call update_delegation_history(p_block_height, l_delegator_address_id, l_delegation_tx_id,
                                           (l_item ->> 'delegationBlockHeight')::bigint,
                                           (l_item ->> 'undelegationReason')::smallint, l_undelegation_tx_id,
                                           (l_item ->> 'undelegationBlockHeight')::bigint);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE update_delegation_history(p_block_height bigint,
                                                      p_delegator_address_id bigint,
                                                      p_delegation_tx_id bigint,
                                                      p_delegation_block_height bigint,
                                                      p_undelegation_reason smallint,
                                                      p_undelegation_tx_id bigint,
                                                      p_undelegation_block_height bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CHANGE_TYPE_DELEGATION_HISTORY CONSTANT smallint = 8;
    l_change_id                             bigint;
    l_prev_delegation_tx_id                 bigint;
    l_prev_delegation_block_height          bigint;
    l_prev_undelegation_reason              smallint;
    l_prev_undelegation_tx_id               bigint;
    l_prev_undelegation_block_height        integer;
    l_prev_is_actual                        boolean;
BEGIN

    SELECT delegation_tx_id,
           delegation_block_height,
           undelegation_reason,
           undelegation_tx_id,
           undelegation_block_height,
           is_actual
    INTO l_prev_delegation_tx_id,
        l_prev_delegation_block_height,
        l_prev_undelegation_reason,
        l_prev_undelegation_tx_id,
        l_prev_undelegation_block_height,
        l_prev_is_actual
    FROM delegation_history
    WHERE delegator_address_id = p_delegator_address_id
      AND is_actual;

    if p_delegation_tx_id is not null then
        if l_prev_delegation_tx_id is not null then
            UPDATE delegation_history
            SET is_actual = null
            WHERE delegator_address_id = p_delegator_address_id
              AND is_actual;

            INSERT INTO changes (block_height, "type")
            VALUES (p_block_height, CHANGE_TYPE_DELEGATION_HISTORY)
            RETURNING id INTO l_change_id;

            INSERT INTO delegation_history_changes (change_id, delegator_address_id, delegation_tx_id,
                                                    delegation_block_height, undelegation_reason, undelegation_tx_id,
                                                    undelegation_block_height, is_actual)
            VALUES (l_change_id, p_delegator_address_id, l_prev_delegation_tx_id, l_prev_delegation_block_height,
                    l_prev_undelegation_reason, l_prev_undelegation_tx_id, l_prev_undelegation_block_height,
                    l_prev_is_actual);
        end if;

        INSERT INTO delegation_history (delegator_address_id, delegation_tx_id, delegation_block_height,
                                        undelegation_reason, undelegation_tx_id, undelegation_block_height, is_actual)
        VALUES (p_delegator_address_id, p_delegation_tx_id, p_delegation_block_height, p_undelegation_reason,
                p_undelegation_tx_id, p_undelegation_block_height, true);

        INSERT INTO changes (block_height, "type")
        VALUES (p_block_height, CHANGE_TYPE_DELEGATION_HISTORY)
        RETURNING id INTO l_change_id;

        INSERT INTO delegation_history_changes (change_id, delegator_address_id, delegation_tx_id)
        VALUES (l_change_id, p_delegator_address_id, p_delegation_tx_id);

    else
        if l_prev_delegation_tx_id is not null then
            UPDATE delegation_history
            SET delegation_tx_id          = coalesce(p_delegation_tx_id, delegation_tx_id),
                delegation_block_height   = coalesce(p_delegation_block_height, delegation_block_height),
                undelegation_reason       = coalesce(p_undelegation_reason, undelegation_reason),
                undelegation_tx_id        = coalesce(p_undelegation_tx_id, undelegation_tx_id),
                undelegation_block_height = coalesce(p_undelegation_block_height, undelegation_block_height),
                is_actual                 = (case
                                                 when ((coalesce(p_delegation_block_height, delegation_block_height) is null and
                                                        coalesce(p_undelegation_block_height, undelegation_block_height) is null)
                                                     or
                                                       coalesce(p_undelegation_block_height, undelegation_block_height) is not null) and
                                                      coalesce(p_undelegation_reason, undelegation_reason) is not null
                                                     then null
                                                 else true end)
            WHERE delegator_address_id = p_delegator_address_id
              AND is_actual;

            INSERT INTO changes (block_height, "type")
            VALUES (p_block_height, CHANGE_TYPE_DELEGATION_HISTORY)
            RETURNING id INTO l_change_id;

            INSERT INTO delegation_history_changes (change_id, delegator_address_id, delegation_tx_id,
                                                    delegation_block_height, undelegation_reason, undelegation_tx_id,
                                                    undelegation_block_height, is_actual)
            VALUES (l_change_id, p_delegator_address_id, l_prev_delegation_tx_id, l_prev_delegation_block_height,
                    l_prev_undelegation_reason, l_prev_undelegation_tx_id, l_prev_undelegation_block_height,
                    l_prev_is_actual);
        end if;
    end if;
END
$$;

CREATE OR REPLACE PROCEDURE reset_delegation_history_changes(p_change_id bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_delegator_address_id      bigint;
    l_delegation_tx_id          bigint;
    l_delegation_block_height   bigint;
    l_undelegation_reason       smallint;
    l_undelegation_tx_id        bigint;
    l_undelegation_block_height integer;
    l_is_actual                 boolean;
BEGIN
    SELECT delegator_address_id,
           delegation_tx_id,
           delegation_block_height,
           undelegation_reason,
           undelegation_tx_id,
           undelegation_block_height,
           is_actual
    INTO l_delegator_address_id, l_delegation_tx_id, l_delegation_block_height, l_undelegation_reason, l_undelegation_tx_id, l_undelegation_block_height, l_is_actual
    FROM delegation_history_changes
    WHERE change_id = p_change_id;

    if l_delegation_block_height is null AND l_undelegation_reason is null AND l_undelegation_tx_id is null AND
       l_undelegation_block_height is null AND l_is_actual is null then
        DELETE
        FROM delegation_history
        WHERE delegator_address_id = l_delegator_address_id
          AND delegation_tx_id = l_delegation_tx_id;
    else
        UPDATE delegation_history
        SET delegation_block_height   = l_delegation_block_height,
            undelegation_reason       = l_undelegation_reason,
            undelegation_tx_id        = l_undelegation_tx_id,
            undelegation_block_height = l_undelegation_block_height,
            is_actual                 = l_is_actual
        WHERE delegator_address_id = l_delegator_address_id
          AND delegation_tx_id = l_delegation_tx_id;
    end if;

    DELETE FROM token_balances_changes WHERE change_id = p_change_id;
END
$$;