// Package matchkratosperiod: Periodic request sampling match function with count-based selection
// Matches one request in each period of N, counted on each route alone
// Can match the first request of the period, and can switch to the last one
// Good fit in load reduction, sampling, and managed traffic shaping scenarios
//
// matchkratosperiod: 周期性请求采样匹配函数，按计数选中请求
// 每 N 个请求匹配一个，每个路由各自计数
// 可以匹配每个周期的第一个请求，也可以匹配最后一个
// 适用于负载降低、采样和受控流量整形场景
package matchkratosperiod

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/go-kratos/kratos/v3/middleware/selector"
	"github.com/yylego/kratos-auth/authkratos"
	"github.com/yylego/syncmap"
)

// Config holds the periodic match configuration: the route scope, the period, and which request of the period matches
//
// Config 保存周期匹配的配置：路由范围、周期，以及周期里匹配哪一个请求
type Config struct {
	routeScope *authkratos.RouteScope
	period     uint32
	matchFirst bool
}

// NewConfig creates a periodic match config that matches one request in each period on each route
// A period of 0 counts as 1, which matches each request; the default matches the first request of each period
//
// NewConfig 创建周期匹配配置，每个路由每 period 个请求匹配一个
// period 为 0 时按 1 算，即每个请求都匹配；默认匹配每个周期的第一个请求
func NewConfig(routeScope *authkratos.RouteScope, period uint32) *Config {
	return &Config{
		routeScope: routeScope,
		period:     period,
		matchFirst: true,
	}
}

// WithMatchFirst picks which request of each period matches: the first (true, the default) and the last (false)
// With a period of 3, true matches requests 1, 4, 7 and false matches requests 3, 6, 9
//
// WithMatchFirst 选每个周期匹配哪一个请求：true 匹配第一个（默认），false 匹配最后一个
// 周期为 3 时，true 匹配第 1、4、7 个请求，false 匹配第 3、6、9 个请求
func (c *Config) WithMatchFirst(matchFirst bool) *Config {
	c.matchFirst = matchFirst
	return c
}

// NewMatchFunc creates a selector.MatchFunc: routes outside the scope do not match, routes in the scope match once in each period
// Pass it to selector.Server(...).Match(...) so the wrapped middleware runs on those requests
// Each match outcome goes to applog at debug level, so the level of applog decides if it shows
//
// NewMatchFunc 创建 selector.MatchFunc：范围外的路由不匹配，范围内的路由每个周期匹配一次
// 把它传给 selector.Server(...).Match(...)，被包裹的中间件就只对这些请求生效
// 每次匹配结果以 debug 级别写进 applog，显示与否由 applog 的日志级别决定
func NewMatchFunc(cfg *Config, applog *slog.Logger) selector.MatchFunc {
	applog.Info("match-kratos-period: new match func", "side", cfg.routeScope.Side, "operations", len(cfg.routeScope.OperationSet), "match-first", cfg.matchFirst, "period", cfg.period)
	applog.Debug("match-kratos-period: new match func route-scope", "route-scope", cfg.routeScope)

	period := uint64(max(cfg.period, 1))
	matchFirst := cfg.matchFirst
	routeCounts := syncmap.New[authkratos.Operation, *atomic.Uint64]()
	return func(ctx context.Context, operation string) bool {
		if !cfg.routeScope.Match(operation) {
			applog.Debug("match-kratos-period: match skip -- outside the route scope", "operation", operation, "side", cfg.routeScope.Side)
			return false
		}
		routeCount, _ := routeCounts.LoadOrStore(operation, &atomic.Uint64{})
		// count is the position of this request on the route, starting at 1
		// count 是这个请求在该路由上的序号，从 1 开始
		count := routeCount.Add(1)
		match := count%period == 0
		if matchFirst {
			match = (count-1)%period == 0
		}
		applog.Debug("match-kratos-period: period match", "operation", operation, "side", cfg.routeScope.Side, "match", match, "count", count)
		return match
	}
}
