select dics.name, count(*) cnt
from (select eiis.address_state_id
      from epoch_identity_interim_states eiis
               join blocks b on b.height = eiis.block_height and b.epoch = $1
      where exists(select 1 from epoch_identities where epoch = $1)
      union
      select s.id address_state_id
      from address_states s
      where $1 - 1 = (select max(epoch) from epoch_identities)
        and s.is_actual
        and not s.state in (0, 5)
      union
      select s.id address_state_id
      from address_states s
               join blocks b on b.height = s.block_height
               left join temporary_identities ti on ti.address_id = s.address_id
      where $1 - 1 = (select max(epoch) from epoch_identities)
        and s.is_actual
        and s.state in (0, 5)
        and b.epoch = $1
        and ti.address_id is null) t
         join address_states s on s.id = t.address_state_id
         join dic_identity_states dics on dics.id = s.state
group by dics.name