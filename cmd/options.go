package cmd

import "github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"

type RenderOption func(*diagram.Config)

func WithMaxWidth(maxWidth int) RenderOption {
	return func(cfg *diagram.Config) {
		cfg.MaxWidth = maxWidth
		if maxWidth > 0 && (cfg.FitPolicy == "" || cfg.FitPolicy == diagram.FitPolicyNone) {
			cfg.FitPolicy = diagram.FitPolicyAuto
		}
	}
}

func WithAscii() RenderOption {
	return func(cfg *diagram.Config) {
		cfg.UseAscii = true
	}
}

func WithFitPolicy(policy string) RenderOption {
	return func(cfg *diagram.Config) {
		cfg.FitPolicy = policy
	}
}
