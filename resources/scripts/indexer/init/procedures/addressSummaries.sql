CREATE OR REPLACE PROCEDURE update_address_summary(p_address_id bigint,
                                                  p_flips_diff integer DEFAULT null,
                                                  p_wrong_words_flips_diff integer DEFAULT null)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_check_address_id bigint;
BEGIN
    UPDATE address_summaries
    SET flips             = flips + coalesce(p_flips_diff, 0),
        wrong_words_flips = wrong_words_flips + coalesce(p_wrong_words_flips_diff, 0)
    WHERE address_id = p_address_id
    RETURNING address_id INTO l_check_address_id;

    if l_check_address_id is null then
        INSERT INTO address_summaries (address_id, flips, wrong_words_flips)
        VALUES (p_address_id, coalesce(p_flips_diff, 0), coalesce(p_wrong_words_flips_diff, 0));
    end if;
END
$$;

CREATE OR REPLACE PROCEDURE restore_address_summaries()
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_epoch bigint;
    rec     record;
BEGIN
    SELECT max(epoch) INTO l_epoch FROM epoch_identities;
    INSERT INTO address_summaries (
        SELECT address_id, made_flips, wrong_words_flips
        FROM (
                 SELECT s.address_id,
                        coalesce(sum(ei.made_flips), 0)        made_flips,
                        coalesce(sum(ei.wrong_words_flips), 0) wrong_words_flips
                 FROM epoch_identities ei
                          JOIN address_states s ON s.id = ei.address_State_id
                 GROUP BY s.address_id
             ) t
        WHERE t.made_flips > 0
           OR t.wrong_words_flips > 0
    );

    for rec in (
        SELECT t.from, count(*) cnt
        FROM flips f
                 JOIN transactions t ON t.id = f.tx_id
                 JOIN blocks b ON b.height = t.block_height AND b.epoch > l_epoch
        WHERE f.delete_tx_id is null
        GROUP BY t.from
    )
        loop
            CALL update_address_summary(p_address_id => rec."from", p_flips_diff => rec.cnt::integer);
        end loop;
END
$$;