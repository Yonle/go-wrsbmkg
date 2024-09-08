package helper

import (
	"codeberg.org/Yonle/go-wrsbmkg"
	"strconv"
	"strings"
)

type Alert struct {
	Subject     string
	Description string
	Headline    string
	Area        string
	Potential   string
	Instruction string

	Coordinates []float64
	Magnitude   float64
	Depth       string
	Shakemap    string
	Felt        string
}

type Realtime struct {
	Place       string
	Time        string
	Magnitude   float64
	Depth       float64
	Coordinates []interface{} // ada yang berbeda disini.
	Phase       int
	Status      string
}

func ParseGempa(g wrsbmkg.DataJSON) Alert {
	i := g["info"].(map[string]interface{})
	point := i["point"].(map[string]interface{})

	coordinates := strings.Split(point["coordinates"].(string), ",")
	var newCoordinates []float64
	for _, co := range coordinates {
		f, _ := strconv.ParseFloat(co, 64)
		newCoordinates = append(newCoordinates, f)
	}

	magnitude, _ := strconv.ParseFloat(i["magnitude"].(string), 64)

	return Alert{
		Subject:     i["subject"].(string),
		Description: i["description"].(string),
		Area:        i["area"].(string),
		Potential:   i["potential"].(string),
		Instruction: i["instruction"].(string),

		Coordinates: newCoordinates,
		Magnitude:   magnitude,
		Depth:       i["depth"].(string),
		Shakemap:    i["shakemap"].(string),
		Felt:        i["felt"].(string),
	}
}

func ParseRealtime(r wrsbmkg.DataJSON) Realtime {
	fc := r["features"].([]interface{})
	f := fc[0].(map[string]interface{})
	p := f["properties"].(map[string]interface{})
	g := f["geometry"].(map[string]interface{})

	magnitude, _ := strconv.ParseFloat(p["mag"].(string), 64)
	depth, _ := strconv.ParseFloat(p["depth"].(string), 64)
	phase, _ := strconv.Atoi(p["fase"].(string))

	return Realtime{
		Place:       p["place"].(string),
		Time:        p["time"].(string),
		Magnitude:   magnitude,
		Depth:       depth,
		Coordinates: g["coordinates"].([]interface{}),
		Phase:       phase,
		Status:      p["status"].(string),
	}
}
