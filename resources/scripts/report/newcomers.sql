CREATE MATERIALIZED VIEW report.mv_newcomers AS
SELECT epoch + 1 AS epoch, count(*) AS cnt
FROM (SELECT s.address_id, min(b.epoch) AS epoch
      FROM indexer.address_states s
               JOIN indexer.blocks b ON b.height = s.block_height
               LEFT JOIN indexer.address_states prevs ON b.epoch = 7 AND prevs.id = s.prev_id
      WHERE s.state = 7 AND (prevs.state IS NULL OR prevs.state <> 7)
      GROUP BY s.address_id) newcomers
GROUP BY epoch
ORDER BY epoch;

CREATE OR REPLACE PROCEDURE report.refresh_newcomers()
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    refresh materialized view report.mv_newcomers;
END
$$;

INSERT INTO report.dynamic_endpoints
("name", refresh_procedure, refresh_period, refresh_delay_minutes, endpoint_method, "limit")
VALUES ('report.mv_newcomers', 'report.refresh_newcomers', 'e', 30, 'Newcomers', null);