package eval

// DecreeConfig holds runtime flags set by decree statements.
type DecreeConfig struct {
	IndexingBase   string // "zero", "one", "weekday" (default)
	DetHashing     bool
	AmbitiousMode  bool
	SoftCasts      bool
	SequentialMood bool
	NoForgiveness  bool
}

// NewDecreeConfig returns a DecreeConfig with defaults.
func NewDecreeConfig() *DecreeConfig {
	return &DecreeConfig{
		IndexingBase: "weekday",
	}
}

// Apply parses a decree string and updates the config.
func (d *DecreeConfig) Apply(decree string) {
	switch decree {
	case "zero_indexed":
		d.IndexingBase = "zero"
	case "one_indexed":
		d.IndexingBase = "one"
	case "deterministic_hashing":
		d.DetHashing = true
	case "soft_casts":
		d.SoftCasts = true
	case "ambitious_mode":
		d.AmbitiousMode = true
	case "sequential_mood":
		d.SequentialMood = true
	case "no_forgiveness":
		d.NoForgiveness = true
	}
}
