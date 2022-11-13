package envconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type subConfig struct {
	Internal    string
	InternalEnv string `env:"value"`
}

type testConfig struct {
	StructValue    subConfig `env:"struct"`
	StructValue2   subConfig
	StringValue    string
	IntValue       int
	StringEnvValue string `env:"string_value"`
	IntEnvValue    int    `env:"int_value"`
}

func Test(t *testing.T) {
	testCases := []struct {
		desc   string
		env    map[string]string
		expect testConfig
		err    bool
	}{
		{
			desc: "noenv",
		},
		{
			desc: "env",
			env: map[string]string{
				"struct_value":     "one",
				"struct_env_value": "two",
				"string_value":     "3",
				"int_value":        "4",
				"string_env_value": "five",
				"int_env_value":    "6",
				"struct":           "se7en",
				"value":            "8",
			},
			expect: testConfig{
				StructValue: subConfig{
					InternalEnv: "one",
				},
				StructValue2: subConfig{
					InternalEnv: "8",
				},
				StringEnvValue: "3",
				IntEnvValue:    4,
			},
		},
		{
			desc: "badtype",
			env: map[string]string{
				"int_value": "six",
			},
			err: true,
		},
	}
	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			cfg := testConfig{}
			for n, v := range c.env {
				require.NoError(t, os.Setenv(n, v))
			}
			err := UnmarshalEnv(&cfg)
			if c.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expect, cfg)
			}
		})
	}
}
