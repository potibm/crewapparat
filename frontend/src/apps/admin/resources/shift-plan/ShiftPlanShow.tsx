import { Show, useGetIdentity, useGetList, useRecordContext } from "react-admin";
import { Button, Card, CardContent, CardHeader, Chip, Stack, Typography } from "@mui/material";
import PrintIcon from "@mui/icons-material/Print";
import { useAppConfig } from "@core/config/useConfig";
import type { ShiftPlanRecord } from "./shift-plan.types";
import { ShiftPlanGrid } from "./ShiftPlanGrid";
import type { TimetableOverlayEvent } from "../organizer/ShiftSchedule";

const ShiftPlanShowContent = () => {
  const record = useRecordContext<ShiftPlanRecord>();
  const config = useAppConfig();
  const { data: identity } = useGetIdentity();
  const start = config.shift?.start ?? "";
  const blockCount = config.shift?.block_count ?? 0;

  const { data: organizers } = useGetList(
    "organizers",
    { pagination: { page: 1, perPage: 1000 }, sort: { field: "handle", order: "ASC" } },
    { enabled: !!record },
  );

  const organizerHandle = (id: string) => {
    const org = (organizers ?? []).find((o: { id: string }) => o.id === id) as
      | { handle?: string }
      | undefined;
    return org?.handle ?? id.slice(0, 8);
  };

  const shiftStartDate = start ? new Date(start) : null;
  const shiftEndDate =
    shiftStartDate && blockCount > 0
      ? new Date(shiftStartDate.getTime() + blockCount * 30 * 60 * 1000)
      : null;

  const toIso = (d: Date) => d.toISOString().replace(/\.\d{3}Z$/, "Z");
  const { data: timetableData } = useGetList(
    "timetable-events",
    {
      pagination: { page: 1, perPage: 1000 },
      filter: {
        start: shiftStartDate ? toIso(shiftStartDate) : "",
        end: shiftEndDate ? toIso(shiftEndDate) : "",
      },
    },
    { enabled: !!record && !!shiftStartDate && blockCount > 0 },
  );

  const timetableEvents: TimetableOverlayEvent[] = (
    (timetableData as
      | { id: number; title: string; start_time: string; end_time: string }[]
      | undefined) ?? []
  ).map((e) => ({
    id: e.id,
    title: e.title,
    start_time: e.start_time,
    end_time: e.end_time,
  }));

  if (!record) return null;

  return (
    <Card className="shift-plan-card">
      <style>{`
        @page {
          size: A4 landscape;
          margin: 8mm;
        }
        @media print {
          html, body { background: #fff; margin: 0; }
          .MuiAppBar-root,
          .MuiDrawer-root,
          nav,
          aside,
          .RaSidebar-root,
          header {
            display: none !important;
          }
          main, .RaLayout-content, .MuiContainer-root {
            margin: 0 !important;
            padding: 0 !important;
            max-width: 100% !important;
            width: 100% !important;
          }
          .shift-plan-card { box-shadow: none; border: none; }
          .no-print { display: none !important; }
          table { font-size: 6px; width: 100%; table-layout: fixed; zoom: 0.85; }
          td, th { padding: 0 1px !important; line-height: 1.1 !important; }
          .MuiTableCell-root { font-size: 6px !important; padding: 0 1px !important; }
          .shift-plan-card .MuiCardHeader-root { padding: 0 0 4px 0 !important; }
          .shift-plan-card .MuiCardContent-root { padding: 0 !important; }
        }
      `}</style>
      <CardHeader
        title={`Shift plan ${record.version}`}
        action={
          <Button
            className="no-print"
            variant="outlined"
            size="small"
            startIcon={<PrintIcon />}
            onClick={() => window.print()}
          >
            Print
          </Button>
        }
        subheader={
          <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
            <Chip label={record.status} size="small" />
            <Chip label={record.event_year_id} size="small" variant="outlined" />
            {record.generated_at && (
              <Typography variant="caption" color="text.secondary">
                Generated {new Date(record.generated_at).toLocaleString()}
              </Typography>
            )}
          </Stack>
        }
      />
      <CardContent>
        {start && blockCount > 0 ? (
          <ShiftPlanGrid
            start={start}
            blockCount={blockCount}
            slots={record.slots ?? []}
            organizerHandle={organizerHandle}
            timetableEvents={timetableEvents}
            currentOrganizerId={identity?.id != null ? String(identity.id) : undefined}
          />
        ) : (
          <Typography variant="body2" color="text.secondary">
            Shift planning is not configured.
          </Typography>
        )}
      </CardContent>
    </Card>
  );
};

const ShiftPlanShow = () => (
  <Show>
    <ShiftPlanShowContent />
  </Show>
);

export default ShiftPlanShow;