select fpo.answer_index, fpo.flip_pic_index
from flip_pic_orders fpo
where fpo.flip_data_id = (select fd.id
                          from flips_data fd
                                   join flips f on f.id = fd.flip_id
                          where lower(f.cid) = lower($1))
order by fpo.pos_index