CREATE OR REPLACE PROCEDURE save_flips_content(p_fails tp_failed_flip_content[],
                                               p_contents jsonb[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_fail       tp_failed_flip_content;
    l_content    jsonb;
    l_flip_tx_id bigint;
    l_cid        text;
    l_encrypted  boolean;
BEGIN
    if p_fails is not null then
        for i in 1..cardinality(p_fails)
            loop
                l_fail := p_fails[i];
                if l_fail.attempts_limit_reached then
                    delete from flips_queue where lower(cid) = lower(l_fail.cid);
                else
                    update flips_queue
                    set attempts              = attempts + 1,
                        next_attempt_timestamp=l_fail.next_attempt_timestamp
                    where lower(cid) = lower(l_fail.cid);
                end if;
            end loop;
    end if;
    if p_contents is not null then
        for i in 1..cardinality(p_contents)
            loop
                l_content := p_contents[i];
                l_cid := lower((l_content ->> 'cid')::text);

                delete from flips_queue where lower(cid) = l_cid;

                select tx_id into l_flip_tx_id from flips where lower(cid) = l_cid;

                insert into flips_data (flip_tx_id)
                values (l_flip_tx_id);

                if l_content -> 'pics' is not null then
                    for j in 0..jsonb_array_length(l_content -> 'pics') - 1
                        loop
                            insert into flip_pics (fd_flip_tx_id, index, data)
                            values (l_flip_tx_id, j, decode(l_content -> 'pics' ->> j, 'hex'));
                        end loop;

                    l_encrypted := jsonb_array_length(l_content -> 'pics') = 2;
                    if l_encrypted then
                        update flip_summaries set encrypted = true where flip_tx_id = l_flip_tx_id;
                    end if;
                end if;

                if l_content -> 'orders' is not null then
                    for l_answer_index in 0..jsonb_array_length(l_content -> 'orders') - 1
                        loop
                            for l_pos_index in 0..jsonb_array_length(l_content -> 'orders' -> l_answer_index) - 1
                                loop
                                    insert into flip_pic_orders (fd_flip_tx_id, answer_index, pos_index, flip_pic_index)
                                    values (l_flip_tx_id, l_answer_index, l_pos_index,
                                            (l_content -> 'orders' -> l_answer_index ->> l_pos_index)::smallint);
                                end loop;
                        end loop;
                end if;

                if l_content -> 'icon' is not null then
                    insert into flip_icons (fd_flip_tx_id, data)
                    values (l_flip_tx_id, decode(l_content ->> 'icon', 'hex'));
                end if;

            end loop;
    end if;
END
$BODY$;