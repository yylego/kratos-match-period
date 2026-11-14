package matchkratosperiod

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/kratos-auth/authkratos"
)

// collectMatches calls the match func on the operation n times and returns the positions that match, starting at 1
// collectMatches 对同一个 operation 调 n 次匹配函数，返回匹配上的序号，从 1 开始
func collectMatches(matchFunc func(ctx context.Context, operation string) bool, operation string, n int) []int {
	var positions []int
	for k := 1; k <= n; k++ {
		if matchFunc(context.Background(), operation) {
			positions = append(positions, k)
		}
	}
	return positions
}

func TestNewConfig(t *testing.T) {
	cfg := NewConfig(authkratos.NewInclude("/api.Demo/Search"), 3)
	require.Equal(t, uint32(3), cfg.period)
	require.True(t, cfg.matchFirst)
}

func TestConfig_WithMatchFirst(t *testing.T) {
	scope := authkratos.NewInclude("/api.Demo/Search")

	matchFunc := NewMatchFunc(NewConfig(scope, 3).WithMatchFirst(true), slog.Default())
	require.Equal(t, []int{1, 4, 7}, collectMatches(matchFunc, "/api.Demo/Search", 9))

	matchFunc = NewMatchFunc(NewConfig(scope, 3).WithMatchFirst(false), slog.Default())
	require.Equal(t, []int{3, 6, 9}, collectMatches(matchFunc, "/api.Demo/Search", 9))
}

func TestNewMatchFunc(t *testing.T) {
	scope := authkratos.NewInclude("/api.Demo/Search", "/api.Demo/Browse")

	t.Run("period-0-and-1-match-each-request", func(t *testing.T) {
		for _, period := range []uint32{0, 1} {
			matchFunc := NewMatchFunc(NewConfig(scope, period), slog.Default())
			require.Equal(t, []int{1, 2, 3, 4}, collectMatches(matchFunc, "/api.Demo/Search", 4))
		}
	})

	t.Run("each-route-counts-on-its-own", func(t *testing.T) {
		matchFunc := NewMatchFunc(NewConfig(scope, 2), slog.Default())
		require.True(t, matchFunc(context.Background(), "/api.Demo/Search"))
		require.True(t, matchFunc(context.Background(), "/api.Demo/Browse"))
		require.False(t, matchFunc(context.Background(), "/api.Demo/Search"))
		require.False(t, matchFunc(context.Background(), "/api.Demo/Browse"))
	})

	t.Run("route-outside-the-scope-does-not-match", func(t *testing.T) {
		matchFunc := NewMatchFunc(NewConfig(scope, 1), slog.Default())
		require.Empty(t, collectMatches(matchFunc, "/api.Demo/Health", 4))
	})

	t.Run("concurrent-requests-match-the-exact-count", func(t *testing.T) {
		matchFunc := NewMatchFunc(NewConfig(scope, 10), slog.Default())
		var count atomic.Int64
		var wg sync.WaitGroup
		for range 1000 {
			wg.Go(func() {
				if matchFunc(context.Background(), "/api.Demo/Search") {
					count.Add(1)
				}
			})
		}
		wg.Wait()
		require.Equal(t, int64(100), count.Load())
	})
}
