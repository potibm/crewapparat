import { defaultTheme } from "react-admin";
import type { ThemeOptions } from "@mui/material";

export const MyTheme: ThemeOptions = {
  ...defaultTheme,
  palette: {
    ...defaultTheme.palette,
    mode: "light",
    primary: {
      main: "#FF6600",
      contrastText: "#FFFFFF",
    },
    secondary: {
      main: "#1E293B",
      contrastText: "#FFFFFF",
    },
    background: {
      default: "#FFF7F2",
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
          backgroundColor: "#FF6600",
          color: "#FFFFFF",
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
      main: "#FF6600",
      light: "#FF8533",
      dark: "#CC5200",
      contrastText: "#FFFFFF",
    },
    secondary: {
      main: "#3B82F6",
      contrastText: "#FFFFFF",
    },
    background: {
      default: "#1A1A1A",
      paper: "#2D2D2D",
    },
    error: { main: "#FF3366" },
    warning: { main: "#FFB300" },
    info: { main: "#00BCD4" },
    success: { main: "#10B981" },
  },
  components: {
    ...defaultTheme.components,
    RaToolbar: {
      styleOverrides: {
        root: {
          backgroundColor: "#424242",
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
    MuiTableCell: {
      styleOverrides: {
        head: {
          backgroundColor: "#1A1A1A",
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        colorPrimary: {
          "&.MuiChip-filled": {
            backgroundColor: "#FF6600",
            color: "#FFFFFF",
          },
        },
        root: {
          backgroundColor: "rgba(26, 26, 26, 0.6)",
          border: "1px solid rgba(255, 102, 0, 0.3)",
        },
        label: {
          color: "#FF6600",
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
