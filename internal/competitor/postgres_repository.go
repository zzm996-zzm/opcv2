package competitor

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type postgresDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	var isAdmin bool
	err := r.db.QueryRow(ctx, `SELECT role = 'admin' FROM users WHERE id = $1 AND status = 'active'`, userID).Scan(&isAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return isAdmin, err
}

func (r *PostgresRepository) ListScriptAccounts(ctx context.Context, platform string, limit int) ([]ScriptAccount, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, platform, account_label, credential_ref <> '', status, cooldown_until,
		       failure_count, max_runs_per_hour, last_used_at, created_at, updated_at
		FROM competitor_script_accounts
		WHERE ($1 = '' OR platform = $1)
		ORDER BY platform, account_label
		LIMIT $2
	`, platform, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ScriptAccount, 0)
	for rows.Next() {
		item, err := scanScriptAccount(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) UpsertScriptAccount(ctx context.Context, account ScriptAccount) (ScriptAccount, error) {
	if account.ID == 0 {
		return scanScriptAccount(r.db.QueryRow(ctx, `
			INSERT INTO competitor_script_accounts (platform, account_label, credential_ref, status, max_runs_per_hour, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $6)
			RETURNING id, platform, account_label, credential_ref <> '', status, cooldown_until,
			          failure_count, max_runs_per_hour, last_used_at, created_at, updated_at
		`, account.Platform, account.AccountLabel, account.CredentialRef, account.Status, account.MaxRunsPerHour, account.CreatedAt))
	}
	item, err := scanScriptAccount(r.db.QueryRow(ctx, `
		UPDATE competitor_script_accounts
		SET platform = $2, account_label = $3,
		    credential_ref = COALESCE(NULLIF($4, ''), credential_ref),
		    status = $5, max_runs_per_hour = $6,
		    cooldown_until = CASE WHEN $5 = 'available' THEN NULL ELSE cooldown_until END,
		    updated_at = $7
		WHERE id = $1
		RETURNING id, platform, account_label, credential_ref <> '', status, cooldown_until,
		          failure_count, max_runs_per_hour, last_used_at, created_at, updated_at
	`, account.ID, account.Platform, account.AccountLabel, account.CredentialRef, account.Status, account.MaxRunsPerHour, account.UpdatedAt))
	if errors.Is(err, pgx.ErrNoRows) {
		return ScriptAccount{}, ErrScriptAccountNotFound
	}
	return item, err
}

func (r *PostgresRepository) CreateScan(ctx context.Context, scan Scan) (Scan, error) {
	targets, err := json.Marshal(scan.Targets)
	if err != nil {
		return Scan{}, err
	}
	competitors, err := json.Marshal(scan.Competitors)
	if err != nil {
		return Scan{}, err
	}
	conclusions, err := json.Marshal(scan.Conclusions)
	if err != nil {
		return Scan{}, err
	}
	evidenceSources, err := json.Marshal(scan.EvidenceSources)
	if err != nil {
		return Scan{}, err
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO competitor_scans (user_id, targets, focus, status, progress_percent, current_step, error_message, competitors, conclusions, evidence_sources, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
		RETURNING id
	`, scan.UserID, targets, scan.Focus, scan.Status, scan.ProgressPercent, scan.CurrentStep, scan.ErrorMessage, competitors, conclusions, evidenceSources, scan.CreatedAt).Scan(&scan.ID)
	return scan, err
}

func (r *PostgresRepository) ListScans(ctx context.Context, userID int64, limit int) ([]Scan, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, targets, focus, status, progress_percent, current_step, error_message, competitors, conclusions, evidence_sources, created_at, updated_at
		FROM competitor_scans
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scans []Scan
	for rows.Next() {
		scan, err := scanScan(rows)
		if err != nil {
			return nil, err
		}
		scans = append(scans, scan)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return scans, nil
}

func (r *PostgresRepository) GetScan(ctx context.Context, userID, id int64) (Scan, error) {
	scan, err := scanScan(r.db.QueryRow(ctx, `
		SELECT id, user_id, targets, focus, status, progress_percent, current_step, error_message, competitors, conclusions, evidence_sources, created_at, updated_at
		FROM competitor_scans
		WHERE user_id = $1 AND id = $2
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Scan{}, ErrScanNotFound
	}
	return scan, err
}

func (r *PostgresRepository) UpdateScanStatus(ctx context.Context, id int64, status string, progressPercent int, currentStep string, errorMessage string) (Scan, error) {
	scan, err := scanScan(r.db.QueryRow(ctx, `
		UPDATE competitor_scans
		SET status = $2, progress_percent = $3, current_step = $4, error_message = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, targets, focus, status, progress_percent, current_step, error_message, competitors, conclusions, evidence_sources, created_at, updated_at
	`, id, status, progressPercent, currentStep, errorMessage))
	if errors.Is(err, pgx.ErrNoRows) {
		return Scan{}, ErrScanNotFound
	}
	return scan, err
}

func (r *PostgresRepository) StoreScanResults(ctx context.Context, id int64, result ScanResult) error {
	competitors, err := json.Marshal(result.Competitors)
	if err != nil {
		return err
	}
	conclusions, err := json.Marshal(result.Conclusions)
	if err != nil {
		return err
	}
	evidenceSources, err := json.Marshal(result.EvidenceSources)
	if err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `
		UPDATE competitor_scans
		SET competitors = $2, conclusions = $3, evidence_sources = $4, updated_at = NOW()
		WHERE id = $1
	`, id, competitors, conclusions, evidenceSources)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrScanNotFound
	}
	if _, err := tx.Exec(ctx, `DELETE FROM competitor_raw_snapshots WHERE scan_id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM competitor_evidence_sources WHERE scan_id = $1`, id); err != nil {
		return err
	}
	for _, snapshot := range result.RawSnapshots {
		payload := snapshot.Payload
		if len(payload) == 0 {
			payload = json.RawMessage(`{}`)
		}
		if !json.Valid(payload) {
			return errors.New("invalid competitor raw snapshot payload")
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO competitor_raw_snapshots (scan_id, platform, raw_payload, object_key, captured_at)
			VALUES ($1, $2, $3, $4, $5)
		`, id, snapshot.Platform, []byte(payload), snapshot.ObjectKey, snapshot.CapturedAt); err != nil {
			return err
		}
	}
	for _, source := range result.EvidenceSources {
		if _, err := tx.Exec(ctx, `
			INSERT INTO competitor_evidence_sources (scan_id, source_type, platform, title, source_url, summary, screenshot_object_key, captured_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, id, source.SourceType, source.Platform, source.Title, source.URL, source.Summary, source.ScreenshotObjectKey, source.CapturedAt); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepository) CreateWatchItem(ctx context.Context, item WatchItem) (WatchItem, error) {
	channels, err := json.Marshal(item.Channels)
	if err != nil {
		return WatchItem{}, err
	}
	return scanWatchItem(r.db.QueryRow(ctx, `
		INSERT INTO competitor_watchlist (user_id, name, category, status, threat, last_seen_at, channels, signal, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING id, name, category, status, threat, last_seen_at, channels, signal
	`, item.UserID, item.Name, item.Category, item.Status, item.Threat, item.LastSeenAt, channels, item.Signal, item.LastSeenAt))
}

func (r *PostgresRepository) GetWatchItem(ctx context.Context, userID, id int64) (WatchItem, error) {
	item, err := scanWatchItem(r.db.QueryRow(ctx, `
		SELECT id, name, category, status, threat, last_seen_at, channels, signal
		FROM competitor_watchlist
		WHERE user_id = $1 AND id = $2
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return WatchItem{}, ErrWatchItemNotFound
	}
	return item, err
}

func (r *PostgresRepository) DeleteWatchItem(ctx context.Context, userID, id int64) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM competitor_watchlist
		WHERE user_id = $1 AND id = $2
	`, userID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrWatchItemNotFound
	}
	return nil
}

func (r *PostgresRepository) ListWatchlist(ctx context.Context, userID int64, limit int) ([]WatchItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, category, status, threat, last_seen_at, channels, signal
		FROM competitor_watchlist
		WHERE user_id = $1
		ORDER BY last_seen_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []WatchItem
	for rows.Next() {
		item, err := scanWatchItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PostgresRepository) ListEvents(ctx context.Context, userID int64, limit int) ([]Event, error) {
	rows, err := r.db.Query(ctx, `
		SELECT occurred_at, company, title, detail, level
		FROM competitor_events
		WHERE user_id = $1
		ORDER BY occurred_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.OccurredAt, &event.Company, &event.Title, &event.Detail, &event.Level); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

type scanScanner interface {
	Scan(dest ...any) error
}

func scanWatchItem(scanner scanScanner) (WatchItem, error) {
	var item WatchItem
	var channels []byte
	if err := scanner.Scan(&item.ID, &item.Name, &item.Category, &item.Status, &item.Threat, &item.LastSeenAt, &channels, &item.Signal); err != nil {
		return WatchItem{}, err
	}
	if err := json.Unmarshal(channels, &item.Channels); err != nil {
		return WatchItem{}, err
	}
	return item, nil
}

func scanScan(scanner scanScanner) (Scan, error) {
	var scan Scan
	var targets []byte
	var competitors []byte
	var conclusions []byte
	var evidenceSources []byte
	if err := scanner.Scan(
		&scan.ID,
		&scan.UserID,
		&targets,
		&scan.Focus,
		&scan.Status,
		&scan.ProgressPercent,
		&scan.CurrentStep,
		&scan.ErrorMessage,
		&competitors,
		&conclusions,
		&evidenceSources,
		&scan.CreatedAt,
		&scan.UpdatedAt,
	); err != nil {
		return Scan{}, err
	}
	if err := json.Unmarshal(targets, &scan.Targets); err != nil {
		return Scan{}, err
	}
	if err := json.Unmarshal(competitors, &scan.Competitors); err != nil {
		return Scan{}, err
	}
	if err := json.Unmarshal(conclusions, &scan.Conclusions); err != nil {
		return Scan{}, err
	}
	if err := json.Unmarshal(evidenceSources, &scan.EvidenceSources); err != nil {
		return Scan{}, err
	}
	return scan, nil
}

func scanScriptAccount(scanner scanScanner) (ScriptAccount, error) {
	var item ScriptAccount
	if err := scanner.Scan(
		&item.ID, &item.Platform, &item.AccountLabel, &item.HasCredential, &item.Status,
		&item.CooldownUntil, &item.FailureCount, &item.MaxRunsPerHour, &item.LastUsedAt,
		&item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return ScriptAccount{}, err
	}
	return item, nil
}
