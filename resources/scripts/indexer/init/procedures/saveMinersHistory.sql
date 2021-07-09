CREATE OR REPLACE PROCEDURE save_miners_history_item(p_height bigint, p_data jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_block_timestamp bigint;
BEGIN
    if p_data is null then
        return;
    end if;
    SELECT "timestamp" INTO l_block_timestamp FROM blocks WHERE height = p_height;
    INSERT INTO miners_history (block_timestamp, online_validators, online_miners)
    VALUES (l_block_timestamp, (p_data ->> 'onlineValidators')::bigint, (p_data ->> 'onlineMiners')::bigint);
END
$$;