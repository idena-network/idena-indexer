select count(*) active_addresses_count
from (select distinct a.id
      from transactions t
               join blocks b
                    on b.height = t.block_height and b."timestamp" > $1
               join addresses a on a.id = t.from and t.type != 1 -- not activation
          or t.to is not null and a.id = t.to and t.type != 2 -- not invite
     ) tt;