import type { TradeHistory } from "../trade/response/trade_res";

export interface EquityCurvePoint {
  date: string;
  daily_pnl: number;
  cumulative_pnl: number;
  balance: number;
}

export interface SymbolStat {
  symbol: string;
  total_trades: number;
  win_trades: number;
  win_rate: number;
  total_pnl: number;
}

export interface SideStat {
  side: string;
  total_trades: number;
  win_trades: number;
  win_rate: number;
  total_pnl: number;
}

export interface SessionStat {
  session: string;
  total_trades: number;
  win_rate: number;
  total_pnl: number;
}

export interface PnlCalendarDay {
  date: string;
  pnl: number;
  trade_count: number;
  win_count: number;
}

export interface DashboardSummary {
  success: boolean;
  initial_balance: number;
  current_balance: number;
  total_withdrawn_usdt?: number;
  total_withdrawn_thb?: number;
  net_pnl: number;
  net_pnl_pct: number;
  today_pnl: number;
  max_daily_loss_usd: number;
  is_daily_loss_alert: boolean;
  monthly_target_pct: number;
  total_trades: number;
  win_trades: number;
  loss_trades: number;
  win_rate: number;
  profit_factor: number;
  avg_risk_reward: number;
  best_trade_pnl: number;
  worst_trade_pnl: number;
  current_streak: number;
  streak_type: string;
  equity_curve: EquityCurvePoint[];
  symbol_stats: SymbolStat[];
  side_stats: SideStat[];
  session_stats: SessionStat[];
  recent_trades: TradeHistory[];
}
