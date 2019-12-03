select t.Hash                    invite_hash,
       a.address                 invite_author,
       b.timestamp               invite_timestamp,
       coalesce(at.Hash, '')     activation_hash,
       coalesce(aa.address, '')  activation_author,
       coalesce(ab.timestamp, 0) activation_timestamp,
       coalesce(eis.state, '')     state
from transactions t
         join addresses a on a.id = t.from
         join blocks b on b.height = t.block_height
         left join used_invites ui on ui.invite_tx_id = t.id
         left join transactions at on at.id = ui.activation_tx_id
         left join blocks ab on ab.height = at.block_height
         left join addresses aa on aa.id = at.to
         left join (select ei.epoch, s.address_id, s.state
                    from epoch_identities ei
                             join address_states s on s.id = ei.address_state_id) eis
                   on eis.address_id = at.to and b.epoch = eis.epoch
where t.type = 'InviteTx'
  and b.epoch = $1
order by b.height desc
limit $3
offset
$2