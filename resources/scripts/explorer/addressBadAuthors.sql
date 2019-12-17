select ''                           address,
       ei.epoch                     epoch,
       ww.address_id is not null as wrong_words,
       coalesce(prevdis.name, '')   prev_state,
       dis.name                     state
from bad_authors ba
         join epoch_identities ei on ei.id = ba.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id
         left join (select distinct b.epoch epoch,
                                    t.from  address_id
                    from flips f
                             join transactions t on t.id = f.tx_id
                             join blocks b on b.height = t.block_height
                    where f.wrong_words) ww on ww.address_id = s.address_id and ww.epoch = ei.epoch
         left join address_states prevs on prevs.id = s.prev_id
         join dic_identity_states dis on dis.id = s.state
         left join dic_identity_states prevdis on prevdis.id = prevs.state
where lower(a.address) = lower($1)
order by ww.address_id nulls last
limit $3
offset
$2