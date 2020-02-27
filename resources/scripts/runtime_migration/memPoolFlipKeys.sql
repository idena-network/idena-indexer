select a.address, mpfk."key"
from mem_pool_flip_keys mpfk
         join epoch_identities ei on ei.address_state_id = mpfk.ei_address_state_id and ei.epoch = $1
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id