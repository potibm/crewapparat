import { Chip, Stack } from "@mui/material";
import { Link } from "react-router";

export interface TeamChipItem {
  id: string;
  name: string;
  color?: string;
}

const fallbackColors = [
  "#FFD6A5",
  "#CAFFBF",
  "#9BF6FF",
  "#A0C4FF",
  "#BDB2FF",
  "#FFC6FF",
  "#FDFFB6",
  "#FFADAD",
  "#B9FBC0",
  "#FFDAC1",
];

const colorFor = (team: TeamChipItem): string => {
  if (team.color) return team.color;

  let hash = 0;
  for (const char of team.name) {
    hash = (hash * 31 + char.charCodeAt(0)) >>> 0;
  }

  return fallbackColors[hash % fallbackColors.length];
};

const teamFilterUrl = (teamId: string): string =>
  `/organizers?filter=${encodeURIComponent(JSON.stringify({ team: teamId }))}`;

export const TeamChips = ({ teams }: { teams?: TeamChipItem[] }) => {
  if (!teams || teams.length === 0) return null;

  return (
    <Stack direction="row" spacing={1} sx={{ flexWrap: "wrap", gap: 1 }}>
      {teams.map((team) => {
        const color = colorFor(team);

        return (
          <Chip
            key={team.id}
            label={team.name}
            size="small"
            clickable
            component={Link}
            to={teamFilterUrl(team.id)}
            onClick={(event) => event.stopPropagation()}
            sx={{
              bgcolor: color,
              color: "#1f2937",
              fontWeight: 600,
              border: "none",
              "& .MuiChip-label": {
                color: "#1f2937",
              },
              "&:hover": { bgcolor: color, filter: "brightness(0.95)" },
            }}
          />
        );
      })}
    </Stack>
  );
};

export default TeamChips;
