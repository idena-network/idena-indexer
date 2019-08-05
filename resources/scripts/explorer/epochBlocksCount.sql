select count(*) block_count
from blocks b
         join epochs e on e.id = b.epoch_id
where e.epoch = $1