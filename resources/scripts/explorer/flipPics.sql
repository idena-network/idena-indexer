select fp.data
from flip_pics fp
where fp.flip_data_id = (select fd.id
                         from flips_data fd
                                  join flips f on f.id = fd.flip_id
                         where lower(f.cid) = lower($1))
order by fp.index