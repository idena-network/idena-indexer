select ''                     cid,
       ad.address,
       ida.name               resp_answer,
       a.wrong_words,
       coalesce(fda.name, '') flip_answer,
       coalesce(f.wrong_words, false),
       coalesce(dfs.name, '') status,
       a.point
from answers a
         join flips f on f.tx_id = a.flip_tx_id and lower(f.cid) = lower($1)
         join address_states s on s.id = a.ei_address_state_id
         join addresses ad on ad.id = s.address_id
         left join dic_flip_statuses dfs on dfs.id = f.status
         left join dic_answers fda on fda.id = f.answer
         join dic_answers ida on ida.id = a.answer
where a.is_short = $2
limit $4
offset
$3