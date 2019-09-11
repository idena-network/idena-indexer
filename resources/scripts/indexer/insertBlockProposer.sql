insert into block_proposers (block_height, address_id)
values ($1, (select id from addresses where lower(address) = lower($2)))