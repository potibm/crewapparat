import { Dialog, DialogContent } from "@mui/material";

interface AvatarLightboxProps {
  src: string | null;
  open: boolean;
  onClose: () => void;
}

export const AvatarLightbox = ({ src, open, onClose }: AvatarLightboxProps) => (
  <Dialog open={open} onClose={onClose} maxWidth="md">
    <DialogContent
      sx={{
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        p: 1,
      }}
    >
      {src && (
        <img
          src={src}
          alt="Avatar"
          style={{ maxWidth: "100%", maxHeight: "80vh", objectFit: "contain" }}
        />
      )}
    </DialogContent>
  </Dialog>
);
