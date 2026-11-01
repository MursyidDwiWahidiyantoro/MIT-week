import { z } from 'zod';

/**
 * Zod mirrors of api/dossier.schema.json. A contract change updates the OpenAPI
 * file, the JSON Schema, the Go structs and this file in the same commit.
 */

export const statusSchema = z.enum(['SAFE', 'SUSPICIOUS', 'TAMPERED', 'INSUFFICIENT']);
export const factStatusSchema = z.enum(['CONTRADICTS', 'SUPPORTS', 'UNVERIFIED']);
export const inputTypeSchema = z.enum(['URL', 'FILE', 'TEXT']);

export const tacticCodeSchema = z.enum([
  'URGENCY',
  'REWARD_BAIT',
  'FALSE_AUTHORITY',
  'FEAR',
  'SECRECY',
  'OUTRAGE_BAIT',
  'IMPERSONATION',
  'FAKE_SOCIAL_PROOF'
]);

export const anomalySchema = z.object({
  bounding_box: z.tuple([z.number(), z.number(), z.number(), z.number()]),
  extracted_text: z.string().optional(),
  reason: z.string().optional()
});

export const sourceSchema = z.object({
  title: z.string(),
  publisher: z.string(),
  url: z.string(),
  match: factStatusSchema
});

export const tacticSchema = z.object({
  code: tacticCodeSchema,
  phrase: z.string(),
  source: z.enum(['text', 'ocr'])
});

export const verdictSchema = z.object({
  score: z.number().int().min(0).max(100).nullable(),
  status: statusSchema,
  short_title: z.string(),
  explanation: z.string().max(400),
  explanation_simple: z.string().max(140),
  basis: z.array(z.enum(['ela', 'ai', 'fact', 'link', 'llm_untrusted'])),
  not_checked: z.array(z.enum(['ela', 'ai', 'fact', 'link'])),
  floors_applied: z.array(z.string())
});

export const dossierSchema = z.object({
  schema_version: z.literal(2),
  id: z.string(),
  timestamp: z.string(),
  input_type: inputTypeSchema,
  source_reference: z.string(),
  degraded: z.boolean(),
  is_demo: z.boolean(),
  verdict: verdictSchema,
  pillars: z.object({
    ela: z.object({
      available: z.boolean(),
      score: z.number(),
      reliability: z.number().min(0).max(1),
      reliability_note: z.string().nullish(),
      // The server serialises an empty Go slice as null. Accept both and
      // normalise here so callers only ever see an array.
      anomalies: z.array(anomalySchema).nullish().transform((v) => v ?? [])
    }),
    ai: z.object({
      available: z.boolean(),
      score: z.number(),
      indicators: z.array(z.string()).nullish().transform((v) => v ?? []),
      caveat: z.string().optional()
    }),
    fact: z.object({
      available: z.boolean(),
      claim: z.string().optional(),
      // An unavailable fact pillar serialises status as "". Treat that as
      // UNVERIFIED rather than failing the whole dossier.
      status: z
        .union([factStatusSchema, z.literal('')])
        .transform((v) => (v === '' ? ('UNVERIFIED' as const) : v)),
      sources: z.array(sourceSchema).nullish().transform((v) => v ?? [])
    }),
    link: z.object({
      available: z.boolean(),
      score: z.number(),
      final_url: z.string().optional(),
      redirect_chain: z.array(z.string()).nullish().transform((v) => v ?? []),
      reasons: z
        .array(z.object({ code: z.string(), text: z.string() }))
        .nullish()
        .transform((v) => v ?? [])
    })
  }),
  tactics: z.array(tacticSchema),
  share_token: z.string()
});

export const errorEnvelopeSchema = z.object({
  error: z.object({
    code: z.string(),
    message: z.string(),
    retryable: z.boolean()
  })
});

export type DossierParsed = z.infer<typeof dossierSchema>;

/** Raised when the server speaks a newer contract than this build understands. */
export class SchemaVersionError extends Error {
  constructor() {
    super('Perbarui aplikasi');
    this.name = 'SchemaVersionError';
  }
}

/**
 * Raised when a payload does not match the contract at all. Kept separate from
 * a network failure so the UI can say "this is our problem" instead of
 * "try again" for something retrying will never fix.
 */
export class ContractError extends Error {
  readonly issues: string[];
  constructor(issues: string[]) {
    super('Payload does not match the Dossier contract');
    this.name = 'ContractError';
    this.issues = issues;
  }
}

/**
 * Parses an unknown payload into a Dossier. A schema_version this build does
 * not know produces a distinct error so the UI can say "Perbarui aplikasi"
 * instead of showing a generic failure.
 */
export function parseDossier(payload: unknown): DossierParsed {
  if (
    typeof payload === 'object' &&
    payload !== null &&
    'schema_version' in payload &&
    (payload as { schema_version: unknown }).schema_version !== 2
  ) {
    throw new SchemaVersionError();
  }
  const parsed = dossierSchema.safeParse(payload);
  if (!parsed.success) {
    throw new ContractError(
      parsed.error.issues.map((i) => `${i.path.join('.') || '(root)'}: ${i.message}`)
    );
  }
  return parsed.data;
}
