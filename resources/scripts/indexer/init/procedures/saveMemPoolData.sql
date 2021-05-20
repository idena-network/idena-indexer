CREATE OR REPLACE PROCEDURE save_mem_pool_data(p_data jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_flip_private_keys jsonb;
    l_item              jsonb;
    l_flip_tx_id        bigint;
BEGIN

    if p_data is null then
        return;
    end if;

    l_flip_private_keys = p_data -> 'flipPrivateKeys';
    if l_flip_private_keys is null then
        return;
    end if;

    for i in 0..jsonb_array_length(l_flip_private_keys) - 1
        loop
            l_item = (l_flip_private_keys ->> i)::jsonb;
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