package envconfig

import (
	"os"
	"reflect"
	"strconv"

	"golang.org/x/xerrors"
)

const (
	envConfigTag = "env"
	yamlTag      = "yaml"
)

func UnmarshalEnv(cfg interface{}) error {
	return unmarshalEnv(reflect.TypeOf(cfg), reflect.ValueOf(cfg), "")
}

func unmarshalEnv(t reflect.Type, v reflect.Value, prefix string) error {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		// no env tags - job is done
		return nil
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)
		if f.Anonymous || f.PkgPath != "" {
			continue
		}
		if !fv.IsValid() || !fv.CanSet() {
			continue
		}
		var tag string
		if v, ok := f.Tag.Lookup(envConfigTag); ok {
			tag = v
			if prefix != "" {
				tag = prefix + "_" + tag
			}
			if token, ok := os.LookupEnv(tag); ok {
				switch fv.Kind() {
				case reflect.String:
					fv.SetString(token)
				case reflect.Int:
					iVal, err := strconv.Atoi(token)
					if err != nil {
						return xerrors.Errorf("type error: %w", err)
					}
					fv.SetInt(int64(iVal))
				}
			}
		}
		if fv.Kind() == reflect.Struct {
			if err := unmarshalEnv(f.Type, fv, tag); err != nil {
				return xerrors.Errorf("dive %s: %w", v, err)
			}
			continue
		}
	}
	return nil
}
