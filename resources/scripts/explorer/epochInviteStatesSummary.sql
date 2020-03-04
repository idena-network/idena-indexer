select coalesce(dis.name, '') state, count(*) cnt
from transactions t
         join blocks b on b.height = t.block_height
         left join activation_txs ui on ui.invite_tx_id = t.id
         left join transactions at on at.id = ui.tx_id
         left join (select ei.epoch, s.address_id, s.state
                    from epoch_identities ei
                             join address_states s on s.id = ei.address_state_id) eis
                   on eis.address_id = at.to and b.epoch = eis.epoch
         left join dic_identity_states dis on dis.id = eis.state
where t.type = (select id from dic_tx_types where name = 'InviteTx')
  and b.epoch = $1
group by dis.name