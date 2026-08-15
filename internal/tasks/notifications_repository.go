package tasks

import (
	"context"
	"time"
)

func (r *PostgresRepository) DispatchTaskNotifications(ctx context.Context, now time.Time, limit int) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		WITH lifecycle_candidates AS MATERIALIZED (
			SELECT activity.id, activity.task_id, activity.user_id, activity.action,
			       activity.before_data, activity.after_data, task.title,
			       COALESCE(preference.notifications_enabled, TRUE) AS notifications_enabled
			FROM task_activities AS activity
			JOIN tasks AS task ON task.id = activity.task_id AND task.user_id = activity.user_id
			LEFT JOIN user_preferences AS preference ON preference.user_id = activity.user_id
			WHERE (
				activity.action IN ('status_changed', 'deleted', 'restored', 'comment_added')
				OR (
					activity.action = 'updated'
					AND activity.before_data->>'assignee' IS DISTINCT FROM activity.after_data->>'assignee'
				)
			)
			AND activity.created_at >= $1::TIMESTAMPTZ - INTERVAL '24 hours'
			AND NOT EXISTS (
				SELECT 1 FROM notifications AS existing
				WHERE existing.user_id = activity.user_id
				  AND existing.dedupe_key = 'task.activity.' || activity.id::TEXT
			)
			ORDER BY activity.created_at, activity.id
			LIMIT $2
		), lifecycle_inserted AS (
			INSERT INTO notifications (
				user_id, type, title, summary, body, source_type, source_id,
				action_label, action_url, dedupe_key, created_at
			)
			SELECT user_id,
			       'task',
			       CASE
			           WHEN action = 'status_changed' THEN '任务状态已变更：' || title
			           WHEN action = 'deleted' THEN '任务已删除：' || title
			           WHEN action = 'restored' THEN '任务已恢复：' || title
			           WHEN action = 'comment_added' AND after_data->>'parent_comment_id' IS NOT NULL THEN '任务评论有新回复：' || title
			           WHEN action = 'comment_added' THEN '任务有新评论：' || title
			           ELSE '任务负责人已变更：' || title
			       END,
			       CASE
			           WHEN action = 'status_changed' THEN '状态已从 ' || COALESCE(before_data->>'status', '未知') || ' 变更为 ' || COALESCE(after_data->>'status', '未知')
			           WHEN action = 'deleted' THEN '任务已进入可恢复状态'
			           WHEN action = 'restored' THEN '任务已重新回到任务中心'
			           WHEN action = 'comment_added' AND after_data->>'parent_comment_id' IS NOT NULL THEN '任务评论收到了新回复'
			           WHEN action = 'comment_added' THEN '任务收到了新评论'
			           ELSE '负责人已从“' || COALESCE(before_data->>'assignee', '未指定') || '”变更为“' || COALESCE(after_data->>'assignee', '未指定') || '”'
			       END,
			       '任务「' || title || '」有新的协作动态，请打开任务详情查看。',
			       'task', task_id, '查看任务', '/tasks?task_id=' || task_id::TEXT,
			       'task.activity.' || id::TEXT, $1
			FROM lifecycle_candidates
			WHERE notifications_enabled
			ON CONFLICT (user_id, dedupe_key) WHERE dedupe_key <> '' DO NOTHING
			RETURNING id
		), due_soon_candidates AS MATERIALIZED (
			SELECT task.id, task.user_id, task.title, task.due_at,
			       COALESCE(preference.notifications_enabled, TRUE) AS notifications_enabled,
			       COALESCE(NULLIF(preference.timezone, ''), 'Asia/Shanghai') AS timezone
			FROM tasks AS task
			LEFT JOIN user_preferences AS preference ON preference.user_id = task.user_id
			WHERE task.deleted_at IS NULL
			  AND task.status NOT IN ('completed', 'cancelled')
			  AND task.due_at > $1::TIMESTAMPTZ AND task.due_at <= $1::TIMESTAMPTZ + INTERVAL '48 hours'
			ORDER BY task.due_at, task.id
			LIMIT $2
		), due_soon_inserted AS (
			INSERT INTO notifications (
				user_id, type, title, summary, body, source_type, source_id,
				action_label, action_url, dedupe_key, created_at
			)
			SELECT user_id, 'task', '任务即将到期：' || title,
			       '任务将在 48 小时内到期',
			       '任务「' || title || '」截止时间为 ' || to_char(due_at AT TIME ZONE timezone, 'YYYY-MM-DD HH24:MI') || '。',
			       'task', id, '查看任务', '/tasks?task_id=' || id::TEXT,
			       'task.due_soon.' || id::TEXT || '.' || EXTRACT(EPOCH FROM due_at)::BIGINT::TEXT, $1
			FROM due_soon_candidates
			WHERE notifications_enabled
			ON CONFLICT (user_id, dedupe_key) WHERE dedupe_key <> '' DO NOTHING
			RETURNING id
		), overdue_candidates AS MATERIALIZED (
			SELECT task.id, task.user_id, task.title, task.due_at,
			       COALESCE(preference.notifications_enabled, TRUE) AS notifications_enabled
			FROM tasks AS task
			LEFT JOIN user_preferences AS preference ON preference.user_id = task.user_id
			WHERE task.deleted_at IS NULL
			  AND task.status NOT IN ('completed', 'cancelled')
			  AND task.due_at < $1::TIMESTAMPTZ - INTERVAL '48 hours'
			ORDER BY task.due_at, task.id
			LIMIT $2
		), overdue_inserted AS (
			INSERT INTO notifications (
				user_id, type, title, summary, body, source_type, source_id,
				action_label, action_url, dedupe_key, created_at
			)
			SELECT user_id, 'task', '任务已逾期：' || title,
			       '任务已超过截止时间 48 小时',
			       '任务「' || title || '」已逾期，请及时处理或调整截止时间。',
			       'task', id, '查看任务', '/tasks?task_id=' || id::TEXT,
			       'task.overdue.' || id::TEXT || '.' || EXTRACT(EPOCH FROM due_at)::BIGINT::TEXT, $1
			FROM overdue_candidates
			WHERE notifications_enabled
			ON CONFLICT (user_id, dedupe_key) WHERE dedupe_key <> '' DO NOTHING
			RETURNING id
		)
		SELECT (
			(SELECT COUNT(*) FROM lifecycle_inserted) +
			(SELECT COUNT(*) FROM due_soon_inserted) +
			(SELECT COUNT(*) FROM overdue_inserted)
		)::INT
	`, now, limit).Scan(&count)
	return count, err
}
