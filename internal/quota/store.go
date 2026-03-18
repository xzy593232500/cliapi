package quota

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Account struct {
	APIKey              string    `json:"api_key"`
	TokenBalance        int64     `json:"token_balance"`
	TokenUsedTotal      int64     `json:"token_used_total"`
	USDUsedTotal        float64   `json:"usd_used_total"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	BalanceUSDEquivalent float64  `json:"balance_usd_equivalent"`
}

type Transaction struct {
	ID                 int64     `json:"id"`
	APIKey             string    `json:"api_key"`
	Type               string    `json:"type"`
	TokenDelta         int64     `json:"token_delta"`
	USDDelta           float64   `json:"usd_delta"`
	BalanceAfterTokens int64     `json:"balance_after_tokens"`
	Model              string    `json:"model"`
	RequestID          string    `json:"request_id"`
	RedemptionCode     string    `json:"redemption_code"`
	Note               string    `json:"note"`
	CreatedAt          time.Time `json:"created_at"`
}

type RedeemCode struct {
	Code        string     `json:"code"`
	USDAmount   float64    `json:"usd_amount"`
	TokenAmount int64      `json:"token_amount"`
	Status      string     `json:"status"`
	UsedByAPIKey string    `json:"used_by_api_key"`
	UsedAt      *time.Time `json:"used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Store struct {
	db             *sql.DB
	tokensPerUSD   int64
	mu             sync.RWMutex
}

func NewStore(ctx context.Context, dsn string, tokensPerUSD int64) (*Store, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("quota: dsn is required")
	}
	if tokensPerUSD <= 0 {
		tokensPerUSD = 500000
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	s := &Store{db: db, tokensPerUSD: tokensPerUSD}
	if err = s.EnsureSchema(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) TokensPerUSD() int64 {
	if s == nil || s.tokensPerUSD <= 0 {
		return 500000
	}
	return s.tokensPerUSD
}

func (s *Store) TokensToUSD(tokens int64) float64 {
	return float64(tokens) / float64(s.TokensPerUSD())
}

func (s *Store) USDToTokens(usd float64) int64 {
	return int64(usd * float64(s.TokensPerUSD()))
}

func (s *Store) EnsureSchema(ctx context.Context) error {
	stmts := []string{
		`create table if not exists quota_accounts (
			api_key text primary key,
			token_balance bigint not null default 0,
			token_used_total bigint not null default 0,
			usd_used_total double precision not null default 0,
			status text not null default 'active',
			created_at timestamptz not null default now(),
			updated_at timestamptz not null default now()
		)`,
		`create table if not exists quota_transactions (
			id bigserial primary key,
			api_key text not null,
			type text not null,
			token_delta bigint not null default 0,
			usd_delta double precision not null default 0,
			balance_after_tokens bigint not null,
			model text,
			request_id text,
			redemption_code text,
			note text,
			created_at timestamptz not null default now()
		)`,
		`create index if not exists idx_quota_transactions_api_key_created_at on quota_transactions(api_key, created_at desc)`,
		`create table if not exists redemption_codes (
			code text primary key,
			usd_amount double precision not null,
			token_amount bigint not null,
			status text not null default 'unused',
			used_by_api_key text,
			used_at timestamptz,
			expires_at timestamptz,
			created_at timestamptz not null default now()
		)`,
		`create index if not exists idx_redemption_codes_status_created_at on redemption_codes(status, created_at desc)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) EnsureAccount(ctx context.Context, apiKey string) error {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return fmt.Errorf("quota: api key required")
	}
	_, err := s.db.ExecContext(ctx, `insert into quota_accounts (api_key) values ($1) on conflict (api_key) do nothing`, apiKey)
	return err
}

func (s *Store) GetAccount(ctx context.Context, apiKey string) (*Account, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("quota: api key required")
	}
	if err := s.EnsureAccount(ctx, apiKey); err != nil {
		return nil, err
	}
	row := s.db.QueryRowContext(ctx, `select api_key, token_balance, token_used_total, usd_used_total, status, created_at, updated_at from quota_accounts where api_key = $1`, apiKey)
	var a Account
	if err := row.Scan(&a.APIKey, &a.TokenBalance, &a.TokenUsedTotal, &a.USDUsedTotal, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, err
	}
	a.BalanceUSDEquivalent = s.TokensToUSD(a.TokenBalance)
	return &a, nil
}

func (s *Store) ListAccounts(ctx context.Context, limit int) ([]Account, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	rows, err := s.db.QueryContext(ctx, `select api_key, token_balance, token_used_total, usd_used_total, status, created_at, updated_at from quota_accounts order by updated_at desc limit $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Account, 0)
	for rows.Next() {
		var a Account
		if err = rows.Scan(&a.APIKey, &a.TokenBalance, &a.TokenUsedTotal, &a.USDUsedTotal, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.BalanceUSDEquivalent = s.TokensToUSD(a.TokenBalance)
		items = append(items, a)
	}
	return items, rows.Err()
}

func (s *Store) ListTransactions(ctx context.Context, apiKey string, limit int) ([]Transaction, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := `select id, api_key, type, token_delta, usd_delta, balance_after_tokens, model, request_id, redemption_code, note, created_at from quota_transactions`
	args := []any{}
	if strings.TrimSpace(apiKey) != "" {
		query += ` where api_key = $1`
		args = append(args, strings.TrimSpace(apiKey))
		query += ` order by created_at desc limit $2`
		args = append(args, limit)
	} else {
		query += ` order by created_at desc limit $1`
		args = append(args, limit)
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Transaction, 0)
	for rows.Next() {
		var t Transaction
		if err = rows.Scan(&t.ID, &t.APIKey, &t.Type, &t.TokenDelta, &t.USDDelta, &t.BalanceAfterTokens, &t.Model, &t.RequestID, &t.RedemptionCode, &t.Note, &t.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func (s *Store) CreateRedeemCodes(ctx context.Context, codes []RedeemCode) error {
	if len(codes) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, code := range codes {
		if strings.TrimSpace(code.Code) == "" {
			return fmt.Errorf("quota: code required")
		}
		if _, err = tx.ExecContext(ctx, `insert into redemption_codes (code, usd_amount, token_amount, status, expires_at) values ($1,$2,$3,'unused',$4)`, strings.TrimSpace(code.Code), code.USDAmount, code.TokenAmount, code.ExpiresAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListRedeemCodes(ctx context.Context, limit int) ([]RedeemCode, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	rows, err := s.db.QueryContext(ctx, `select code, usd_amount, token_amount, status, used_by_api_key, used_at, expires_at, created_at from redemption_codes order by created_at desc limit $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]RedeemCode, 0)
	for rows.Next() {
		var item RedeemCode
		if err = rows.Scan(&item.Code, &item.USDAmount, &item.TokenAmount, &item.Status, &item.UsedByAPIKey, &item.UsedAt, &item.ExpiresAt, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) RedeemCode(ctx context.Context, apiKey, code string) (*Account, error) {
	apiKey = strings.TrimSpace(apiKey)
	code = strings.TrimSpace(code)
	if apiKey == "" || code == "" {
		return nil, fmt.Errorf("quota: api key and code required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `insert into quota_accounts (api_key) values ($1) on conflict (api_key) do nothing`, apiKey); err != nil {
		return nil, err
	}
	var usdAmount float64
	var tokenAmount int64
	var status string
	var expiresAt sql.NullTime
	if err = tx.QueryRowContext(ctx, `select usd_amount, token_amount, status, expires_at from redemption_codes where code = $1 for update`, code).Scan(&usdAmount, &tokenAmount, &status, &expiresAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("quota: invalid code")
		}
		return nil, err
	}
	if status != "unused" {
		return nil, fmt.Errorf("quota: code not available")
	}
	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("quota: code expired")
	}
	if _, err = tx.ExecContext(ctx, `update redemption_codes set status = 'used', used_by_api_key = $1, used_at = now() where code = $2`, apiKey, code); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `update quota_accounts set token_balance = token_balance + $1, updated_at = now() where api_key = $2`, tokenAmount, apiKey); err != nil {
		return nil, err
	}
	var balanceAfter int64
	if err = tx.QueryRowContext(ctx, `select token_balance from quota_accounts where api_key = $1`, apiKey).Scan(&balanceAfter); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `insert into quota_transactions (api_key, type, token_delta, usd_delta, balance_after_tokens, redemption_code, note) values ($1,'redeem',$2,$3,$4,$5,$6)`, apiKey, tokenAmount, usdAmount, balanceAfter, code, "redeem code topup"); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetAccount(ctx, apiKey)
}

func (s *Store) AdjustBalance(ctx context.Context, apiKey string, tokenDelta int64, usdDelta float64, note string) (*Account, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("quota: api key required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `insert into quota_accounts (api_key) values ($1) on conflict (api_key) do nothing`, apiKey); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `update quota_accounts set token_balance = token_balance + $1, updated_at = now() where api_key = $2`, tokenDelta, apiKey); err != nil {
		return nil, err
	}
	var balanceAfter int64
	if err = tx.QueryRowContext(ctx, `select token_balance from quota_accounts where api_key = $1`, apiKey).Scan(&balanceAfter); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `insert into quota_transactions (api_key, type, token_delta, usd_delta, balance_after_tokens, note) values ($1,'admin_adjust',$2,$3,$4,$5)`, apiKey, tokenDelta, usdDelta, balanceAfter, strings.TrimSpace(note)); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetAccount(ctx, apiKey)
}

func (s *Store) HasEnoughBalance(ctx context.Context, apiKey string, requiredTokens int64) (bool, *Account, error) {
	account, err := s.GetAccount(ctx, apiKey)
	if err != nil {
		return false, nil, err
	}
	if account.Status != "active" {
		return false, account, nil
	}
	return account.TokenBalance >= requiredTokens, account, nil
}

func (s *Store) ConsumeTokens(ctx context.Context, apiKey string, tokens int64, model, requestID string) (*Account, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, nil
	}
	if tokens < 0 {
		tokens = 0
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `insert into quota_accounts (api_key) values ($1) on conflict (api_key) do nothing`, apiKey); err != nil {
		return nil, err
	}
	var balance int64
	var status string
	if err = tx.QueryRowContext(ctx, `select token_balance, status from quota_accounts where api_key = $1 for update`, apiKey).Scan(&balance, &status); err != nil {
		return nil, err
	}
	if status != "active" {
		return nil, fmt.Errorf("quota: account disabled")
	}
	if balance < tokens {
		return nil, fmt.Errorf("quota: insufficient token balance")
	}
	usdCost := s.TokensToUSD(tokens)
	if _, err = tx.ExecContext(ctx, `update quota_accounts set token_balance = token_balance - $1, token_used_total = token_used_total + $1, usd_used_total = usd_used_total + $2, updated_at = now() where api_key = $3`, tokens, usdCost, apiKey); err != nil {
		return nil, err
	}
	balanceAfter := balance - tokens
	if _, err = tx.ExecContext(ctx, `insert into quota_transactions (api_key, type, token_delta, usd_delta, balance_after_tokens, model, request_id, note) values ($1,'consume',$2,$3,$4,$5,$6,$7)`, apiKey, -tokens, -usdCost, balanceAfter, strings.TrimSpace(model), strings.TrimSpace(requestID), "usage consume"); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetAccount(ctx, apiKey)
}

var (
	defaultStoreMu sync.RWMutex
	defaultStore   *Store
)

func SetDefaultStore(store *Store) {
	defaultStoreMu.Lock()
	defaultStore = store
	defaultStoreMu.Unlock()
}

func DefaultStore() *Store {
	defaultStoreMu.RLock()
	defer defaultStoreMu.RUnlock()
	return defaultStore
}
