import type { JSX } from "react";
import { useMemo } from "react";
import { Center, Loader, Table, Text } from "@mantine/core";
import type { LatestQuery } from "../types/__generated__/graphql";
import { AQIBadge } from "./aqibadge";
import { METRICS, METRIC_ORDER } from "../lib/metrics";
import type { MetricKey } from "../lib/metrics";
import { useRelativeTime } from "../lib/time";

function DeviceCell({
  deviceId,
  timestamp,
}: {
  deviceId: string;
  timestamp: string;
}): JSX.Element {
  const relativeTime = useRelativeTime(timestamp);
  return (
    <Table.Td>
      <Text size="sm" fw={500}>
        {deviceId}
      </Text>
      <Text size="xs" c="dimmed">
        {relativeTime}
      </Text>
    </Table.Td>
  );
}

function MetricCellContent({
  metric,
  measurement,
}: {
  metric: MetricKey;
  measurement: LatestMeasurement;
}): JSX.Element {
  if (measurement[metric] == null) {
    return (
      <Text size="sm" c="dimmed">
        —
      </Text>
    );
  }

  if (metric === "aqi") {
    return <AQIBadge aqi={measurement[metric] as number} />;
  }

  return <Text size="sm">{(measurement[metric] as number).toFixed(2)}</Text>;
}

function MetricCell({
  metric,
  measurement,
}: {
  metric: MetricKey;
  measurement: LatestMeasurement;
}): JSX.Element {
  return (
    <Table.Td key={metric}>
      <MetricCellContent metric={metric} measurement={measurement} />
    </Table.Td>
  );
}

type LatestMeasurement = LatestQuery["latest"][number];

interface LatestTableProps {
  measurements: LatestMeasurement[];
  loading?: boolean;
}

export function LatestMeasurementsTable({
  measurements,
  loading,
}: LatestTableProps): JSX.Element {
  const rows = useMemo(
    () =>
      [...measurements].sort((a, b) => a.deviceId.localeCompare(b.deviceId)),
    [measurements],
  );

  const presentMetrics = useMemo(
    () => METRIC_ORDER.filter((m) => measurements.some((r) => r[m] != null)),
    [measurements],
  );

  return (
    <div style={{ position: "relative" }}>
      <Table striped highlightOnHover>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Device</Table.Th>
            {presentMetrics.map((m) => (
              <Table.Th key={m}>
                {METRICS[m].label}
                {METRICS[m].unit && (
                  <Text span c="dimmed" size="xs">
                    {" "}
                    ({METRICS[m].unit})
                  </Text>
                )}
              </Table.Th>
            ))}
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {rows.map((row) => (
            <Table.Tr key={row.deviceId}>
              <DeviceCell deviceId={row.deviceId} timestamp={row.timestamp} />
              {presentMetrics.map((m) => (
                <MetricCell metric={m} measurement={row} />
              ))}
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
      {loading && (
        <Center
          style={{
            position: "absolute",
            inset: 0,
            backgroundColor:
              "color-mix(in srgb, var(--mantine-color-body) 60%, transparent)",
          }}
        >
          <Loader size="sm" />
        </Center>
      )}
    </div>
  );
}
