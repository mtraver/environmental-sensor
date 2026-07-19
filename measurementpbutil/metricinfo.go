package measurementpbutil

type MetricInfo struct {
	// The metric's name, e.g. "PM2.5"
	Name string

	// The metric's unit, e.g. "μg/m³".
	Unit string
}

var Metrics = map[string]MetricInfo{
	"temp": {
		Name: "temp",
		Unit: "°C",
	},
	"pm25": {
		Name: "PM2.5",
		Unit: "μg/m³",
	},
	"pm10": {
		Name: "PM10",
		Unit: "μg/m³",
	},
	"rh": {
		Name: "RH",
		Unit: "%",
	},
}
