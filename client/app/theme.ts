import {
  createTheme,
  darken,
  defaultVariantColorsResolver,
  parseThemeColor,
  rgba,
} from "@mantine/core";
import type { VariantColorsResolver } from "@mantine/core";

const variantColorResolver: VariantColorsResolver = (input) => {
  const defaultResolvedColors = defaultVariantColorsResolver(input);
  const parsedColor = parseThemeColor({
    color: input.color || input.theme.primaryColor,
    theme: input.theme,
  });

  if (input.variant === "light-with-border") {
    return {
      background: rgba(parsedColor.value, 0.1),
      hover: rgba(parsedColor.value, 0.15),
      border: `1px solid ${parsedColor.value}`,
      color: darken(parsedColor.value, 0.1),
    };
  }

  return defaultResolvedColors;
};

export const theme = createTheme({
  fontFamily: "Open Sans, sans-serif",
  variantColorResolver: variantColorResolver,
});
