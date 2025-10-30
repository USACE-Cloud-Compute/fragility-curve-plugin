package compute

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"

	"github.com/HydrologicEngineeringCenter/go-statistics/paireddata"
	"github.com/HydrologicEngineeringCenter/go-statistics/statistics"
)

type Model struct {
	Name      string                   `json:"name"`
	Locations []FragilityCurveLocation `json:"locations"`
}
type FragilityCurveLocation struct {
	Hydraulic_Model_Unit string                           `json:"hydraulic_model_unit"`
	Name                 string                           `json:"location"`
	Breach_FID           string                           `json:"breach_fid"`
	NLD_System_ID        string                           `json:"nld_system_id"`
	NLD_System_Name      string                           `json:"nld_system_name"`
	NLD_Segment_ID       string                           `json:"nld_segment_id"`
	NIDID                string                           `json:"nidid"`
	NID_Dam_Name         string                           `json:"nid_dam_name"`
	FailureMode          string                           `json:"failure_mode"`
	Source               string                           `json:"source"`
	Top_Elev             string                           `json:"top_elev"`
	Toe_Elev             string                           `json:"toe_elev"`
	FragilityCurve       paireddata.UncertaintyPairedData `json:"probability-stage"`
}
type FragilityCurveLocationResult struct {
	Name             string  `json:"location"`
	FailureElevation float64 `json:"failure_elevation"`
}
type ModelResult struct {
	Results []FragilityCurveLocationResult `json:"results"`
}

func InitFragilityCurveLocation(locationbytes []byte) FragilityCurveLocation {
	file := string(locationbytes)
	file = strings.ReplaceAll(file, "\r", "")
	lines := strings.Split(file, "\n")
	hydraulic_model_unit := parseLine(lines[0])
	name := parseLine(lines[1])
	breachfid := parseLine(lines[2])
	NLD_System_Id := parseLine(lines[3])
	NLD_System_Name := parseLine(lines[4])
	NLD_Segment_ID := parseLine(lines[5])
	NIDID := parseLine(lines[6])
	nid_dam_name := parseLine(lines[7])
	failure_mode := parseLine(lines[8])
	source := parseLine(lines[9])
	top_elev := parseLine(lines[10])
	toe_elev := parseLine(lines[11])
	lines = lines[13:] //skip stage,probability
	xvals := make([]float64, 0)
	yvals := make([]statistics.ContinuousDistribution, 0)

	for _, line := range lines {
		row := strings.Split(line, ",")
		if len(row) >= 2 {
			xstring := row[0]
			ystring := row[1]
			xval, err := strconv.ParseFloat(xstring, 64)
			if err != nil {
				return FragilityCurveLocation{}
			}
			yval, err := strconv.ParseFloat(ystring, 64)
			if err != nil {
				return FragilityCurveLocation{}
			}
			xvals = append(xvals, xval)
			yvals = append(yvals, statistics.DeterministicDistribution{Value: yval})
		}

	}
	return FragilityCurveLocation{
		Hydraulic_Model_Unit: hydraulic_model_unit,
		Name:                 name,
		Breach_FID:           breachfid,
		NLD_System_ID:        NLD_System_Id,
		NLD_System_Name:      NLD_System_Name,
		NLD_Segment_ID:       NLD_Segment_ID,
		NIDID:                NIDID,
		NID_Dam_Name:         nid_dam_name,
		FailureMode:          failure_mode,
		Source:               source,
		Top_Elev:             top_elev,
		Toe_Elev:             toe_elev,
		FragilityCurve:       paireddata.UncertaintyPairedData{Xvals: xvals, Yvals: yvals},
	}
}
func parseLine(line string) string {
	parts := strings.Split(line, ",")
	if len(parts) < 1 {
		return ""
	} else {
		if len(parts) == 1 {
			return ""
		} else {
			if len(parts) == 2 {
				return parts[1]
			} else { //what if the name has commas in it?
				fullname := ""
				parts = parts[1:]
				for _, p := range parts {
					fullname += p
				}
				return fullname
			}

		}
	}
}
func (fcm Model) Compute(variabilitySeed int64, uncertaintySeed int64) (ModelResult, error) {
	uncertaintyRandom := rand.New(rand.NewSource(uncertaintySeed))
	variabilityRandom := rand.New(rand.NewSource(variabilitySeed))
	results := ModelResult{
		Results: make([]FragilityCurveLocationResult, len(fcm.Locations)),
	}
	for idx, fcl := range fcm.Locations {
		//sample fragility curve for a location with knowledge uncertianty
		pd := fcl.FragilityCurve.SampleValueSampler(uncertaintyRandom.Float64())
		pd2, ok := pd.(paireddata.PairedData)
		if ok {
			//invert the paired data because we will be sampling probability to derive a stage.
			pd3 := paireddata.PairedData{
				Xvals: pd2.Yvals,
				Yvals: pd2.Xvals,
			}
			//sample sampledfragility curve at a location with natural variability
			locationResult := FragilityCurveLocationResult{
				Name:             fcl.Name,
				FailureElevation: pd3.SampleValue(variabilityRandom.Float64()),
			}
			results.Results[idx] = locationResult
		} else {
			return ModelResult{}, errors.New("failed to convert to paired data")
		}

	}
	return results, nil
}
func (fcm Model) ComputeAll(seeds []SeedSet) ([]ModelResult, error) {
	results := make([]ModelResult, 0)
	for _, seed := range seeds {
		result, err := fcm.Compute(seed.BlockSeed, seed.RealizationSeed)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}
