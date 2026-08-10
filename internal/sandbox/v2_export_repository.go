package sandbox

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func (r *PostgresRepository) CreateV2Export(ctx context.Context, export V2Export, userID int64, payload []byte) (V2Export, error) {
	var data []byte
	if len(payload) == 0 {
		data = []byte(`{}`)
	} else {
		data = payload
	}
	err := r.db.QueryRow(ctx, `
		INSERT INTO sandbox_exports (run_id, user_id, format, payload, expires_at, created_at)
		SELECT $1, $2, $3, $4, $5, COALESCE($6, NOW())
		WHERE EXISTS(SELECT 1 FROM sandbox_sessions WHERE id = $1 AND user_id = $2 AND sandbox_version = 2)
		RETURNING id, created_at
	`, export.RunID, userID, export.Format, data, export.ExpiresAt, nullableTime(export.CreatedAt)).Scan(&export.ID, &export.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return V2Export{}, ErrV2RunNotFound
	}
	if err != nil {
		return V2Export{}, err
	}
	export.DownloadURL = "/api/v1/sandbox-runs/" + formatInt(export.RunID) + "/exports/" + formatInt(export.ID) + "/download"
	return export, nil
}

func (r *PostgresRepository) GetV2Export(ctx context.Context, userID, exportID int64) (V2Export, []byte, error) {
	var export V2Export
	var payload []byte
	err := r.db.QueryRow(ctx, `
		SELECT id, run_id, format, payload, expires_at, created_at
		FROM sandbox_exports WHERE id = $1 AND user_id = $2
	`, exportID, userID).Scan(&export.ID, &export.RunID, &export.Format, &payload, &export.ExpiresAt, &export.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return V2Export{}, nil, ErrV2RunNotFound
	}
	if err != nil {
		return V2Export{}, nil, err
	}
	export.DownloadURL = "/api/v1/sandbox-runs/" + formatInt(export.RunID) + "/exports/" + formatInt(export.ID) + "/download"
	return export, payload, nil
}

func formatInt(value int64) string {
	return strconv.FormatInt(value, 10)
}
