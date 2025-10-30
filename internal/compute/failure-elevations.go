package compute

import (
	"errors"

	"github.com/usace-cloud-compute/cc-go-sdk"
)

func WriteFailureElevationsToTiledb(iomanager cc.IOManager, storeName string, path string, results []ModelResult) error {
	//get the store
	ds, err := iomanager.GetStore(storeName)
	if err != nil {
		return err
	}
	locations := make([]string, 0)
	for _, l := range results[0].Results {
		locations = append(locations, l.Name)
	}
	//check if it is the right store type
	tdbds, ok := ds.Session.(cc.MultiDimensionalArrayStore)
	if ok {
		arrayInput := cc.CreateArrayInput{
			ArrayPath: path,
			Dimensions: []cc.ArrayDimension{
				{
					Name:          "Events", //row
					DimensionType: cc.DIMENSION_INT,
					Domain:        []int64{1, int64(len(results))},
					TileExtent:    1,
				}, {
					Name:          "fragility_locations", //column
					DimensionType: cc.DIMENSION_INT,
					Domain:        []int64{1, int64(len(results[0].Results))},
					TileExtent:    int64(len(results[0].Results)),
				},
			},
			Attributes: []cc.ArrayAttribute{
				{Name: "failure_elevation", DataType: cc.ATTR_FLOAT64},
			},
			ArrayType:  cc.ARRAY_DENSE,
			CellLayout: cc.ROWMAJOR,
			TileLayout: cc.COLMAJOR,
		}
		err = tdbds.CreateArray(arrayInput)
		if err != nil {
			return err
		}

		//now make a put array input and put the data properly arranged.
		elevationdata := make([]float64, (len(results[0].Results))*(len(results)))
		for i, result := range results {
			for j, loc := range result.Results {
				elevationdata[(i*len(results[0].Results))+j] = loc.FailureElevation
			}
		}
		//create a buffer
		buffer := []cc.PutArrayBuffer{
			{
				AttrName: "failure_elevation",
				Buffer:   elevationdata,
			},
		}
		//create an input
		input := cc.PutArrayInput{
			Buffers:   buffer,
			DataPath:  path,
			ArrayType: cc.ARRAY_DENSE,
		}
		err = tdbds.PutArray(input)
		if err != nil {
			return err
		}
		mds, ok := ds.Session.(cc.MetadataStore)
		if ok {
			err = mds.PutMetadata("fragility_locations", locations)
			if err != nil {
				return err
			}

		} else {
			return errors.New("could not store metadata on which locations store failure elevations")
		}
	} else {
		//store does not support this data.
		return errors.New("store session is not a multidimensional array store")
	}
	return nil
}
