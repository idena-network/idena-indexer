select count(*) invites_count
from transactions t
         join blocks b on b.height = t.block_height
where t.type = (select id from dic_tx_types where name = 'InviteTx')
  and b.epoch = $1