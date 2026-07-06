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
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
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
	err = r.db.QueryRow(ctx, `
		INSERT INTO competitor_scans (user_id, targets, focus, status, progress_percent, current_step, error_message, competitors, conclusions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
		RETURNING id
	`, scan.UserID, targets, scan.Focus, scan.Status, scan.ProgressPercent, scan.CurrentStep, scan.ErrorMessage, competitors, conclusions, scan.CreatedAt).Scan(&scan.ID)
	return scan, err
}

func (r *PostgresRepository) ListScans(ctx context.Context, userID int64, limit int) ([]Scan, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, targets, focus, status, progress_percent, current_step, error_message, competitors, conclusions, created_at, updated_at
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
		SELECT id, user_id, targets, focus, status, progress_percent, current_step, error_message, competitors, conclusions, created_at, updated_at
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
		RETURNING id, user_id, targets, focus, status, progress_percent, current_step, error_message, competitors, conclusions, created_at, updated_at
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
	tag, err := r.db.Exec(ctx, `
		UPDATE competitor_scans
		SET competitors = $2, conclusions = $3, updated_at = NOW()
		WHERE id = $1
	`, id, competitors, conclusions)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrScanNotFound
	}
	return nil
}

func (r *PostgresRepository) ListWatchlist(ctx context.Context, userID int64, limit int) ([]WatchItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT name, category, status, threat, last_seen_at, channels, signal
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
		var item WatchItem
		var channels []byte
		if err := rows.Scan(&item.Name, &item.Category, &item.Status, &item.Threat, &item.LastSeenAt, &channels, &item.Signal); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(channels, &item.Channels); err != nil {
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

func scanScan(scanner scanScanner) (Scan, error) {
	var scan Scan
	var targets []byte
	var competitors []byte
	var conclusions []byte
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
	return scan, nil
}
