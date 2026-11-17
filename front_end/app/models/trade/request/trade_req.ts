export interface TradeCreateRequest {
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
  screenshot_url?: string;
}
