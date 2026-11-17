package service

import (
	"backend/model/model"
	"backend/repository"
	"backend/utils"
	"context"
	"mime/multipart"
)

type AiCoachService interface {
	GetMarketData(symbol, timeframe string) (*model.MarketDataSummary, error)
	AnalyzeChart(ctx context.Context, memberID int32, symbol, timeframe, question string, teachMode bool, fh *multipart.FileHeader) (*model.AiAnalysisResponse, error)
	ReviewJournal(ctx context.Context, memberID int32) (*model.AiAnalysisResponse, error)
	ListLogs(ctx context.Context, memberID int32) ([]model.AiAnalysisResponse, error)
}

type aiCoachServiceImpl struct {
	aiRepo  repository.AiLogRepository
	dashSvc DashboardService
	tradeRepo repository.TradeRepository
}

func NewAiCoachService(aiRepo repository.AiLogRepository, dashSvc DashboardService, tradeRepo repository.TradeRepository) AiCoachService {
	return &aiCoachServiceImpl{
		aiRepo:    aiRepo,
		dashSvc:   dashSvc,
		tradeRepo: tradeRepo,
	}
}

func (s *aiCoachServiceImpl) GetMarketData(symbol, timeframe string) (*model.MarketDataSummary, error) {
	return utils.FetchBinanceMarketData(symbol, timeframe)
}

func (s *aiCoachServiceImpl) AnalyzeChart(ctx context.Context, memberID int32, symbol, timeframe, question string, teachMode bool, fh *multipart.FileHeader) (*model.AiAnalysisResponse, error) {
	normSym := utils.NormalizeBinanceSymbol(symbol)
	normTF := utils.NormalizeTimeframe(timeframe)

	market, err := utils.FetchBinanceMarketData(normSym, normTF)
	if err != nil && fh == nil {
		return nil, err
	}

	aiText, signal, score, err := utils.AnalyzeChartAndTeach(market, fh, question, teachMode)
	if err != nil {
		return nil, err
	}

	entry := &model.AiAnalysisLog{
		MemberID:        memberID,
		Symbol:          normSym,
		Timeframe:       normTF,
		AnalysisType:    "Chart",
		PromptSummary:   question,
		AiResponse:      aiText,
		SignalDirection: signal,
		SentimentScore:  score,
	}
	_ = s.aiRepo.Create(ctx, entry)

	return &model.AiAnalysisResponse{
		ID:              entry.ID,
		Symbol:          entry.Symbol,
		Timeframe:       entry.Timeframe,
		AnalysisType:    entry.AnalysisType,
		AiResponse:      entry.AiResponse,
		SignalDirection: entry.SignalDirection,
		SentimentScore:  entry.SentimentScore,
		CreatedAt:       entry.CreatedAt.Format("02/01/2006 15:04"),
	}, nil
}

func (s *aiCoachServiceImpl) ReviewJournal(ctx context.Context, memberID int32) (*model.AiAnalysisResponse, error) {
	trades, err := s.tradeRepo.ListByMember(ctx, memberID, "", "")
	if err != nil {
		return nil, err
	}
	summary, err := s.dashSvc.GetSummary(ctx, memberID)
	if err != nil {
		return nil, err
	}

	aiText, err := utils.ReviewUserTradingHabits(trades, summary)
	if err != nil {
		return nil, err
	}

	entry := &model.AiAnalysisLog{
		MemberID:        memberID,
		Symbol:          "PORTFOLIO",
		Timeframe:       "ALL",
		AnalysisType:    "TradeJournalReview",
		PromptSummary:   "วิเคราะห์พฤติกรรมและผลการเทรดในพอร์ต",
		AiResponse:      aiText,
		SignalDirection: "NEUTRAL",
		SentimentScore:  0,
	}
	_ = s.aiRepo.Create(ctx, entry)

	return &model.AiAnalysisResponse{
		ID:              entry.ID,
		Symbol:          entry.Symbol,
		Timeframe:       entry.Timeframe,
		AnalysisType:    entry.AnalysisType,
		AiResponse:      entry.AiResponse,
		SignalDirection: entry.SignalDirection,
		SentimentScore:  entry.SentimentScore,
		CreatedAt:       entry.CreatedAt.Format("02/01/2006 15:04"),
	}, nil
}

func (s *aiCoachServiceImpl) ListLogs(ctx context.Context, memberID int32) ([]model.AiAnalysisResponse, error) {
	logs, err := s.aiRepo.ListByMember(ctx, memberID, 15)
	if err != nil {
		return nil, err
	}
	res := make([]model.AiAnalysisResponse, 0, len(logs))
	for _, l := range logs {
		res = append(res, model.AiAnalysisResponse{
			ID:              l.ID,
			Symbol:          l.Symbol,
			Timeframe:       l.Timeframe,
			AnalysisType:    l.AnalysisType,
			AiResponse:      l.AiResponse,
			SignalDirection: l.SignalDirection,
			SentimentScore:  l.SentimentScore,
			CreatedAt:       l.CreatedAt.Format("02/01/2006 15:04"),
		})
	}
	return res, nil
}
