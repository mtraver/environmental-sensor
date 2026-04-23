import type { JSX } from "react";
import { useMemo, useState } from "react";
import { gql } from "@apollo/client";
import type { TypedDocumentNode } from "@apollo/client";
import { useQuery } from "@apollo/client/react";
import { Card, Group, Stack, Text } from "@mantine/core";
import type {
  GetMeasurementsQuery,
  LatestQuery,
} from "../types/__generated__/graphql.ts";
import { LatestMeasurementsTable } from "../components/latestmeasurementstable";
import { MeasurementChart } from "../components/measurementchart";
import type { TimeRange } from "../components/timerangeselector";
import { TimeRangeSelector } from "../components/timerangeselector";

const GET_MEASUREMENTS: TypedDocumentNode<GetMeasurementsQuery> = gql`
  query GetMeasurements($startTime: DateTime!, $endTime: DateTime) {
    measurements(startTime: $startTime, endTime: $endTime) {
      deviceId
      timestamp
      uploadTimestamp
      temp
      pm25
      pm10
      rh
      aqi
    }
  }
`;

const LATEST: TypedDocumentNode<LatestQuery> = gql`
  query Latest {
    latest {
      deviceId
      timestamp
      uploadTimestamp
      temp
      pm25
      pm10
      rh
      aqi
    }
  }
`;

const DEFAULT_TIME_RANGE: TimeRange = {
  type: "relative",
  ms: 12 * 60 * 60 * 1000,
  label: "12h",
};

export default function Index(): JSX.Element {
  const [timeRange, setTimeRange] = useState<TimeRange>(DEFAULT_TIME_RANGE);

  const queryParams = useMemo(() => {
    if (timeRange.type === "absolute") {
      return {
        startTime: new Date(timeRange.from),
        endTime: new Date(timeRange.to),
      };
    } else {
      return { startTime: new Date(Date.now() - timeRange.ms) };
    }
  }, [timeRange]);

  // Get chart data.
  const { error, loading, data, previousData } = useQuery(GET_MEASUREMENTS, {
    variables: queryParams,
  });
  const displayData = data ?? previousData;

  // Get latest measurement data.
  const {
    error: latestError,
    loading: latestLoading,
    data: latestData,
  } = useQuery(LATEST);

  return (
    <>
      <Card withBorder radius="md" p="md">
        <Stack gap="md">
          <Group>
            <TimeRangeSelector value={timeRange} onChange={setTimeRange} />
          </Group>

          {error ? (
            <Text c="red">Error: {error.message}</Text>
          ) : (
            <MeasurementChart
              measurements={displayData?.measurements ?? []}
              loading={loading}
            />
          )}
        </Stack>
      </Card>

      <Card withBorder radius="md" p="md">
        <Stack gap="sm">
          <Text fw={600} mb="sm">
            Latest Readings
          </Text>

          {latestError ? (
            <Text c="red">Error: {latestError.message}</Text>
          ) : (
            <LatestMeasurementsTable
              measurements={latestData?.latest ?? []}
              loading={latestLoading}
            />
          )}
        </Stack>
      </Card>
    </>
  );
}
