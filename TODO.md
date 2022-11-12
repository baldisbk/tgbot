* generic config with envvars (reflect+structtags)
    ```
    type SomeConfig struct {
        SomeValue  string `yaml:"some_value" env:"SOME_VALUE"`
        OtherValue string `yaml:"other_value"`
    }
    type Config struct {
        SomeConfig EnvConfig `yaml:"some_config"`
    }
    cfg := Config{
        SomeConfig: EnvConfig{
            Config: SomeConfig{}
        }
    }
    type EnvConfig struct {
        Config interface{}
    }
    func (c *EnvConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	    if err := unmarshal(&c.Config); err != nil {
		    return xerrors.Errorf("unmarshal: %w", err)
	    }
        // some reflect + env lookup magic
        return nil
    }
    ```
* structured logging
* webhook server
* tgapi mock
* strike add
* configurable postpones
* configurable check times (with timezones)
