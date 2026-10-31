/**
 * Mirrors api/openapi.yaml. Keep in sync in the same commit as any contract change.
 */
export type Status = 'SAFE' | 'SUSPICIOUS' | 'TAMPERED' | 'INSUFFICIENT';
export type FactStatus = 'CONTRADICTS' | 'SUPPORTS' | 'UNVERIFIED';
export type InputType = 'URL' | 'FILE' | 'TEXT';
export type TacticCode =
  | 'URGENCY'
  | 'REWARD_BAIT'
  | 'FALSE_AUTHORITY'
  | 'FEAR'
  | 'SECRECY'
  | 'OUTRAGE_BAIT'
  | 'IMPERSONATION'
  | 'FAKE_SOCIAL_PROOF';

export interface Anomaly {
  bounding_box: [number, number, number, number];
  extracted_text?: string;
  reason?: string;
}

export interface Source {
  title: string;
  publisher: string;
  url: string;
  match: FactStatus;
}

export interface Tactic {
  code: TacticCode;
  phrase: string;
  source: 'text' | 'ocr';
}

export interface Verdict {
  score: number | null;
  status: Status;
  short_title: string;
  explanation: string;
  explanation_simple: string;
  basis: Array<'ela' | 'ai' | 'fact' | 'link' | 'llm_untrusted'>;
  not_checked: Array<'ela' | 'ai' | 'fact' | 'link'>;
  floors_applied: string[];
}

export interface Dossier {
  schema_version: 2;
  id: string;
  timestamp: string;
  input_type: InputType;
  source_reference: string;
  degraded: boolean;
  is_demo: boolean;
  verdict: Verdict;
  pillars: {
    ela: { available: boolean; score: number; reliability: number; reliability_note: string | null; anomalies: Anomaly[] };
    ai: { available: boolean; score: number; indicators: string[]; caveat: string };
    fact: { available: boolean; claim: string; status: FactStatus; sources: Source[] };
    link: { available: boolean; score: number; final_url: string; redirect_chain: string[]; reasons: { code: string; text: string }[] };
  };
  tactics: Tactic[];
  share_token: string;
}

export interface ErrorEnvelope {
  error: { code: string; message: string; retryable: boolean };
}
