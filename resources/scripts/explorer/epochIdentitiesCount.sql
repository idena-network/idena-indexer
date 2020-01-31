select case
           when $2::smallint[] is null then (
               select count(*)
               from epoch_identity_states eis
               where eis.epoch = $1
                 and ($3::smallint[] is null or eis.state = any ($3::smallint[]))
           )
           else (
               select count(*)
               from epoch_identity_states eis
                        left join address_states prevs on prevs.id = eis.prev_id
               where eis.epoch = $1
                 and (prevs.state = any ($2::smallint[]))
                 and ($3::smallint[] is null or eis.state = any ($3::smallint[]))
           )
           end identity_count