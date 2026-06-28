// Package matchkratosperiod: Periodic request sampling middleware with count-based throttling
// Provides predictive request sampling with configurable period count
// Passes each Nth request while blocking others with first-match support if needed
// Good fit in load reduction, sampling, and managed traffic shaping scenarios
//
// matchkratosperiod: 周期性请求采样中间件，基于计数器的限流
// 提供可配置周期数的确定性请求采样
// 每 N 个请求放行一个，其余阻断，支持可选的首次匹配
// 适用于负载降低、采样和受控流量整形场景
package matchkratosperiod

import (
	"context"
	"log/slog"
	"sync"

	"github.com/go-kratos/kratos/v3/middleware/selector"
	"github.com/yylego/kratos-auth/authkratos"
	"github.com/yylego/neatjson/neatjsons"
	"github.com/yylego/syncmap"
)

type Config struct {
	routeScope *authkratos.RouteScope
	n          uint32
	matchFirst bool
	spanHooks  []authkratos.NewSpanHookFunc // 追踪 span 钩子列表，为空时不启动追踪
	debugMode  bool
}

func NewConfig(routeScope *authkratos.RouteScope, n uint32) *Config {
	return &Config{
		routeScope: routeScope,
		n:          n,
		matchFirst: true,
		spanHooks:  nil,
		debugMode:  false,
	}
}

func (c *Config) WithMatchFirst(matchFirst bool) *Config {
	c.matchFirst = matchFirst
	return c
}

func (c *Config) WithDebugMode(debugMode bool) *Config {
	c.debugMode = debugMode
	return c
}

// WithNewSpanHook appends a span hook factory to the hook list
// When hooks are present, each hook is started for every match evaluation
//
// WithNewSpanHook 向钩子列表追加一个 span 钩子工厂
// 存在钩子时，每次匹配评估都会启动所有钩子
func (c *Config) WithNewSpanHook(fn authkratos.NewSpanHookFunc) *Config {
	c.spanHooks = append(c.spanHooks, fn)
	return c
}

func NewMatchFunc(cfg *Config, applog *slog.Logger) selector.MatchFunc {
	applog.Info("match-kratos-period: new match func", "side", cfg.routeScope.Side, "operations", len(cfg.routeScope.OperationSet), "match-first", cfg.matchFirst, "period", cfg.n)
	if cfg.debugMode {
		applog.Debug("match-kratos-period: new match func route-scope", "route-scope", neatjsons.S(cfg.routeScope))
	}

	type countBox struct {
		mutex *sync.Mutex
		count uint64
	}
	mp := syncmap.New[authkratos.Operation, *countBox]()
	return func(ctx context.Context, operation string) bool {
		defer authkratos.RunSpanHooks(ctx, cfg.spanHooks, "match-kratos-period")()

		if match := cfg.routeScope.Match(operation); !match {
			if cfg.debugMode {
				applog.Debug("match-kratos-period: match next -> skip period", "operation", operation, "side", cfg.routeScope.Side, "match", authkratos.BooleanToNum(match))
			}
			return false
		}
		value, loaded := mp.LoadOrStore(operation, &countBox{&sync.Mutex{}, 0})
		if !loaded && cfg.matchFirst {
			if cfg.debugMode {
				applog.Debug("match-kratos-period: match next -> match first (count=0)", "operation", operation, "side", cfg.routeScope.Side, "match", authkratos.BooleanToNum(true))
			}
			return true
		}
		value.mutex.Lock()
		value.count = (value.count + 1) % uint64(max(cfg.n, 1))
		count := value.count
		value.mutex.Unlock()
		match := count == 0
		if cfg.debugMode {
			if match {
				applog.Debug("match-kratos-period: match next -> period matched", "operation", operation, "side", cfg.routeScope.Side, "match", authkratos.BooleanToNum(match), "count", count)
			} else {
				applog.Debug("match-kratos-period: match skip -- period skipped", "operation", operation, "side", cfg.routeScope.Side, "match", authkratos.BooleanToNum(match), "count", count)
			}
		}
		return match
	}
}
