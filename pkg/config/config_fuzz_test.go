package config

import (
	"encoding/json"
	"testing"
)

func FuzzConfigDecode(f *testing.F) {
	f.Add([]byte(`{"activePreset":"quick","enabledCategories":["temp","logs"],"language":"en"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(``))
	f.Add([]byte(`{"activePreset":"` + string(make([]byte, 10000)) + `"}`))
	f.Add([]byte(`{"enabledCategories":["` + string(make([]byte, 5000)) + `"]}`))
	f.Add([]byte(`{"autoRiskOverrides":{"temp":"safe"},"exclusions":["..\\..\\etc\\passwd"]}`))
	f.Add([]byte(`{"unknownField":true,"activePreset":123}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var cfg Config
		err := json.Unmarshal(data, &cfg)
		_ = err // should not panic

		if cfg.ActivePreset != "" &&
			cfg.ActivePreset != "quick" &&
			cfg.ActivePreset != "standard" &&
			cfg.ActivePreset != "deep" &&
			cfg.ActivePreset != "custom" {
			t.Logf("unknown activePreset value: %q", cfg.ActivePreset)
		}
		if len(cfg.Exclusions) > 1000 {
			t.Errorf("config has %d exclusions — suspiciously large", len(cfg.Exclusions))
		}
	})
}
