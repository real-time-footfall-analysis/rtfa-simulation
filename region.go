package main

import (
	"bytes"
	"encoding/json"
	"golang.org/x/exp/errors/fmt"
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
	Radius  int     `json:"radius,omitempty"`
	EventID int32   `json:"eventID"`
	X       float64 `json:"X"`
	Y       float64 `json:"Y"`
	sqRad   float64
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

	newList := make([]update, 0)
	bulkUpdate = &newList

	if s.BulkSend {
		updateChannel = make(chan []byte, 50)
		startBulkConsumer(updateChannel)
	}

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

func UpdateServer(regions *[]Region, individual *Individual, time time.Time, bulk bool) {
	for _, r := range *regions {
		x, y := individual.Loc.GetLatestXY()
		dx := x - r.X
		dy := y - r.Y
		distanceSquared := math.Pow(dx, 2) + math.Pow(dy, 2)
		if r.sqRad > distanceSquared {
			// this individual is in this region
			//_, knownInside := individual.RegionIds[r.ID]
			if !individual.RegionIds[r.ID] {
				individual.RegionIds[r.ID] = true
				// we must send update to backend
				u := update{EventID: r.EventID, RegionID: r.ID, UUID: individual.UUID, Entering: true, OccurredAt: time.Unix()}
				if bulk {
					*bulkUpdate = append(*bulkUpdate, u)
				} else {
					sendUpdate(&u)
				}

			}
		} else {
			// this individual is not in this region
			//_, knownInside := individual.RegionIds[r.ID]
			if individual.RegionIds[r.ID] {
				individual.RegionIds[r.ID] = false
				// we must send update to backend to say this individual is no longer in the region.
				u := update{EventID: r.EventID, RegionID: r.ID, UUID: individual.UUID, Entering: false, OccurredAt: time.Unix()}
				if bulk {
					*bulkUpdate = append(*bulkUpdate, u)
				} else {
					sendUpdate(&u)
				}
			}
		}

	}
}

func LeaveAllRegions(regions *[]Region, individual *Individual, time time.Time, bulk bool) {
	for rID, b := range individual.RegionIds {
		if b {
			r := (*regions)[rID]
			u := update{EventID: r.EventID, RegionID: r.ID, UUID: individual.UUID, Entering: false, OccurredAt: time.Unix()}
			if bulk {
				*bulkUpdate = append(*bulkUpdate, u)
			} else {
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
		err := resp.Body.Close()
		if err != nil {
			log.Println("cannot close http response, don't care")
		}
	}
}

const bulkUrl = "http://api.jackchorley.club/bulkUpdate"

var bulkUpdate *[]update
var updateChannel chan []byte

func SendBulk() {
	if bulkUpdate == nil || len(*bulkUpdate) < 10 {
		return
	}
	updateList := bulkUpdate
	newList := make([]update, 0)
	bulkUpdate = &newList

	var jsonStr, err = json.Marshal(*updateList)
	if err != nil {
		log.Fatal("Cannot marshal bulk update:")
	}
	updateChannel <- jsonStr
}

func startBulkConsumer(jsonChannel chan []byte) {
	go func() {
		for {
			jsonStr := <-jsonChannel
			log.Println(len(jsonChannel), " updates buffered")
			continue
			buffer := bytes.NewBuffer(jsonStr)
			req, err := http.NewRequest("POST", bulkUrl, buffer)
			//req.Header.Set("X-Custom-Header", "myvalue")
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Print("Cannot connect to backend")
			} else {
				log.Println(resp.Status)
				err := resp.Body.Close()
				if err != nil {
					log.Println("cannot close http response, don't care")
				}
			}
		}
	}()
}
