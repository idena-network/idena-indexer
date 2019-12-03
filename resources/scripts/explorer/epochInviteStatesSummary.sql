select coalesce(eis.state, '') state, count(*) cnt
from transactions t
         join blocks b on b.height = t.block_height
         left join used_invites ui on ui.invite_tx_id = t.id
         left join transactions at on at.id = ui.activation_tx_id
         left join (select ei.epoch, s.address_id, s.state
                    from epoch_identities ei
                             join address_states s on s.id = ei.address_state_id) eis
                   on eis.address_id = at.to and b.epoch = eis.epoch
where t.type = 'InviteTx'
  and b.epoch = $1
group by eis.state