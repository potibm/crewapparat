import { z } from "zod";

const SentrySchema = z.object({
  dsn: z.string(),
  environment: z.string(),
  version: z.string(),
  replay_session_sample_rate: z.number().min(0).max(1).default(0),
  replay_error_sample_rate: z.number().min(0).max(1).default(1),
});

export const AppConfigSchema = z.object({
  version: z.string(),
  environment: z.string(),
  environment_message: z.string().optional(),
  sentry: SentrySchema,
  date_locale: z.string().min(2),
  date_options: z.record(z.string(), z.any()).default({
    weekday: "long",
    hour: "2-digit",
    minute: "2-digit",
  }),
  pronoun_options: z
    .array(z.string())
    .default(["she/her", "he/him", "they/them"]),
  language_options: z
    .array(z.string())
    .default(["de", "en", "fr", "hu", "no", "nl", "ru", "pl"]),
  tshirt_size_options: z
    .array(z.string())
    .default(["XXS", "XS", "S", "M", "L", "XL", "XXL", "3XL"]),
  shift: z
    .object({
      start: z.string().default(""),
      block_count: z.number().int().nonnegative().default(0),
    })
    .default({ start: "", block_count: 0 }),
  auth: z
    .object({
      type: z.enum(["oidc"]),
      name: z.string(),
      authority: z.string(),
      client_id: z.string(),
    })
    .optional(),
});

export type AppConfig = z.infer<typeof AppConfigSchema>;
