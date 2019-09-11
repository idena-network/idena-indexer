-- epochs
insert into epochs (select *
                    from OLD_SCHEMA_TAG.epochs
                    where epoch in (select distinct epoch from OLD_SCHEMA_TAG.blocks where height <= $1));

-- blocks
insert into blocks (select * from OLD_SCHEMA_TAG.blocks where height <= $1);

-- block_flags
insert into block_flags (select * from OLD_SCHEMA_TAG.block_flags where block_height <= $1);
-- block_flags sequence
select setval('block_flags_id_seq', max(id))
from block_flags;

-- addresses
insert into addresses (select * from OLD_SCHEMA_TAG.addresses where block_height <= $1);
-- addresses sequence
select setval('addresses_id_seq', max(id))
from addresses;

-- temporary_identities
insert into temporary_identities (select * from OLD_SCHEMA_TAG.temporary_identities where block_height <= $1);

-- block_validators
insert into block_validators (select * from OLD_SCHEMA_TAG.block_validators where block_height <= $1);

-- block_proposers
insert into block_proposers (select * from OLD_SCHEMA_TAG.block_proposers where block_height <= $1);

-- transactions
insert into transactions (select * from OLD_SCHEMA_TAG.transactions where block_height <= $1);
-- transactions sequence
select setval('transactions_id_seq', max(id))
from transactions;

-- flip_keys
insert into flip_keys (select *
                       from OLD_SCHEMA_TAG.flip_keys
                       where tx_id in (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));
-- flip_keys sequence
select setval('flip_keys_id_seq', max(id))
from flip_keys;

-- flips
insert into flips (select *
                   from OLD_SCHEMA_TAG.flips
                   where tx_id in (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));
update flips
set status_block_height=null,
    status=null,
    answer=null
where status_block_height > $1;
-- flips sequence
select setval('flips_id_seq', max(id))
from flips;

-- flips_data
insert into flips_data (select * from OLD_SCHEMA_TAG.flips_data where block_height <= $1);
-- flips_data sequence
select setval('flips_data_id_seq', max(id))
from flips_data;

-- flip_pic_orders
insert into flip_pic_orders (select *
                             from OLD_SCHEMA_TAG.flip_pic_orders
                             where flip_data_id in (select id from OLD_SCHEMA_TAG.flips_data where block_height <= $1));

-- flip_icons
insert into flip_icons (select *
                        from OLD_SCHEMA_TAG.flip_icons
                        where flip_data_id in (select id from OLD_SCHEMA_TAG.flips_data where block_height <= $1));

-- flip_pics
insert into flip_pics (select *
                       from OLD_SCHEMA_TAG.flip_pics
                       where flip_data_id in (select id from OLD_SCHEMA_TAG.flips_data where block_height <= $1));

-- address_states
insert into address_states (select * from OLD_SCHEMA_TAG.address_states where block_height <= $1);
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
-- address_states sequence
select setval('address_states_id_seq', max(id))
from address_states;

-- epoch_identities
insert into epoch_identities (select *
                              from OLD_SCHEMA_TAG.epoch_identities
                              where address_state_id in
                                    (select id from OLD_SCHEMA_TAG.address_states where block_height <= $1));
-- epoch_identities sequence
select setval('epoch_identities_id_seq', max(id))
from epoch_identities;

-- answers
insert into answers
    (select *
     from OLD_SCHEMA_TAG.answers
     where epoch_identity_id in (select id
                                 from OLD_SCHEMA_TAG.epoch_identities
                                 where address_state_id in
                                       (select id from OLD_SCHEMA_TAG.address_states where block_height <= $1)));
-- answers sequence
select setval('answers_id_seq', max(id))
from answers;

-- flips_to_solve
insert into flips_to_solve
    (select *
     from OLD_SCHEMA_TAG.flips_to_solve
     where epoch_identity_id in (select id
                                 from OLD_SCHEMA_TAG.epoch_identities
                                 where address_state_id in
                                       (select id from OLD_SCHEMA_TAG.address_states where block_height <= $1)));
-- flips_to_solve sequence
select setval('flips_to_solve_id_seq', max(id))
from flips_to_solve;

-- coins
insert into coins (select * from OLD_SCHEMA_TAG.coins where block_height <= $1);

-- epoch_summaries
insert into epoch_summaries (select * from OLD_SCHEMA_TAG.epoch_summaries where block_height <= $1);