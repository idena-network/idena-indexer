select count(*) invite_count
from transactions t
         join addresses a on a.id = t.from
where t.type = (select id from dic_tx_types where name = 'InviteTx')
  and lower(a.address) = lower($1)