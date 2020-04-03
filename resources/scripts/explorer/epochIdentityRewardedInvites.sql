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
       coalesce(kitb.epoch, 0)     kill_invitee_epoch,
       coalesce(dicrt.name, '')    reward_type
from transactions t
         join addresses a on a.id = t.from and lower(a.address) = lower($2)
         join blocks b on b.height = t.block_height and b.epoch >= ($1 - 2) and b.epoch <= $1
         left join activation_txs ui on ui.invite_tx_id = t.id
         left join transactions at on at.id = ui.tx_id
         left join blocks ab on ab.height = at.block_height
         left join addresses aa on aa.id = at.to
         left join (select ei.epoch, s.address_id, s.state
                    from epoch_identities ei
                             join address_states s on s.id = ei.address_state_id) eis
                   on eis.address_id = at.to and eis.epoch = $1
         left join dic_identity_states dis on dis.id = eis.state
         left join kill_invitee_txs kit on kit.invite_tx_id = t.id
         left join transactions kitt on kitt.id = kit.tx_id
         left join blocks kitb on kitb.height = kitt.block_height
         left join (select ri.reward_type, ri.invite_tx_id, b.epoch
                    from rewarded_invitations ri
                             join blocks b on b.height = ri.block_height) ri on ri.invite_tx_id = t.id and ri.epoch = $1
         left join dic_epoch_reward_types dicrt on dicrt.id = ri.reward_type
where t.type = (select id from dic_tx_types where name = 'InviteTx')
order by t.id desc