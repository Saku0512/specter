package config

import "fmt"

func BuiltinLatencyProfiles() map[string]LatencyProfile {
	return map[string]LatencyProfile{
		"fast":      {DelayMin: 20, DelayMax: 80},
		"mobile-4g": {DelayMin: 150, DelayMax: 450},
		"slow-api":  {DelayMin: 800, DelayMax: 1600},
	}
}

func ResolveLatencyProfile(cfg *Config, name string) (LatencyProfile, bool) {
	if name == "" {
		return LatencyProfile{}, false
	}
	if cfg != nil {
		if profile, ok := cfg.LatencyProfiles[name]; ok {
			return profile, true
		}
	}
	profile, ok := BuiltinLatencyProfiles()[name]
	return profile, ok
}

func ValidateLatencyProfile(name string, profile LatencyProfile) []string {
	var errs []string
	prefix := fmt.Sprintf("latency profile %q", name)
	if profile.Delay < 0 {
		errs = append(errs, prefix+": delay must be non-negative")
	}
	if profile.DelayMin < 0 {
		errs = append(errs, prefix+": delay_min must be non-negative")
	}
	if profile.DelayMax < 0 {
		errs = append(errs, prefix+": delay_max must be non-negative")
	}
	if profile.DelayMax > 0 && profile.DelayMin > profile.DelayMax {
		errs = append(errs, prefix+": delay_min must be <= delay_max")
	}
	return errs
}
