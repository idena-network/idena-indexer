select f.cid,
       ''                             address,
       ida.name                       resp_answer,
       a.wrong_words,
       coalesce(fda.name, '')         flip_answer,
       COALESCE(f.wrong_words, false) wrongWords,
       coalesce(dfs.name, '')         status,
       a.point
from answers a
         join flips f on f.id = a.flip_id
         join epoch_identities ei on ei.id = a.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses ad on ad.id = s.address_id
         left join dic_flip_statuses dfs on dfs.id = f.status
         left join dic_answers fda on fda.id = f.answer
         join dic_answers ida on ida.id = a.answer
where ei.epoch = $1
  and lower(ad.address) = lower($2)
  and a.is_short = $3