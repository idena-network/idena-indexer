select (select max(epoch) from epochs) - coalesce(
        (select ((select max(epoch) from epochs) - max(b.epoch)) age
         from address_states s
                  join blocks b on b.height = s.block_height
                  join addresses a on a.id = s.address_id
         where s.state = 'Candidate'
           and lower(a.address) = lower($1)), 0)