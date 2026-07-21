package metric

type Key string

const (
	Temp     Key = "temp"
	PM1      Key = "pm1"
	PM25     Key = "pm25"
	PM4      Key = "pm4"
	PM10     Key = "pm10"
	RH       Key = "rh"
	VOCIndex Key = "vocIndex"
	NOxIndex Key = "noxIndex"
	HCHO     Key = "hcho"
	CO2      Key = "co2"
	AQI      Key = "aqi"
)

type Info struct {
	// The metric's name, e.g. "PM2.5"
	Name string

	// The metric's unit, e.g. "μg/m³".
	Unit string
}

var All = map[Key]Info{
	Temp: {
		Name: "temp",
		Unit: "°C",
	},
	PM1: {
		Name: "PM1.0",
		Unit: "μg/m³",
	},
	PM25: {
		Name: "PM2.5",
		Unit: "μg/m³",
	},
	PM4: {
		Name: "PM4",
		Unit: "μg/m³",
	},
	PM10: {
		Name: "PM10",
		Unit: "μg/m³",
	},
	RH: {
		Name: "RH",
		Unit: "%",
	},
	VOCIndex: {
		Name: "VOCIndex",
		Unit: "",
	},
	NOxIndex: {
		Name: "NOₓIndex",
		Unit: "",
	},
	HCHO: {
		Name: "HCHO",
		Unit: "ppb",
	},
	CO2: {
		Name: "CO₂",
		Unit: "ppm",
	},
}
