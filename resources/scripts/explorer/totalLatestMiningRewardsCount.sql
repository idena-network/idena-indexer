select count(*)
from (select distinct mr.address_id
      from mining_rewards mr
               join blocks b on b.height = mr.block_height
      where b."timestamp" > $1) lr
         full outer join (select a.address, a.id
                          from address_states s
                                   join addresses a on a.id = s.address_id
                          where is_actual
                            -- 'Verified', 'Newbie', 'Human'
                            and "state" in (3, 7, 8)) identities on identities.id = lr.address_id