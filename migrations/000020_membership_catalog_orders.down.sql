DROP TABLE IF EXISTS membership_orders;
DROP TABLE IF EXISTS membership_usage;
DROP TABLE IF EXISTS membership_plan_quotas;

ALTER TABLE membership_plans
    DROP COLUMN IF EXISTS display_order,
    DROP COLUMN IF EXISTS active,
    DROP COLUMN IF EXISTS recommended,
    DROP COLUMN IF EXISTS features,
    DROP COLUMN IF EXISTS billing_cycle,
    DROP COLUMN IF EXISTS price_cents;
