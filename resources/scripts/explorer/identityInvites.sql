select t.Hash, a.address author
from transactions t
         join addresses a on a.id = t.from
         join blocks b on b.id = t.block_id
         join epochs e on e.id = epoch_id
where t.type = 'InviteTx'
  and a.id = $1