package world

import (
	"encoding/json"
	"golang.org/x/exp/errors/fmt"
	"log"
	"math"
	"os"
)

type update struct {
	UUID       string `json:"uuid"`
	EventID    int    `json:"eventId"`
	RegionID   int    `json:"regionId"`
	Entering   bool   `json:"entering"`
	OccurredAt int    `json:"occurredAt"`
}

type Region struct {
	ID      int32   `json:"regionID,omitempty"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Lat     float64 `json:"lat,omitempty"`
	Lng     float64 `json:"lng,omitempty"`
	Radius  int32   `json:"radius,omitempty"`
	EventID int32   `json:"eventID"`
	x       float64
	y       float64
}

type RegionList struct {
	list []Region
}

func (s *State) LoadRegions(path string, lat, lng float64) {
	configFile, err := os.Open(path)
	if err != nil {
		log.Fatal("opening region file", err.Error())
	}

	var regions []Region

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&regions); err != nil {
		log.Fatal("parsing config file", err.Error())
	}

	for i, _ := range regions {
		regions[i].x, regions[i].y = latLngToCoords(regions[i].Lat, regions[i].Lng, lat, lng)
		fmt.Println(i, " - ", regions[i])
	}
	s.Regions = regions

}

func latLngToCoords(lat, lng, latOrigin, lngOrigin float64) (float64, float64) {
	const (
		MetersPerLat          = 111034.60528834906 // at 45 deg
		MetersPerLngAtEquator = 111319.458
	)
	lat -= latOrigin
	lng -= lngOrigin
	y := lat * MetersPerLat
	latRad := lat * math.Pi / 180
	x := lng * math.Cos(latRad) * MetersPerLngAtEquator
	return x, y
}
