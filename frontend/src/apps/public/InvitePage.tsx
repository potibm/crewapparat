import { useEffect, useState } from "react";
import { BrowserRouter, Route, Routes, useParams } from "react-router";
import {
  Box,
  Button,
  Card,
  CardContent,
  CardHeader,
  Checkbox,
  FormControlLabel,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { MarkdownContent } from "@admin/components/fields/MarkdownContent";

const API_HOST = import.meta.env.VITE_API_HOST ?? "http://localhost:3400";

interface InviteData {
  email: string;
  handle: string;
  event_name: string;
  expires_at: string;
  code_of_conduct: string;
  code_of_conduct_version: string;
}

const InviteForm = () => {
  const { token } = useParams();
  const [invite, setInvite] = useState<InviteData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitted, setSubmitted] = useState(false);

  const [notes, setNotes] = useState("");
  const [cocAccepted, setCocAccepted] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!token) return;
    const controller = new AbortController();
    fetch(`${API_HOST}/api/invites/${token}`, { signal: controller.signal })
      .then(async (res) => {
        if (!res.ok) throw new Error(await res.text());
        return res.json();
      })
      .then((data) => setInvite(data))
      .catch((e) => {
        if (e instanceof DOMException && e.name === "AbortError") return;
        setError(e instanceof Error ? e.message : "Failed to load invite");
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });
    return () => controller.abort();
  }, [token]);

  const submit = async () => {
    if (!token) return;
    setSubmitting(true);
    try {
      const res = await fetch(`${API_HOST}/api/invites/${token}/submit`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          notes,
          coc_accepted: cocAccepted,
          coc_version: invite?.code_of_conduct_version,
        }),
      });
      if (!res.ok) throw new Error(await res.text());
      setSubmitted(true);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to submit");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) return <Typography>Loading...</Typography>;
  if (error) return <Typography color="error">{error}</Typography>;
  if (submitted) {
    return (
      <Box sx={{ maxWidth: 600, mx: "auto", mt: 4 }}>
        <Typography variant="h5">Thank you!</Typography>
        <Typography>Your details have been submitted for review.</Typography>
      </Box>
    );
  }
  if (!invite) return null;

  return (
    <Box sx={{ maxWidth: 600, mx: "auto", mt: 4 }}>
      <Card>
        <CardHeader title={`Invitation to ${invite.event_name}`} />
        <CardContent>
          <Stack spacing={2}>
            <Typography>Email: {invite.email}</Typography>
            {invite.handle ? (
              <Typography>Handle: {invite.handle}</Typography>
            ) : null}
            <TextField
              label="What you would like us to know"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              fullWidth
              multiline
            />
            {invite.code_of_conduct ? (
              <Box
                sx={{
                  maxHeight: 300,
                  overflow: "auto",
                  border: 1,
                  borderColor: "divider",
                  p: 1,
                }}
              >
                <MarkdownContent value={invite.code_of_conduct} />
              </Box>
            ) : null}
            <FormControlLabel
              control={
                <Checkbox
                  checked={cocAccepted}
                  onChange={(e) => setCocAccepted(e.target.checked)}
                />
              }
              label="I accept the Code of Conduct"
            />
            <Button
              variant="contained"
              onClick={submit}
              disabled={submitting || !cocAccepted}
            >
              Submit
            </Button>
          </Stack>
        </CardContent>
      </Card>
    </Box>
  );
};

const InvitePage = () => (
  <BrowserRouter>
    <Routes>
      <Route path="/invite/:token" element={<InviteForm />} />
    </Routes>
  </BrowserRouter>
);

export default InvitePage;
