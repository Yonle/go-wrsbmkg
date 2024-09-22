package wrsbmkg

// Struct ini bisa disimplikasi dengan [codeberg.org/Yonle/go-wrsbmkg/helper].
type Raw_DataGempa struct {
	Code       string `json:"code"`
	Identifier string `json:"identifier"`
	Info       struct {
		Area        string `json:"area"`
		Date        string `json:"date"`
		Depth       string `json:"depth"`
		Description string `json:"description"`
		Event       string `json:"event"`
		EventID     string `json:"eventid"`
		Felt        string `json:"felt"`
		Headline    string `json:"headline"`
		Instruction string `json:"instruction"`
		Latitude    string `json:"latitude"`
		Longitude   string `json:"longitude"`
		Magnitude   string `json:"magnitude"`
		Point       struct {
			Coordinates string `json:"coordinates"`
		} `json:"point"`
		Potential string `json:"potential"`
		Shakemap  string `json:"shakemap"`
		Subject   string `json:"subject"`
		Time      string `json:"time"`
		Timesent  string `json:"timesent"`
	} `json:"info"`
	MsgType string `json:"msgType"`
	Scope   string `json:"scope"`
	Sender  string `json:"sender"`
	Sent    string `json:"sent"`
	Status  string `json:"status"`
}

// Struct ini bisa disimplikasi dengan [codeberg.org/Yonle/go-wrsbmkg/helper].
type Raw_QL struct {
	Features []Raw_QL_Feature `json:"features"`
	Type     string           `json:"type"`
}

type Raw_QL_Feature struct {
	Geometry struct {
		Coordinates []any  `json:"coordinates"`
		Type        string `json:"type"`
	} `json:"geometry"`
	Properties struct {
		Depth  string `json:"depth"`
		Fase   string `json:"fase"`
		ID     string `json:"id"`
		Mag    string `json:"mag"`
		Place  string `json:"place"`
		Status string `json:"status"`
		Time   string `json:"time"`
	} `json:"properties"`
	Type string `json:"type"`
}
