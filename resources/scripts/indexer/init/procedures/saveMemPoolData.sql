CREATE OR REPLACE PROCEDURE save_mem_pool_data(p_data jsonb)
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    if p_data is null then
        return;
    end if;

    call save_mem_pool_flip_private_keys_package_timestamps(p_data -> 'flipPrivateKeysPackageTimestamps');
    call save_mem_pool_flip_key_timestamps(p_data -> 'flipKeyTimestamps');
    call save_mem_pool_answers_hash_tx_timestamps(p_data -> 'answersHashTxTimestamps');
    call save_mem_pool_short_answers_tx_timestamps(p_data -> 'shortAnswersTxTimestamps');
    call save_flip_private_keys(p_data -> 'flipPrivateKeys');

END
$$;

CREATE OR REPLACE PROCEDURE save_mem_pool_flip_private_keys_package_timestamps(p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item jsonb;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            INSERT INTO flip_private_keys_package_timestamps (address, epoch, "timestamp")
            VALUES ((l_item ->> 'address')::text, (l_item ->> 'epoch')::smallint, (l_item ->> 'timestamp')::bigint)
            ON CONFLICT DO NOTHING;
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_mem_pool_flip_key_timestamps(p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item jsonb;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            INSERT INTO flip_key_timestamps (address, epoch, "timestamp")
            VALUES ((l_item ->> 'address')::text, (l_item ->> 'epoch')::bigint, (l_item ->> 'timestamp')::bigint)
            ON CONFLICT DO NOTHING;
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_mem_pool_answers_hash_tx_timestamps(p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item jsonb;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            INSERT INTO answers_hash_tx_timestamps (address, epoch, "timestamp")
            VALUES ((l_item ->> 'address')::text, (l_item ->> 'epoch')::bigint, (l_item ->> 'timestamp')::bigint)
            ON CONFLICT DO NOTHING;
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_mem_pool_short_answers_tx_timestamps(p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item jsonb;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            INSERT INTO short_answers_tx_timestamps (address, epoch, "timestamp")
            VALUES ((l_item ->> 'address')::text, (l_item ->> 'epoch')::smallint, (l_item ->> 'timestamp')::bigint)
            ON CONFLICT DO NOTHING;
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_flip_private_keys(p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item       jsonb;
    l_flip_tx_id bigint;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            SELECT tx_id INTO l_flip_tx_id FROM flips WHERE lower(cid) = lower((l_item ->> 'cid')::text);
            if l_flip_tx_id is null then
                continue;
            end if;

            INSERT INTO flip_private_keys (flip_tx_id, "key")
            VALUES (l_flip_tx_id, decode(l_item ->> 'key', 'hex'))
            ON CONFLICT DO NOTHING;
        end loop;
END
$$;