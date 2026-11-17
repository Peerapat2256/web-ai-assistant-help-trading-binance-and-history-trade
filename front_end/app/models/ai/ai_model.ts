export interface MarketKline {
  open_time: number;
  time_str: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface MarketDataSummary {
  symbol: string;
  timeframe: string;
  current_price: number;
  change_24h_pct: number;
  high_24h: number;
  low_24h: number;
  volume_24h: number;
  rsi_14: number;
  ema_20: number;
  ema_50: number;
  support: number;
  resistance: number;
  klines: MarketKline[];
}

export interface AiAnalysisResponse {
  id: number;
  symbol: string;
  timeframe: string;
  analysis_type: string;
  prompt_summary?: string;
  ai_response: string;
  signal_direction: string;
  sentiment_score: number;
  created_at: string;
}
