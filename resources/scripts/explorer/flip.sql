select f.id,
       f.size,
       b.timestamp,
       coalesce(f.answer, '')                                answer,
       coalesce(f.status, '')                                status,
       coalesce(f.data, coalesce(f.mempool_data, ''::bytea)) "data",
       t.hash                                                tx_hash,
       b.hash                                                block_hash,
       b.height                                              block_height,
       e.epoch                                               epoch
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.id = t.block_id
         join epochs e on e.id = b.epoch_id
where LOWER(f.cid) = LOWER($1)