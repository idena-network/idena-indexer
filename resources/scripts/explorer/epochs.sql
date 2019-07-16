select e.epoch,
       COALESCE(ei.verifiedCount, 0) verifiedCount,
       COALESCE(b.blockCount, 0)     blockCount,
       COALESCE(f.flipsCount, 0)     flipsCount
from epochs e
         left join (select ei.epoch_id, count(*) verifiedCount
                    from epoch_identities ei
                    where ei.state = 'Verified'
                    group by ei.epoch_id) ei on ei.epoch_id = e.id
         left join (select b.epoch_id, count(*) blockCount from blocks b group by b.epoch_id) b on b.epoch_id = e.id
         left join (select b.epoch_id, count(*) flipsCount
                    from flips f,
                         transactions t,
                         blocks b
                    where f.tx_id = t.id
                      and t.block_id = b.id
                    group by b.epoch_id) f on f.epoch_id = e.id
order by e.epoch