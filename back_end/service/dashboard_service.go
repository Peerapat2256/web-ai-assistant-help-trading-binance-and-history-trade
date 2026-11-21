package service

import (
	"backend/model/model"
	"backend/repository"
	"backend/utils"
	"context"
	"math"
	"sort"
	"time"
)

type DashboardService interface {
	GetSummary(ctx context.Context, memberID int32) (*model.DashboardSummaryResponse, error)
	GetCalendar(ctx context.Context, memberID int32, year int, month int) ([]model.PnlCalendarDay, error)
}

type dashboardServiceImpl struct {
	dashRepo       repository.DashboardRepository
	portfolioRepo  repository.PortfolioRepository
	withdrawalRepo repository.WithdrawalRepository
}

func NewDashboardService(dashRepo repository.DashboardRepository, portfolioRepo repository.PortfolioRepository, withdrawalRepo repository.WithdrawalRepository) DashboardService {
	return &dashboardServiceImpl{
		dashRepo:       dashRepo,
		portfolioRepo:  portfolioRepo,
		withdrawalRepo: withdrawalRepo,
	}
}

func (s *dashboardServiceImpl) GetSummary(ctx context.Context, memberID int32) (*model.DashboardSummaryResponse, error) {
	trades, err := s.dashRepo.GetAllTradesAsc(ctx, memberID)
	if err != nil {
		return nil, err
	}

	setting, err := s.portfolioRepo.GetByMemberID(ctx, memberID)
	initialBalance := 1000.0
	maxDailyLoss := 50.0
	monthlyTarget := 10.0
	if err == nil && setting != nil {
		if setting.InitialBalance > 0 {
			initialBalance = setting.InitialBalance
		}
		maxDailyLoss = setting.MaxDailyLossUSD
		monthlyTarget = setting.MonthlyTargetPct
	}

	var netPnl, grossProfit, grossLoss, todayPnl float64
	var winTrades, lossTrades int
	var bestTrade, worstTrade float64
	todayStr := time.Now().Format("2006-01-02")

	dailyMap := map[string]float64{}
	dailyOrder := []string{}
	symbolMap := map[string]*model.SymbolStat{}
	sideMap := map[string]*model.SideStat{}
	sessionMap := map[string]*struct {
		total int
		wins  int
		pnl   float64
	}{}

	for i, t := range trades {
		netPnl += t.Pnl
		if t.Pnl > 0 {
			winTrades++
			grossProfit += t.Pnl
		} else if t.Pnl < 0 {
			lossTrades++
			grossLoss += math.Abs(t.Pnl)
		}

		if i == 0 || t.Pnl > bestTrade {
			bestTrade = t.Pnl
		}
		if i == 0 || t.Pnl < worstTrade {
			worstTrade = t.Pnl
		}

		dKey := t.TradeTime.Format("2006-01-02")
		if dKey == todayStr {
			todayPnl += t.Pnl
		}
		if _, exists := dailyMap[dKey]; !exists {
			dailyOrder = append(dailyOrder, dKey)
		}
		dailyMap[dKey] += t.Pnl

		symStat, ok := symbolMap[t.Symbol]
		if !ok {
			symStat = &model.SymbolStat{Symbol: t.Symbol}
			symbolMap[t.Symbol] = symStat
		}
		symStat.TotalTrades++
		if t.Pnl > 0 {
			symStat.WinTrades++
		}
		symStat.TotalPnl += t.Pnl

		sideStat, ok := sideMap[t.PositionSide]
		if !ok {
			sideStat = &model.SideStat{Side: t.PositionSide}
			sideMap[t.PositionSide] = sideStat
		}
		sideStat.TotalTrades++
		if t.Pnl > 0 {
			sideStat.WinTrades++
		}
		sideStat.TotalPnl += t.Pnl

		sessName := classifySession(t.TradeTime.Hour())
		sess, ok := sessionMap[sessName]
		if !ok {
			sess = &struct {
				total int
				wins  int
				pnl   float64
			}{}
			sessionMap[sessName] = sess
		}
		sess.total++
		if t.Pnl > 0 {
			sess.wins++
		}
		sess.pnl += t.Pnl
	}

	totalTrades := len(trades)
	winRate := 0.0
	if totalTrades > 0 {
		winRate = math.Round((float64(winTrades)/float64(totalTrades))*10000) / 100
	}

	profitFactor := 0.0
	if grossLoss > 0 {
		profitFactor = math.Round((grossProfit/grossLoss)*100) / 100
	} else if grossProfit > 0 {
		profitFactor = 99.99
	}

	avgRR := 0.0
	if winTrades > 0 && lossTrades > 0 {
		avgWin := grossProfit / float64(winTrades)
		avgLoss := grossLoss / float64(lossTrades)
		if avgLoss > 0 {
			avgRR = math.Round((avgWin/avgLoss)*100) / 100
		}
	}

	streak := 0
	streakType := "NONE"
	for i := len(trades) - 1; i >= 0; i-- {
		pnl := trades[i].Pnl
		if streak == 0 {
			if pnl > 0 {
				streakType = "WIN"
				streak = 1
			} else if pnl < 0 {
				streakType = "LOSS"
				streak = 1
			}
		} else {
			if streakType == "WIN" && pnl > 0 {
				streak++
			} else if streakType == "LOSS" && pnl < 0 {
				streak++
			} else {
				break
			}
		}
	}

	equityCurve := make([]model.EquityCurvePoint, 0, len(dailyOrder))
	cumPnl := 0.0
	for _, d := range dailyOrder {
		dp := dailyMap[d]
		cumPnl += dp
		equityCurve = append(equityCurve, model.EquityCurvePoint{
			Date:          d,
			DailyPnl:      math.Round(dp*100) / 100,
			CumulativePnl: math.Round(cumPnl*100) / 100,
			Balance:       math.Round((initialBalance+cumPnl)*100) / 100,
		})
	}

	symbolStats := make([]model.SymbolStat, 0, len(symbolMap))
	for _, st := range symbolMap {
		if st.TotalTrades > 0 {
			st.WinRate = math.Round((float64(st.WinTrades)/float64(st.TotalTrades))*10000) / 100
		}
		st.TotalPnl = math.Round(st.TotalPnl*100) / 100
		symbolStats = append(symbolStats, *st)
	}
	sort.Slice(symbolStats, func(i, j int) bool {
		return symbolStats[i].TotalPnl > symbolStats[j].TotalPnl
	})

	sideStats := make([]model.SideStat, 0, len(sideMap))
	for _, st := range sideMap {
		if st.TotalTrades > 0 {
			st.WinRate = math.Round((float64(st.WinTrades)/float64(st.TotalTrades))*10000) / 100
		}
		st.TotalPnl = math.Round(st.TotalPnl*100) / 100
		sideStats = append(sideStats, *st)
	}

	sessionStats := make([]model.SessionStat, 0, len(sessionMap))
	for name, sess := range sessionMap {
		wr := 0.0
		if sess.total > 0 {
			wr = math.Round((float64(sess.wins)/float64(sess.total))*10000) / 100
		}
		sessionStats = append(sessionStats, model.SessionStat{
			Session:     name,
			TotalTrades: sess.total,
			WinRate:     wr,
			TotalPnl:    math.Round(sess.pnl*100) / 100,
		})
	}

	recent := make([]model.TradeHistory, 0)
	for i := len(trades) - 1; i >= 0 && len(recent) < 5; i-- {
		recent = append(recent, trades[i])
	}

	// คำนวณยอดถอนเงินทั้งหมด (เฉพาะรายการสำเร็จ) เพื่อหักออกจากยอดพอร์ตปัจจุบัน
	totalWithdrawnUSDT := 0.0
	totalWithdrawnTHB := 0.0
	if s.withdrawalRepo != nil {
		if withdrawals, wErr := s.withdrawalRepo.ListByMember(ctx, memberID, 0); wErr == nil {
			liveRate := 0.0
			for _, w := range withdrawals {
				if w.Status != "สำเร็จ" && w.Status != "Completed" {
					continue
				}
				if w.Currency == "THB" {
					totalWithdrawnTHB += w.Amount
				}
				if w.AmountUSDT > 0 {
					totalWithdrawnUSDT += w.AmountUSDT
				} else if w.Currency == "USDT" || w.Currency == "USD" {
					totalWithdrawnUSDT += w.Amount
				} else if w.Currency == "THB" {
					if liveRate <= 0 {
						liveRate = utils.GetUSDTTHBRate()
					}
					totalWithdrawnUSDT += w.Amount / liveRate
				}
			}
		}
	}

	currentBalance := initialBalance + netPnl - totalWithdrawnUSDT
	netPnlPct := 0.0
	if initialBalance > 0 {
		netPnlPct = math.Round((netPnl/initialBalance)*10000) / 100
	}

	isAlert := maxDailyLoss > 0 && todayPnl <= -maxDailyLoss

	return &model.DashboardSummaryResponse{
		Success:            true,
		InitialBalance:     initialBalance,
		CurrentBalance:     math.Round(currentBalance*100) / 100,
		TotalWithdrawnUSDT: math.Round(totalWithdrawnUSDT*100) / 100,
		TotalWithdrawnTHB:  math.Round(totalWithdrawnTHB*100) / 100,
		NetPnl:             math.Round(netPnl*100) / 100,
		NetPnlPct:        netPnlPct,
		TodayPnl:         math.Round(todayPnl*100) / 100,
		MaxDailyLossUSD:  maxDailyLoss,
		IsDailyLossAlert: isAlert,
		MonthlyTargetPct: monthlyTarget,
		TotalTrades:      totalTrades,
		WinTrades:        winTrades,
		LossTrades:       lossTrades,
		WinRate:          winRate,
		ProfitFactor:     profitFactor,
		AvgRiskReward:    avgRR,
		BestTradePnl:     math.Round(bestTrade*100) / 100,
		WorstTradePnl:    math.Round(worstTrade*100) / 100,
		CurrentStreak:    streak,
		StreakType:       streakType,
		EquityCurve:      equityCurve,
		SymbolStats:      symbolStats,
		SideStats:        sideStats,
		SessionStats:     sessionStats,
		RecentTrades:     recent,
	}, nil
}

func classifySession(hour int) string {
	switch {
	case hour >= 6 && hour < 13:
		return "ช่วงเช้า (ตลาดเอเชีย)"
	case hour >= 13 && hour < 19:
		return "ช่วงบ่าย (ตลาดยุโรป)"
	default:
		return "ช่วงค่ำ-ดึก (ตลาดอเมริกา)"
	}
}

func (s *dashboardServiceImpl) GetCalendar(ctx context.Context, memberID int32, year int, month int) ([]model.PnlCalendarDay, error) {
	trades, err := s.dashRepo.GetTradesByMonth(ctx, memberID, year, month)
	if err != nil {
		return nil, err
	}

	dayMap := map[string]*model.PnlCalendarDay{}
	keys := []string{}
	for _, t := range trades {
		d := t.TradeTime.Format("2006-01-02")
		entry, ok := dayMap[d]
		if !ok {
			entry = &model.PnlCalendarDay{Date: d}
			dayMap[d] = entry
			keys = append(keys, d)
		}
		entry.Pnl += t.Pnl
		entry.TradeCount++
		if t.Pnl > 0 {
			entry.WinCount++
		}
	}

	result := make([]model.PnlCalendarDay, 0, len(keys))
	for _, k := range keys {
		e := dayMap[k]
		e.Pnl = math.Round(e.Pnl*100) / 100
		result = append(result, *e)
	}
	return result, nil
}
