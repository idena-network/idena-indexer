select f.cid
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.height = t.block_height
         join addresses a on a.id = t.from
where b.epoch = (select epoch from epochs order by epoch desc limit 1)
  and lower(a.address) = lower($1)