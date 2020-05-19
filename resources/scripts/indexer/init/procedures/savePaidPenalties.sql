CREATE OR REPLACE PROCEDURE save_paid_penalties(p_height bigint, p_paid_penalties tp_paid_penalty[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    l_id           bigint;
    l_paid_penalty tp_paid_penalty;
    l_full_amount  numeric;
BEGIN
    for i in 1..cardinality(p_paid_penalties)
        loop
            l_paid_penalty := p_paid_penalties[i];
            select p.id, p.penalty
            into l_id, l_full_amount
            from penalties p
                     join addresses a on a.id = p.address_id and lower(a.address) = lower(l_paid_penalty.address)
            order by p.id desc
            limit 1;

            if l_id is null then
                raise exception 'there is no penalty to close for address %', l_paid_penalty.address;
            end if;

            if (select exists(select 1
                              from paid_penalties
                              where penalty_id = l_id)) then
                raise exception 'latest penalty is already closed for address %', l_paid_penalty.address;
            end if;

            insert into paid_penalties (penalty_id, penalty, block_height)
            values (l_id, l_full_amount - l_paid_penalty.burnt_penalty_amount, p_height);
        end loop;
END
$BODY$;