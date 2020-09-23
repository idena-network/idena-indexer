-- epochs
insert into epochs (select *
                    from OLD_SCHEMA_TAG.epochs
                    where epoch in (select distinct epoch from OLD_SCHEMA_TAG.blocks where height <= $1));

-- blocks
insert into blocks (select * from OLD_SCHEMA_TAG.blocks where height <= $1);

-- block_flags
insert into block_flags (select * from OLD_SCHEMA_TAG.block_flags where block_height <= $1);

-- addresses
insert into addresses (select * from OLD_SCHEMA_TAG.addresses where block_height <= $1);
-- addresses sequence
select setval('addresses_id_seq', max(id))
from addresses;

-- temporary_identities
insert into temporary_identities (select * from OLD_SCHEMA_TAG.temporary_identities where block_height <= $1);

-- block_proposers
insert into block_proposers (select * from OLD_SCHEMA_TAG.block_proposers where block_height <= $1);

-- block_proposer_vrf_scores
insert into block_proposer_vrf_scores (select * from OLD_SCHEMA_TAG.block_proposer_vrf_scores where block_height <= $1);

-- mining_rewards
insert into mining_rewards (select * from OLD_SCHEMA_TAG.mining_rewards where block_height <= $1);

insert into transactions (select * from OLD_SCHEMA_TAG.transactions where block_height <= $1);
select setval('transactions_id_seq', max(id))
from transactions;

insert into rewarded_invitations (select * from OLD_SCHEMA_TAG.rewarded_invitations where block_height <= $1);

insert into transaction_raws (select *
                              from OLD_SCHEMA_TAG.transaction_raws
                              where tx_id in (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

insert into activation_txs (select *
                            from OLD_SCHEMA_TAG.activation_txs
                            where tx_id in (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

insert into kill_invitee_txs (select *
                              from OLD_SCHEMA_TAG.kill_invitee_txs
                              where tx_id in (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

insert into become_online_txs (select *
                               from OLD_SCHEMA_TAG.become_online_txs
                               where tx_id in (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

insert into become_offline_txs (select *
                                from OLD_SCHEMA_TAG.become_offline_txs
                                where tx_id in (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

insert into activation_tx_transfers (select *
                                     from OLD_SCHEMA_TAG.activation_tx_transfers
                                     where tx_id in
                                           (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

insert into kill_tx_transfers (select *
                               from OLD_SCHEMA_TAG.kill_tx_transfers
                               where tx_id in (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

insert into kill_invitee_tx_transfers (select *
                                       from OLD_SCHEMA_TAG.kill_invitee_tx_transfers
                                       where tx_id in
                                             (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

-- burnt_coins
insert into burnt_coins (select * from OLD_SCHEMA_TAG.burnt_coins where block_height <= $1);

-- flip_keys
insert into flip_keys (select *
                       from OLD_SCHEMA_TAG.flip_keys
                       where tx_id in (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

-- flips
insert into flips (select tx_id,
                          cid,
                          size,
                          pair,
                          (case when status_block_height <= $1 then status_block_height else null end),
                          (case when status_block_height <= $1 then answer else null end),
                          (case when status_block_height <= $1 then status else null end),
                          (case
                               when delete_tx_id <= (select max(id) from transactions) then delete_tx_id
                               else null end),
                          (case when status_block_height <= $1 then grade else null end)
                   from OLD_SCHEMA_TAG.flips
                   where tx_id in (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

--flip_words
insert into flip_words (select *
                        from OLD_SCHEMA_TAG.flip_words
                        where tx_id in (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

-- flips_data
insert into flips_data (select *
                        from OLD_SCHEMA_TAG.flips_data
                        where flip_tx_id in (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

-- flip_pic_orders
insert into flip_pic_orders (select *
                             from OLD_SCHEMA_TAG.flip_pic_orders
                             where fd_flip_tx_id in
                                   (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

-- flip_icons
insert into flip_icons (select *
                        from OLD_SCHEMA_TAG.flip_icons
                        where fd_flip_tx_id in
                              (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

-- flip_pics
insert into flip_pics (select *
                       from OLD_SCHEMA_TAG.flip_pics
                       where fd_flip_tx_id in
                             (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

insert into rewarded_flips (select *
                            from OLD_SCHEMA_TAG.rewarded_flips
                            where flip_tx_id in
                                  (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

insert into reported_flip_rewards (select *
                                   from OLD_SCHEMA_TAG.reported_flip_rewards
                                   where flip_tx_id in
                                         (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

insert into flip_summaries (select *
                            from OLD_SCHEMA_TAG.flip_summaries
                            where flip_tx_id in
                                  (select id from OLD_SCHEMA_TAG.transactions where block_height <= $1));

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

insert into epoch_identity_interim_states (select *
                                           from OLD_SCHEMA_TAG.epoch_identity_interim_states
                                           where block_height <= $1);

-- epoch_identities
insert into epoch_identities (select *
                              from OLD_SCHEMA_TAG.epoch_identities
                              where address_state_id in
                                    (select id from OLD_SCHEMA_TAG.address_states where block_height <= $1));

-- mem_pool_flip_keys
insert into mem_pool_flip_keys
    (select *
     from OLD_SCHEMA_TAG.mem_pool_flip_keys
     where ei_address_state_id in (select id from OLD_SCHEMA_TAG.address_states where block_height <= $1));

-- answers
insert into answers
    (select *
     from OLD_SCHEMA_TAG.answers
     where ei_address_state_id in (select id from OLD_SCHEMA_TAG.address_states where block_height <= $1));

-- flips_to_solve
insert into flips_to_solve
    (select *
     from OLD_SCHEMA_TAG.flips_to_solve
     where ei_address_state_id in (select id from OLD_SCHEMA_TAG.address_states where block_height <= $1));

-- coins
insert into coins (select * from OLD_SCHEMA_TAG.coins where block_height <= $1);

insert into epoch_summaries (select * from OLD_SCHEMA_TAG.epoch_summaries where block_height <= $1);

-- penalties
insert into penalties (select * from OLD_SCHEMA_TAG.penalties where block_height <= $1);

-- penalties sequence
select setval('penalties_id_seq', max(id))
from penalties;

-- paid_penalties
insert into paid_penalties (select * from OLD_SCHEMA_TAG.paid_penalties where block_height <= $1);

-- total_rewards
insert into total_rewards (select * from OLD_SCHEMA_TAG.total_rewards where block_height <= $1);

-- fund_rewards
insert into fund_rewards (select * from OLD_SCHEMA_TAG.fund_rewards where block_height <= $1);

-- bad_authors
insert into bad_authors
    (select *
     from OLD_SCHEMA_TAG.bad_authors
     where ei_address_state_id in (select id from OLD_SCHEMA_TAG.address_states where block_height <= $1));

-- total_rewards
insert into good_authors
    (select *
     from OLD_SCHEMA_TAG.good_authors
     where ei_address_state_id in (select id from OLD_SCHEMA_TAG.address_states where block_height <= $1));

-- validation_rewards
insert into validation_rewards
    (select *
     from OLD_SCHEMA_TAG.validation_rewards
     where ei_address_state_id in (select id from OLD_SCHEMA_TAG.address_states where block_height <= $1));

-- reward_ages
insert into reward_ages
    (select *
     from OLD_SCHEMA_TAG.reward_ages
     where ei_address_state_id in (select id from OLD_SCHEMA_TAG.address_states where block_height <= $1));

insert into saved_invite_rewards
    (select *
     from OLD_SCHEMA_TAG.saved_invite_rewards
     where ei_address_state_id in (select id from OLD_SCHEMA_TAG.address_states where block_height <= $1));

-- failed_validations
insert into failed_validations (select * from OLD_SCHEMA_TAG.failed_validations where block_height <= $1);

-- flip_key_timestamps (mem pool data)
insert into flip_key_timestamps (select *
                                 from OLD_SCHEMA_TAG.flip_key_timestamps
                                 where epoch in (select epoch from epochs));

-- answers_hash_tx_timestamps (mem pool data)
insert into answers_hash_tx_timestamps (select *
                                        from OLD_SCHEMA_TAG.answers_hash_tx_timestamps
                                        where epoch in (select epoch from epochs));

call migrate_balance_updates($1, 'OLD_SCHEMA_TAG');

call restore_coins_summary();

call restore_epoch_summary($1);

call restore_address_summaries();