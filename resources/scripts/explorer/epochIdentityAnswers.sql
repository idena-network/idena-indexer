select f.cid,
       ''                       address,
       ida.name                 resp_answer,
       a.grade = 1              resp_reported,
       coalesce(fda.name, '')   flip_answer,
       coalesce(f.grade, 0) = 1 flip_reported,
       coalesce(dfs.name, '')   status,
       a.point,
       a.grade                  resp_grade,
       coalesce(f.grade, 0)     flip_grade
from answers a
         join flips f on f.tx_id = a.flip_tx_id
         join epoch_identities ei on ei.address_state_id = a.ei_address_state_id and ei.epoch = $1
         join address_states s on s.id = ei.address_state_id
         join addresses ad on ad.id = s.address_id and lower(ad.address) = lower($2)
         left join dic_flip_statuses dfs on dfs.id = f.status
         left join dic_answers fda on fda.id = f.answer
         join dic_answers ida on ida.id = a.answer
where a.is_short = $3