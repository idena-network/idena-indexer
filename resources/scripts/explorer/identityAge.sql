select case
           when (select s.state
                 from address_states s
                          join addresses a on a.id = s.address_id
                 where lower(a.address) = lower($1)
                   and s.is_actual) in
               -- 'Verified', 'Suspended', 'Zombie', 'Newbie', 'Human'
                (3, 4, 6, 7, 8)
               then (select max(epoch) from epochs) -
                    coalesce((select bd.birth_epoch
                              from birthdays bd
                                       join addresses a on a.id = bd.address_id
                              where lower(a.address) = lower($1)), 0)
           else
               0
           end