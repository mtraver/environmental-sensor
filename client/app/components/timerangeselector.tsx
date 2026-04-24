import type { JSX } from "react";
import { useCallback, useEffect, useState } from "react";
import {
  Box,
  Divider,
  Group,
  Popover,
  ScrollArea,
  Stack,
  Text,
  TextInput,
  UnstyledButton,
  useMantineTheme,
} from "@mantine/core";
import { useMediaQuery } from "@mantine/hooks";
import { DateTimePicker } from "@mantine/dates";
import { Calendar } from "@phosphor-icons/react";
import { formatDate, parseRelativeTime } from "../lib/time";
import "@mantine/dates/styles.css";
import classes from "../css/UnstyledButton.module.css";

export type TimeRange =
  | { type: "relative"; ms: number; label: string }
  | { type: "absolute"; from: number; to: number };

const PRESET_GROUPS = [
  {
    label: "Minutes",
    presets: [
      { label: "Past 5 minutes", shortLabel: "5m", ms: 5 * 60_000 },
      { label: "Past 15 minutes", shortLabel: "15m", ms: 15 * 60_000 },
      { label: "Past 30 minutes", shortLabel: "30m", ms: 30 * 60_000 },
    ],
  },
  {
    label: "Hours",
    presets: [
      { label: "Past 1 hour", shortLabel: "1h", ms: 60 * 60_000 },
      { label: "Past 3 hours", shortLabel: "3h", ms: 3 * 60 * 60_000 },
      { label: "Past 6 hours", shortLabel: "6h", ms: 6 * 60 * 60_000 },
      { label: "Past 12 hours", shortLabel: "12h", ms: 12 * 60 * 60_000 },
    ],
  },
  {
    label: "Days",
    presets: [
      { label: "Past 1 day", shortLabel: "1d", ms: 24 * 60 * 60_000 },
      { label: "Past 2 days", shortLabel: "2d", ms: 2 * 24 * 60 * 60_000 },
      { label: "Past 7 days", shortLabel: "7d", ms: 7 * 24 * 60 * 60_000 },
    ],
  },
  {
    label: "Months",
    presets: [
      { label: "Past 1 month", shortLabel: "1mo", ms: 30 * 24 * 60 * 60_000 },
      { label: "Past 3 months", shortLabel: "3mo", ms: 90 * 24 * 60 * 60_000 },
    ],
  },
];

const DATE_TIME_PICKER_VALUE_FORMAT = "YYYY-MM-DD HH:mm";

function formatAbsoluteRange(from: number, to: number, narrow = false): string {
  return `${formatDate(from, narrow)} – ${formatDate(to, narrow)}`;
}

function rangeToShortLabel(range: TimeRange): string {
  if (range.type === "relative") return range.label;

  // Absolute ranges have no short label.
  return "";
}

function rangeToDisplayString(range: TimeRange, narrow = false): string {
  if (range.type === "relative") {
    const now = Date.now();
    return formatAbsoluteRange(now - range.ms, now, narrow);
  }

  return formatAbsoluteRange(range.from, range.to, narrow);
}

export function TimeRangeSelector({
  value,
  onChange,
}: {
  value: TimeRange;
  onChange: (range: TimeRange) => void;
}): JSX.Element {
  const theme = useMantineTheme();
  const isNarrow = useMediaQuery(`(max-width: ${theme.breakpoints.sm})`);

  const [open, setOpen] = useState(false);
  const [textInput, setTextInput] = useState("");
  const [textError, setTextError] = useState(false);
  const [fromDate, setFromDate] = useState<string | null>(null);
  const [toDate, setToDate] = useState<string | null>(null);

  const applyRelative = useCallback(
    (ms: number, shortLabel: string) => {
      onChange({ type: "relative", ms, label: shortLabel });
      setOpen(false);
    },
    [onChange],
  );

  const handleTextKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key !== "Enter") return;
      const ms = parseRelativeTime(textInput);
      if (ms === null) {
        setTextError(true);
        return;
      }
      applyRelative(ms, textInput.trim());
    },
    [textInput, applyRelative],
  );

  const handleTextChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setTextInput(e.currentTarget.value);
      setTextError(false);
    },
    [],
  );

  const handleFromChange = useCallback(
    (dateStr: string | null) => {
      setFromDate(dateStr);
      if (dateStr && toDate) {
        onChange({
          type: "absolute",
          from: new Date(dateStr).getTime(),
          to: new Date(toDate).getTime(),
        });
        setOpen(false);
      }
    },
    [toDate, onChange],
  );

  const handleToChange = useCallback(
    (dateStr: string | null) => {
      setToDate(dateStr);
      if (fromDate && dateStr) {
        onChange({
          type: "absolute",
          from: new Date(fromDate).getTime(),
          to: new Date(dateStr).getTime(),
        });
        setOpen(false);
      }
    },
    [fromDate, onChange],
  );

  const shortLabel = rangeToShortLabel(value);
  const displayString = rangeToDisplayString(value, isNarrow);

  // Update fromDate and toDate when value changes.
  useEffect(() => {
    if (value.type === "absolute") {
      setFromDate(new Date(value.from).toISOString());
      setToDate(new Date(value.to).toISOString());
    } else {
      setFromDate(null);
      setToDate(null);
    }
  }, [value]);

  return (
    <Popover
      opened={open}
      onDismiss={() => {
        setOpen(false);

        // Clear the relative time input when the popover closes.
        setTextInput("");
        setTextError(false);
      }}
      position="bottom-start"
      width={520}
      withArrow={false}
      shadow="md"
      trapFocus
    >
      <Popover.Target>
        {/* The trigger "field" */}
        <UnstyledButton
          onClick={() => setOpen((o) => !o)}
          style={{
            display: "inline-flex",
            alignItems: "center",
            border: "1px solid var(--mantine-color-default-border)",
            borderRadius: "var(--mantine-radius-sm)",
            backgroundColor: "var(--mantine-color-default)",
            height: 36,
            cursor: "pointer",
            overflow: "hidden",
          }}
        >
          {/* Left badge: short label */}
          {shortLabel && (
            <Text
              size="sm"
              fw={600}
              px="sm"
              style={{
                borderRight: "1px solid var(--mantine-color-default-border)",
                height: "100%",
                display: "flex",
                alignItems: "center",
                backgroundColor: "var(--mantine-color-default-hover)",
                whiteSpace: "nowrap",
              }}
            >
              {shortLabel}
            </Text>
          )}
          {/* Right: resolved date range */}
          <Group gap="xs" px="sm">
            <Calendar
              size={14}
              style={{ color: "var(--mantine-color-dimmed)" }}
            />
            <Text size="sm" style={{ whiteSpace: "nowrap" }}>
              {displayString}
            </Text>
          </Group>
        </UnstyledButton>
      </Popover.Target>

      <Popover.Dropdown p={0}>
        <Group align="flex-start" gap={0} style={{ height: 340 }}>
          {/* Left panel: presets + text input */}
          <Stack
            gap={0}
            style={{
              width: 200,
              borderRight: "1px solid var(--mantine-color-default-border)",
              height: "100%",
            }}
          >
            {/* Text input for custom relative range */}
            <Box p="xs">
              <TextInput
                size="xs"
                placeholder="e.g. 3h, 6h30m, 2d"
                value={textInput}
                onChange={handleTextChange}
                onKeyDown={handleTextKeyDown}
                error={textError ? "Invalid format" : undefined}
                data-autofocus
              />
            </Box>

            <Divider />

            {/* Preset groups */}
            <ScrollArea style={{ flex: 1 }}>
              {PRESET_GROUPS.map((group) => (
                <Stack key={group.label} gap={0}>
                  <Text
                    size="xs"
                    fw={600}
                    c="dimmed"
                    px="sm"
                    pt="xs"
                    pb={4}
                    style={{
                      textTransform: "uppercase",
                      letterSpacing: "0.05em",
                    }}
                  >
                    {group.label}
                  </Text>
                  {group.presets.map((preset) => (
                    <UnstyledButton
                      key={preset.shortLabel}
                      className={classes.presetButton}
                      px="sm"
                      py={6}
                      onClick={() =>
                        applyRelative(preset.ms, preset.shortLabel)
                      }
                      style={(theme) => ({
                        fontSize: theme.fontSizes.sm,
                        backgroundColor:
                          value.type === "relative" && value.ms === preset.ms
                            ? "var(--mantine-color-blue-light)"
                            : undefined,
                      })}
                    >
                      {preset.label}
                    </UnstyledButton>
                  ))}
                </Stack>
              ))}
            </ScrollArea>
          </Stack>

          {/* Right panel: absolute date pickers */}
          <Stack gap="sm" p="md" style={{ flex: 1 }}>
            <Text
              size="xs"
              fw={600}
              c="dimmed"
              style={{ textTransform: "uppercase", letterSpacing: "0.05em" }}
            >
              Custom Range
            </Text>
            <DateTimePicker
              size="xs"
              label="From"
              placeholder="Start date & time"
              value={fromDate}
              valueFormat={DATE_TIME_PICKER_VALUE_FORMAT}
              maxDate={toDate ?? undefined}
              onChange={handleFromChange}
              popoverProps={{ withinPortal: false }}
              clearable
            />
            <DateTimePicker
              size="xs"
              label="To"
              placeholder="End date & time"
              value={toDate}
              valueFormat={DATE_TIME_PICKER_VALUE_FORMAT}
              minDate={fromDate ?? undefined}
              onChange={handleToChange}
              popoverProps={{ withinPortal: false }}
              clearable
            />
            {fromDate && !toDate && (
              <Text size="xs" c="dimmed">
                Pick an end date to apply
              </Text>
            )}
          </Stack>
        </Group>
      </Popover.Dropdown>
    </Popover>
  );
}
