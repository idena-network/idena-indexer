-- balances
delete
from balances
where block_id in
      (select id from blocks where height > $1);

-- flips_to_solve
delete
from flips_to_solve
where epoch_identity_id in
      (select id
       from epoch_identities
       where address_state_id in
             (select id
              from address_states
              where block_id in (select id from blocks where height > $1)));

-- answers
delete
from answers
where epoch_identity_id in
      (select id
       from epoch_identities
       where address_state_id in
             (select id
              from address_states
              where block_id in (select id from blocks where height > $1)));

-- epoch_identities
delete
from epoch_identities
where address_state_id in
      (select id
       from address_states
       where block_id in (select id from blocks where height > $1));

-- address_states
delete
from address_states
where block_id in (select id from blocks where height > $1);
-- restore actual states
update address_states
set is_actual = true
where id in
      (select s.id
       from address_states s
                join blocks b on b.id = s.block_id
       where (s.address_id, b.height) in
             (select s.address_id, max(b.height)
              from address_states s
                       join blocks b on b.id = s.block_id
              group by address_id)
         and not s.is_actual);

-- flip_pics
delete
from flip_pics
where flip_data_id in
      (select id
       from flips_data
       where block_id in
             (select id from blocks where height > $1));

-- flip_pic_orders
delete
from flip_pic_orders
where flip_data_id in
      (select id
       from flips_data
       where block_id in
             (select id from blocks where height > $1));

-- flips_data
delete
from flips_data
where block_id in
      (select id from blocks where height > $1);

-- flips
delete
from flips
where tx_id in
      (select id
       from transactions
       where block_id in
             (select id from blocks where height > $1));
update flips
set status_block_id=null,
    status=null,
    answer=null
where status_block_id in (select id from blocks where height > $1);

-- flip_keys
delete
from flip_keys
where tx_id in
      (select t.id
       from transactions t
                join blocks b on b.id = t.block_id
       where b.height > $1);

-- transactions
delete
from transactions
where block_id in
      (select id from blocks where height > $1);

-- proposers
delete
from proposers
where block_id in (select id from blocks where height > $1);

-- temporary_identities
delete
from temporary_identities
where block_id in (select id from blocks where height > $1);

-- addresses
delete
from addresses
where block_id in (select id from blocks where height > $1);

-- block_flags
delete
from block_flags
where block_id in (select id from blocks where height > $1);

-- blocks
delete
from blocks
where height > $1;

-- epochs
delete
from epochs
where id not in (select distinct epoch_id from blocks);