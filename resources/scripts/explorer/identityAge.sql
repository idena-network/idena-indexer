select case
           when ((select s.state
                  from address_states s
                           join addresses a on a.id = s.address_id
                  where lower(a.address) = lower($1)
                    and s.is_actual) not in ('Verified', 'Suspended', 'Zombie', 'Newbie'))
               then 0
           else
                       (select max(epoch) from epochs) -
                       coalesce((select max(b.epoch) age
                                 from address_states s
                                          join blocks b on b.height = s.block_height
                                          join addresses a on a.id = s.address_id
                                 where s.state = 'Candidate'
                                   and lower(a.address) = lower($1)), 0)
           end