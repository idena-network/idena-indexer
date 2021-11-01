ALTER TABLE transactions
    ADD COLUMN nonce integer;

DROP TYPE tp_tx CASCADE;

DO
$$
    DECLARE
        l_record record;
        l_nonce  smallint;
        l_epoch  smallint;
    BEGIN
        CREATE TABLE tmp_nonces
        (
            address_id bigint   NOT NULL,
            epoch      smallint NOT NULL,
            nonce      integer  NOT NULL
        );
        CREATE UNIQUE INDEX tmp_nonces_pk ON tmp_nonces (address_id);
        for l_record in SELECT t.id, t.from as address_d, b.epoch
                        FROM transactions t
                                 LEFT JOIN blocks b on b.height = t.block_height
                        ORDER BY id
            loop
                SELECT epoch, nonce INTO l_epoch, l_nonce FROM tmp_nonces WHERE address_id = l_record.address_d;
                if l_nonce is null then
                    UPDATE transactions SET nonce = 1 WHERE id = l_record.id;
                    INSERT INTO tmp_nonces (address_id, epoch, nonce) VALUES (l_record.address_d, l_record.epoch, 1);
                    continue;
                end if;
                if l_epoch <> l_record.epoch then
                    UPDATE transactions SET nonce = 1 WHERE id = l_record.id;
                    UPDATE tmp_nonces SET epoch = l_record.epoch, nonce = 1 WHERE address_id = l_record.address_d;
                    continue;
                end if;
                UPDATE transactions SET nonce = l_nonce + 1 WHERE id = l_record.id;
                UPDATE tmp_nonces SET nonce = l_nonce + 1 WHERE address_id = l_record.address_d;
            end loop;

        DROP TABLE tmp_nonces;
    END
$$;

ALTER TABLE transactions
    ALTER COLUMN nonce SET NOT NULL;