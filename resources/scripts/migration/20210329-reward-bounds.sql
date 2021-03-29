DROP PROCEDURE save_epoch_result;

CREATE TABLE IF NOT EXISTS epoch_reward_bounds
(
    epoch          bigint          NOT NULL,
    bound_type     smallint        NOT NULL,
    min_amount     numeric(30, 18) NOT NULL,
    min_address_id bigint          NOT NULL,
    max_amount     numeric(30, 18) NOT NULL,
    max_address_id bigint          NOT NULL,
    CONSTRAINT epoch_reward_bounds_pkey PRIMARY KEY (epoch, bound_type)
);

DO
$$
    DECLARE
        l_rec           record;
        l_address_id    bigint;
        l_amount        numeric;
        l_type          smallint;
        l_min_amount    numeric;
        l_max_amount    numeric;
        l_god_address_1 bigint;
        l_god_address_2 bigint;
    BEGIN

        select id
        into l_god_address_1
        from addresses
        where lower(address) = lower('0x4d60dC6A2CbA8c3EF1Ba5e1Eba5c12c54cEE6B61');
        select id
        into l_god_address_2
        from addresses
        where lower(address) = lower('0xcbb98843270812eeCE07BFb82d26b4881a33aA91');

        CREATE TABLE IF NOT EXISTS epoch_reward_bounds_l
        (
            epoch          bigint          NOT NULL,
            bound_type     smallint        NOT NULL,
            min_amount     numeric(30, 18) NOT NULL,
            min_address_id bigint          NOT NULL,
            max_amount     numeric(30, 18) NOT NULL,
            max_address_id bigint          NOT NULL,
            CONSTRAINT epoch_reward_bounds_l_pkey PRIMARY KEY (epoch, bound_type)
        );

        for l_rec in select * from epoch_identities
            loop
                select sum(balance + stake)
                into l_amount
                from validation_rewards
                where ei_address_state_id = l_rec.address_state_id;

                if l_amount is null or l_amount = 0 then
                    continue;
                end if;

                l_address_id = (select address_id from address_states where id = l_rec.address_state_id);

                -- exclude god addresses
                if l_address_id = l_god_address_1 or l_address_id = l_god_address_2 then
                    continue;
                end if;

                l_type = l_rec.epoch - l_rec.birth_epoch + 1;
                if l_type > 6 then
                    l_type = 6;
                end if;

                select min_amount, max_amount
                into l_min_amount, l_max_amount
                from epoch_reward_bounds_l
                where epoch = l_rec.epoch
                  and bound_type = l_type;

                if l_min_amount is null then
                    insert into epoch_reward_bounds_l
                    values (l_rec.epoch, l_type, l_amount, l_address_id, l_amount, l_address_id);
                    continue;
                end if;

                if l_min_amount > l_amount then
                    update epoch_reward_bounds_l
                    set min_amount     = l_amount,
                        min_address_id = l_address_id
                    where epoch = l_rec.epoch
                      and bound_type = l_type;
                end if;

                if l_max_amount < l_amount then
                    update epoch_reward_bounds_l
                    set max_amount     = l_amount,
                        max_address_id = l_address_id
                    where epoch = l_rec.epoch
                      and bound_type = l_type;
                end if;

            end loop;

        insert into epoch_reward_bounds select * from epoch_reward_bounds_l order by epoch, bound_type;

        DROP TABLE epoch_reward_bounds_l;
    END
$$