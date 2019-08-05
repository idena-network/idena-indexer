select ((select max(epoch) from epochs) - max(e.epoch)) age
from address_states s
         join blocks b on b.id = s.block_id
         join epochs e on e.id = b.epoch_id
         join addresses a on a.id = s.address_id
where s.state = 'Candidate'
  and lower(a.address) = lower($1)