export interface PortfolioSetting {
  id: number;
  member_id: number;
  initial_balance: number;
  currency: string;
  monthly_target_pct: number;
  max_daily_loss_usd: number;
}

export interface SheetConfig {
  id: number;
  member_id: number;
  spreadsheet_id: string;
  sheet_name: string;
  is_auto_sync: boolean;
  last_synced_at: string | null;
}
