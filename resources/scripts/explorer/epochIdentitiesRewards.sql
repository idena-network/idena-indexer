select a.address,
       vr.balance,
       vr.stake,
       vr.type
from validation_rewards vr
         join epoch_identities ei on ei.id = vr.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id
where vr.epoch_identity_id in
      (
          select distinct vr.epoch_identity_id
          from validation_rewards vr
                   join epoch_identities ei on ei.id = vr.epoch_identity_id
          where ei.epoch = $1
          order by vr.epoch_identity_id
          limit $3
          offset
          $2
      )
order by a.address