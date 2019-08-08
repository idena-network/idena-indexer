select f.cid
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.id = t.block_id
         join addresses a on a.id = t.from
where b.epoch_id = (select id from epochs order by epoch desc limit 1)
  and f.data is null
limit $1