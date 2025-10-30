package actions

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/usace-cloud-compute/cc-go-sdk"
	"github.com/usace-cloud-compute/fragility-curves/internal/compute"
)

const (
	singleSampleActionName string = "single-sample"
)

func init() {
	cc.ActionRegistry.RegisterAction(singleSampleActionName, &SingleSampleAction{})
}

type SingleSampleAction struct {
	cc.ActionRunnerBase
}

func (ssa *SingleSampleAction) Run() error {
	a := ssa.Action

	if len(a.Outputs) != 1 {
		err := errors.New("more than one output was defined")
		return err
	}

	var fcm compute.Model
	modelReader, err := a.GetReader(cc.DataSourceOpInput{DataSourceName: "fragilitycurve", PathKey: "default"})
	if err != nil {
		return err
	}
	defer modelReader.Close()
	err = json.NewDecoder(modelReader).Decode(&fcm)
	if err != nil {
		return err
	}
	var seedSet compute.SeedSet
	var ec compute.EventConfiguration
	eventConfigurationReader, err := a.GetReader(cc.DataSourceOpInput{DataSourceName: "seeds", PathKey: "default"})
	if err != nil {
		return err
	}
	defer eventConfigurationReader.Close()
	err = json.NewDecoder(eventConfigurationReader).Decode(&ec)
	if err != nil {
		return err
	}

	seedSetName := "fragilitycurveplugin"
	seedSet, seedsFound := ec.Seeds[seedSetName]
	if !seedsFound {
		return fmt.Errorf("no seeds found by name of %v", seedSetName)
	}
	modelResult, err := fcm.Compute(seedSet.BlockSeed, seedSet.RealizationSeed)
	if err != nil {
		return err
	}
	data, err := json.Marshal(modelResult)
	if err != nil {
		return err
	}
	input := cc.PutOpInput{
		SrcReader:         bytes.NewReader(data),
		DataSourceOpInput: cc.DataSourceOpInput{DataSourceName: a.Outputs[0].Name, PathKey: "default"},
	}
	_, err = a.Put(input)
	return err
}
