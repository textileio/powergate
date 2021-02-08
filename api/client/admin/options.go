package admin

var (
	defaultCfgGetMiners = cfgGetMiners{
		withPower: true,
	}
)

type cfgGetMiners struct {
	withPower bool
}

type GetMinersOption func(*cfgGetMiners) error

func WithPowerGreaterThanZero(withPower bool) GetMinersOption {
	return func(cfg *cfgGetMiners) error {
		cfg.withPower = withPower
		return nil
	}
}
