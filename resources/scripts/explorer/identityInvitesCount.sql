select count(*) invite_count
from transactions t
         join addresses a on a.id = t.from
         join blocks b on b.id = t.block_id
where t.type = 'InviteTx'
  and lower(a.address) = lower($1)