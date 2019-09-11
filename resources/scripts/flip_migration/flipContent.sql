select fd.id
from flips_data fd
         join flips f on f.id = fd.flip_id
where lower(f.cid) = lower($1)