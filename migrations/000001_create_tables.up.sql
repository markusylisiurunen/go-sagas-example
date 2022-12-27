BEGIN;

CREATE TABLE events (
  event_id                BIGSERIAL PRIMARY KEY,
  event_uuid              TEXT NOT NULL,
  event_published_at      TIMESTAMPTZ NOT NULL,
  event_topic             TEXT NOT NULL,
  event_queue             TEXT NOT NULL,
  event_name              TEXT NOT NULL,
  event_status            TEXT NOT NULL,
  event_deliver_at        TIMESTAMPTZ NOT NULL,
  event_delivery_attempts INTEGER NOT NULL DEFAULT 0,
  event_data              JSON NOT NULL,

  UNIQUE (event_queue, event_uuid),

  CONSTRAINT events_event_status_check CHECK (event_status IN ('pending', 'processed', 'dropped'))
);

CREATE INDEX events_status_queue_published_at_idx
ON events (event_status, event_queue, event_published_at);

CREATE INDEX events_uuid_queue_idx
ON events (event_uuid, event_queue);

CREATE FUNCTION notify_of_new_events() RETURNS TRIGGER AS $$
DECLARE
  notification JSON;
BEGIN
  notification = json_build_object(
    'topic', NEW .event_topic,
    'queue', NEW .event_queue,
    'uuid',  NEW .event_uuid
  );

  PERFORM pg_notify('__events', notification::TEXT);

  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER notify_of_new_events_trigger AFTER INSERT ON events
FOR EACH ROW EXECUTE PROCEDURE notify_of_new_events();

COMMIT;
