select b.height, b.timestamp, (select count(*) from transactions where block_id = b.id) TX_COUNT
from blocks b
         join epochs e on e.id = b.epoch_id
where e.epoch = $1