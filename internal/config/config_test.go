package config

import (
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const goldenCfgPath = "testdata/golden.yaml"

type goldenConfig struct {
	Int      int
	String   string
	Duration time.Duration
}

func Test(t *testing.T) {
	testCases := []struct {
		name  string
		flags map[string]string
		devel bool
	}{
		// TODO add checks
		{
			name:  "default",
			flags: map[string]string{"config": goldenCfgPath},
		},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			for n, v := range c.flags {
				flag.Set(n, v)
			}
			var out goldenConfig
			check := goldenConfig{
				Int:      10,
				String:   "string",
				Duration: 10 * time.Minute,
			}
			flags, err := ParseCustomConfig(&out)
			require.NoError(t, err)
			require.Equal(t, goldenCfgPath, flags.Path)
			require.Equal(t, c.devel, flags.Devel)
			require.Equal(t, check, out)
		})
	}
}
