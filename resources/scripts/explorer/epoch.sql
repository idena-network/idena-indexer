select e.epoch,
       COALESCE(b.blockCount, 0) blockCount
from epochs e
         left join (select b.epoch_id, count(*) blockCount from blocks b group by b.epoch_id) b on b.epoch_id = e.id
where e.epoch = $1