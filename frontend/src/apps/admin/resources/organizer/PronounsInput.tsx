import { useInput } from "react-admin";
import {
  FormControl,
  FormControlLabel,
  Radio,
  RadioGroup,
  TextField,
  Typography,
} from "@mui/material";

const FIXED_PRONOUNS = ["she/her", "he/him", "they/them"];

export const PronounsInput = (props: { source: string; label?: string }) => {
  const { field } = useInput(props);
  const value = (field.value as string) ?? "";

  const isFixed = FIXED_PRONOUNS.includes(value);
  const isOther = value !== "" && !isFixed;

  const selected = isFixed ? value : isOther ? "other" : "";

  const handleRadioChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const next = event.target.value;
    if (next === "other") {
      field.onChange(isOther ? value : "");
    } else {
      field.onChange(next);
    }
  };

  return (
    <FormControl component="fieldset" fullWidth>
      <Typography variant="caption" color="text.secondary">
        Pronouns
      </Typography>
      <RadioGroup value={selected} onChange={handleRadioChange} row>
        {FIXED_PRONOUNS.map((option) => (
          <FormControlLabel
            key={option}
            value={option}
            control={<Radio size="small" />}
            label={option}
          />
        ))}
        <FormControlLabel
          value="other"
          control={<Radio size="small" />}
          label="Other"
        />
      </RadioGroup>
      {(selected === "other" || isOther) && (
        <TextField
          size="small"
          fullWidth
          value={value}
          onChange={(event) => field.onChange(event.target.value)}
          placeholder="e.g. she/they"
        />
      )}
    </FormControl>
  );
};
