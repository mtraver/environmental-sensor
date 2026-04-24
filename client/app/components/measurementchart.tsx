import type { JSX } from "react";
import { useEffect, useMemo, useState } from "react";
import {
  Center,
  Checkbox,
  Group,
  Loader,
  SegmentedControl,
  Stack,
} from "@mantine/core";
import { LineChart } from "@mantine/charts";
import fnv1a from "@sindresorhus/fnv1a";
import type { Measurement } from "../types/__generated__/graphql";
import { DEFAULT_METRIC, METRICS, METRIC_ORDER } from "../lib/metrics";
import type { MetricKey } from "../lib/metrics";
import { formatDate } from "../lib/time";
import "@mantine/charts/styles.css";

type SeriesPoint = {
  timestamp: number;
  value: number;
};

type DeviceSeries = Record<MetricKey, SeriesPoint[]>;

type ChartDataByDevice = Record<string, DeviceSeries>;

type AxisDomainFn = (
  domain: readonly [number, number],
  allowDataOverflow: boolean,
) => readonly [number, number];

const Y_DOMAIN: Record<MetricKey, readonly number[] | AxisDomainFn> = {
  temp: ([min, max]) => [
    Math.floor(min * 2) / 2 - 0.5,
    Math.ceil(max * 2) / 2 + 0.5,
  ],
  rh: ([min, max]) => [
    Math.max(0, Math.floor(min * 2) / 2 - 1),
    Math.min(100, Math.ceil(max * 2) / 2 + 1),
  ],
  pm25: ([min, max]) => [Math.floor(min * 0.9), Math.ceil(max * 1.1)],
  pm10: ([min, max]) => [Math.floor(min * 0.9), Math.ceil(max * 1.1)],
  aqi: ([min, max]) => {
    const low = Math.max(0, Math.floor(min * 0.9));
    const high = Math.ceil(max * 1.1);

    // Ensure at least an 8-point span so the axis isn't super tight.
    const mid = (low + high) / 2;
    return high - low < 8
      ? [Math.max(0, Math.floor(mid - 4)), Math.ceil(mid + 4)]
      : [low, high];
  },
};

function normalizeMeasurements(measurements: Measurement[]): ChartDataByDevice {
  const result: ChartDataByDevice = {};

  for (const m of measurements) {
    const ts = new Date(m.timestamp).getTime();
    if (Number.isNaN(ts)) continue;

    if (!result[m.deviceId]) {
      result[m.deviceId] = {
        temp: [],
        rh: [],
        pm25: [],
        pm10: [],
        aqi: [],
      };
    }

    if (m.temp != null)
      result[m.deviceId].temp.push({ timestamp: ts, value: m.temp });
    if (m.rh != null)
      result[m.deviceId].rh.push({ timestamp: ts, value: m.rh });
    if (m.pm25 != null)
      result[m.deviceId].pm25.push({ timestamp: ts, value: m.pm25 });
    if (m.pm10 != null)
      result[m.deviceId].pm10.push({ timestamp: ts, value: m.pm10 });
    if (m.aqi != null)
      result[m.deviceId].aqi.push({ timestamp: ts, value: m.aqi });
  }

  return result;
}

// mergeMetricSeries merges per-device series into a single array of chart data points,
// grouping together timestamps that are within a given tolerance of each other as
// a single point. That allows very slight differences in timestamp to be ignored
// for the purpose of a cleaner plot. For example, devices might all be supposed to
// report data at 10:02, but the actual timestamps are 10:02:00, 10:02:03, and 10:02:01.
// All three of those belong to the logical 10:02:00 timestamp.
function mergeMetricSeries(
  dataByDevice: ChartDataByDevice,
  metric: MetricKey,
  presentDevices: string[],
  toleranceMs = 10_000,
) {
  const points: Array<{ timestamp: number; device: string; value: number }> =
    [];

  for (const device of presentDevices) {
    for (const p of dataByDevice[device]?.[metric] ?? []) {
      points.push({ timestamp: p.timestamp, device, value: p.value });
    }
  }

  if (points.length === 0) return [];

  points.sort((a, b) => a.timestamp - b.timestamp);

  // Assign each point to a cluster ID based on proximity to neighbors.
  const clusterIds = new Array<number>(points.length);
  clusterIds[0] = 0;
  let clusterId = 0;

  for (let i = 1; i < points.length; i++) {
    if (points[i].timestamp - points[i - 1].timestamp > toleranceMs) {
      clusterId++;
    }
    clusterIds[i] = clusterId;
  }

  // Group by cluster.
  const clusters = new Map<
    number,
    { sumTs: number; count: number; values: Record<string, number> }
  >();

  for (let i = 0; i < points.length; i++) {
    const cid = clusterIds[i];
    if (!clusters.has(cid)) {
      clusters.set(cid, { sumTs: 0, count: 0, values: {} });
    }
    const c = clusters.get(cid)!;
    c.sumTs += points[i].timestamp;
    c.count++;
    c.values[points[i].device] = points[i].value;
  }

  return Array.from(clusters.values())
    .map((c) => ({ timestamp: Math.round(c.sumTs / c.count), ...c.values }))
    .sort((a, b) => a.timestamp - b.timestamp);
}

const formatTimeTick = (value: number) => {
  const d = new Date(value);
  return d.toLocaleTimeString(undefined, {
    hour12: false,
    hour: "2-digit",
    minute: "2-digit",
  });
};

const DEVICE_COLORS: string[] = [
  "red",
  "pink",
  "grape",
  "violet",
  "indigo",
  "blue",
  "cyan",
  "teal",
  "green",
  "lime",
  "yellow",
  "orange",
];

function colorForDevice(id: string): string {
  const idx = Number(fnv1a(id, { size: 32 })) % DEVICE_COLORS.length;
  return DEVICE_COLORS[idx];
}

function useMeasurements(measurements: Measurement[]) {
  return useMemo(() => normalizeMeasurements(measurements), [measurements]);
}

export function MeasurementChart({
  measurements,
  loading,
}: {
  measurements: Measurement[];
  loading?: boolean;
}): JSX.Element {
  const [metric, setMetric] = useState<MetricKey>(DEFAULT_METRIC);

  const dataByDevice = useMeasurements(measurements);

  const devices = useMemo(
    () => Object.keys(dataByDevice).sort(),
    [dataByDevice],
  );

  const [presentDevices, setPresentDevices] = useState<string[]>([]);

  // Derive present devices from data.
  useEffect(() => {
    setPresentDevices((prev) => {
      if (prev.length === 0) return devices;
      return prev.filter((d) => devices.includes(d));
    });
  }, [devices]);

  // Derive present metrics from data.
  const presentMetrics = useMemo<MetricKey[]>(() => {
    const set = new Set<MetricKey>();

    for (const device of Object.values(dataByDevice)) {
      (Object.keys(device) as MetricKey[]).forEach((m) => {
        if (device[m].length > 0) set.add(m);
      });
    }

    return Array.from(set)
      .slice()
      .sort((a, b) => METRIC_ORDER.indexOf(a) - METRIC_ORDER.indexOf(b));
  }, [dataByDevice]);

  // Ensure the selected metric is a valid metric.
  const validatedMetric = presentMetrics.includes(metric)
    ? metric
    : (presentMetrics[0] ?? DEFAULT_METRIC);

  const chartData = useMemo(
    () => mergeMetricSeries(dataByDevice, validatedMetric, presentDevices),
    [dataByDevice, validatedMetric, presentDevices],
  );

  const series = useMemo(
    () =>
      presentDevices.map((device) => ({
        name: device,
        dataKey: device,
        color: colorForDevice(device),
      })),
    [presentDevices],
  );

  return (
    <Stack gap="md">
      {/* Metric selector */}
      <SegmentedControl
        value={validatedMetric}
        onChange={(v) => setMetric(v as MetricKey)}
        data={presentMetrics.map((k) => ({
          value: k,
          label: METRICS[k].label,
        }))}
        style={{ alignSelf: "flex-start" }}
      />

      {/* Device toggles */}
      <Group>
        {devices.map((device) => (
          <Checkbox
            key={device}
            label={device}
            checked={presentDevices.includes(device)}
            onChange={(e) =>
              setPresentDevices((prev) =>
                e.currentTarget.checked
                  ? [...prev, device]
                  : prev.filter((d) => d !== device),
              )
            }
          />
        ))}
      </Group>

      <div style={{ position: "relative" }}>
        <LineChart
          h={360}
          data={chartData}
          series={series}
          dataKey="timestamp"
          xAxisProps={{
            scale: "time",
            type: "number",
            domain: ["dataMin", "dataMax"],
            tickFormatter: formatTimeTick,
          }}
          yAxisProps={{
            domain: Y_DOMAIN[validatedMetric],
          }}
          yAxisLabel={METRICS[validatedMetric].unit}
          valueFormatter={(value) => value.toFixed(2)}
          strokeWidth={1}
          withDots={false}
          curveType="monotone"
          withTooltip
          tooltipProps={{
            labelFormatter: (value) =>
              value != null ? formatDate(value as number) : "",
            cursor: { strokeDasharray: "3 3" },
          }}
          withLegend
          legendProps={{
            verticalAlign: "bottom",
          }}
        />

        {loading && (
          <Center
            style={{
              position: "absolute",
              inset: 0,
              backgroundColor:
                "color-mix(in srgb, var(--mantine-color-body) 60%, transparent)",
              borderRadius: "var(--mantine-radius-md)",
              animation: "fadeIn 300ms ease",
            }}
          >
            <Loader />
          </Center>
        )}
      </div>
    </Stack>
  );
}
