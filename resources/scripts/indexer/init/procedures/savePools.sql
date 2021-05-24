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