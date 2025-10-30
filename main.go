package main

import (
	"fmt"
	"log"

	"github.com/usace-cloud-compute/cc-go-sdk"
	tiledb "github.com/usace-cloud-compute/cc-go-sdk/tiledb-store"
)

func main() {
	fmt.Println("Starting Fragility Curves")
	cc.DataStoreTypeRegistry.Register("TILEDB", tiledb.TileDbEventStore{})
	pm, err := cc.InitPluginManager()
	if err != nil {
		log.Fatalf("Unable to initialize the plugin manager: %s\n", err)
	}

	err = pm.RunActions()
	if err != nil {
		pm.Logger.Error(err.Error())
	}

}
