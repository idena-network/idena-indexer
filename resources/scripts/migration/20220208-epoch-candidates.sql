DROP PROCEDURE update_epoch_summary;
DROP FUNCTION save_txs;

ALTER TABLE epoch_summaries
    ADD COLUMN candidate_count bigint;

UPDATE epoch_summaries
SET candidate_count=t.candidate_count
FROM (SELECT b.epoch, count(*) candidate_count
      FROM epoch_identity_interim_states eiis
               JOIN address_states s ON s.id = eiis.address_state_id
               JOIN blocks b ON b.height = eiis.block_height
      WHERE s.state = 2
      GROUP BY b.epoch) t
WHERE epoch_summaries.epoch = t.epoch;

UPDATE epoch_summaries
SET candidate_count = (SELECT count(*) FROM address_states WHERE is_actual AND state = 2)
WHERE epoch = (SELECT max(epoch) FROM epochs);

ALTER TABLE epoch_summaries
    ALTER COLUMN candidate_count SET NOT NULL;
