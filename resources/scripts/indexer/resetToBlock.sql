-- failed_validations
delete
from failed_validations
where block_height > $1;

-- fund_rewards
delete
from fund_rewards
where block_height > $1;

-- total_rewards
delete
from total_rewards
where block_height > $1;

-- paid_penalties
delete
from paid_penalties
where block_height > $1;

-- penalties
delete
from penalties
where block_height > $1;

-- epoch_summaries
delete
from epoch_summaries
where block_height > $1;

-- balances
delete from balances;

-- birthdays
delete from birthdays;

-- coins
delete
from coins
where block_height > $1;

-- flips_to_solve
delete
from flips_to_solve
where epoch_identity_id in
      (select id
       from epoch_identities
       where address_state_id in
             (select id
              from address_states
              where block_height > $1));

-- answers
delete
from answers
where epoch_identity_id in
      (select id
       from epoch_identities
       where address_state_id in
             (select id
              from address_states
              where block_height > $1));

-- validation_rewards
delete
from validation_rewards
where epoch_identity_id in
      (select id
       from epoch_identities
       where address_state_id in
             (select id
              from address_states
              where block_height > $1));

-- reward_ages
delete
from reward_ages
where epoch_identity_id in
      (select id
       from epoch_identities
       where address_state_id in
             (select id
              from address_states
              where block_height > $1));

-- bad_authors
delete
from bad_authors
where epoch_identity_id in
      (select id
       from epoch_identities
       where address_state_id in
             (select id
              from address_states
              where block_height > $1));

-- bad_authors
delete
from good_authors
where epoch_identity_id in
      (select id
       from epoch_identities
       where address_state_id in
             (select id
              from address_states
              where block_height > $1));

-- epoch_identities
delete
from epoch_identities
where address_state_id in
      (select id
       from address_states
       where block_height > $1);

-- address_states
delete
from address_states
where block_height > $1;
-- restore actual states
update address_states
set is_actual = true
where id in
      (select s.id
       from address_states s
       where (s.address_id, s.block_height) in
             (select s.address_id, max(s.block_height)
              from address_states s
              group by address_id)
         and not s.is_actual);

-- flip_pics
delete
from flip_pics
where flip_data_id in
      (select id
       from flips_data
       where block_height > $1);

-- flip_icons
delete
from flip_icons
where flip_data_id in
      (select id
       from flips_data
       where block_height > $1);

-- flip_pic_orders
delete
from flip_pic_orders
where flip_data_id in
      (select id
       from flips_data
       where block_height > $1);

-- flips_data
delete
from flips_data
where block_height > $1;

-- flip_words
delete
from flip_words
where tx_id in
      (select t.id
       from transactions t
       where t.block_height > $1);

-- flips
delete
from flips
where tx_id in
      (select id
       from transactions
       where block_height > $1);
update flips
set status_block_height=null,
    status=null,
    answer=null
where status_block_height > $1;

-- flip_keys
delete
from flip_keys
where tx_id in
      (select t.id
       from transactions t
       where t.block_height > $1);

-- transactions
delete
from transactions
where block_height > $1;

-- block_proposers
delete
from block_proposers
where block_height > $1;

-- block_validators
delete
from block_validators
where block_height > $1;

-- temporary_identities
delete
from temporary_identities
where block_height > $1;

-- addresses
delete
from addresses
where block_height > $1;

-- block_flags
delete
from block_flags
where block_height > $1;

-- blocks
delete
from blocks
where height > $1;

-- epochs
delete
from epochs
where epoch not in (select distinct epoch from blocks);