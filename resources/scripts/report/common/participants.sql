CREATE MATERIALIZED VIEW report.mv_participants AS
SELECT p.epoch + 1                     AS epoch,
       p.address_id,
       p.num_of_trans                  AS num_of_validtxs,
       (CASE
            WHEN s.state = ANY (ARRAY [0, 4, 5, 6]) THEN
                (CASE
                     WHEN ei.missed IS TRUE AND ei.short_answers > 0 THEN 1
                     WHEN ei.missed IS TRUE AND ei.short_answers = 0 THEN 2
                     WHEN (ei.short_point / ei.short_flips::double precision) < 0.6::double precision OR
                          (ei.short_point / ei.short_flips::double precision) >= 0.6::double precision AND
                          (ei.long_point / ei.long_flips::double precision) < 0.75::double precision THEN 3
                     WHEN ei.total_short_flips > 6 AND
                          (ei.total_short_point / ei.total_short_flips::double precision) <
                          0.75::double precision AND
                          (ei.short_point / ei.short_flips::double precision) >= 0.6::double precision AND
                          (ei.long_point / ei.long_flips::double precision) >= 0.75::double precision THEN 4
                     ELSE 0
                    END)
            ELSE 0 END)                AS fail_reason,
       COALESCE(ans_fl.cnt, 0::bigint) AS unloaded_flips
FROM (SELECT b.epoch,
             t."from" AS address_id,
             count(*) AS num_of_trans
      FROM indexer.transactions t
               JOIN indexer.blocks b ON b.height = t.block_height
      WHERE t.id >= (SELECT min_tx_id
                     FROM indexer.epoch_summaries
                     WHERE epoch = (SELECT max(epochs.epoch) - 21 FROM indexer.epochs))
        AND (t.type = ANY (ARRAY [5, 6, 7, 8]))
      GROUP BY b.epoch, t."from") p
         JOIN indexer.epoch_identities ei ON ei.address_id = p.address_id AND ei.epoch = p.epoch
         JOIN indexer.address_states s ON s.id = ei.address_state_id
         LEFT JOIN (SELECT ei.epoch,
                           ei.address_id,
                           count(*) AS cnt
                    FROM indexer.answers ans
                             JOIN indexer.epoch_identities ei ON ei.address_state_id = ans.ei_address_state_id AND
                                                                 ei.epoch >=
                                                                 (SELECT max(epochs.epoch) - 21 FROM indexer.epochs)
                    WHERE ans.considered = true
                      AND ans.answer = 0
                      AND ans.is_short
                    GROUP BY ei.epoch, ei.address_id) ans_fl
                   ON ans_fl.address_id = p.address_id AND ans_fl.epoch = p.epoch;

CREATE OR REPLACE PROCEDURE report.refresh_participants()
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    if (SELECT exists(SELECT 1
                      FROM report.mv_participants
                      WHERE epoch = (SELECT max(epoch) FROM indexer.epochs))) then
        return;
    end if;
    refresh materialized view report.mv_participants;
END
$$;