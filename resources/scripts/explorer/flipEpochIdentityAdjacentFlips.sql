select coalesce(prev_cid, last_cid)  prev_cid,
       prev_cid is null as           prev_cycled,
       coalesce(next_cid, first_cid) next_cid,
       next_cid is null as           next_cycled
from (select f.cid,
             lead(f.cid) over (order by t.id desc)                                              next_cid,
             first_value(f.cid)
             over (order by t.id desc ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING) first_cid,
             lag(f.cid) over (order by t.id desc)                                               prev_cid,
             last_value(f.cid)
             over (order by t.id DESC ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING) last_cid
      from flips f
               join transactions t on t.id = f.tx_id
               join blocks b on b.height = t.block_height
      where (b.epoch, t.from) in (select b.epoch, t.from
                                  from flips f
                                           join transactions t on t.id = f.tx_id
                                           join blocks b on b.height = t.block_height
                                  where lower(f.cid) = lower($1))) rel
where lower(rel.cid) = lower($1)