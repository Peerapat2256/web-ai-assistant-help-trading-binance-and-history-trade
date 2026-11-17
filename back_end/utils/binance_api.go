package utils

import (
	"backend/model/model"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func NormalizeBinanceSymbol(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "-", "")
	if s == "" {
		return "BTCUSDT"
	}
	if !strings.HasSuffix(s, "USDT") && !strings.HasSuffix(s, "BUSD") && !strings.HasSuffix(s, "FDUSD") {
		s += "USDT"
	}
	return s
}

func NormalizeTimeframe(tf string) string {
	allowed := map[string]bool{
		"5m": true, "15m": true, "30m": true, "1h": true, "4h": true, "1d": true,
	}
	clean := strings.ToLower(strings.TrimSpace(tf))
	if allowed[clean] {
		return clean
	}
	return "1h"
}

func FetchBinanceMarketData(symbol, timeframe string) (*model.MarketDataSummary, error) {
	normSym := NormalizeBinanceSymbol(symbol)
	normTF := NormalizeTimeframe(timeframe)

	client := &http.Client{Timeout: 12 * time.Second}

	klineURL := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&limit=80", normSym, normTF)
	resp, err := client.Get(klineURL)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถดึงข้อมูลกราฟจาก Binance ได้: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Binance API ตอบกลับสถานะ %d (ตรวจสอบชื่อคู่เหรียญ %s)", resp.StatusCode, normSym)
	}

	var rawKlines [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawKlines); err != nil {
		return nil, fmt.Errorf("อ่านข้อมูลแท่งเทียนไม่สำเร็จ: %v", err)
	}

	klines := make([]model.MarketKline, 0, len(rawKlines))
	closes := make([]float64, 0, len(rawKlines))

	for _, item := range rawKlines {
		if len(item) < 6 {
			continue
		}
		openTimeMs, _ := item[0].(float64)
		openVal := parseFloatInterface(item[1])
		highVal := parseFloatInterface(item[2])
		lowVal := parseFloatInterface(item[3])
		closeVal := parseFloatInterface(item[4])
		volVal := parseFloatInterface(item[5])

		t := time.UnixMilli(int64(openTimeMs)).In(time.FixedZone("ICT", 7*3600))
		klines = append(klines, model.MarketKline{
			OpenTime: int64(openTimeMs),
			TimeStr:  t.Format("02/01 15:04"),
			Open:     openVal,
			High:     highVal,
			Low:      lowVal,
			Close:    closeVal,
			Volume:   volVal,
		})
		closes = append(closes, closeVal)
	}

	if len(klines) == 0 {
		return nil, fmt.Errorf("ไม่พบข้อมูลแท่งเทียนสำหรับ %s", normSym)
	}

	tickerURL := fmt.Sprintf("https://api.binance.com/api/v3/ticker/24hr?symbol=%s", normSym)
	tResp, err := client.Get(tickerURL)
	var changePct, high24, low24, vol24 float64
	if err == nil {
		defer tResp.Body.Close()
		var ticker struct {
			PriceChangePercent string `json:"priceChangePercent"`
			HighPrice          string `json:"highPrice"`
			LowPrice           string `json:"lowPrice"`
			QuoteVolume        string `json:"quoteVolume"`
		}
		if json.NewDecoder(tResp.Body).Decode(&ticker) == nil {
			changePct, _ = strconv.ParseFloat(ticker.PriceChangePercent, 64)
			high24, _ = strconv.ParseFloat(ticker.HighPrice, 64)
			low24, _ = strconv.ParseFloat(ticker.LowPrice, 64)
			vol24, _ = strconv.ParseFloat(ticker.QuoteVolume, 64)
		}
	}

	currentPrice := closes[len(closes)-1]
	rsi14 := calculateRSI(closes, 14)
	ema20 := calculateEMA(closes, 20)
	ema50 := calculateEMA(closes, 50)
	support, resistance := calculateSupportResistance(klines, 24)

	return &model.MarketDataSummary{
		Symbol:       normSym,
		Timeframe:    normTF,
		CurrentPrice: currentPrice,
		Change24hPct: math.Round(changePct*100) / 100,
		High24h:      high24,
		Low24h:       low24,
		Volume24h:    vol24,
		RSI14:        math.Round(rsi14*100) / 100,
		EMA20:        math.Round(ema20*10000) / 10000,
		EMA50:        math.Round(ema50*10000) / 10000,
		Support:      support,
		Resistance:   resistance,
		Klines:       klines,
	}, nil
}

func parseFloatInterface(val interface{}) float64 {
	s, ok := val.(string)
	if !ok {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func calculateEMA(prices []float64, period int) float64 {
	if len(prices) == 0 {
		return 0
	}
	if len(prices) < period {
		sum := 0.0
		for _, p := range prices {
			sum += p
		}
		return sum / float64(len(prices))
	}

	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	ema := sum / float64(period)
	k := 2.0 / float64(period+1)
	for i := period; i < len(prices); i++ {
		ema = prices[i]*k + ema*(1-k)
	}
	return ema
}

func calculateRSI(prices []float64, period int) float64 {
	if len(prices) <= period {
		return 50
	}
	gains := 0.0
	losses := 0.0
	for i := 1; i <= period; i++ {
		diff := prices[i] - prices[i-1]
		if diff >= 0 {
			gains += diff
		} else {
			losses -= diff
		}
	}
	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	for i := period + 1; i < len(prices); i++ {
		diff := prices[i] - prices[i-1]
		if diff >= 0 {
			avgGain = (avgGain*float64(period-1) + diff) / float64(period)
			avgLoss = (avgLoss * float64(period-1)) / float64(period)
		} else {
			avgGain = (avgGain * float64(period-1)) / float64(period)
			avgLoss = (avgLoss*float64(period-1) - diff) / float64(period)
		}
	}

	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

func calculateSupportResistance(klines []model.MarketKline, lookback int) (float64, float64) {
	start := len(klines) - lookback
	if start < 0 {
		start = 0
	}
	minLow := klines[start].Low
	maxHigh := klines[start].High
	for i := start; i < len(klines); i++ {
		if klines[i].Low < minLow {
			minLow = klines[i].Low
		}
		if klines[i].High > maxHigh {
			maxHigh = klines[i].High
		}
	}
	return minLow, maxHigh
}
