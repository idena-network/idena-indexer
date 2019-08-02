select e.epoch,
       (select count(*) from blocks b where b.epoch_id = e.id) block_count,
       (select count(*)
        from transactions t
                 join blocks b on b.id = t.block_id
        where b.epoch_id = e.id)                               tx_count
from epochs e
where e.epoch = $1