CREATE OR REPLACE PROCEDURE save_tokens(p_block_height bigint,
                                        p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                jsonb;
    l_contract_address_id bigint;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            l_contract_address_id = get_address_id_or_insert(p_block_height, (l_item ->> 'contractAddress')::text);

            INSERT INTO tokens (contract_address_id, "name", symbol, decimals)
            VALUES (l_contract_address_id, limited_text(l_item ->> 'name', 50), limited_text(l_item ->> 'symbol', 10),
                    (l_item ->> 'decimals')::smallint);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_token_balance_updates(p_block_height bigint,
                                                       p_items jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CHANGE_TYPE_TOKEN_BALANCES CONSTANT smallint = 7;
    l_item                              jsonb;
    l_contract_address_id               bigint;
    l_address                           text;
    l_prev_balance                      numeric;
    l_change_id                         bigint;
    l_balance                           numeric;
BEGIN
    if p_items is null then
        return;
    end if;
    for i in 0..jsonb_array_length(p_items) - 1
        loop
            l_item = (p_items ->> i)::jsonb;
            l_contract_address_id = get_address_id_or_insert(p_block_height, (l_item ->> 'contractAddress')::text);
            l_address = (l_item ->> 'address')::text;
            l_balance = l_item ->> 'balance';

            if not exists(SELECT 1 FROM tokens WHERE contract_address_id = l_contract_address_id) then
                INSERT INTO tokens (contract_address_id, "name", symbol, decimals)
                VALUES (l_contract_address_id, null, null, null);
            end if;

            SELECT balance
            INTO l_prev_balance
            FROM token_balances
            WHERE contract_address_id = l_contract_address_id
              AND lower(address) = lower(l_address);

            if l_prev_balance is null then
                if l_balance = 0 then
                    continue;
                end if;
                INSERT INTO token_balances (contract_address_id, address, balance)
                VALUES (l_contract_address_id, l_address, l_balance);
            else
                if l_balance = 0 then
                    DELETE
                    FROM token_balances
                    WHERE contract_address_id = l_contract_address_id
                      AND lower(address) = lower(l_address);
                else
                    UPDATE token_balances
                    SET balance = l_balance
                    WHERE contract_address_id = l_contract_address_id
                      AND lower(address) = lower(l_address);
                end if;
            end if;

            INSERT INTO changes (block_height, "type")
            VALUES (p_block_height, CHANGE_TYPE_TOKEN_BALANCES)
            RETURNING id INTO l_change_id;

            INSERT INTO token_balances_changes (change_id, contract_address_id, address, balance)
            VALUES (l_change_id, l_contract_address_id, l_address, l_prev_balance);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE reset_token_balances_changes(p_change_id bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_contract_address_id bigint;
    l_address             text;
    l_balance             numeric;
BEGIN
    SELECT contract_address_id, address, balance
    INTO l_contract_address_id, l_address, l_balance
    FROM token_balances_changes
    WHERE change_id = p_change_id;

    if l_balance is null then
        DELETE
        FROM token_balances
        WHERE contract_address_id = l_contract_address_id
          AND lower(address) = lower(l_address);
    else
        INSERT INTO token_balances (contract_address_id, address, balance)
        VALUES (l_contract_address_id, l_address, l_balance)
        ON CONFLICT (contract_address_id, lower(address)) DO UPDATE SET balance = l_balance;
    end if;

    DELETE FROM token_balances_changes WHERE change_id = p_change_id;
END
$$;