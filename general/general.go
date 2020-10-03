package general

var Version string
var BuildTime string

type Versions struct {
	Version   string `json:"version"`
	BuildTime string `json:"build-time"`
}
