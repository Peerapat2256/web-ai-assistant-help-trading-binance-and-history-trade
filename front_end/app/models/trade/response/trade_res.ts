export interface TradeHistory {
  trade_id: number;
  trade_code: string;
  member_id: number;
  symbol: string;
  position_side: string;
  leverage: number;
  entry_price: number;
  exit_price: number;
  quantity: number;
  margin_used: number;
  pnl: number;
  pnl_percentage: number;
  fee: number;
  trade_time: string;
  setup_tag: string;
  emotion_tag: string;
  notes: string;
  screenshot_url: string;
  sync_sheet_status: string;
}

export interface ParsedScreenshotData {
  symbol: string;
  position_side: string;
  leverage: number;
  entry_price: number;
  exit_price: number;
  quantity: number;
  margin_used: number;
  pnl: number;
  pnl_percentage: number;
  fee: number;
  trade_time: string;
  screenshot_url: string;
  confidence: string;
  raw_summary: string;
}
