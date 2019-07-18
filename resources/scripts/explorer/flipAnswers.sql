select ad.address, a.answer, a.is_short
from answers a
         join epoch_identities ei on ei.id = a.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses ad on ad.id = s.address_id
where a.flip_id = $1