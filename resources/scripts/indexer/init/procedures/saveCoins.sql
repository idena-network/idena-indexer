CREATE OR REPLACE PROCEDURE save_coins(p_block_height bigint,
                                       p_burnt numeric,
                                       p_minted numeric,
                                       p_total_balance numeric,
                                       p_total_stake numeric)
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_check_value numeric;
BEGIN
    insert into coins (block_height, burnt, minted, total_balance, total_stake)
    values (p_block_height, p_burnt, p_minted, p_total_balance, p_total_stake);

    update coins_summary
    set total_burnt  = total_burnt + p_burnt,
        total_minted = total_minted + p_minted
    returning total_burnt into l_check_value;
    if l_check_value is null then
        insert into coins_summary (total_burnt, total_minted) values (p_burnt, p_minted);
    end if;

    call update_epoch_summary(p_block_height => p_block_height,
                              p_burnt_diff => p_burnt,
                              p_minted_diff => p_minted,
                              p_total_balance => p_total_balance,
                              p_total_stake => p_total_stake);
END
$BODY$;

CREATE OR REPLACE PROCEDURE restore_coins_summary()
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_burnt  numeric;
    l_minted numeric;
BEGIN
    delete from coins_summary;

    select sum(burnt), sum(minted)
    into l_burnt, l_minted
    from coins;

    if l_burnt is not null then
        insert into coins_summary (total_burnt, total_minted) values (l_burnt, l_minted);
    end if;
END
$BODY$;