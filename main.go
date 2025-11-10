package main

import (
	"log"

	"github.com/usace-cloud-compute/cc-go-sdk"
	tiledb "github.com/usace-cloud-compute/cc-go-sdk/tiledb-store"
	_ "github.com/usace-cloud-compute/fragility-curves/internal/actions"
)

var commit string
var date string

func main() {
	cc.DataStoreTypeRegistry.Register("TILEDB", tiledb.TileDbEventStore{})
	pm, err := cc.InitPluginManager()
	if err != nil {
		log.Fatalf("Unable to initialize the plugin manager: %s\n", err)
	}

	pm.Logger.Info("Fragility Curves", "version", commit, "build-date", date)

	err = pm.RunActions()
	if err != nil {
		pm.Logger.Error(err.Error())
	}

}
