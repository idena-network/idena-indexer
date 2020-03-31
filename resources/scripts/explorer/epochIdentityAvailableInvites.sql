select ei.epoch + 1, ei.next_epoch_invites
from epoch_identities ei
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id and lower(a.address) = lower($2)
where ei.epoch <= ($1 - 1)
  and ei.epoch >= ($1 - 3)
order by ei.address_state_id desc