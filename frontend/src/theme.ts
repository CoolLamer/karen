import { createTheme, MantineColorsTuple } from "@mantine/core";

// Teal - Primary color (trust, tech, calm)
const teal: MantineColorsTuple = [
  "#E6FFFA",
  "#B2F5EA",
  "#81E6D9",
  "#4FD1C5",
  "#38B2AC",
  "#319795",
  "#2C7A7B",
  "#285E61",
  "#234E52",
  "#1D4044",
];

// Purple - Secondary color (accents, CTAs)
const purple: MantineColorsTuple = [
  "#FAF5FF",
  "#E9D8FD",
  "#D6BCFA",
  "#B794F4",
  "#9F7AEA",
  "#805AD5",
  "#6B46C1",
  "#553C9A",
  "#44337A",
  "#322659",
];

export const zvednuTheme = createTheme({
  // Colors
  colors: {
    teal,
    purple,
  },
  primaryColor: "teal",
  primaryShade: 5,

  // Typography
  fontFamily:
    '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
  headings: {
    fontFamily: "inherit",
    fontWeight: "600",
  },

  // Border radius - slightly more rounded for modern feel
  radius: {
    xs: "4px",
    sm: "6px",
    md: "8px",
    lg: "12px",
    xl: "16px",
  },
  defaultRadius: "md",

  // Shadows - softer, more modern
  shadows: {
    xs: "0 1px 2px rgba(0, 0, 0, 0.05)",
    sm: "0 1px 3px rgba(0, 0, 0, 0.1), 0 1px 2px rgba(0, 0, 0, 0.06)",
    md: "0 4px 6px rgba(0, 0, 0, 0.1), 0 2px 4px rgba(0, 0, 0, 0.06)",
    lg: "0 10px 15px rgba(0, 0, 0, 0.1), 0 4px 6px rgba(0, 0, 0, 0.05)",
    xl: "0 20px 25px rgba(0, 0, 0, 0.1), 0 10px 10px rgba(0, 0, 0, 0.04)",
  },

  // Component defaults
  components: {
    Button: {
      defaultProps: {
        radius: "md",
      },
      styles: {
        root: {
          fontWeight: 500,
        },
      },
    },
    Paper: {
      defaultProps: {
        shadow: "sm",
        radius: "md",
      },
    },
    Card: {
      defaultProps: {
        shadow: "sm",
        radius: "md",
      },
    },
    Badge: {
      defaultProps: {
        radius: "sm",
      },
    },
    TextInput: {
      defaultProps: {
        radius: "md",
      },
    },
    Textarea: {
      defaultProps: {
        radius: "md",
      },
    },
    Alert: {
      defaultProps: {
        radius: "md",
      },
    },
    Modal: {
      defaultProps: {
        radius: "md",
      },
    },
  },
});
