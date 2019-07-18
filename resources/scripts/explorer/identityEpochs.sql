select e.epoch,
       s.state,
       ei.approved,
       ei.missed,
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
         join address_states s on s.id=ei.address_state_id
         join epochs e on e.id = ei.epoch_id

         left join (select e.id e_id, ad.id a_id, count(*) fCount
                    from flips f
                             join transactions t on t.id = f.tx_id
                             join blocks b on b.id = t.block_id
                             join epochs e on e.id = b.epoch_id
                             join addresses ad on ad.id = t.from
                    where f.status = 'Qualified'
                    group by ad.id, e.id) qf on qf.a_id = s.address_id and qf.e_id = e.id

         left join (select e.id e_id, ad.id a_id, count(*) fCount
                    from flips f
                             join transactions t on t.id = f.tx_id
                             join blocks b on b.id = t.block_id
                             join epochs e on e.id = b.epoch_id
                             join addresses ad on ad.id = t.from
                    where f.status = 'WeaklyQualified'
                    group by ad.id, e.id) wqf on wqf.a_id = s.address_id and wqf.e_id = e.id

         left join (select e.id e_id, ad.id a_id, count(*) fCount
                    from flips f
                             join transactions t on t.id = f.tx_id
                             join blocks b on b.id = t.block_id
                             join epochs e on e.id = b.epoch_id
                             join addresses ad on ad.id = t.from
                    where f.status is not NULL
                    group by ad.id, e.id) allf on allf.a_id = s.address_id and allf.e_id = e.id
where s.address_id = $1
order by e.epoch