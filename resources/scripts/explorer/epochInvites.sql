select t.Hash, a.address author, b.timestamp
from transactions t
         join addresses a on a.id = t.from
         join blocks b on b.id = t.block_id
         join epochs e on e.id = epoch_id
where t.type = 'InviteTx'
  and e.epoch = $1
limit $3
    offset $2