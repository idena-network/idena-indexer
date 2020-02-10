select a.address                  address,
       0                          epoch,
       ww.id is not null as       wrong_words,
       coalesce(prevdis.name, '') prev_state,
       dis.name                   state
from bad_authors ba
         join epoch_identities ei on ei.id = ba.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id
         left join (select distinct t.from id
                    from flips f
                             join transactions t on t.id = f.tx_id
                    where f.wrong_words
                      and f.tx_id in (select id
                                      from transactions
                                      where block_height in
                                            (select height from blocks where epoch = $1))) ww on ww.id = a.id
         left join address_states prevs on prevs.id = s.prev_id
         join dic_identity_states dis on dis.id = s.state
         left join dic_identity_states prevdis on prevdis.id = prevs.state
where ei.epoch = $1
order by ww.id nulls last
limit $3
offset
$2