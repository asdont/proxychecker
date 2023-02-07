package app

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/ip2location/ip2location-go"
)

var errFileEmpty = errors.New("file is empty")

func DBGeo(fileName string) (*ip2location.DB, error) {
	if err := checkFile(fileName); err != nil {
		return nil, fmt.Errorf("check %s: %w", fileName, err)
	}

	currentAppPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	geoDB, err := ip2location.OpenDB(currentAppPath + "/" + fileName)
	if err != nil {
		log.Fatal(err)
	}

	return geoDB, nil
}

func checkFile(fileName string) error {
	fileStat, err := os.Stat(fileName)
	if err != nil {
		return fmt.Errorf("stat %s: %w", fileName, err)
	}

	if fileStat.Size() == 0 {
		return fmt.Errorf("%s: %w", fileName, errFileEmpty)
	}

	return nil
}
