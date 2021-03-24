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
                                       p_upgrade integer)
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_empty_count smallint;
BEGIN
    INSERT INTO blocks (height, hash, epoch, timestamp, is_empty, validators_count, pool_validators_count, body_size, vrf_proposer_threshold,
                        full_size, fee_rate, upgrade)
    VALUES (p_height, p_hash, p_epoch, p_timestamp, p_is_empty, p_original_validators_count, p_pool_validators_count,
            p_body_size, p_vrf_proposer_threshold, p_full_size, p_fee_rate, p_upgrade);

    if p_is_empty then
        l_empty_count = 1;
    end if;

    call update_epoch_summary(p_block_height => p_height,
                              p_block_count_diff => 1,
                              p_empty_block_count_diff => l_empty_count);
END
$BODY$;