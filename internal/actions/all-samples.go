package actions

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/usace-cloud-compute/cc-go-sdk"
	"github.com/usace-cloud-compute/fragility-curves/internal/compute"
)

const (
	allSamplesActionName string = "all-samples"
)

func init() {
	cc.ActionRegistry.RegisterAction(allSamplesActionName, &AllSamplesAction{})
}

type AllSamplesAction struct {
	cc.ActionRunnerBase
}

func (asa *AllSamplesAction) Run() error {
	a := asa.Action
	readSeedsFromTiledb := a.Attributes.GetBooleanOrDefault("seeds_format", false)
	writeSamplesToTiledb := a.Attributes.GetBooleanOrDefault("elevations_format", false)
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
	//seeds
	seeds := make([]compute.SeedSet, 0)
	if readSeedsFromTiledb {
		seeds, err = compute.ReadSeedsFromTiledb(a.IOManager, "store", "seeds", "fragilitycurveplugin") //improve this to not be hard coded.
		if err != nil {
			return err
		}
	} else {
		//json
		eventConfigurationReader, err := a.GetReader(cc.DataSourceOpInput{DataSourceName: "seeds", PathKey: "default"})
		if err != nil {
			return err
		}
		var ecs []compute.EventConfiguration
		defer eventConfigurationReader.Close()
		err = json.NewDecoder(eventConfigurationReader).Decode(&ecs)
		if err != nil {
			return err
		}
		for _, ec := range ecs {
			seeds = append(seeds, ec.Seeds["fragilitycurveplugin"])
		}
	}

	modelResult, err := fcm.ComputeAll(seeds)
	if err != nil {
		return err
	}
	if writeSamplesToTiledb {
		err = compute.WriteFailureElevationsToTiledb(a.IOManager, "store", "failure_elevations", modelResult)
		if err != nil {
			return err
		}
	} else {
		strdatab := strings.Builder{}
		pathPattern := a.Outputs[0].Paths["event"]
		tenpercent := len(modelResult) / 10
		percent_complete := 0
		for i, r := range modelResult {
			istring := fmt.Sprintf("%v", i+1)
			if i%tenpercent == 0 {
				asa.Log("Progress", "percent", percent_complete)
				percent_complete += 10
			}
			if i == 0 {
				strdatab.WriteString("event_number")
				for _, elev := range r.Results {
					strdatab.WriteString(fmt.Sprintf(",%s", elev.Name))
				}
				strdatab.WriteString("\n")
			}
			strdatab.WriteString(istring)
			for _, elev := range r.Results {
				strdatab.WriteString(fmt.Sprintf(",%v", elev.FailureElevation))
			}
			strdatab.WriteString("\n")

			a.Outputs[0].Paths["event"] = strings.ReplaceAll(pathPattern, "${VAR::eventnumber}", istring)
			data, err := json.Marshal(r)
			if err != nil {
				return err
			}
			input := cc.PutOpInput{
				SrcReader:         bytes.NewReader(data),
				DataSourceOpInput: cc.DataSourceOpInput{DataSourceName: a.Outputs[0].Name, PathKey: "event"},
			}
			_, err = a.Put(input)
			if err != nil {
				return err
			}
		}
		data := []byte(strdatab.String())
		input := cc.PutOpInput{
			SrcReader:         bytes.NewReader(data),
			DataSourceOpInput: cc.DataSourceOpInput{DataSourceName: a.Outputs[0].Name, PathKey: "default"},
		}
		_, err = a.Put(input)
		if err != nil {
			return err
		}
	}
	return nil
}
