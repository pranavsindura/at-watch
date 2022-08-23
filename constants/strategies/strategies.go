package strategyConstants

const (
	StrategyPositional string = "POSITIONAL"
	StrategyMarketOpen string = "MARKETOPEN"
)

var Strategies = map[string]struct{}{
	StrategyPositional: {},
	StrategyMarketOpen: {},
}
