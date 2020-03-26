-- flips_queue
delete
from flips_queue
where lower(cid) in (
    select f.cid
    from flips f,
         transactions t,
         blocks b
    where f.tx_id = t.id
      and t.block_height = b.height
      and b.epoch + 1 > (select epoch from blocks where height = greatest(2, $1))
);

-- flip_pics
delete
from flip_pics
where fd_flip_tx_id in
      (select t.id
       from transactions t
                join blocks b on b.height = t.block_height and
                                 b.epoch + 1 > (select epoch from blocks where height = greatest(2, $1)));

-- flip_icons
delete
from flip_icons
where fd_flip_tx_id in
      (select t.id
       from transactions t
                join blocks b on b.height = t.block_height and
                                 b.epoch + 1 > (select epoch from blocks where height = greatest(2, $1)));

-- flip_pic_orders
delete
from flip_pic_orders
where fd_flip_tx_id in
      (select t.id
       from transactions t
                join blocks b on b.height = t.block_height and
                                 b.epoch + 1 > (select epoch from blocks where height = greatest(2, $1)));

-- flips_data
delete
from flips_data
where flip_tx_id in
      (select t.id
       from transactions t
                join blocks b on b.height = t.block_height and
                                 b.epoch + 1 > (select epoch from blocks where height = greatest(2, $1)));

delete
from rewarded_flips
where flip_tx_id in
      (select t.id
       from transactions t
                join blocks b on b.height = t.block_height and
                                 b.epoch + 1 > (select epoch from blocks where height = greatest(2, $1)));

-- block_proposer_vrf_scores
delete
from block_proposer_vrf_scores
where block_height > $1;

-- burnt_coins
delete
from burnt_coins
where block_height > $1;

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
delete
from balances;

-- birthdays
delete
from birthdays;

-- coins
delete
from coins
where block_height > $1;

-- mem_pool_flip_keys
delete
from mem_pool_flip_keys
where ei_address_state_id in
      (select id
       from address_states
       where block_height > $1);

-- flips_to_solve
delete
from flips_to_solve
where ei_address_state_id in
      (select id
       from address_states
       where block_height > $1);

-- answers
delete
from answers
where ei_address_state_id in
      (select id
       from address_states
       where block_height > $1);

-- validation_rewards
delete
from validation_rewards
where ei_address_state_id in
      (select id
       from address_states
       where block_height > $1);

-- reward_ages
delete
from reward_ages
where ei_address_state_id in
      (select id
       from address_states
       where block_height > $1);

-- bad_authors
delete
from bad_authors
where ei_address_state_id in
      (select id
       from address_states
       where block_height > $1);

-- bad_authors
delete
from good_authors
where ei_address_state_id in
      (select id
       from address_states
       where block_height > $1);

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
    answer=null,
    wrong_words=null
where status_block_height > $1;
update flips
set delete_tx_id=null
where delete_tx_id in (select t.id
                       from transactions t
                       where t.block_height > $1);

-- flip_keys
delete
from flip_keys
where tx_id in
      (select t.id
       from transactions t
       where t.block_height > $1);

-- activation_txs
delete
from activation_txs
where tx_id in
      (select t.id
       from transactions t
       where t.block_height > $1);

-- kill_invitee_txs
delete
from kill_invitee_txs
where tx_id in
      (select t.id
       from transactions t
       where t.block_height > $1);

delete
from become_online_txs
where tx_id in
      (select t.id
       from transactions t
       where t.block_height > $1);

delete
from become_offline_txs
where tx_id in
      (select t.id
       from transactions t
       where t.block_height > $1);

-- activation_tx_transfers
delete
from activation_tx_transfers
where tx_id in
      (select t.id
       from transactions t
       where t.block_height > $1);

-- kill_tx_transfers
delete
from kill_tx_transfers
where tx_id in
      (select t.id
       from transactions t
       where t.block_height > $1);

-- kill_invitee_tx_transfers
delete
from kill_invitee_tx_transfers
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

-- mining_rewards
delete
from mining_rewards
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