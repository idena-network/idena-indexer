select ''                     cid,
       ad.address,
       a.answer,
       a.wrong_words,
       coalesce(f.answer),
       coalesce(f.wrong_words, false),
       coalesce(dfs.name, '') status,
       a.point
from answers a
         join epoch_identities ei on ei.id = a.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses ad on ad.id = s.address_id
         join flips f on f.id = a.flip_id
         left join dic_flip_statuses dfs on dfs.id = f.status
where lower(f.cid) = lower($1)
  and a.is_short = $2
limit $4
offset
$3