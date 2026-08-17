package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zzm/opcv2/internal/ai"
)

type pendingTranslation struct {
	dataset    string
	recordKey  string
	sourceHash string
	payload    json.RawMessage
}

type translationItem struct {
	Dataset   string         `json:"dataset"`
	RecordKey string         `json:"record_key"`
	Payload   map[string]any `json:"translation"`
}

type translationResponse struct {
	Items []translationItem `json:"items"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	databaseURL := flag.String("database-url", envOrDefault("OPCV2_DATABASE_URL", "postgres://opcv2:opcv2@127.0.0.1:5432/opcv2?sslmode=disable"), "OPC PostgreSQL URL")
	apiKey := flag.String("api-key", os.Getenv("OPCV2_DEEPSEEK_API_KEY"), "DeepSeek API key (prefer OPCV2_DEEPSEEK_API_KEY)")
	baseURL := flag.String("base-url", envOrDefault("OPCV2_DEEPSEEK_BASE_URL", "https://api.deepseek.com"), "OpenAI-compatible API base URL")
	model := flag.String("model", envOrDefault("OPCV2_DEEPSEEK_MODEL", "deepseek-chat"), "translation model")
	dataset := flag.String("dataset", "", "only translate one dataset: startup, rebuild_plan, idea")
	limit := flag.Int("limit", 100, "maximum records to translate; explicitly set 0 to translate all pending records")
	batchSize := flag.Int("batch-size", 10, "records per model request")
	retryFailed := flag.Bool("retry-failed", false, "retry records previously marked failed")
	shardCount := flag.Int("shard-count", 1, "number of numeric record-key shards")
	shardIndex := flag.Int("shard-index", 0, "zero-based shard index")
	flag.Parse()
	if strings.TrimSpace(*apiKey) == "" {
		logger.Error("missing DeepSeek API key", "env", "OPCV2_DEEPSEEK_API_KEY")
		os.Exit(2)
	}
	if *batchSize < 1 || *batchSize > 25 {
		logger.Error("batch-size must be between 1 and 25")
		os.Exit(2)
	}
	if *limit < 0 {
		logger.Error("limit must be 0 or greater")
		os.Exit(2)
	}
	if *shardCount < 1 || *shardIndex < 0 || *shardIndex >= *shardCount {
		logger.Error("invalid shard configuration")
		os.Exit(2)
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, *databaseURL)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		logger.Error("ping postgres", "error", err)
		os.Exit(1)
	}
	provider := ai.NewOpenAICompatibleProvider(ai.OpenAICompatibleConfig{
		BaseURL: *baseURL,
		APIKey:  *apiKey,
		Model:   *model,
		Timeout: 90 * time.Second,
	})
	translated, failed, err := translatePending(ctx, db, provider, strings.TrimSpace(*dataset), *limit, *batchSize, *model, *retryFailed, *shardCount, *shardIndex, logger)
	if err != nil {
		logger.Error("translation failed", "error", err)
		os.Exit(1)
	}
	logger.Info("translation completed", "translated", translated, "failed", failed)
}

func translatePending(ctx context.Context, db *pgxpool.Pool, provider ai.Provider, dataset string, limit, batchSize int, model string, retryFailed bool, shardCount, shardIndex int, logger *slog.Logger) (int, int, error) {
	if retryFailed {
		if _, err := db.Exec(ctx, `
			UPDATE lootdrop_translations
			SET status = 'pending', error_message = '', updated_at = NOW()
			WHERE language = 'zh-CN' AND status = 'failed' AND ($1 = '' OR dataset = $1)
		`, dataset); err != nil {
			return 0, 0, err
		}
	}
	totalTranslated, totalFailed := 0, 0
	remaining := limit
	for {
		fetchLimit := batchSize
		if remaining > 0 && fetchLimit > remaining {
			fetchLimit = remaining
		}
		items, err := loadPending(ctx, db, dataset, fetchLimit, retryFailed, shardCount, shardIndex)
		if err != nil {
			return totalTranslated, totalFailed, err
		}
		if len(items) == 0 {
			return totalTranslated, totalFailed, nil
		}
		translated, err := translateBatch(ctx, db, provider, items, model)
		if err != nil {
			for _, item := range items {
				if markTranslationFailed(ctx, db, item, model, err) != nil {
					logger.Warn("mark translation failed", "dataset", item.dataset, "record_key", item.recordKey)
				}
			}
			totalFailed += len(items)
			logger.Warn("translation batch failed", "count", len(items), "error", err)
			if isProviderError(err) {
				return totalTranslated, totalFailed, err
			}
		} else {
			totalTranslated += translated
			if translated == 0 {
				for _, item := range items {
					_ = markTranslationFailed(ctx, db, item, model, errors.New("model returned no matching translations"))
				}
				totalFailed += len(items)
			}
		}
		if remaining > 0 {
			remaining -= len(items)
			if remaining <= 0 {
				return totalTranslated, totalFailed, nil
			}
		}
	}
}

func isProviderError(err error) bool {
	return errors.Is(err, ai.ErrProviderTimeout) ||
		errors.Is(err, ai.ErrProviderRateLimited) ||
		errors.Is(err, ai.ErrProviderAuthentication) ||
		errors.Is(err, ai.ErrProviderPermission) ||
		errors.Is(err, ai.ErrProviderModelNotFound) ||
		errors.Is(err, ai.ErrProviderBadRequest) ||
		errors.Is(err, ai.ErrProviderUnavailable)
}

func loadPending(ctx context.Context, db *pgxpool.Pool, dataset string, limit int, retryFailed bool, shardCount, shardIndex int) ([]pendingTranslation, error) {
	query := `
		SELECT t.dataset, t.record_key, t.source_hash,
			CASE t.dataset
				WHEN 'startup' THEN s.source_payload
				WHEN 'rebuild_plan' THEN r.source_payload
				WHEN 'idea' THEN i.source_payload
			END AS payload
		FROM lootdrop_translations t
		LEFT JOIN lootdrop_startups s ON t.dataset = 'startup' AND s.source_id::TEXT = t.record_key
		LEFT JOIN lootdrop_rebuild_plans r ON t.dataset = 'rebuild_plan' AND r.source_id::TEXT = t.record_key
		LEFT JOIN lootdrop_ideas i ON t.dataset = 'idea' AND i.source_id::TEXT = t.record_key
		WHERE t.language = 'zh-CN' AND t.status = ANY($2::text[])
		  AND ($1 = '' OR t.dataset = $1)
		  AND ($4 = 1 OR (t.record_key ~ '^[0-9]+$' AND MOD(t.record_key::BIGINT, $4) = $5))
		ORDER BY t.dataset, t.record_key
		LIMIT $3`
	statuses := []string{"pending"}
	if retryFailed {
		statuses = append(statuses, "failed")
	}
	rows, err := db.Query(ctx, query, dataset, statuses, limit, shardCount, shardIndex)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]pendingTranslation, 0, limit)
	for rows.Next() {
		var item pendingTranslation
		if err := rows.Scan(&item.dataset, &item.recordKey, &item.sourceHash, &item.payload); err != nil {
			return nil, err
		}
		if len(item.payload) == 0 || string(item.payload) == "null" {
			item.payload = json.RawMessage(`{}`)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func translateBatch(ctx context.Context, db *pgxpool.Pool, provider ai.Provider, items []pendingTranslation, model string) (int, error) {
	input := make([]map[string]any, 0, len(items))
	for _, item := range items {
		var rawPayload map[string]any
		if err := json.Unmarshal(item.payload, &rawPayload); err != nil {
			rawPayload = map[string]any{}
		}
		input = append(input, map[string]any{"dataset": item.dataset, "record_key": item.recordKey, "payload": translatablePayload(item.dataset, rawPayload)})
	}
	requestPayload, err := json.Marshal(input)
	if err != nil {
		return 0, err
	}
	result, err := provider.Generate(ctx, ai.ProviderRequest{
		Feature:      "lootdrop.translation",
		Model:        model,
		SchemaName:   "lootdrop_translation_batch",
		SystemPrompt: "你是严格的数据翻译器。把给定的 Loot Drop 英文创业数据翻译成简体中文。保留 dataset 和 record_key，返回 JSON 对象 {\"items\":[{\"dataset\":\"...\",\"record_key\":\"...\",\"translation\":{...}}]}。translation 只放翻译后的字段，字段名使用 name_zh、title_zh、description_zh、sector_zh、country_zh、*_zh；数组逐项翻译，对象保留结构。数字、年份、URL、ID、哈希和代码不要翻译。不要添加原数据没有的事实。只输出 JSON。",
		UserPrompt:   string(requestPayload),
	})
	if err != nil {
		return 0, err
	}
	var response translationResponse
	if err := json.Unmarshal(result.Content, &response); err != nil {
		return 0, fmt.Errorf("decode model JSON: %w", err)
	}
	byKey := make(map[string]translationItem, len(response.Items))
	for _, item := range response.Items {
		if item.Dataset != "" && item.RecordKey != "" && item.Payload != nil {
			byKey[item.Dataset+"\x00"+item.RecordKey] = item
		}
	}
	completed := 0
	projectOpportunityIDs := make([]int64, 0, len(items))
	for _, source := range items {
		item, ok := byKey[source.dataset+"\x00"+source.recordKey]
		if !ok {
			continue
		}
		payload, err := json.Marshal(item.Payload)
		if err != nil {
			continue
		}
		commandTag, err := db.Exec(ctx, `
			UPDATE lootdrop_translations
			SET status = 'completed', translated_payload = $1, model = $2, translated_at = NOW(),
				error_message = '', updated_at = NOW()
			WHERE dataset = $3 AND record_key = $4 AND language = 'zh-CN'
			  AND status = 'pending' AND source_hash = $5
		`, payload, model, source.dataset, source.recordKey, source.sourceHash)
		if err != nil {
			return completed, err
		}
		if commandTag.RowsAffected() == 1 {
			completed++
			if source.dataset == "startup" {
				_ = updateProjectedCase(ctx, db, source.recordKey, item.Payload)
			}
			if source.dataset == "rebuild_plan" {
				if sourceID, parseErr := strconv.ParseInt(source.recordKey, 10, 64); parseErr == nil && sourceID > 0 {
					projectOpportunityIDs = append(projectOpportunityIDs, sourceID)
				}
			}
		}
	}
	if len(projectOpportunityIDs) > 0 {
		if _, err := db.Exec(ctx, `SELECT refresh_lootdrop_project_opportunities($1)`, projectOpportunityIDs); err != nil {
			return completed, fmt.Errorf("refresh translated project opportunities: %w", err)
		}
		if _, err := db.Exec(ctx, `SELECT normalize_lootdrop_project_opportunities($1)`, projectOpportunityIDs); err != nil {
			return completed, fmt.Errorf("normalize translated project opportunities: %w", err)
		}
		if _, err := db.Exec(ctx, `SELECT index_lootdrop_project_search_content($1)`, projectOpportunityIDs); err != nil {
			return completed, fmt.Errorf("index translated project opportunities: %w", err)
		}
	}
	return completed, nil
}

func translatablePayload(dataset string, payload map[string]any) map[string]any {
	fields := map[string][]string{
		"startup":      {"name", "description", "sector", "country", "cause_of_death", "primary_cause_of_death", "market_analysis", "product_type", "condensed_value_prop", "condensed_cause_of_death"},
		"rebuild_plan": {"name", "sector", "product_type", "country", "primary_cause_of_death", "market_potential", "pivot_idea", "the_loot"},
		"idea":         {"title", "description", "model", "effort", "speed", "category", "tags", "persona"},
	}
	selected := make(map[string]any)
	for _, field := range fields[dataset] {
		if value, ok := payload[field]; ok {
			selected[field] = value
		}
	}
	return selected
}

func updateProjectedCase(ctx context.Context, db *pgxpool.Pool, recordKey string, payload map[string]any) error {
	name := firstPayloadString(payload, "name_zh", "title_zh")
	description := firstPayloadString(payload, "description_zh", "summary_zh")
	cause := firstPayloadString(payload, "cause_of_death_zh", "primary_cause_of_death_zh")
	if name == "" && description == "" && cause == "" {
		return nil
	}
	_, err := db.Exec(ctx, `
		UPDATE project_cases
		SET title = COALESCE(NULLIF($1, ''), title),
			summary = COALESCE(NULLIF($2, ''), summary),
			result_summary = COALESCE(NULLIF($3, ''), result_summary),
			outcome = COALESCE(NULLIF($3, ''), outcome),
			content_md = CONCAT_WS(E'\n\n', NULLIF($1, ''), NULLIF($2, ''), NULLIF($3, '')),
			updated_at = NOW()
		WHERE slug = 'lootdrop-startup-' || $4
	`, name, description, cause, recordKey)
	if err != nil {
		return err
	}
	claimTranslations := map[string]string{
		"company_name":           firstPayloadString(payload, "name_zh", "title_zh"),
		"description":            firstPayloadString(payload, "description_zh", "summary_zh"),
		"sector":                 firstPayloadString(payload, "sector_zh"),
		"country":                firstPayloadString(payload, "country_zh"),
		"primary_cause_of_death": firstPayloadString(payload, "primary_cause_of_death_zh"),
		"cause_of_death":         firstPayloadString(payload, "cause_of_death_zh"),
		"market_analysis":        firstPayloadString(payload, "market_analysis_zh"),
	}
	for field, value := range claimTranslations {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if _, err := db.Exec(ctx, `
			UPDATE project_case_claims c
			SET value_text = $1
			FROM project_cases p
			WHERE c.case_id = p.id AND p.slug = 'lootdrop-startup-' || $2 AND c.field_name = $3
		`, value, recordKey, field); err != nil {
			return err
		}
	}
	return nil
}

func markTranslationFailed(ctx context.Context, db *pgxpool.Pool, item pendingTranslation, model string, cause error) error {
	_, err := db.Exec(ctx, `
		UPDATE lootdrop_translations
		SET status = 'failed', model = $1, error_message = LEFT($2, 1000), updated_at = NOW()
		WHERE dataset = $3 AND record_key = $4 AND language = 'zh-CN' AND status = 'pending' AND source_hash = $5
	`, model, cause.Error(), item.dataset, item.recordKey, item.sourceHash)
	return err
}

func firstPayloadString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
