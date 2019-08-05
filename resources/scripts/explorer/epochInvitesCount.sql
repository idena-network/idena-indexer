select count(*) invites_count
from transactions t
         join blocks b on b.id = t.block_id
         join epochs e on e.id = b.epoch_id
where t.type = 'InviteTx'
  and e.epoch = $1