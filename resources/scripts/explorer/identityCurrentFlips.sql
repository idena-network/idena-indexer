select f.cid
from flips f
         join transactions t on t.id = f.tx_id
         join addresses a on a.id = t.from
         join blocks b on b.height = t.block_height
where f.status_block_height is null
  and b.epoch = (select max(epoch) from epochs)
  and lower(a.address) = lower($1)