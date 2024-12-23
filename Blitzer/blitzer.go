package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
)

type Blitzer struct {
	Vmax     int32
	City     string
	Street   string
	Distance float64
}

type BlitzerDEResponse struct {
	Pois []struct {
		ID      string `json:"id"`
		Lat     string `json:"lat"`
		Lng     string `json:"lng"`
		Address struct {
			Country      string `json:"country"`
			State        string `json:"state"`
			ZipCode      string `json:"zip_code"`
			City         string `json:"city"`
			CityDistrict string `json:"city_district"`
			Street       string `json:"street"`
		} `json:"address"`
		Content     string `json:"content"`
		Backend     string `json:"backend"`
		Type        string `json:"type"`
		Vmax        string `json:"vmax"`
		Counter     string `json:"counter"`
		CreateDate  string `json:"create_date"`
		ConfirmDate string `json:"confirm_date"`
		Info        struct {
			QltyCountryRoad int    `json:"qltyCountryRoad"`
			Confirmed       int    `json:"confirmed"`
			Gesperrt        int    `json:"gesperrt"`
			Quality         int    `json:"quality"`
			Label           string `json:"label"`
			Tags            []any  `json:"tags"`
			AlertType       int    `json:"alert_type"`
			Precheck        string `json:"precheck"`
			Desc            string `json:"desc"`
			Fixed           int    `json:"fixed"`
			Reason          string `json:"reason"`
			Length          int    `json:"length"`
			Duration        string `json:"duration"`
			LatEnd          string `json:"lat_end"`
			LngEnd          string `json:"lng_end"`
		} `json:"info,omitempty"`
		Polyline string `json:"polyline"`
		Style    int    `json:"style"`
	} `json:"pois"`
	Grid  []any `json:"grid"`
	Infos []any `json:"infos"`
}

type Point struct {
	x float64
	y float64
	z float64
}

func decode(resp *http.Response) BlitzerDEResponse {
	decoder := json.NewDecoder(resp.Body)
	var t BlitzerDEResponse
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	return t
}

func getBlitzer(blitzers BlitzerDEResponse, currPos [2]float64) []Blitzer {
	a := []Blitzer{}
	for _, blitzer := range blitzers.Pois {
		if blitzer.Vmax != "" {
			lat, _ := strconv.ParseFloat(blitzer.Lat, 64)
			lng, _ := strconv.ParseFloat(blitzer.Lng, 64)
			vmax, _ := strconv.ParseInt(blitzer.Vmax, 0, 32)			
			a = append(a, Blitzer{int32(vmax), blitzer.Address.City, blitzer.Address.Street, getDist(currPos, [2]float64{lat, lng})})
		}
	}
	return a
}

func getDist(start [2]float64, end [2]float64) float64 {
	R := 6371.0
	dlat := end[0] - start[0]
	dlon := end[1] - start[1]
	a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(end[0])*math.Cos(start[0])*math.Pow(math.Sin(dlon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c
	return distance / 100
}

func getDirection(coord1 [2]float64, coord2 [2]float64) float64 {
	lat1 := coord1[0] * math.Pi / 180
	lat2 := coord2[0] * math.Pi / 180
	dLon := (coord2[1] - coord1[1]) * math.Pi / 180
	y := math.Sin(dLon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLon)
	brng := math.Atan2(y, x) * 180 / math.Pi
	if brng < 0 {
		brng += 360
	}
	return brng
}

// SPECIAL THANKS TO TIM SIEFKEN (I588350)
func getScanBox(lastPos [2]float64, currPos [2]float64) [4][2]float64 {
	L1 := 1.0
	L2 := 0.4
	p0 := get3DPos(lastPos)
	p1 := get3DPos(currPos)

	p01 := Point{p1.x - p0.x, p1.y - p0.y, p1.z - p0.z}
	lenp01 := math.Sqrt(math.Pow(p01.x, 2) + math.Pow(p01.y, 2) + math.Pow(p01.z, 2))

	p2 := Point{p1.x + p01.x*L1/lenp01, p1.y + p01.y*L1/lenp01, p1.z + p01.z*L1/lenp01}

	v := Point{p1.y*p01.z - p1.z*p01.y, p1.z*p01.x - p1.x*p01.z, p1.x*p01.y - p1.y*p01.x}
	lenV := math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
	vNorm := Point{v.x / lenV, v.y / lenV, v.z / lenV}

	A := Point{p1.x + vNorm.x*L2, p1.y + vNorm.y*L2, p1.z + vNorm.z*L2}
	B := Point{p1.x - vNorm.x*L2, p1.y - vNorm.y*L2, p1.z - vNorm.z*L2}
	C := Point{p2.x + vNorm.x*L2, p2.y + vNorm.y*L2, p2.z + vNorm.z*L2}
	D := Point{p2.x - vNorm.x*L2, p2.y - vNorm.y*L2, p2.z - vNorm.z*L2}

	return [4][2]float64{get2DPos(A), get2DPos(B), get2DPos(C), get2DPos(D)}
}

// SPECIAL THANKS TO TIM SIEFKEN (I588350)
func getBoundingBox(points [4][2]float64) ([2]float64, [2]float64) {
	len_min := points[0][0]
	len_max := points[0][0]
	lat_min := points[0][1]
	lat_max := points[0][1]
	for _, point := range points {
		if point[0] < len_min {
			len_min = point[0]
		} else if point[0] > len_max {
			len_max = point[0]
		}
		if point[1] < lat_min {
			lat_min = point[1]
		} else if point[1] > lat_max {
			lat_max = point[1]
		}
	}
	return [2]float64{len_min, lat_min}, [2]float64{len_max, lat_max}
}

// SPECIAL THANKS TO TIM SIEFKEN (I588350)
func get3DPos(pos [2]float64) Point {
	R := 6371.00
	lat := pos[0] * math.Pi / 180.0
	lon := pos[1] * math.Pi / 180.0
	return Point{
		math.Cos(lat) * math.Cos(lon) * R,
		math.Cos(lat) * math.Sin(lon) * R,
		math.Sin(lat) * R,
	}
}

func get2DPos(point Point) [2]float64 {
	R := 6371.00
	lat := math.Asin(point.z / R) * 180.0 / math.Pi
    lon := math.Atan2(point.y, point.x) * 180.0 / math.Pi
    return [2]float64{lat, lon}}

func main() {
	lastPos := [2]float64{49.0161, 8.3980}
 	currPos := [2]float64{49.0189, 8.3974}
	//lastPos := [2]float64{49.01880678328532, 8.389688331453078}

	scanBox := getScanBox(lastPos, currPos)
	// print(scanBox[0][0], scanBox[0][1], scanBox[1][0], scanBox[1][1], scanBox[2][0], scanBox[2][1], scanBox[3][0], scanBox[3][1], "\n")
	boxStart, boxEnd := getBoundingBox(scanBox)

	url := fmt.Sprintf("https://cdn2.atudo.net/api/4.0/pois.php?type=22,26,20,101,102,103,104,105,106,107,108,109,110,111,112,113,115,117,114,ts,0,1,2,3,4,5,6,21,23,24,25,29,vwd,traffic&z=17&box=%f,%f,%f,%f",
		boxStart[0], boxStart[1], boxEnd[0], boxEnd[1])

	resp, err := http.Get(url)
	print(url, "\n")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	response := decode(resp)
	// still need to get distance with vectors instead of lat/long
	// also, need to find a way to store current Blitzers and compare them to the last ones && decide which to show in ui
	Blitzers := getBlitzer(response, currPos)
	for _, blitzer := range Blitzers {
		println(fmt.Sprintf("%d limit in %s %s in %fm", blitzer.Vmax, blitzer.Street, blitzer.City, blitzer.Distance))
	}
	// print(fmt.Sprintf("Direction: %.2f°\n", getDirection(currPos, lastPos)))
}
