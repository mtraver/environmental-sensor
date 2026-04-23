import type { JSX } from "react";
import { Badge } from "@mantine/core";
import { aqiCategory } from "../lib/aqi";

export function AQIBadge({ aqi }: { aqi: number }): JSX.Element {
  const aqiCat = aqiCategory(aqi);
  return (
    <Badge color={aqiCat.color} variant="light-with-border">
      {aqi} ({aqiCat.label})
    </Badge>
  );
}
