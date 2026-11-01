import { Avatar } from "@mui/material";

interface OrganizerAvatarProps {
  src?: string | null;
  handle?: string;
  realname?: string;
  size?: number;
  onClick?: () => void;
}

export const OrganizerAvatar = ({
  src,
  handle,
  realname,
  size = 40,
  onClick,
}: OrganizerAvatarProps) => {
  const initial = (handle || realname || "?").charAt(0).toUpperCase();

  return (
    <Avatar
      src={src || undefined}
      onClick={onClick}
      sx={{
        width: size,
        height: size,
        bgcolor: "primary.main",
        cursor: src ? "pointer" : "default",
        fontSize: size > 64 ? 36 : undefined,
      }}
    >
      {initial}
    </Avatar>
  );
};
