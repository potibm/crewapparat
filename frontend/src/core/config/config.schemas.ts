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
