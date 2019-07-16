select t.Hash, t.From
from transactions t
         join blocks b on b.id = t.block_id
         join epochs e on e.id = epoch_id
where t.type = 'InviteTx'
  and e.epoch = $1