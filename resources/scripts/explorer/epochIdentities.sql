select i.address,
       ei.state,
       (case
            when ei.short_flips != 0 then ei.short_point / ei.short_flips
            else 0
           end) respScore,
       (case
            when COALESCE(allf.fCount, 0) != 0 then (COALESCE(qf.fCount, 0) + 0.5 * COALESCE(wqf.fCount, 0)) /
                                                    allf.fCount
            else 0
           end) authorScore

from epoch_identities ei
         join epochs e on e.id = ei.epoch_id
         join identities i on i.id = ei.identity_id

         left join (select e.id e_id, i.id i_id, count(*) fCount
                    from flips f
                             join transactions t on t.id = f.tx_id
                             join blocks b on b.id = t.block_id
                             join epochs e on e.id = b.epoch_id
                             join identities i on i.address = t.from
                    where f.status = 'Qualified'
                    group by i.id, e.id) qf on qf.i_id = i.id and qf.e_id = e.id

         left join (select e.id e_id, i.id i_id, count(*) fCount
                    from flips f
                             join transactions t on t.id = f.tx_id
                             join blocks b on b.id = t.block_id
                             join epochs e on e.id = b.epoch_id
                             join identities i on i.address = t.from
                    where f.status = 'WeaklyQualified'
                    group by i.id, e.id) wqf on wqf.i_id = i.id and wqf.e_id = e.id

         left join (select e.id e_id, i.id i_id, count(*) fCount
                    from flips f
                             join transactions t on t.id = f.tx_id
                             join blocks b on b.id = t.block_id
                             join epochs e on e.id = b.epoch_id
                             join identities i on i.address = t.from
                    group by i.id, e.id) allf on allf.i_id = i.id and allf.e_id = e.id
where e.epoch = $1;