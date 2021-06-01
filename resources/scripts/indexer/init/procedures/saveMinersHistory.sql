CREATE OR REPLACE PROCEDURE save_miners_history_item(p_height bigint, p_data jsonb)
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    if p_data is null then
        return;
    end if;
    INSERT INTO miners_history (block_height, online_validators, online_miners)
    VALUES (p_height, (p_data ->> 'onlineValidators')::bigint, (p_data ->> 'onlineMiners')::bigint);
END
$$;