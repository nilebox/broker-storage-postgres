CREATE TABLE IF NOT EXISTS instance (
   id BIGSERIAL PRIMARY KEY,
   instance_id TEXT NOT NULL UNIQUE,
   service_id TEXT NOT NULL,
   plan_id TEXT NOT NULL,
   parameters TEXT NOT NULL,
   outputs TEXT NOT NULL,
   state TEXT NOT NULL,
   error TEXT NOT NULL,
   created TIMESTAMP NOT NULL,
   modified TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS instance_state_modified ON instance (state, modified);