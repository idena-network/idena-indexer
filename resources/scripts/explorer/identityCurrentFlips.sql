select f.cid
from flips f
         join transactions t on t.id = f.tx_id
         join addresses a on a.id = t.from and lower(a.address) = lower($1)
         join blocks b on b.height = t.block_height and b.epoch = (select max(epoch) from epochs)
where f.status_block_height is null and f.delete_tx_id is null