select e.epoch,
       COALESCE(ei.verifiedCount, 0) verifiedCount,
       COALESCE(b.blockCount, 0)     blockCount,
       COALESCE(f.flipsCount, 0)     flipsCount,
       COALESCE(qf.flipsCount, 0)    qFlipsCount,
       COALESCE(wqf.flipsCount, 0)   wqFlipsCount
from epochs e
         left join (select ei.epoch_id, count(*) verifiedCount
                    from epoch_identities ei
                             join address_states s on s.id = ei.address_state_id
                    where s.state = 'Verified'
                    group by ei.epoch_id) ei on ei.epoch_id = e.id
         left join (select b.epoch_id, count(*) blockCount from blocks b group by b.epoch_id) b on b.epoch_id = e.id
         left join (select b.epoch_id, count(*) flipsCount
                    from flips f,
                         transactions t,
                         blocks b
                    where f.tx_id = t.id
                      and t.block_id = b.id
                    group by b.epoch_id) f on f.epoch_id = e.id
         left join (select b.epoch_id, count(*) flipsCount
                    from flips f,
                         transactions t,
                         blocks b
                    where f.tx_id = t.id
                      and t.block_id = b.id
                      and f.status = 'Qualified'
                    group by b.epoch_id) qf on qf.epoch_id = e.id
         left join (select b.epoch_id, count(*) flipsCount
                    from flips f,
                         transactions t,
                         blocks b
                    where f.tx_id = t.id
                      and t.block_id = b.id
                      and f.status = 'WeaklyQualified'
                    group by b.epoch_id) wqf on wqf.epoch_id = e.id
where e.epoch = $1