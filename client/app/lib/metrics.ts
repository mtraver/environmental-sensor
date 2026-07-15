export const METRIC_ORDER = [
  "temp",
  "rh",
  "aqi",
  "pm25",
  "pm10",
] as const satisfies readonly string[];
export type MetricKey = (typeof METRIC_ORDER)[number];

export const DEFAULT_METRIC: MetricKey = "temp";

type CellRenderer = "number" | "aqiBadge";

type AxisDomainFn = (
  domain: readonly [number, number],
  allowDataOverflow: boolean,
) => readonly [number, number];

type YDomain = readonly [number, number] | AxisDomainFn;

export interface MetricConfig {
  label: string;
  shortLabel: string;
  unit: string;
  decimals: number;
  yDomain: YDomain;
  // Optional custom cell renderer for the latest readings table.
  // If omitted, falls back to a plain number.
  cellRenderer?: CellRenderer;
}

export const METRICS: Record<MetricKey, MetricConfig> = {
  temp: {
    label: "Temperature",
    shortLabel: "Temp",
    unit: "°C",
    decimals: 2,
    yDomain: ([min, max]) => [
      Math.floor(min * 2) / 2 - 0.5,
      Math.ceil(max * 2) / 2 + 0.5,
    ],
  },
  rh: {
    label: "Relative Humidity",
    shortLabel: "RH",
    unit: "%",
    decimals: 2,
    yDomain: ([min, max]) => [
      Math.max(0, Math.floor(min * 2) / 2 - 1),
      Math.min(100, Math.ceil(max * 2) / 2 + 1),
    ],
  },
  pm25: {
    label: "PM2.5",
    shortLabel: "PM2.5",
    unit: "µg/m³",
    decimals: 2,
    yDomain: ([min, max]) => [Math.floor(min * 0.9), Math.ceil(max * 1.1)],
  },
  pm10: {
    label: "PM10",
    shortLabel: "PM10",
    unit: "µg/m³",
    decimals: 2,
    yDomain: ([min, max]) => [Math.floor(min * 0.9), Math.ceil(max * 1.1)],
  },
  aqi: {
    label: "AQI",
    shortLabel: "AQI",
    unit: "",
    decimals: 0,
    yDomain: ([min, max]) => {
      const low = Math.max(0, Math.floor(min * 0.9));
      const high = Math.ceil(max * 1.1);

      // Ensure at least an 8-point span so the axis isn't super tight.
      const mid = (low + high) / 2;
      return high - low < 8
        ? [Math.max(0, Math.floor(mid - 4)), Math.ceil(mid + 4)]
        : [low, high];
    },
    cellRenderer: "aqiBadge",
  },
};
