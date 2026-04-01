package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type cachedRate struct {
	rate      decimal.Decimal
	fetchedAt time.Time
}

type ExchangeService struct {
	baseCurrency string
	mu           sync.RWMutex
	cache        map[string]cachedRate
	client       *http.Client
	ttl          time.Duration
}

func NewExchangeService(baseCurrency string) *ExchangeService {
	return &ExchangeService{
		baseCurrency: baseCurrency,
		cache:        make(map[string]cachedRate),
		client:       &http.Client{Timeout: 10 * time.Second},
		ttl:          1 * time.Hour,
	}
}

func (s *ExchangeService) BaseCurrency() string {
	return s.baseCurrency
}

// GetRate returns the exchange rate from one currency to another for a given month.
func (s *ExchangeService) GetRate(ctx context.Context, from, to, month string) (decimal.Decimal, error) {
	if from == to {
		return decimal.NewFromInt(1), nil
	}

	key := fmt.Sprintf("%s:%s:%s", month, from, to)

	s.mu.RLock()
	if cached, ok := s.cache[key]; ok && time.Since(cached.fetchedAt) < s.ttl {
		s.mu.RUnlock()
		return cached.rate, nil
	}
	s.mu.RUnlock()

	// Use the first day of the month for historical rates
	date := month + "-01"
	url := fmt.Sprintf("https://api.frankfurter.app/%s?from=%s&to=%s", date, from, to)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("creating request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("fetching exchange rate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return decimal.Decimal{}, fmt.Errorf("frankfurter API returned status %d", resp.StatusCode)
	}

	var result struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return decimal.Decimal{}, fmt.Errorf("decoding response: %w", err)
	}

	rateFloat, ok := result.Rates[to]
	if !ok {
		return decimal.Decimal{}, fmt.Errorf("rate for %s not found in response", to)
	}

	rate := decimal.NewFromFloat(rateFloat)

	s.mu.Lock()
	s.cache[key] = cachedRate{rate: rate, fetchedAt: time.Now()}
	s.mu.Unlock()

	return rate, nil
}

// Convert converts an amount from one currency to another for a given month.
func (s *ExchangeService) Convert(ctx context.Context, amount decimal.Decimal, from, to, month string) (decimal.Decimal, error) {
	if from == to {
		return amount, nil
	}
	rate, err := s.GetRate(ctx, from, to, month)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return amount.Mul(rate), nil
}
