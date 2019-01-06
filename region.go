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

type UpdateChan struct {
	updates chan update
	exited  chan bool
}

type NetworkStats struct {
	totalUpdates   int
	runningUpdates chan bool
	queuedUpdates  chan int
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

	if s.BulkSend {
		newList := make([]update, 0)
		bulkUpdate = &newList
		updateChannel = make(chan []byte, 50)
		if s.SendUpdates {
			startBulkConsumer(updateChannel)
		} else {
			startVoidBulkConsumer(updateChannel)
		}
	}

	networkStats.runningUpdates = make(chan bool, 10)
	networkStats.queuedUpdates = make(chan int, 10)
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

func UpdateRegions(regions *[]Region, individual *Individual, time time.Time, bulk bool) {
	if !individual.UpdateSender {
		return
	}
	x, y := individual.Loc.GetLatestXY()
	for _, r := range *regions {
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
				networkStats.totalUpdates++
				networkStats.queuedUpdates <- 1
				if bulk {
					*bulkUpdate = append(*bulkUpdate, u)
				} else {
					individual.UpdateChan.updates <- u
				}

			}
		} else {
			// this individual is not in this region
			//_, knownInside := individual.RegionIds[r.ID]
			if individual.RegionIds[r.ID] {
				individual.RegionIds[r.ID] = false
				// we must send update to backend to say this individual is no longer in the region.
				u := update{EventID: r.EventID, RegionID: r.ID, UUID: individual.UUID, Entering: false, OccurredAt: time.Unix()}
				networkStats.totalUpdates++
				networkStats.queuedUpdates <- 1
				if bulk {
					*bulkUpdate = append(*bulkUpdate, u)
				} else {
					individual.UpdateChan.updates <- u
				}
			}
		}

	}
}

func LeaveAllRegions(state *State, individual *Individual, time time.Time, bulk bool) {
	if !individual.UpdateSender {
		return
	}
	for rID, b := range individual.RegionIds {
		if b {
			r := state.FindRegion(rID)
			u := update{EventID: r.EventID, RegionID: r.ID, UUID: individual.UUID, Entering: false, OccurredAt: time.Unix()}
			networkStats.totalUpdates++
			networkStats.queuedUpdates <- 1
			if bulk {
				*bulkUpdate = append(*bulkUpdate, u)
			} else {
				individual.UpdateChan.updates <- u
			}
		}
	}
	individual.UpdateChan.exited <- true
}

func PersonalUpdateDemon(sender *UpdateChan, send bool) {
	for {
		select {
		case u := <-sender.updates:
			handleUpdate(send, &u)
		case <-sender.exited:
			for len(sender.updates) > 0 {
				u := <-sender.updates
				handleUpdate(send, &u)
			}
			return
		}
	}
}
func handleUpdate(send bool, u *update) {
	networkStats.queuedUpdates <- -1
	networkStats.runningUpdates <- true
	if send {
		sendUpdate(u)
	} else {
		time.Sleep(1000 * time.Millisecond)
	}
	networkStats.runningUpdates <- false
}

const url = "http://api.jackchorley.club/update"

func sendUpdate(u *update) {
	var jsonStr, err = json.Marshal(*u)
	if err != nil {
		log.Fatal("Cannot marshal update: ", *u)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Print("Cannot endcode update to json", err)
	}
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("Cannot connect to backend", err)
	} else {
		if resp.StatusCode != http.StatusOK {
			log.Println("Error sending update: ", resp.Status)
		}
		err := resp.Body.Close()
		if err != nil {
			log.Println("cannot close http response, don't care")
		}
	}
}

const bulkUrl = "http://api.jackchorley.club/bulkUpdate"

var bulkUpdate *[]update
var updateChannel chan []byte
var networkStats NetworkStats

func GetTotalUpdates() int {
	return networkStats.totalUpdates
}

func SendBulk() {
	if bulkUpdate == nil {
		log.Println("total updates so far:", networkStats.totalUpdates)
		return
	}
	if len(*bulkUpdate) < 10 {
		return
	}
	networkStats.queuedUpdates <- -len(*bulkUpdate)
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
			networkStats.runningUpdates <- true
			log.Println(len(jsonChannel), " updates buffered")
			buffer := bytes.NewBuffer(jsonStr)
			req, err := http.NewRequest("POST", bulkUrl, buffer)
			//req.Header.Set("X-Custom-Header", "myvalue")
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Print("Cannot connect to backend")
			} else {
				if resp.StatusCode != http.StatusOK {
					log.Println("Error sending update: ", resp.Status)
				}
				err := resp.Body.Close()
				if err != nil {
					log.Println("cannot close http response, don't care")
				}
			}
			networkStats.runningUpdates <- false
		}
	}()
}

func startVoidBulkConsumer(jsonChannel chan []byte) {
	go func() {
		for {
			<-jsonChannel
			log.Println("Bulk update Voided")
		}
	}()
}
