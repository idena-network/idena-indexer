insert into penalties (address_id, penalty, penalty_seconds, inherited_from_address_id, block_height)
values ((select id from addresses where lower(address) = lower($1)), $2, $3,
        (case
             when $4::text is NULL or char_length($4::text) = 0 then NULL
             else get_address_id_or_insert($5::bigint, $4::text) end)
           , $5)