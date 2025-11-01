package config

type StrategyConfig interface {
	Validate() error
}
