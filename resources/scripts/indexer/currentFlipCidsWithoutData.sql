select f.cid
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.height = t.block_height
         left join flips_data fd on fd.flip_id = f.id
where b.epoch = (select epoch from epochs order by epoch desc limit 1)
  and fd.id is null
limit $1