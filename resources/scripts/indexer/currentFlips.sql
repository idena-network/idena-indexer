select f.id, f.cid, f.pair
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.height = t.block_height and b.epoch = (select epoch from epochs order by epoch desc limit 1)
         join addresses a on a.id = t.from and lower(a.address) = lower($1)
where f.delete_tx_id is null
