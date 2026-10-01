import { defaultTheme } from "react-admin";
import type { ThemeOptions } from "@mui/material";

export const MyTheme: ThemeOptions = {
  ...defaultTheme,
  palette: {
    ...defaultTheme.palette,
    mode: "light",
    primary: {
      main: "#1E293B",
      contrastText: "#FFFFFF",
    },
    secondary: {
      main: "#3B82F6",
      contrastText: "#FFFFFF",
    },
    background: {
      default: "#F8FAFC",
      paper: "#FFFFFF",
    },
    error: { main: "#FF3366" },
    warning: { main: "#FF9800" },
    info: { main: "#00BCD4" },
    success: { main: "#10B981" },
  },
  components: {
    ...defaultTheme.components,
    MuiAppBar: {
      styleOverrides: {
        root: {
          backgroundColor: "#CCFF00",
          color: "#000000",
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: "6px",
          fontWeight: 600,
          textTransform: "none",
        },
      },
    },
  },
};

export const MyDarkTheme: ThemeOptions = {
  ...defaultTheme,
  palette: {
    ...defaultTheme.palette,
    mode: "dark",
    primary: {
      main: "#CCFF00",
      light: "#D9FF33",
      dark: "#99CC00",
      contrastText: "#000000",
    },
    secondary: {
      main: "#00E5FF",
      contrastText: "#000000",
    },
    background: {
      default: "#0F172A",
      paper: "#1E293B",
    },
    error: { main: "#FF3366" },
    warning: { main: "#FFB300" },
    info: { main: "#00E5FF" },
    success: { main: "#00E676" },
  },
  components: {
    ...defaultTheme.components,
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: "6px",
          fontWeight: 600,
          textTransform: "none",
        },
      },
    },
    MuiTableCell: {
      styleOverrides: {
        head: {
          backgroundColor: "#0F172A",
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        filledPrimary: {
          backgroundColor: "#CCFF00",
          color: "#000000",
        },
        root: {
          backgroundColor: "rgba(15, 23, 42, 0.6)",
          border: "1px solid rgba(204, 255, 0, 0.3)",
        },
        label: {
          color: "#CCFF00",
        },
      },
    },
    MuiTableRow: {
      styleOverrides: {
        root: {
          "&:hover": {
            backgroundColor: "rgba(255, 255, 255, 0.04) !important",
          },
        },
      },
    },
  },
};
