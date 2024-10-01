package helper

import (
	"codeberg.org/Yonle/go-wrsbmkg"
	"html"
	"regexp"
	"strconv"
	"strings"
)

var htmlBrRegExp = regexp.MustCompile(`<br>`)
var htmlRegExp = regexp.MustCompile(`<[^>]*>`)

type Alert struct {
	Identifier string // Indonesia-InaTEWS-YYYYMMDDHHMMSS
	EventID    string // YYYYMMDDHHMMSS
	Date       string
	Time       string

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

	// Properti-properti dibawah ini hanya tersedia saat Tsunami
	WZMap        string
	TTMap        string
	SSHMap       string
	Instruction1 string
	Instruction2 string
	Instruction3 string
	WZAreas      []wrsbmkg.WZArea
}

type Realtime struct {
	Place       string
	Time        string
	Magnitude   float64
	Depth       float64
	Coordinates []any // ada yang berbeda disini.
	Phase       int
	Status      string
}

func ParseGempa(g *wrsbmkg.Raw_DataGempa) *Alert {
	i := g.Info
	coordinates := strings.Split(i.Point.Coordinates, ",")

	var newCoordinates []float64
	for _, co := range coordinates {
		f, _ := strconv.ParseFloat(co, 64)
		newCoordinates = append(newCoordinates, f)
	}

	magnitude, _ := strconv.ParseFloat(i.Magnitude, 64)

	return &Alert{
		Identifier: g.Identifier,
		EventID:    i.EventID,
		Date:       i.Date,
		Time:       i.Time,

		Subject:     i.Subject,
		Description: i.Description,
		Headline:    i.Headline,
		Area:        i.Area,
		Potential:   i.Potential,
		Instruction: i.Instruction,

		Coordinates: newCoordinates,
		Magnitude:   magnitude,
		Depth:       i.Depth,
		Shakemap:    i.Shakemap,
		Felt:        i.Felt,

		WZMap:        i.WZMap,
		TTMap:        i.TTMap,
		SSHMap:       i.SSHMap,
		Instruction1: i.Instruction1,
		Instruction2: i.Instruction2,
		Instruction3: i.Instruction3,
		WZAreas:      i.WZArea,
	}
}

func parseRealtimeProperty(f *wrsbmkg.Raw_QL_Feature) *Realtime {
	p := f.Properties
	g := f.Geometry

	magnitude, _ := strconv.ParseFloat(p.Mag, 64)
	depth, _ := strconv.ParseFloat(p.Depth, 64)
	phase, _ := strconv.Atoi(p.Fase)

	return &Realtime{
		Place:       p.Place,
		Time:        p.Time,
		Magnitude:   magnitude,
		Depth:       depth,
		Coordinates: g.Coordinates,
		Phase:       phase,
		Status:      p.Status,
	}
}

func ParseRealtime(r *wrsbmkg.Raw_QL) *Realtime {
	f := r.Features[0]
	return parseRealtimeProperty(&f)
}

func ParseRiwayatGempa(r *wrsbmkg.Raw_QL) []*Realtime {
	var history []*Realtime
	for _, f := range r.Features {
		parsed := parseRealtimeProperty(&f)
		history = append(history, parsed)
	}

	return history
}

// Membersihkan elemen-elemen HTML dari teks narasi.
func CleanNarasi(narasi string) string {
	lined := htmlBrRegExp.ReplaceAllString(narasi, "\n")
	cleaned := htmlRegExp.ReplaceAllString(lined, "")
	unescaped := html.UnescapeString(cleaned)

	return unescaped
}
