select t.Hash                      invite_hash,
       a.address                   invite_author,
       b.timestamp                 invite_timestamp,
       b.epoch                     invite_epoch,
       coalesce(at.Hash, '')       activation_hash,
       coalesce(aa.address, '')    activation_author,
       coalesce(ab.timestamp, 0)   activation_timestamp,
       coalesce(dis.name, '')      state,
       coalesce(kitt.Hash, '')     kill_invitee_hash,
       coalesce(kitb.timestamp, 0) kill_invitee_timestamp,
       coalesce(kitb.epoch, 0)     kill_invitee_epoch
from transactions t
         join addresses a on a.id = t.from
         join blocks b on b.height = t.block_height
         left join activation_txs ui on ui.invite_tx_id = t.id
         left join transactions at on at.id = ui.tx_id
         left join blocks ab on ab.height = at.block_height
         left join addresses aa on aa.id = at.to
         left join (select ei.epoch, s.address_id, s.state
                    from epoch_identities ei
                             join address_states s on s.id = ei.address_state_id) eis
                   on eis.address_id = at.to and b.epoch = eis.epoch
         left join dic_identity_states dis on dis.id = eis.state
         left join kill_invitee_txs kit on kit.invite_tx_id = t.id
         left join transactions kitt on kitt.id = kit.tx_id
         left join blocks kitb on kitb.height = kitt.block_height
where t.type = (select id from dic_tx_types where name = 'InviteTx')
  and b.epoch = $1
order by b.height desc
limit $3
    offset
    $2