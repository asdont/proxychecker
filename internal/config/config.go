package config

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
)

var errInvalidField = errors.New("invalid field")

type Conf struct {
	Server  server
	Checker checker
}

type server struct {
	Port                   int `toml:"Port" validate:"gte=1,lte=65536"`
	ReadTimeoutSeconds     int `toml:"ReadTimeoutSeconds" validate:"gte=1,lte=3000"`
	WriteTimeoutSeconds    int `toml:"WriteTimeoutSeconds" validate:"gte=1,lte=3000"`
	ShutdownMaxTimeSeconds int `toml:"ShutdownMaxTimeSeconds" validate:"gte=1,lte=3000"`
}

type checker struct {
	HeaderUserAgent             string `toml:"HeaderUserAgent" validate:"g"`
	ClientTimeoutSeconds        int    `toml:"ClientTimeoutSeconds" validate:"gte=1,lte=300"`
	AttemptsGetResponse         int    `toml:"AttemptsGetResponse" validate:"gte=1,lte=100"`
	DelayBetweenAttemptsSeconds int    `toml:"DelayBetweenAttemptsSeconds" validate:"gte=1,lte=300"`
}

func GetConfig(fileName string) (Conf, error) {
	var conf Conf
	if _, err := toml.DecodeFile(fileName, &conf); err != nil {
		return Conf{}, fmt.Errorf("decode file: %w", err)
	}

	if err := validator.New().Struct(conf); err != nil {
		var vErrors validator.ValidationErrors
		if errors.As(err, &vErrors) {
			if err := checkValidatorErrs(vErrors); err != nil {
				return Conf{}, fmt.Errorf("validator: check err: %w", err)
			}

			return Conf{}, fmt.Errorf("validator: %w", err)
		}
	}

	return conf, nil
}

func checkValidatorErrs(errs validator.ValidationErrors) error {
	for _, err := range errs {
		return fmt.Errorf("%w: %s(%s): see it <%v> want <%s=%s>",
			errInvalidField,
			err.StructNamespace(),
			err.Type(),
			err.Value(),
			err.ActualTag(),
			err.Param())
	}

	return nil
}
