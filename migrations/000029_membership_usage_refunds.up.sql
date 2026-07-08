ALTER TABLE membership_usage_events
    DROP CONSTRAINT IF EXISTS membership_usage_events_amount_check;

ALTER TABLE membership_usage_events
    ADD CONSTRAINT membership_usage_events_amount_check CHECK (amount <> 0);
