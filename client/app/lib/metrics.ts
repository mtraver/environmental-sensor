export type MetricKey = "temp" | "rh" | "pm25" | "pm10" | "aqi";
export const METRIC_ORDER: MetricKey[] = ["temp", "rh", "aqi", "pm25", "pm10"];
export const DEFAULT_METRIC: MetricKey = "temp";

export const METRICS: Record<
  MetricKey,
  { label: string; shortLabel: string; unit: string }
> = {
  temp: { label: "Temperature", shortLabel: "Temp", unit: "°C" },
  rh: { label: "Relative Humidity", shortLabel: "RH", unit: "%" },
  pm25: { label: "PM2.5", shortLabel: "PM2.5", unit: "µg/m³" },
  pm10: { label: "PM10", shortLabel: "PM10", unit: "µg/m³" },
  aqi: { label: "AQI", shortLabel: "AQI", unit: "" },
};
