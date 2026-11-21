export interface WithdrawalHistory {
  withdrawal_id: number;
  member_id: number;
  transaction_id: string;
  amount: number;
  currency: string;
  status: string;
  payment_method: string;
  withdrawal_time: string;
  notes: string;
  screenshot_url: string;
  created_at: string;
}

export interface WithdrawalCreateRequest {
  transaction_id: string;
  amount: number;
  currency: string;
  status: string;
  payment_method: string;
  withdrawal_time: string;
  notes: string;
  screenshot_url?: string;
}

export interface ParsedWithdrawalData {
  transaction_id: string;
  amount: number;
  currency: string;
  status: string;
  payment_method: string;
  withdrawal_time: string;
  notes: string;
  screenshot_url: string;
  confidence: string;
  raw_summary: string;
  is_duplicate?: boolean;
  duplicate_reason?: string;
}
