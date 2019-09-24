select a.address,
       ga.strong_flips,
       ga.weak_flips,
       ga.successful_invites
from good_authors ga
         join epoch_identities ei on ei.id = ga.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id
where ei.epoch = $1
limit $3
offset
$2