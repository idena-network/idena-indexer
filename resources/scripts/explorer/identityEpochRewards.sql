select ei.epoch,
       vr.balance,
       vr.stake,
       dert.name                  "type",
       coalesce(prevdis.name, '') prev_state,
       dis.name                   state,
       coalesce(ra.age, 0)
from validation_rewards vr
         join epoch_identities ei on ei.id = vr.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id
         left join address_states prevs on prevs.id = s.prev_id
         left join reward_ages ra on ra.epoch_identity_id = vr.epoch_identity_id
         join dic_identity_states dis on dis.id = s.state
         left join dic_identity_states prevdis on prevdis.id = prevs.state
         join dic_epoch_reward_types dert on dert.id = vr.type
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