export interface AQICategory {
  min: number;
  max: number;
  label: string;
  abbrv: string;
  color: string;
}

export function aqiCategory(aqi: number): AQICategory {
  if (aqi <= 50) {
    return {
      min: 0,
      max: 50,
      label: "Good",
      abbrv: "G",
      // Green; airnow.gov color: "rgb(0, 228, 0)"
      color: "green.7",
    };
  } else if (aqi <= 100) {
    return {
      min: 51,
      max: 100,
      label: "Moderate",
      abbrv: "M",
      // Yellow; airnow.gov color: "rgb(255, 255, 0)"
      color: "yellow.5",
    };
  } else if (aqi <= 150) {
    return {
      min: 101,
      max: 150,
      label: "Unhealthy for Sensitive Groups",
      abbrv: "USG",
      // Orange; airnow.gov color: "rgb(255, 126, 0)"
      color: "orange.7",
    };
  } else if (aqi <= 200) {
    return {
      min: 151,
      max: 200,
      label: "Unhealthy",
      abbrv: "U",
      // Red; airnow.gov color: "rgb(255, 0, 0)"
      color: "red.8",
    };
  } else if (aqi <= 300) {
    return {
      min: 201,
      max: 300,
      label: "Very Unhealthy",
      abbrv: "VU",
      // Purple; airnow.gov color: "rgb(143, 63, 151)"
      color: "violet.8",
    };
  } else {
    return {
      min: 301,
      max: 500,
      label: "Hazardous",
      abbrv: "H",
      // Maroon; airnow.gov color: "rgb(126, 0, 35)"
      color: "grape.9",
    };
  }
}
