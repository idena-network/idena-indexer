select a.address,
       ga.strong_flips,
       ga.weak_flips,
       ga.successful_invites
from good_authors ga
         join epoch_identities ei on ei.address_state_id = ga.ei_address_state_id and ei.epoch = $1
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id
limit $3
offset
$2