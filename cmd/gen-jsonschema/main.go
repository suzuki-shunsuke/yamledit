package main

import (
	"fmt"
	"log"

	"github.com/suzuki-shunsuke/gen-go-jsonschema/jsonschema"
	"github.com/suzuki-shunsuke/yamledit/pkg/config"
)

func main() {
	if err := jsonschema.Write(&config.Ruleset{}, "json-schema/migration.json"); err != nil {
		log.Fatal(fmt.Errorf("create or update a JSON Schema: %w", err))
	}
}
