package utils

import (
	"encoding/json"
	"log"

	"github.com/davecgh/go-spew/spew"
)

func JsonString(val interface{}) (string, error) {
	json, err := json.Marshal(val)
	if err != nil {
		log.Fatal("Failed to generate json", err)
		return "", err
	}

	return string(json), nil
}

func PrintWithStruct(val interface{}) {
	spew.Dump(val)
}
