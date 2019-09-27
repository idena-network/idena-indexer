select ei.epoch,
       vr.balance,
       vr.stake,
       vr.type
from validation_rewards vr
         join epoch_identities ei on ei.id = vr.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id
where ei.epoch in
      (
          select distinct ei.epoch
          from validation_rewards vr
                   join epoch_identities ei on ei.id = vr.epoch_identity_id
                   join address_states s on s.id = ei.address_state_id
                   join addresses a on a.id = s.address_id
          where lower(a.address) = lower($1)
          order by ei.epoch desc
          limit $3
          offset
          $2
      )
  and lower(a.address) = lower($1)
order by ei.epoch desc