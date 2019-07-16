select f.cid
from flips f
         join transactions t on t.id = f.tx_id
         join identities i on i.address = t.from
         join blocks b on b.id = t.block_id
         join (select id, epoch from epochs order by epoch desc limit 1) e on e.id = b.epoch_id
where f.status is null
  and i.id = $1