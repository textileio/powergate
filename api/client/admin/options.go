package admin

var (
	defaultCfgGetMiners = cfgGetMiners{
		withPower: true,
	}
)

type cfgGetMiners struct {
	withPower bool
}

// GetMinersOption configures filters for getting miners.
type GetMinersOption func(*cfgGetMiners) error

// WithPowerGreaterThanZero filters miners that have power greater
// than zero.
func WithPowerGreaterThanZero(withPower bool) GetMinersOption {
	return func(cfg *cfgGetMiners) error {
		cfg.withPower = withPower
		return nil
	}
}
