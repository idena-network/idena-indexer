CREATE OR REPLACE PROCEDURE save_upgrades_votes(p_block_height bigint,
                                                p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item    jsonb;
    l_upgrade smallint;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            l_upgrade = (l_item ->> 'upgrade')::smallint;
            INSERT INTO upgrade_voting_history (block_height, upgrade, votes, "timestamp")
            VALUES (p_block_height, l_upgrade, (l_item ->> 'votes')::bigint,
                    (l_item ->> 'timestamp')::bigint);

            INSERT INTO upgrade_voting_history_summary (upgrade, items)
            VALUES (l_upgrade, 1)
            ON CONFLICT (upgrade) DO UPDATE SET items = upgrade_voting_history_summary.items + 1;

        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE update_upgrade_voting_short_history(p_data jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_upgrade     smallint;
    l_last_height bigint;
    l_last_step   integer;
    l_history     jsonb;
    l_item        jsonb;
BEGIN

    if p_data is null then
        return;
    end if;

    l_upgrade = (p_data ->> 'upgrade')::smallint;

    DELETE FROM upgrade_voting_short_history WHERE upgrade = l_upgrade;

    l_last_height = (p_data ->> 'lastHeight')::bigint;
    l_last_step = (p_data ->> 'lastStep')::integer;

    INSERT INTO upgrade_voting_short_history_summary (upgrade, last_height, last_step)
    VALUES (l_upgrade, l_last_height, l_last_step)
    ON CONFLICT (upgrade) DO UPDATE SET last_height = l_last_height, last_step = l_last_step;

    l_history = p_data -> 'history';
    if l_history is null then
        return;
    end if;

    for i in 0..jsonb_array_length(l_history) - 1
        loop
            l_item = (l_history ->> i)::jsonb;
            INSERT INTO upgrade_voting_short_history (block_height, upgrade, votes)
            VALUES ((l_item ->> 'blockHeight')::bigint, l_upgrade, (l_item ->> 'votes')::bigint);
        end loop;
END
$$;