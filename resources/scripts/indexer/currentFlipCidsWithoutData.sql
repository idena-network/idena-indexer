select f.cid, a.address
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.height = t.block_height
         join addresses a on a.id = t.from
         left join flips_data fd on fd.flip_id = f.id
where b.epoch = (select max(epoch) from epochs)
  and fd.id is null
limit $1