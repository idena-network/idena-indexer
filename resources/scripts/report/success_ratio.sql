CREATE MATERIALIZED VIEW report.mv_success_ratio AS
SELECT identities.epoch + 1                                                 AS epoch,
       (CASE
            WHEN participants.total = 0 THEN 0
            ELSE identities.total::double precision / participants.total::real END) AS success
FROM (SELECT b.epoch, count(*) AS total
      FROM indexer.address_states s
               JOIN indexer.blocks b ON b.height = s.block_height AND s.block_height > 181001
      WHERE s.state IN (3, 7, 8)
      GROUP BY b.epoch) identities
         JOIN (SELECT participants.epoch, count(*) AS total
               FROM (SELECT DISTINCT b.epoch, t.from address_id
                     FROM indexer.transactions t
                              JOIN indexer.blocks b ON b.height = t.block_height
                     WHERE t.type IN (5, 6, 7, 8)) participants
               GROUP BY epoch) participants ON participants.epoch = identities.epoch
ORDER BY identities.epoch;

CREATE OR REPLACE PROCEDURE report.refresh_success_ratio()
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    refresh materialized view report.mv_success_ratio;
END
$$;

INSERT INTO report.dynamic_endpoints
("name", refresh_procedure, refresh_period, refresh_delay_minutes, endpoint_method, "limit")
VALUES ('report.mv_success_ratio', 'report.refresh_success_ratio', 'e', 30, 'SuccessRatio', null);