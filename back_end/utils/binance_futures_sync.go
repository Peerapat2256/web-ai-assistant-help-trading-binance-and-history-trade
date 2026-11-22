package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const binanceFuturesBaseURL = "https://fapi.binance.com"

type SyncedBinanceTrade struct {
	ExternalOrderID string
	Symbol          string
	PositionSide    string
	Leverage        int32
	EntryPrice      float64
	ExitPrice       float64
	Quantity        float64
	MarginUsed      float64
	Pnl             float64
	PnlPercentage   float64
	Fee             float64
	TradeTime       time.Time
}

func signBinanceQuery(params url.Values, secret string) string {
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "10000")
	queryStr := params.Encode()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(queryStr))
	sig := hex.EncodeToString(mac.Sum(nil))
	return queryStr + "&signature=" + sig
}

func doBinanceSignedGet(endpoint string, params url.Values, apiKey, apiSecret string) ([]byte, error) {
	if params == nil {
		params = url.Values{}
	}
	signedQuery := signBinanceQuery(params, apiSecret)
	fullURL := fmt.Sprintf("%s%s?%s", binanceFuturesBaseURL, endpoint, signedQuery)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-MBX-APIKEY", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ Binance API ได้: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusUnavailableForLegalReasons {
		return nil, fmt.Errorf("Binance บล็อก IP ของเซิร์ฟเวอร์โซนสหรัฐฯ (HTTP 451) กรุณาตั้ง Render Region เป็น Singapore")
	}
	if resp.StatusCode != http.StatusOK {
		var errObj struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
		if json.Unmarshal(body, &errObj) == nil && errObj.Msg != "" {
			return nil, fmt.Errorf("Binance API Error (%d): %s", errObj.Code, errObj.Msg)
		}
		return nil, fmt.Errorf("Binance API HTTP %d", resp.StatusCode)
	}
	return body, nil
}

// TestBinanceFuturesConnection ทดสอบ API Key และดึงยอด Wallet Balance (USDT)
func TestBinanceFuturesConnection(apiKey, apiSecret string) (float64, error) {
	body, err := doBinanceSignedGet("/fapi/v2/account", nil, apiKey, apiSecret)
	if err != nil {
		return 0, err
	}
	var acc struct {
		TotalWalletBalance string `json:"totalWalletBalance"`
	}
	if err := json.Unmarshal(body, &acc); err != nil {
		return 0, fmt.Errorf("อ่านข้อมูลบัญชี Binance ไม่สำเร็จ")
	}
	bal, _ := strconv.ParseFloat(acc.TotalWalletBalance, 64)
	return math.Round(bal*100) / 100, nil
}

// FetchClosedFuturesTrades ดึงออเดอร์ Futures ที่ปิดแล้วทั้งหมดในช่วง lookbackDays วันล่าสุด
func FetchClosedFuturesTrades(apiKey, apiSecret string, lookbackDays int) ([]SyncedBinanceTrade, error) {
	if lookbackDays <= 0 {
		lookbackDays = 7
	}
	if lookbackDays > 90 {
		lookbackDays = 90
	}

	startMs := time.Now().Add(-time.Duration(lookbackDays) * 24 * time.Hour).UnixMilli()

	// 1) ดึงรายการ Realized PnL เพื่อหาคู่เหรียญทั้งหมดที่เพิ่งปิดออเดอร์
	incParams := url.Values{}
	incParams.Set("incomeType", "REALIZED_PNL")
	incParams.Set("startTime", strconv.FormatInt(startMs, 10))
	incParams.Set("limit", "1000")

	incBody, err := doBinanceSignedGet("/fapi/v1/income", incParams, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}

	var incomes []struct {
		Symbol string `json:"symbol"`
		Income string `json:"income"`
		Time   int64  `json:"time"`
		TranID int64  `json:"tranId"`
	}
	if err := json.Unmarshal(incBody, &incomes); err != nil {
		return nil, fmt.Errorf("อ่านประวัติ Realized PnL จาก Binance ไม่สำเร็จ")
	}
	if len(incomes) == 0 {
		return []SyncedBinanceTrade{}, nil
	}

	symbolSet := map[string]bool{}
	for _, inc := range incomes {
		if inc.Symbol != "" {
			symbolSet[inc.Symbol] = true
		}
	}

	// 2) ดึง Leverage ปัจจุบันของแต่ละคู่เหรียญ (ถ้าดึงไม่สำเร็จใช้ 10x)
	leverageMap := map[string]int32{}
	if riskBody, err := doBinanceSignedGet("/fapi/v2/positionRisk", nil, apiKey, apiSecret); err == nil {
		var risks []struct {
			Symbol   string `json:"symbol"`
			Leverage string `json:"leverage"`
		}
		if json.Unmarshal(riskBody, &risks) == nil {
			for _, r := range risks {
				if lv, err := strconv.Atoi(r.Leverage); err == nil && lv > 0 {
					leverageMap[r.Symbol] = int32(lv)
				}
			}
		}
	}

	type rawFill struct {
		Symbol       string `json:"symbol"`
		ID           int64  `json:"id"`
		OrderID      int64  `json:"orderId"`
		Side         string `json:"side"`
		PositionSide string `json:"positionSide"`
		Price        string `json:"price"`
		Qty          string `json:"qty"`
		RealizedPnl  string `json:"realizedPnl"`
		Commission   string `json:"commission"`
		Time         int64  `json:"time"`
	}

	type orderAgg struct {
		OrderID      int64
		Symbol       string
		PositionSide string
		TotalQty     float64
		WeightedExit float64
		TotalPnl     float64
		TotalFee     float64
		LastTimeMs   int64
	}

	// 3) ดึง userTrades ของแต่ละคู่เหรียญ และรวมการปิดออเดอร์ที่มี OrderID เดียวกัน
	results := []SyncedBinanceTrade{}
	for sym := range symbolSet {
		tParams := url.Values{}
		tParams.Set("symbol", sym)
		tParams.Set("startTime", strconv.FormatInt(startMs, 10))
		tParams.Set("limit", "1000")

		tBody, err := doBinanceSignedGet("/fapi/v1/userTrades", tParams, apiKey, apiSecret)
		if err != nil {
			continue
		}
		var fills []rawFill
		if json.Unmarshal(tBody, &fills) != nil {
			continue
		}

		aggMap := map[int64]*orderAgg{}
		for _, f := range fills {
			pnl, _ := strconv.ParseFloat(f.RealizedPnl, 64)
			if pnl == 0 {
				continue // ข้ามไม้เปิด (Opening Fill) นับเฉพาะไม้ที่ปิดและเกิด Realized PnL
			}
			price, _ := strconv.ParseFloat(f.Price, 64)
			qty, _ := strconv.ParseFloat(f.Qty, 64)
			fee, _ := strconv.ParseFloat(f.Commission, 64)

			posSide := strings.ToUpper(strings.TrimSpace(f.PositionSide))
			if posSide != "LONG" && posSide != "SHORT" {
				// One-Way Mode: ปิดด้วย SELL = เดิมถือ LONG, ปิดด้วย BUY = เดิมถือ SHORT
				if strings.ToUpper(f.Side) == "SELL" {
					posSide = "LONG"
				} else {
					posSide = "SHORT"
				}
			}

			agg, exists := aggMap[f.OrderID]
			if !exists {
				agg = &orderAgg{
					OrderID:      f.OrderID,
					Symbol:       NormalizeBinanceSymbol(f.Symbol),
					PositionSide: posSide,
					LastTimeMs:   f.Time,
				}
				aggMap[f.OrderID] = agg
			}
			agg.TotalQty += qty
			agg.WeightedExit += price * qty
			agg.TotalPnl += pnl
			agg.TotalFee += math.Abs(fee)
			if f.Time > agg.LastTimeMs {
				agg.LastTimeMs = f.Time
			}
		}

		for _, agg := range aggMap {
			if agg.TotalQty <= 0 {
				continue
			}
			exitPrice := agg.WeightedExit / agg.TotalQty
			entryPrice := exitPrice
			if agg.PositionSide == "LONG" {
				entryPrice = exitPrice - (agg.TotalPnl / agg.TotalQty)
			} else {
				entryPrice = exitPrice + (agg.TotalPnl / agg.TotalQty)
			}
			if entryPrice < 0 {
				entryPrice = exitPrice
			}

			lev := leverageMap[sym]
			if lev <= 0 {
				lev = 10
			}
			notional := entryPrice * agg.TotalQty
			marginUsed := notional / float64(lev)
			roiPct := 0.0
			if marginUsed > 0 {
				roiPct = (agg.TotalPnl / marginUsed) * 100
			}

			results = append(results, SyncedBinanceTrade{
				ExternalOrderID: fmt.Sprintf("BNF-%s-%d", agg.Symbol, agg.OrderID),
				Symbol:          agg.Symbol,
				PositionSide:    agg.PositionSide,
				Leverage:        lev,
				EntryPrice:      math.Round(entryPrice*10000) / 10000,
				ExitPrice:       math.Round(exitPrice*10000) / 10000,
				Quantity:        math.Round(agg.TotalQty*1000000) / 1000000,
				MarginUsed:      math.Round(marginUsed*100) / 100,
				Pnl:             math.Round(agg.TotalPnl*100) / 100,
				PnlPercentage:   math.Round(roiPct*100) / 100,
				Fee:             math.Round(agg.TotalFee*10000) / 10000,
				TradeTime:       time.UnixMilli(agg.LastTimeMs).In(time.FixedZone("ICT", 7*3600)),
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].TradeTime.Before(results[j].TradeTime)
	})
	return results, nil
}
