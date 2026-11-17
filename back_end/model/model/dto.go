package model

type LoginResponse struct {
	ID          int    `json:"id"`
	Fullname    string `json:"fullname"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
	Type        string `json:"type"`
	HasPassword bool   `json:"has_password"`
	GoogleID    string `json:"google_id"`
}

type AuthResponse struct {
	Success      bool          `json:"success"`
	Message      string        `json:"message,omitempty"`
	Token        string        `json:"token"`
	RefreshToken string        `json:"refresh_token"`
	User         LoginResponse `json:"user"`
}

type PortfolioUpdateRequest struct {
	InitialBalance   float64 `json:"initial_balance"`
	Currency         string  `json:"currency"`
	MonthlyTargetPct float64 `json:"monthly_target_pct"`
	MaxDailyLossUSD  float64 `json:"max_daily_loss_usd"`
}

type TradeCreateRequest struct {
	Symbol        string  `json:"symbol" binding:"required"`
	PositionSide  string  `json:"position_side" binding:"required"`
	Leverage      int32   `json:"leverage"`
	EntryPrice    float64 `json:"entry_price"`
	ExitPrice     float64 `json:"exit_price"`
	Quantity      float64 `json:"quantity"`
	MarginUsed    float64 `json:"margin_used"`
	Pnl           float64 `json:"pnl"`
	PnlPercentage float64 `json:"pnl_percentage"`
	Fee           float64 `json:"fee"`
	TradeTime     string  `json:"trade_time"`
	SetupTag      string  `json:"setup_tag"`
	EmotionTag    string  `json:"emotion_tag"`
	Notes         string  `json:"notes"`
	ScreenshotURL string  `json:"screenshot_url"`
}

type TradeUpdateRequest struct {
	Symbol        string  `json:"symbol"`
	PositionSide  string  `json:"position_side"`
	Leverage      int32   `json:"leverage"`
	EntryPrice    float64 `json:"entry_price"`
	ExitPrice     float64 `json:"exit_price"`
	Quantity      float64 `json:"quantity"`
	MarginUsed    float64 `json:"margin_used"`
	Pnl           float64 `json:"pnl"`
	PnlPercentage float64 `json:"pnl_percentage"`
	Fee           float64 `json:"fee"`
	TradeTime     string  `json:"trade_time"`
	SetupTag      string  `json:"setup_tag"`
	EmotionTag    string  `json:"emotion_tag"`
	Notes         string  `json:"notes"`
}

type ParsedScreenshotData struct {
	Symbol        string  `json:"symbol"`
	PositionSide  string  `json:"position_side"`
	Leverage      int32   `json:"leverage"`
	EntryPrice    float64 `json:"entry_price"`
	ExitPrice     float64 `json:"exit_price"`
	Quantity      float64 `json:"quantity"`
	MarginUsed    float64 `json:"margin_used"`
	Pnl           float64 `json:"pnl"`
	PnlPercentage float64 `json:"pnl_percentage"`
	Fee           float64 `json:"fee"`
	TradeTime     string  `json:"trade_time"`
	ScreenshotURL string  `json:"screenshot_url"`
	Confidence    string  `json:"confidence"`
	RawSummary    string  `json:"raw_summary"`
}

type EquityCurvePoint struct {
	Date          string  `json:"date"`
	DailyPnl      float64 `json:"daily_pnl"`
	CumulativePnl float64 `json:"cumulative_pnl"`
	Balance       float64 `json:"balance"`
}

type SymbolStat struct {
	Symbol     string  `json:"symbol"`
	TotalTrades int    `json:"total_trades"`
	WinTrades  int     `json:"win_trades"`
	WinRate    float64 `json:"win_rate"`
	TotalPnl   float64 `json:"total_pnl"`
}

type SideStat struct {
	Side        string  `json:"side"`
	TotalTrades int     `json:"total_trades"`
	WinTrades   int     `json:"win_trades"`
	WinRate     float64 `json:"win_rate"`
	TotalPnl    float64 `json:"total_pnl"`
}

type SessionStat struct {
	Session     string  `json:"session"`
	TotalTrades int     `json:"total_trades"`
	WinRate     float64 `json:"win_rate"`
	TotalPnl    float64 `json:"total_pnl"`
}

type PnlCalendarDay struct {
	Date       string  `json:"date"`
	Pnl        float64 `json:"pnl"`
	TradeCount int     `json:"trade_count"`
	WinCount   int     `json:"win_count"`
}

type DashboardSummaryResponse struct {
	Success           bool               `json:"success"`
	InitialBalance    float64            `json:"initial_balance"`
	CurrentBalance    float64            `json:"current_balance"`
	NetPnl            float64            `json:"net_pnl"`
	NetPnlPct         float64            `json:"net_pnl_pct"`
	TodayPnl          float64            `json:"today_pnl"`
	MaxDailyLossUSD   float64            `json:"max_daily_loss_usd"`
	IsDailyLossAlert  bool               `json:"is_daily_loss_alert"`
	MonthlyTargetPct  float64            `json:"monthly_target_pct"`
	TotalTrades       int                `json:"total_trades"`
	WinTrades         int                `json:"win_trades"`
	LossTrades        int                `json:"loss_trades"`
	WinRate           float64            `json:"win_rate"`
	ProfitFactor      float64            `json:"profit_factor"`
	AvgRiskReward     float64            `json:"avg_risk_reward"`
	BestTradePnl      float64            `json:"best_trade_pnl"`
	WorstTradePnl     float64            `json:"worst_trade_pnl"`
	CurrentStreak     int                `json:"current_streak"`
	StreakType        string             `json:"streak_type"`
	EquityCurve       []EquityCurvePoint `json:"equity_curve"`
	SymbolStats       []SymbolStat       `json:"symbol_stats"`
	SideStats         []SideStat         `json:"side_stats"`
	SessionStats      []SessionStat      `json:"session_stats"`
	RecentTrades      []TradeHistory     `json:"recent_trades"`
}

type MarketKline struct {
	OpenTime int64   `json:"open_time"`
	TimeStr  string  `json:"time_str"`
	Open     float64 `json:"open"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Close    float64 `json:"close"`
	Volume   float64 `json:"volume"`
}

type MarketDataSummary struct {
	Symbol       string        `json:"symbol"`
	Timeframe    string        `json:"timeframe"`
	CurrentPrice float64       `json:"current_price"`
	Change24hPct float64       `json:"change_24h_pct"`
	High24h      float64       `json:"high_24h"`
	Low24h       float64       `json:"low_24h"`
	Volume24h    float64       `json:"volume_24h"`
	RSI14        float64       `json:"rsi_14"`
	EMA20        float64       `json:"ema_20"`
	EMA50        float64       `json:"ema_50"`
	Support      float64       `json:"support"`
	Resistance   float64       `json:"resistance"`
	Klines       []MarketKline `json:"klines"`
}

type AiChartAnalyzeRequest struct {
	Symbol       string `json:"symbol"`
	Timeframe    string `json:"timeframe"`
	Question     string `json:"question"`
	TeachMode    bool   `json:"teach_mode"`
}

type AiAnalysisResponse struct {
	ID              int32   `json:"id"`
	Symbol          string  `json:"symbol"`
	Timeframe       string  `json:"timeframe"`
	AnalysisType    string  `json:"analysis_type"`
	AiResponse      string  `json:"ai_response"`
	SignalDirection string  `json:"signal_direction"`
	SentimentScore  float64 `json:"sentiment_score"`
	CreatedAt       string  `json:"created_at"`
}

type NewsItem struct {
	Title           string  `json:"title"`
	SummaryThai     string  `json:"summary_thai"`
	Source          string  `json:"source"`
	PublishedAt     string  `json:"published_at"`
	URL             string  `json:"url"`
	Symbol          string  `json:"symbol"`
	Sentiment       string  `json:"sentiment"`
	SentimentReason string  `json:"sentiment_reason"`
	ImpactScore     float64 `json:"impact_score"`
}

type SheetConfigRequest struct {
	SpreadsheetID string `json:"spreadsheet_id" binding:"required"`
	SheetName     string `json:"sheet_name"`
	IsAutoSync    bool   `json:"is_auto_sync"`
}
