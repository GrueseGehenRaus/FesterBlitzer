package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
)

const R = 6371.0 // Earth's radius in km

var currPos = [2]float64{49.0161, 8.3980}

// Function to calculate distance in km
func getDist(start, end [2]float64) float64 {
	dLat := (end[0] - start[0]) * math.Pi / 180
	dLon := (end[1] - start[1]) * math.Pi / 180
	lat1 := start[0] * math.Pi / 180
	lat2 := end[0] * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c // Distance in km
}

// Function to calculate direction in degrees
func getDirection(coord1, coord2 [2]float64) float64 {
	lat1, lon1 := radians(coord1[0]), radians(coord1[1])
	lat2, lon2 := radians(coord2[0]), radians(coord2[1])
	deltaLon := lon2 - lon1

	x := math.Sin(deltaLon) * math.Cos(lat2)
	y := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(deltaLon)
	bearing := math.Atan2(x, y)

	return math.Mod(degrees(bearing)+360, 360)
}

// Helper functions for degree-radian conversion
func radians(deg float64) float64 { return deg * math.Pi / 180 }
func degrees(rad float64) float64 { return rad * 180 / math.Pi }

// Main function
func main() {
	direction := getDirection(currPos, [2]float64{49.018955, 8.397428})
	fmt.Printf("Direction: %.2f degrees\n", direction)

	url := fmt.Sprintf("https://cdn2.atudo.net/api/4.0/pois.php?type=22,26,20,101,102,103,104,105,106,107,108,109,110,111,112,113,115,117,114,ts,0,1,2,3,4,5,6,21,23,24,25,29,vwd,traffic&z=17&box=%f,%f,%f,%f",
		currPos[0]-0.015, currPos[1]-0.015, currPos[0]+0.015, currPos[1]+0.015)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			fmt.Println("Error decoding JSON:", err)
			return
		}
		pois := data["pois"].([]interface{})
		if len(pois) == 0 {
			fmt.Println("No Blitzer found")
			return
		}

		for _, blitzer := range pois {
			b := blitzer.(map[string]interface{})
			if vmax, ok := b["vmax"].(float64); ok {
				address := b["address"].(map[string]interface{})
				lat := b["lat"].(string)
				lng := b["lng"].(string)
				distance := getDist(currPos, [2]float64{parseFloat(lat), parseFloat(lng)})
				fmt.Printf("%.0f limit in %s %s, %.2f km away\n", vmax, address["city"], address["street"], distance)
			}
		}
	} else {
		fmt.Println("Failed to fetch data:", resp.Status)
	}
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
