select coalesce(sum(1), 0)                                                      all_count,
       coalesce(sum((case when invite_tx_id is not null then 1 else 0 end)), 0) used_count
from transactions t
         join blocks b on b.height = t.block_height and b.epoch = $1
         left join activation_txs ui on ui.invite_tx_id = t.id
where t.type = (select id from dic_tx_types where name = 'InviteTx')