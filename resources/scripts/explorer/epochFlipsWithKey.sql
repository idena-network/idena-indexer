select f.Cid, fk.Key
from flips f,
     transactions t,
     blocks b,
     epochs e,
     blocks ib,
     transactions it,
     flip_keys fk
where f.tx_id = t.id
  and t.block_id = b.id
  and b.epoch_id = e.id
  and e.id = ib.epoch_id
  and ib.id = it.block_id
  and it.id = fk.tx_id
  and t.from = it.from
  and e.epoch = $1