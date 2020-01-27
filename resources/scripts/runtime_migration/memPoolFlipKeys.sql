select a.address, mpfk."key"
from mem_pool_flip_keys mpfk
         join epoch_identities ei on ei.id = mpfk.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id
where ei.epoch = $1