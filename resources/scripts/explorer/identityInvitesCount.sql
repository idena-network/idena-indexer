select count(*) invite_count
from transactions t
         join addresses a on a.id = t.from
where t.type = 'InviteTx'
  and lower(a.address) = lower($1)