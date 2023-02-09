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
	Port                   int      `toml:"Port" validate:"gte=1,lte=65536"`
	ReadTimeoutSeconds     int      `toml:"ReadTimeoutSeconds" validate:"gte=1,lte=3000"`
	WriteTimeoutSeconds    int      `toml:"WriteTimeoutSeconds" validate:"gte=1,lte=3000"`
	ShutdownMaxTimeSeconds int      `toml:"ShutdownMaxTimeSeconds" validate:"gte=1,lte=3000"`
	CORSAllowMethods       []string `toml:"CORSAllowMethods" validate:"min=1"`
	CORSAllowHeaders       []string `toml:"CORSAllowHeaders" validate:"min=1"`
	CORSAllowOrigins       []string `toml:"CORSAllowOrigins" validate:"min=1"`
}

type checker struct {
	ServiceMyIP                 string `toml:"ServiceMyIP" validate:"url"`
	HeaderUserAgent             string `toml:"HeaderUserAgent"`
	RequestTimeoutSeconds       int    `toml:"RequestTimeoutSeconds" validate:"gte=1,lte=300"`
	DelayBetweenAttemptsSeconds int    `toml:"DelayBetweenAttemptsSeconds" validate:"gte=1,lte=300"`
}

func GetConfig(fileName string) (Conf, error) {
	var conf Conf
	if _, err := toml.DecodeFile(fileName, &conf); err != nil {
		return Conf{}, fmt.Errorf("decode file: %w", err)
	}

	if err := validator.New().Struct(conf); err != nil {
		var validatorErrors validator.ValidationErrors
		if errors.As(err, &validatorErrors) {
			if err := checkValidatorErrs(validatorErrors); err != nil {
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
