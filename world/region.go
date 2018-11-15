package world

import (
	"bytes"
	"encoding/json"
	"github.com/real-time-footfall-analysis/rtfa-simulation/individual"
	"golang.org/x/exp/errors/fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

type update struct {
	UUID       string `json:"uuid"`
	EventID    int32  `json:"eventId"`
	RegionID   int32  `json:"regionId"`
	Entering   bool   `json:"entering"`
	OccurredAt int64  `json:"occurredAt"`
}

type Region struct {
	ID      int32   `json:"regionID,omitempty"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Lat     float64 `json:"lat,omitempty"`
	Lng     float64 `json:"lng,omitempty"`
	Radius  int32   `json:"radius,omitempty"`
	EventID int32   `json:"eventID"`
	X       float64 `json:"X"`
	Y       float64 `json:"Y"`
	sqRad   float64
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
		if regions[i].X == 0 || regions[i].Y == 0 {
			regions[i].X, regions[i].Y = latLngToCoords(regions[i].Lat, regions[i].Lng, lat, lng)
		}
		regions[i].sqRad = math.Pow(float64(regions[i].Radius), 2)
		r := regions[i]
		fmt.Println(i, " - ", r)
		if r.X < 0 || r.X > float64(s.GetWidth()) ||
			r.Y < 0 || r.Y > float64(s.GetHeight()) {
			fmt.Printf("Warning! Region: %v - %s is outside the boundries of the world at X,Y: %v, %v\n",
				r.ID, r.Name, r.X, r.Y)
		}
	}
	s.Regions = regions

}

func latLngToCoords(lat, lng, latOrigin, lngOrigin float64) (float64, float64) {
	const (
		MetersPerLat          = 111034.60528834906 // at 45 deg lat
		MetersPerLngAtEquator = 111319.458
	)
	lat -= latOrigin
	lng -= lngOrigin
	y := lat * MetersPerLat
	latRad := lat * math.Pi / 180
	x := lng * math.Cos(latRad) * MetersPerLngAtEquator
	return x, y
}

func UpdateServer(regions *[]Region, actor *individual.Individual, time time.Time) {
	for i, r := range *regions {
		fmt.Println(i, " - ", r)
		x, y := actor.Loc.GetLatestXY()
		dx := x - r.X
		dy := y - r.Y
		distanceSquared := math.Pow(dx, 2) + math.Pow(dy, 2)
		if r.sqRad > distanceSquared {
			// this actor is in this region
			_, knownInside := actor.RegionIds[r.ID]
			if !knownInside {
				// we must send update to backend
				u := update{EventID: r.EventID, RegionID: r.ID, UUID: actor.UUID, Entering: true, OccurredAt: time.Unix()}
				fmt.Println(u)
				sendUpdate(&u)
			}
		} else {
			// this actor is not in this region
			_, knownInside := actor.RegionIds[r.ID]
			if knownInside {
				// we must send update to backend to say this actor is no longer in the region.
				u := update{EventID: r.EventID, RegionID: r.ID, UUID: actor.UUID, Entering: false, OccurredAt: time.Unix()}
				fmt.Println(u)
				sendUpdate(&u)
			}
		}

	}
}

const url = "http://api.jackchorley.club/update"

func sendUpdate(u *update) {
	var jsonStr, err = json.Marshal(*u)
	if err != nil {
		log.Fatal("Cannot marshal update: ", *u)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("Cannot connect to backend")
	} else {
		defer resp.Body.Close()

		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("response Body:", string(body))
	}
}
