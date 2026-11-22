package model

import "time"

type Member struct {
	MemberID      int32     `gorm:"column:member_id;primaryKey;autoIncrement" json:"member_id"`
	Fullname      string    `gorm:"column:fullname;type:varchar(150);not null" json:"fullname"`
	Email         string    `gorm:"column:email;type:varchar(191);uniqueIndex:idx_email;not null" json:"email"`
	Password      string    `gorm:"column:password;type:varchar(255);default:''" json:"-"`
	PhoneNumber   string    `gorm:"column:phone_number;type:varchar(50);default:''" json:"phone_number"`
	GoogleID      string    `gorm:"column:google_id;type:varchar(191);index:idx_google_id;default:''" json:"google_id"`
	IsPasswordSet bool      `gorm:"column:is_password_set;default:true" json:"is_password_set"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Member) TableName() string {
	return "member"
}

type PortfolioSetting struct {
	ID               int32     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MemberID         int32     `gorm:"column:member_id;uniqueIndex;not null" json:"member_id"`
	InitialBalance   float64   `gorm:"column:initial_balance;type:decimal(18,4);default:1000" json:"initial_balance"`
	Currency         string    `gorm:"column:currency;type:varchar(20);default:'USDT'" json:"currency"`
	MonthlyTargetPct float64   `gorm:"column:monthly_target_pct;type:decimal(6,2);default:10" json:"monthly_target_pct"`
	MaxDailyLossUSD  float64   `gorm:"column:max_daily_loss_usd;type:decimal(18,4);default:50" json:"max_daily_loss_usd"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (PortfolioSetting) TableName() string {
	return "portfolio_setting"
}

type TradeHistory struct {
	TradeID         int32     `gorm:"column:trade_id;primaryKey;autoIncrement" json:"trade_id"`
	TradeCode       string    `gorm:"column:trade_code;type:varchar(50);uniqueIndex;not null" json:"trade_code"`
	ExternalOrderID string    `gorm:"column:external_order_id;type:varchar(100);index:idx_trade_ext_order;default:''" json:"external_order_id"`
	MemberID        int32     `gorm:"column:member_id;index:idx_trade_member;not null" json:"member_id"`
	Symbol          string    `gorm:"column:symbol;type:varchar(30);index:idx_trade_symbol;not null" json:"symbol"`
	PositionSide    string    `gorm:"column:position_side;type:varchar(10);not null" json:"position_side"`
	Leverage        int32     `gorm:"column:leverage;default:1" json:"leverage"`
	EntryPrice      float64   `gorm:"column:entry_price;type:decimal(18,6);not null" json:"entry_price"`
	ExitPrice       float64   `gorm:"column:exit_price;type:decimal(18,6);not null" json:"exit_price"`
	Quantity        float64   `gorm:"column:quantity;type:decimal(18,6);default:0" json:"quantity"`
	MarginUsed      float64   `gorm:"column:margin_used;type:decimal(18,4);default:0" json:"margin_used"`
	Pnl             float64   `gorm:"column:pnl;type:decimal(18,4);not null" json:"pnl"`
	PnlPercentage   float64   `gorm:"column:pnl_percentage;type:decimal(10,4);not null" json:"pnl_percentage"`
	Fee             float64   `gorm:"column:fee;type:decimal(18,4);default:0" json:"fee"`
	TradeTime       time.Time `gorm:"column:trade_time;index:idx_trade_time;not null" json:"trade_time"`
	SetupTag        string    `gorm:"column:setup_tag;type:varchar(100);default:''" json:"setup_tag"`
	EmotionTag      string    `gorm:"column:emotion_tag;type:varchar(100);default:''" json:"emotion_tag"`
	Notes           string    `gorm:"column:notes;type:text" json:"notes"`
	ScreenshotURL   string    `gorm:"column:screenshot_url;type:varchar(255);default:''" json:"screenshot_url"`
	SyncSheetStatus string    `gorm:"column:sync_sheet_status;type:varchar(20);default:'Pending'" json:"sync_sheet_status"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (TradeHistory) TableName() string {
	return "trade_history"
}

type AiAnalysisLog struct {
	ID              int32     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MemberID        int32     `gorm:"column:member_id;index:idx_ai_member;not null" json:"member_id"`
	Symbol          string    `gorm:"column:symbol;type:varchar(30);not null" json:"symbol"`
	Timeframe       string    `gorm:"column:timeframe;type:varchar(20);default:'1h'" json:"timeframe"`
	AnalysisType    string    `gorm:"column:analysis_type;type:varchar(50);not null" json:"analysis_type"`
	PromptSummary   string    `gorm:"column:prompt_summary;type:text" json:"prompt_summary"`
	AiResponse      string    `gorm:"column:ai_response;type:longtext;not null" json:"ai_response"`
	SignalDirection string    `gorm:"column:signal_direction;type:varchar(20);default:'NEUTRAL'" json:"signal_direction"`
	SentimentScore  float64   `gorm:"column:sentiment_score;type:decimal(4,2);default:0" json:"sentiment_score"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (AiAnalysisLog) TableName() string {
	return "ai_analysis_log"
}

type SheetConfig struct {
	ID            int32      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MemberID      int32      `gorm:"column:member_id;uniqueIndex;not null" json:"member_id"`
	SpreadsheetID string     `gorm:"column:spreadsheet_id;type:varchar(191);not null" json:"spreadsheet_id"`
	SheetName     string     `gorm:"column:sheet_name;type:varchar(100);default:'Trades'" json:"sheet_name"`
	IsAutoSync    bool       `gorm:"column:is_auto_sync;default:true" json:"is_auto_sync"`
	LastSyncedAt  *time.Time `gorm:"column:last_synced_at" json:"last_synced_at"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (SheetConfig) TableName() string {
	return "sheet_config"
}

type WithdrawalHistory struct {
	WithdrawalID    int32     `gorm:"column:withdrawal_id;primaryKey;autoIncrement" json:"withdrawal_id"`
	MemberID        int32     `gorm:"column:member_id;index:idx_withdrawal_member;not null" json:"member_id"`
	TransactionType string    `gorm:"column:transaction_type;type:varchar(20);default:'WITHDRAW'" json:"transaction_type"`
	TransactionID   string    `gorm:"column:transaction_id;type:varchar(120);index:idx_withdrawal_tx;default:''" json:"transaction_id"`
	Amount         float64   `gorm:"column:amount;type:decimal(18,2);not null" json:"amount"`
	Currency       string    `gorm:"column:currency;type:varchar(20);default:'THB'" json:"currency"`
	AmountUSDT     float64   `gorm:"column:amount_usdt;type:decimal(18,4);default:0" json:"amount_usdt"`
	ExchangeRate   float64   `gorm:"column:exchange_rate;type:decimal(10,4);default:0" json:"exchange_rate"`
	Status         string    `gorm:"column:status;type:varchar(30);default:'สำเร็จ'" json:"status"`
	PaymentMethod  string    `gorm:"column:payment_method;type:varchar(60);default:'BANK_TRANSFER'" json:"payment_method"`
	WithdrawalTime time.Time `gorm:"column:withdrawal_time;index:idx_withdrawal_time;not null" json:"withdrawal_time"`
	Notes          string    `gorm:"column:notes;type:text" json:"notes"`
	ScreenshotURL  string    `gorm:"column:screenshot_url;type:varchar(255);default:''" json:"screenshot_url"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (WithdrawalHistory) TableName() string {
	return "withdrawal_history"
}

type BinanceConfig struct {
	ID            int32      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MemberID      int32      `gorm:"column:member_id;uniqueIndex;not null" json:"member_id"`
	APIKeyEnc     string     `gorm:"column:api_key_enc;type:text;not null" json:"-"`
	APISecretEnc  string     `gorm:"column:api_secret_enc;type:text;not null" json:"-"`
	APIKeyMasked  string     `gorm:"column:api_key_masked;type:varchar(60);default:''" json:"api_key_masked"`
	IsAutoSync    bool       `gorm:"column:is_auto_sync;default:true" json:"is_auto_sync"`
	LastSyncedAt  *time.Time `gorm:"column:last_synced_at" json:"last_synced_at"`
	LastSyncError string     `gorm:"column:last_sync_error;type:text" json:"last_sync_error"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (BinanceConfig) TableName() string {
	return "binance_config"
}
