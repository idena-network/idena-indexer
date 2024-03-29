CREATE OR REPLACE PROCEDURE save_block(p_height bigint,
                                       p_hash text,
                                       p_epoch bigint,
                                       p_timestamp bigint,
                                       p_is_empty boolean,
                                       p_original_validators_count integer,
                                       p_pool_validators_count integer,
                                       p_body_size integer,
                                       p_vrf_proposer_threshold double precision,
                                       p_full_size integer,
                                       p_fee_rate numeric,
                                       p_upgrade integer,
                                       p_offline_address text,
                                       p_flags text[],
                                       p_used_gas integer)
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_empty_count               smallint;
    l_flip_lottery_block_height bigint;
    l_offline_address_id        bigint;
BEGIN
    if p_offline_address is not null and p_offline_address <> '' then
        SELECT id INTO l_offline_address_id FROM addresses WHERE lower(address) = lower(p_offline_address);
    end if;
    INSERT INTO blocks (height, hash, epoch, timestamp, is_empty, validators_count, pool_validators_count, body_size,
                        vrf_proposer_threshold, full_size, fee_rate, upgrade, offline_address_id, used_gas)
    VALUES (p_height, p_hash, p_epoch, p_timestamp, p_is_empty, p_original_validators_count, p_pool_validators_count,
            p_body_size, p_vrf_proposer_threshold, p_full_size, p_fee_rate, p_upgrade, l_offline_address_id,
            p_used_gas);

    if p_is_empty then
        l_empty_count = 1;
    end if;

    if p_flags is not null then
        for i in 1..cardinality(p_flags)
            loop
                INSERT INTO block_flags (block_height, flag) VALUES (p_height, p_flags[i]);
                if p_flags[i] = 'FlipLotteryStarted' then
                    l_flip_lottery_block_height = p_height;
                end if;
            end loop;
    end if;

    call update_epoch_summary(p_block_height => p_height,
                              p_block_count_diff => 1,
                              p_empty_block_count_diff => l_empty_count,
                              p_flip_lottery_block_height => l_flip_lottery_block_height);
END
$BODY$;