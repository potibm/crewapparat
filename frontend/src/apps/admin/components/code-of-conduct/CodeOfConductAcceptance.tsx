import { useState } from "react";
import {
  Box,
  Button,
  Checkbox,
  FormControlLabel,
  Paper,
  Typography,
} from "@mui/material";
import { Logo } from "@core/logo/Logo";
import { useAppConfig } from "@core/config/useConfig";
import { MarkdownContent } from "../fields/MarkdownContent";
import type { MyCodeOfConduct } from "./useMyCodeOfConduct";

export interface CodeOfConductAcceptanceProps {
  codeOfConduct: MyCodeOfConduct;
  onAccept: () => Promise<void>;
}

export const CodeOfConductAcceptance = ({
  codeOfConduct,
  onAccept,
}: CodeOfConductAcceptanceProps) => {
  const { date_locale } = useAppConfig();
  const [confirmed, setConfirmed] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const publishedAt = codeOfConduct.published_at
    ? new Intl.DateTimeFormat(date_locale, {
        year: "numeric",
        month: "long",
        day: "numeric",
      }).format(new Date(codeOfConduct.published_at))
    : null;

  const handleAccept = async () => {
    setSubmitting(true);
    setError(null);

    try {
      await onAccept();
    } catch (acceptError) {
      setError(
        acceptError instanceof Error
          ? acceptError.message
          : "Failed to accept the Code of Conduct",
      );
      setSubmitting(false);
    }
  };

  return (
    <Box
      sx={{
        minHeight: "100vh",
        bgcolor: "background.default",
        display: "flex",
        justifyContent: "center",
        alignItems: "flex-start",
        py: { xs: 2, md: 6 },
        px: 2,
      }}
    >
      <Paper sx={{ width: "100%", maxWidth: 900, p: { xs: 2, md: 4 } }}>
        <Box sx={{ display: "flex", alignItems: "center", gap: 2, mb: 2 }}>
          <Logo style={{ height: "40px", width: "40px" }} />
          <Box>
            <Typography variant="h5">Code of Conduct</Typography>
            <Typography variant="body2" color="text.secondary">
              Version {codeOfConduct.version}
              {publishedAt ? ` · published ${publishedAt}` : ""}
            </Typography>
          </Box>
        </Box>

        <Box sx={{ "& .wmde-markdown": { p: { xs: 1.5, md: 2 } } }}>
          <MarkdownContent value={codeOfConduct.text} />
        </Box>

        <Box sx={{ mt: 3 }}>
          <FormControlLabel
            control={
              <Checkbox
                checked={confirmed}
                onChange={(event) => setConfirmed(event.target.checked)}
                disabled={submitting}
              />
            }
            label="I confirm that I am fine with the content of this Code of Conduct and act accordingly."
          />

          {error && (
            <Typography color="error" variant="body2" sx={{ mt: 1 }}>
              {error}
            </Typography>
          )}

          <Box sx={{ mt: 2 }}>
            <Button
              variant="contained"
              disabled={!confirmed || submitting}
              onClick={handleAccept}
            >
              I confirm
            </Button>
          </Box>
        </Box>
      </Paper>
    </Box>
  );
};
