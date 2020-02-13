select f.cid
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.height = t.block_height and b.epoch = $1
where f.size = 0
  and f.delete_tx_id is null
limit $2;