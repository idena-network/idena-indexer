insert into proposers (block_id, address_id)
values ($1, (select id from addresses where address=$2))