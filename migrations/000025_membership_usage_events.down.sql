DROP TABLE IF EXISTS membership_usage_events;

DELETE FROM membership_plan_quotas
WHERE key = 'competitor_scans'
   OR (plan_code = 'free' AND key = 'sandbox_runs');
