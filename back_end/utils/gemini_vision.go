package utils

import (
	"backend/model/model"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

var (
	geminiModelLock    sync.RWMutex
	geminiWorkingModel string
	geminiCachedModels []string
	geminiLastFetch    time.Time
)

type geminiModelInfo struct {
	Name                       string   `json:"name"`
	SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
}

type geminiListModelsResp struct {
	Models []geminiModelInfo `json:"models"`
}

func getCandidateGeminiModels(apiKey string) []string {
	geminiModelLock.RLock()
	if time.Since(geminiLastFetch) < 30*time.Minute && len(geminiCachedModels) > 0 {
		models := make([]string, len(geminiCachedModels))
		copy(models, geminiCachedModels)
		curr := geminiWorkingModel
		geminiModelLock.RUnlock()

		if curr != "" {
			for i, m := range models {
				if m == curr {
					models = append([]string{curr}, append(models[:i], models[i+1:]...)...)
					break
				}
			}
		}
		return models
	}
	geminiModelLock.RUnlock()

	defaultFallbacks := []string{
		"gemini-3.8-flash",
		"gemini-2.5-flash",
		"gemini-1.5-flash",
		"gemini-1.5-flash-latest",
		"gemini-3.8-pro",
		"gemini-2.5-pro",
		"gemini-1.5-pro",
		"gemini-flash-latest",
		"gemini-pro-latest",
		"gemini-2.0-flash-exp",
	}

	// Dynamic Discovery: ค้นหาโมเดลที่มีให้ใช้จริงจากบัญชี Google AI Studio
	client := &http.Client{Timeout: 6 * time.Second}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)
	req, err := http.NewRequest("GET", url, nil)
	discovered := []string{}
	if err == nil {
		req.Header.Set("x-goog-api-key", apiKey)
		if resp, err := client.Do(req); err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var listResp geminiListModelsResp
				if body, err := io.ReadAll(resp.Body); err == nil {
					if json.Unmarshal(body, &listResp) == nil {
						for _, m := range listResp.Models {
							canGen := false
							for _, method := range m.SupportedGenerationMethods {
								if method == "generateContent" {
									canGen = true
									break
								}
							}
							if !canGen {
								continue
							}
							cleanName := strings.TrimPrefix(m.Name, "models/")
							if strings.HasPrefix(cleanName, "gemini") {
								discovered = append(discovered, cleanName)
							}
						}
					}
				}
			}
		}
	}

	geminiModelLock.Lock()
	defer geminiModelLock.Unlock()

	merged := []string{}
	seen := make(map[string]bool)

	// เรียงลำดับโมเดล: นำตระกูล Flash มาก่อน ตามด้วย Pro เพื่อความเร็วและประหยัด quota
	sort.SliceStable(discovered, func(i, j int) bool {
		flashI := strings.Contains(discovered[i], "flash")
		flashJ := strings.Contains(discovered[j], "flash")
		if flashI && !flashJ {
			return true
		}
		if !flashI && flashJ {
			return false
		}
		return discovered[i] > discovered[j]
	})

	for _, m := range discovered {
		if !seen[m] {
			seen[m] = true
			merged = append(merged, m)
		}
	}
	for _, m := range defaultFallbacks {
		if !seen[m] {
			seen[m] = true
			merged = append(merged, m)
		}
	}

	if len(merged) > 0 {
		geminiCachedModels = merged
		geminiLastFetch = time.Now()
	}

	result := make([]string, len(merged))
	copy(result, merged)
	if geminiWorkingModel != "" {
		for i, m := range result {
			if m == geminiWorkingModel {
				result = append([]string{geminiWorkingModel}, append(result[:i], result[i+1:]...)...)
				break
			}
		}
	}
	return result
}

func callGemini(prompt string, imageBase64 string, mimeType string) (string, error) {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apiKey == "" {
		return "", fmt.Errorf("ยังไม่ได้ตั้งค่า GEMINI_API_KEY")
	}

	parts := []geminiPart{{Text: prompt}}
	if imageBase64 != "" {
		if mimeType == "" {
			mimeType = "image/jpeg"
		}
		parts = append(parts, geminiPart{
			InlineData: &geminiInlineData{
				MimeType: mimeType,
				Data:     imageBase64,
			},
		})
	}

	reqBody := geminiRequest{
		Contents: []geminiContent{{Parts: parts}},
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	models := getCandidateGeminiModels(apiKey)
	client := &http.Client{Timeout: 35 * time.Second}

	var lastErr error
	for _, m := range models {
		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", m, apiKey)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-goog-api-key", apiKey)

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			log.Printf("[Gemini Auto] Model %s network error: %v, trying next...", m, err)
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var gemResp geminiResponse
		if err := json.Unmarshal(bodyBytes, &gemResp); err != nil {
			lastErr = fmt.Errorf("อ่านผลลัพธ์จาก AI ไม่สำเร็จ")
			continue
		}
		if gemResp.Error != nil {
			lastErr = fmt.Errorf("Gemini API (%s) error (%d): %s", m, gemResp.Error.Code, gemResp.Error.Message)
			log.Printf("[Gemini Auto] Model %s failed (%d: %s), auto-switching...", m, gemResp.Error.Code, gemResp.Error.Message)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("Gemini (%s) HTTP %d", m, resp.StatusCode)
			log.Printf("[Gemini Auto] Model %s returned HTTP %d, auto-switching...", m, resp.StatusCode)
			continue
		}
		if len(gemResp.Candidates) > 0 && len(gemResp.Candidates[0].Content.Parts) > 0 {
			// จำโมเดลที่ทำงานได้จริงไว้ใช้ต่อทันที ไม่ต้องวนหาใหม่ทุกครั้ง
			geminiModelLock.Lock()
			geminiWorkingModel = m
			geminiModelLock.Unlock()
			log.Printf("[Gemini Auto] Successfully connected and used model: %s", m)
			return gemResp.Candidates[0].Content.Parts[0].Text, nil
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("ไม่สามารถติดต่อ Gemini API ได้จากทุกโมเดลที่ทดสอบ")
	}
	return "", lastErr
}

func cleanJSONBlock(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	startArr := strings.Index(s, "[")
	endArr := strings.LastIndex(s, "]")
	if startArr >= 0 && endArr > startArr {
		return s[startArr : endArr+1]
	}
	return strings.TrimSpace(s)
}

func readMultipartBase64(fh *multipart.FileHeader) (string, string, error) {
	file, err := fh.Open()
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", "", err
	}
	mime := fh.Header.Get("Content-Type")
	if mime == "" {
		mime = http.DetectContentType(data)
	}
	return base64.StdEncoding.EncodeToString(data), mime, nil
}

func normalizeParsedTrade(p *model.ParsedScreenshotData, savedURL string) {
	if p.Symbol == "" {
		p.Symbol = "BTCUSDT"
	}
	p.Symbol = NormalizeBinanceSymbol(p.Symbol)

	p.PositionSide = strings.ToUpper(strings.TrimSpace(p.PositionSide))
	if p.PositionSide == "S" || p.PositionSide == "SELL" {
		p.PositionSide = "SHORT"
	} else if p.PositionSide == "B" || p.PositionSide == "BUY" {
		p.PositionSide = "LONG"
	}
	if p.PositionSide == "" || (p.PositionSide != "LONG" && p.PositionSide != "SHORT" && p.PositionSide != "SPOT") {
		if p.EntryPrice > 0 && p.ExitPrice > 0 && p.Pnl != 0 {
			priceUp := p.ExitPrice > p.EntryPrice
			if (priceUp && p.Pnl > 0) || (!priceUp && p.Pnl < 0) {
				p.PositionSide = "LONG"
			} else {
				p.PositionSide = "SHORT"
			}
		} else {
			p.PositionSide = "LONG"
		}
	}

	if p.MarginUsed <= 0 && p.Pnl != 0 && p.PnlPercentage != 0 {
		p.MarginUsed = math.Round((math.Abs(p.Pnl)/(math.Abs(p.PnlPercentage)/100))*100) / 100
	}
	if p.Leverage <= 0 {
		if p.MarginUsed > 0 && p.EntryPrice > 0 && p.Quantity > 0 {
			estLev := math.Round((p.EntryPrice * p.Quantity) / p.MarginUsed)
			if estLev < 1 {
				estLev = 1
			}
			p.Leverage = int32(estLev)
		} else {
			p.Leverage = 10
		}
	}

	if strings.TrimSpace(p.TradeTime) == "" {
		p.TradeTime = time.Now().Format("2006-01-02 15:04")
	}
	p.ScreenshotURL = savedURL
	p.Confidence = "ai_verified"
}

func ParseBinanceScreenshot(fh *multipart.FileHeader, savedURL string) (*model.ParsedScreenshotBatchResponse, error) {
	b64, mime, err := readMultipartBase64(fh)
	if err != nil {
		return nil, err
	}

	prompt := `คุณคือผู้เชี่ยวชาญการอ่านภาพหน้าจอการเทรดจากแอปและเว็บ Binance (ทั้งหน้า PnL Card เดี่ยว, Futures Position History ที่มีหลายรายการ, Order History และ Spot Trade)

คำแนะนำสำคัญ:
1. ตรวจสอบก่อนว่าภาพนี้มี "1 ออเดอร์" หรือ "หลายออเดอร์" (เช่น หน้า Position History ที่เรียงลงมาหลายบล็อก)
   - หากมีหลายออเดอร์ จงสกัดข้อมูลออกมา **ทุกออเดอร์ที่มองเห็นในภาพ** เรียงจากบนลงล่างใส่ในอาเรย์ "trades" และตั้ง "mode": "multi"
   - หากมีแค่ออเดอร์เดียว ให้ใส่ 1 รายการในอาเรย์ "trades" และตั้ง "mode": "single"
2. วิธีอ่านหน้า Position History ของ Binance:
   - สัญลักษณ์ตัว [S] สีแดงข้างชื่อเหรียญ = "SHORT", ตัว [B] สีเขียว = "LONG"
   - ป้าย "Cross 90X" หรือ "Isolated 20X" คือ leverage (เช่น 90 หรือ 20)
   - ช่อง "Realized PNL" หรือ "Closing PNL" คือ pnl (ใส่เครื่องหมายลบถ้าติดลบ เช่น -114.17)
   - ช่อง "ROI" หรือ "ROE" คือ pnl_percentage (เป็น % เช่น 63.43 หรือ -105.11)
   - ช่อง "Closed Vol." คือ quantity (เช่น 0.132)
   - ช่อง "Avg. Entry Price" หรือ "Entry Price" คือ entry_price และ "Avg. Close Price" คือ exit_price
   - วัน-เวลา (trade_time): ใช้เวลาจากช่อง "Closed" เป็นหลัก (รูปแบบ YYYY-MM-DD HH:MM:SS หรือ YYYY-MM-DD HH:MM)

จงตอบกลับเป็น JSON ที่ถูกต้องเท่านั้น ห้ามมีข้อความอื่นปน:
{
  "mode": "single หรือ multi",
  "raw_summary": "สรุปสั้นๆ เป็นภาษาไทย เช่น ตรวจพบประวัติการเทรด 3 ออเดอร์จากหน้า Position History",
  "trades": [
    {
      "symbol": "BTCUSDT",
      "position_side": "LONG หรือ SHORT หรือ SPOT",
      "leverage": 90,
      "entry_price": 0.0,
      "exit_price": 0.0,
      "quantity": 0.0,
      "margin_used": 0.0,
      "pnl": 0.0,
      "pnl_percentage": 0.0,
      "fee": 0.0,
      "trade_time": "YYYY-MM-DD HH:MM:SS"
    }
  ]
}`

	fallbackSingle := func(summary string) *model.ParsedScreenshotBatchResponse {
		item := model.ParsedScreenshotData{
			Symbol:        "BTCUSDT",
			PositionSide:  "LONG",
			Leverage:      10,
			TradeTime:     time.Now().Format("2006-01-02 15:04"),
			ScreenshotURL: savedURL,
			Confidence:    "manual_review",
			RawSummary:    summary,
		}
		return &model.ParsedScreenshotBatchResponse{
			Mode:          "single",
			ScreenshotURL: savedURL,
			Confidence:    "manual_review",
			RawSummary:    summary,
			TotalFound:    1,
			Trades:        []model.ParsedScreenshotData{item},
		}
	}

	text, err := callGemini(prompt, b64, mime)
	if err != nil {
		log.Printf("[Gemini Vision] ParseBinanceScreenshot fallback: %v", err)
		return fallbackSingle(fmt.Sprintf("อัปโหลดรูปสำเร็จ แต่ระบบ AI Vision ยังไม่สามารถสแกนอัตโนมัติได้ในขณะนี้ (%v) กรุณากรอกหรือตรวจสอบตัวเลขก่อนบันทึก", err)), nil
	}

	type aiMultiResp struct {
		Mode       string                       `json:"mode"`
		RawSummary string                       `json:"raw_summary"`
		Trades     []model.ParsedScreenshotData `json:"trades"`
	}

	cleaned := cleanJSONBlock(text)
	var multi aiMultiResp
	if err := json.Unmarshal([]byte(cleaned), &multi); err == nil && len(multi.Trades) > 0 {
		for i := range multi.Trades {
			normalizeParsedTrade(&multi.Trades[i], savedURL)
		}
		mode := strings.ToLower(strings.TrimSpace(multi.Mode))
		if len(multi.Trades) > 1 {
			mode = "multi"
		} else if mode != "multi" {
			mode = "single"
		}
		summary := strings.TrimSpace(multi.RawSummary)
		if summary == "" {
			if mode == "multi" {
				summary = fmt.Sprintf("ตรวจพบประวัติการเทรดหลายออเดอร์ทั้งหมด %d รายการ", len(multi.Trades))
			} else {
				summary = "AI สแกนข้อมูลการเทรดสำเร็จ กรุณาตรวจสอบตัวเลขก่อนบันทึก"
			}
		}
		return &model.ParsedScreenshotBatchResponse{
			Mode:          mode,
			ScreenshotURL: savedURL,
			Confidence:    "ai_verified",
			RawSummary:    summary,
			TotalFound:    len(multi.Trades),
			Trades:        multi.Trades,
		}, nil
	}

	// Fallback หาก AI ตอบกลับมาเป็น JSON object เดี่ยว
	var single model.ParsedScreenshotData
	if err := json.Unmarshal([]byte(cleaned), &single); err == nil && (single.Symbol != "" || single.Pnl != 0 || single.EntryPrice != 0) {
		normalizeParsedTrade(&single, savedURL)
		summary := strings.TrimSpace(single.RawSummary)
		if summary == "" {
			summary = "AI สแกนข้อมูลการเทรดสำเร็จ กรุณาตรวจสอบตัวเลขก่อนบันทึก"
		}
		return &model.ParsedScreenshotBatchResponse{
			Mode:          "single",
			ScreenshotURL: savedURL,
			Confidence:    "ai_verified",
			RawSummary:    summary,
			TotalFound:    1,
			Trades:        []model.ParsedScreenshotData{single},
		}, nil
	}

	return fallbackSingle("AI อ่านข้อมูลบางส่วนได้ไม่สมบูรณ์ กรุณาตรวจสอบตัวเลขก่อนบันทึก"), nil
}

func ParseWithdrawalScreenshot(fh *multipart.FileHeader, savedURL string) (*model.ParsedWithdrawalData, error) {
	b64, mime, err := readMultipartBase64(fh)
	if err != nil {
		return nil, err
	}

	prompt := `คุณคือผู้เชี่ยวชาญการอ่านภาพหน้าจอ "รายละเอียดการถอนเงิน" (Withdrawal Details) จากแอป Binance TH / Binance / Bitkub / แอปธนาคาร
คำแนะนำในการอ่าน:
- จำนวนเงิน (amount): อ่านเป็นตัวเลขบวกเสมอ เช่น ในภาพเขียน "-10,232.09 THB" ให้ตอบ amount เป็น 10232.09 และ currency เป็น "THB"
- สถานะ (status): เช่น "สำเร็จ", "รอดำเนินการ", "ล้มเหลว" (หากภาษาอังกฤษ Completed/Success ให้ตอบ "สำเร็จ")
- วันที่ (withdrawal_time): รูปแบบ "YYYY-MM-DD HH:MM:SS" เช่น "2026-11-21 13:57:15"
- วิธีการชำระเงิน (payment_method): เช่น "BANK_TRANSFER", "PROMPTPAY", "P2P"
- หมายเหตุการทำธุรกรรม (notes): หากเป็น "--" หรือ "-" หรือว่าง ให้ใส่ ""
- หมายเลขธุรกรรม (transaction_id): รหัสอ้างอิงทั้งหมด เช่น "W1330793249660076034"

จงตอบกลับเป็น JSON ที่ถูกต้องเท่านั้น ห้ามมีข้อความอื่นปน:
{
  "transaction_id": "W1330793249660076034",
  "amount": 10232.09,
  "currency": "THB",
  "status": "สำเร็จ",
  "payment_method": "BANK_TRANSFER",
  "withdrawal_time": "2026-11-21 13:57:15",
  "notes": "",
  "raw_summary": "สรุปสั้นๆ เช่น สแกนสลิปถอนเงิน 10,232.09 THB สำเร็จ"
}`

	text, err := callGemini(prompt, b64, mime)
	if err != nil {
		log.Printf("[Gemini Vision] ParseWithdrawalScreenshot fallback: %v", err)
		return &model.ParsedWithdrawalData{
			Currency:       "THB",
			Status:         "สำเร็จ",
			PaymentMethod:  "BANK_TRANSFER",
			WithdrawalTime: time.Now().Format("2006-01-02 15:04:05"),
			ScreenshotURL:  savedURL,
			Confidence:     "manual_review",
			RawSummary:     fmt.Sprintf("อัปโหลดรูปสำเร็จ แต่ AI ไม่สามารถอ่านอัตโนมัติได้ (%v) กรุณากรอกข้อมูลเอง", err),
		}, nil
	}

	var parsed model.ParsedWithdrawalData
	cleaned := cleanJSONBlock(text)
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		return &model.ParsedWithdrawalData{
			Currency:       "THB",
			Status:         "สำเร็จ",
			PaymentMethod:  "BANK_TRANSFER",
			WithdrawalTime: time.Now().Format("2006-01-02 15:04:05"),
			ScreenshotURL:  savedURL,
			Confidence:     "manual_review",
			RawSummary:     "AI อ่านข้อมูลบางส่วนได้ไม่สมบูรณ์ กรุณาตรวจสอบตัวเลขก่อนบันทึก",
		}, nil
	}

	parsed.Amount = math.Abs(parsed.Amount)
	if strings.TrimSpace(parsed.Currency) == "" {
		parsed.Currency = "THB"
	}
	parsed.Currency = strings.ToUpper(strings.TrimSpace(parsed.Currency))
	if strings.TrimSpace(parsed.Status) == "" {
		parsed.Status = "สำเร็จ"
	}
	if strings.TrimSpace(parsed.PaymentMethod) == "" {
		parsed.PaymentMethod = "BANK_TRANSFER"
	}
	if parsed.Notes == "--" || parsed.Notes == "-" {
		parsed.Notes = ""
	}
	if strings.TrimSpace(parsed.WithdrawalTime) == "" {
		parsed.WithdrawalTime = time.Now().Format("2006-01-02 15:04:05")
	}
	parsed.TransactionID = strings.TrimSpace(parsed.TransactionID)
	parsed.ScreenshotURL = savedURL
	parsed.Confidence = "ai_verified"
	if strings.TrimSpace(parsed.RawSummary) == "" {
		parsed.RawSummary = fmt.Sprintf("สแกนรายละเอียดการถอน %.2f %s สำเร็จ", parsed.Amount, parsed.Currency)
	}
	return &parsed, nil
}

func AnalyzeChartAndTeach(market *model.MarketDataSummary, fh *multipart.FileHeader, userQuestion string, teachMode bool) (string, string, float64, error) {
	var b64, mime string
	if fh != nil {
		b64, mime, _ = readMultipartBase64(fh)
	}

	signal := "NEUTRAL"
	score := 0.0
	if market != nil {
		if market.CurrentPrice > market.EMA20 && market.EMA20 > market.EMA50 && market.RSI14 < 70 {
			signal = "BULLISH"
			score = 0.65
		} else if market.CurrentPrice < market.EMA20 && market.EMA20 < market.EMA50 && market.RSI14 > 30 {
			signal = "BEARISH"
			score = -0.65
		}
	}

	teachInstruction := ""
	if teachMode {
		teachInstruction = "\n5. โหมดครูสอนเทรด: อธิบายอย่างละเอียดเป็นกันเองว่าค่า RSI, EMA20, EMA50 และโครงสร้างราคาในตอนนี้หมายความว่าอย่างไร และเทรดเดอร์ควรระวังจุดไหนเพื่อไม่ให้ติดดอยหรือโดนหลอก (Fakeout)"
	}

	marketCtx := ""
	if market != nil {
		marketCtx = fmt.Sprintf(
			"ข้อมูลตลาดสดจาก Binance:\n- คู่เหรียญ: %s (Timeframe: %s)\n- ราคาปัจจุบัน: %.4f\n- เปลี่ยนแปลง 24 ชม.: %.2f%%\n- สูงสุด/ต่ำสุด 24 ชม.: %.4f / %.4f\n- RSI (14): %.2f\n- EMA 20: %.4f\n- EMA 50: %.4f\n- แนวรับล่าสุด: %.4f\n- แนวต้านล่าสุด: %.4f\n",
			market.Symbol, market.Timeframe, market.CurrentPrice, market.Change24hPct,
			market.High24h, market.Low24h, market.RSI14, market.EMA20, market.EMA50,
			market.Support, market.Resistance,
		)
	}

	qText := strings.TrimSpace(userQuestion)
	if qText == "" {
		qText = "ช่วยวิเคราะห์แนวโน้มและวางแผนการเทรดให้หน่อย"
	}

	prompt := fmt.Sprintf(`คุณคือโค้ชสอนเทรดคริปโตมืออาชีพ พูดคุยเป็นภาษาไทยที่เข้าใจง่าย ชัดเจน ไม่ใช้ศัพท์ซับซ้อนเกินไป
%s
คำถามจากผู้ใช้งาน: %s

กรุณาวิเคราะห์และจัดรูปแบบหัวข้อดังนี้:
1. ภาพรวมแนวโน้มหลัก (Trend Direction)
2. โซนแนวรับ - แนวต้านสำคัญที่ต้องจับตา
3. แผนการเทรดตัวอย่าง (จุดพิจารณาเข้า, จุดตัดขาดทุน Stop Loss, เป้าทำกำไร Take Profit และ Risk:Reward)
4. ข้อควรระวังในการบริหารความเสี่ยง (Money Management)%s`, marketCtx, qText, teachInstruction)

	aiText, err := callGemini(prompt, b64, mime)
	if err != nil {
		log.Printf("[Gemini Chart] Fallback rule-based analysis: %v", err)
		aiText = buildFallbackChartAnalysis(market, signal, teachMode)
	} else {
		lower := strings.ToLower(aiText)
		if strings.Contains(lower, "ขาขึ้น") || strings.Contains(lower, "bullish") || strings.Contains(lower, "long") {
			signal = "BULLISH"
			score = 0.7
		} else if strings.Contains(lower, "ขาลง") || strings.Contains(lower, "bearish") || strings.Contains(lower, "short") {
			signal = "BEARISH"
			score = -0.7
		}
	}

	return aiText, signal, score, nil
}

func buildFallbackChartAnalysis(market *model.MarketDataSummary, signal string, teachMode bool) string {
	if market == nil {
		return "ไม่สามารถดึงข้อมูลกราฟได้ในขณะนี้ กรุณาลองใหม่อีกครั้ง"
	}

	trendThai := "ไซด์เวย์ (แกว่งตัวออกข้าง)"
	plan := fmt.Sprintf("รอราคาทดสอบแนวรับ %.4f หรือทะลุแนวต้าน %.4f อย่างชัดเจนก่อนตัดสินใจ", market.Support, market.Resistance)
	switch signal {
	case "BULLISH":
		trendThai = "แนวโน้มขาขึ้นระยะสั้น (ราคาอยู่เหนือ EMA20 และ EMA50)"
		plan = fmt.Sprintf("พิจารณาจังหวะย่อตัวใกล้โซน EMA20 (%.4f) เพื่อหาโอกาส Long โดยวาง Stop Loss ใต้แนวรับ %.4f และเป้าทำกำไรที่แนวต้าน %.4f", market.EMA20, market.Support, market.Resistance)
	case "BEARISH":
		trendThai = "แนวโน้มขาลงระยะสั้น (ราคาอยู่ต่ำกว่า EMA20 และ EMA50)"
		plan = fmt.Sprintf("พิจารณาจังหวะเด้งทดสอบโซน EMA20 (%.4f) เพื่อหาโอกาส Short โดยวาง Stop Loss เหนือแนวต้าน %.4f และเป้าทำกำไรที่แนวรับ %.4f", market.EMA20, market.Resistance, market.Support)
	}

	rsiNote := "RSI อยู่ในโซนสมดุล"
	if market.RSI14 >= 70 {
		rsiNote = "RSI เข้าเขต Overbought (ซื้อมากเกินไป) ระวังแรงเทขายทำกำไรระยะสั้น"
	} else if market.RSI14 <= 30 {
		rsiNote = "RSI เข้าเขต Oversold (ขายมากเกินไป) มีโอกาสเกิดแรงดีดกลับ (Rebound)"
	}

	res := fmt.Sprintf(`### วิเคราะห์กราฟ %s (Timeframe %s)
1. **ภาพรวมแนวโน้ม:** %s (ราคาปัจจุบัน %.4f, เปลี่ยนแปลง 24 ชม. %.2f%%)
2. **แนวรับ - แนวต้านสำคัญ:**
   - แนวรับ: %.4f
   - แนวต้าน: %.4f
3. **แผนการเทรดเบื้องต้น:**
   - %s
4. **ความแข็งแรงโมเมนตัม (RSI 14 = %.2f):**
   - %s`,
		market.Symbol, market.Timeframe, trendThai, market.CurrentPrice, market.Change24hPct,
		market.Support, market.Resistance, plan, market.RSI14, rsiNote)

	if teachMode {
		res += fmt.Sprintf(`

### บทเรียนสำหรับเทรดเดอร์ (โหมดสอนเทรด)
- **EMA20 (%.4f) vs EMA50 (%.4f):** เส้นค่าเฉลี่ยช่วยบอกทิศทางเทรนด์ ถ้าราคายืนเหนือทั้ง 2 เส้นแสดงว่าฝั่งซื้อคุมตลาด แต่หากเข้าเทรดกลางทางโดยไม่รอจังหวะย่อ อาจจะโดนสะบัดกิน Stop Loss ได้ง่าย
- **การบริหารความเสี่ยง:** ก่อนกดเปิดออเดอร์ทุกครั้ง แนะนำให้คำนวณขนาดไม้จากเครื่องมือคำนวณความเสี่ยง เพื่อให้เสียไม่เกิน 1-2%% ของพอร์ตหากราคาผิดทาง`, market.EMA20, market.EMA50)
	}
	return res
}

func ReviewUserTradingHabits(trades []model.TradeHistory, summary *model.DashboardSummaryResponse) (string, error) {
	if len(trades) == 0 {
		return "ยังไม่มีประวัติการเทรดในระบบ กรุณาบันทึกการเทรดอย่างน้อย 3-5 ไม้ก่อน เพื่อให้ AI ช่วยวิเคราะห์จุดแข็งและจุดอ่อนในพอร์ตของคุณครับ", nil
	}

	limit := len(trades)
	if limit > 25 {
		limit = 25
	}
	var sb strings.Builder
	for i := 0; i < limit; i++ {
		t := trades[i]
		sb.WriteString(fmt.Sprintf("- %s %s %dx | PnL: %.2f USDT (%.2f%%) | กลยุทธ์: %s | อารมณ์: %s\n",
			t.Symbol, t.PositionSide, t.Leverage, t.Pnl, t.PnlPercentage, t.SetupTag, t.EmotionTag))
	}

	prompt := fmt.Sprintf(`คุณคือโค้ชจิตวิทยาและการบริหารพอร์ตเทรดคริปโต จงวิเคราะห์ประวัติการเทรดของผู้ใช้ต่อไปนี้เป็นภาษาไทยแบบตรงไปตรงมาและสร้างสรรค์:
สถิติภาพรวม:
- จำนวนไม้ทั้งหมด: %d (ชนะ %d / แพ้ %d)
- Win Rate: %.2f%%
- Profit Factor: %.2f
- กำไร/ขาดทุนสุทธิ: %.2f USDT
- ชนะสูงสุด: %.2f USDT | แพ้หนักสุด: %.2f USDT

ประวัติ %d ไม้ล่าสุด:
%s

กรุณาสรุปเป็นหัวข้อภาษาไทย:
1. จุดแข็งที่ทำได้ดีอยู่แล้ว
2. จุดอ่อนหรือพฤติกรรมเสี่ยงที่ทำให้พอร์ตโตช้า (ดูจากสัดส่วนกำไร/ขาดทุน และแท็กอารมณ์/กลยุทธ์)
3. คำแนะนำเชิงปฏิบัติ 3 ข้อสำหรับสัปดาห์ถัดไป`,
		summary.TotalTrades, summary.WinTrades, summary.LossTrades, summary.WinRate,
		summary.ProfitFactor, summary.NetPnl, summary.BestTradePnl, summary.WorstTradePnl,
		limit, sb.String())

	aiText, err := callGemini(prompt, "", "")
	if err != nil {
		log.Printf("[Gemini Review] Fallback habit analysis: %v", err)
		advice := "วินัยการตัดขาดทุนและทำตามแผนคือหัวใจสำคัญ"
		if summary.WinRate >= 50 && summary.NetPnl < 0 {
			advice = "สังเกตว่า Win Rate ของคุณสูงกว่า 50% แต่ภาพรวมยังติดลบ แสดงว่าเวลาได้กำไรมักปิดเร็วเกินไป แต่เวลาผิดทางปล่อยให้ขาดทุนลากยาว ควรคุม Risk:Reward ให้ได้อย่างน้อย 1:2 และตั้ง Stop Loss เสมอ"
		} else if summary.WinRate < 40 {
			advice = "Win Rate ยังค่อนข้างต่ำ แนะนำให้คัดเลือกเฉพาะ Setup ที่เข้าเงื่อนไขชัดเจน ลดความถี่ในการเทรด และหลีกเลี่ยงการเทรดตอนอารมณ์ FOMO หรือหัวร้อน"
		} else {
			advice = "ภาพรวมพอร์ตของคุณอยู่ในเกณฑ์ดี มีความสมดุลระหว่างอัตราชนะและกำไร ควรรักษาวินัยและจำกัดความเสี่ยงต่อไม้ไม่ให้เกิน 2% เพื่อเติบโตอย่างยั่งยืน"
		}

		return fmt.Sprintf(`### สรุปผลการวิเคราะห์พอร์ตส่วนตัว (อ้างอิงจาก %d ไม้)
1. **สถานะพอร์ตปัจจุบัน:** เทรดทั้งหมด %d ไม้ (Win Rate: %.2f%%, Profit Factor: %.2f, กำไรสุทธิ: %.2f USDT)
2. **จุดที่ต้องโฟกัส:** %s
3. **คำแนะนำสำหรับไม้ถัดไป:**
   - ตรวจสอบเช็คลิสต์ก่อนเข้าเทรดทุกครั้ง
   - หลีกเลี่ยงการเพิ่ม Leverage เกินความจำเป็นเมื่อเจอลำดับแพ้ต่อเนื่อง
   - จดบันทึกแท็ก "กลยุทธ์" และ "อารมณ์" ทุกไม้เพื่อให้สถิติแม่นยำยิ่งขึ้น`,
			len(trades), summary.TotalTrades, summary.WinRate, summary.ProfitFactor, summary.NetPnl, advice), nil
	}

	return aiText, nil
}

type cryptoCompareNewsResp struct {
	Data []struct {
		Title       string `json:"title"`
		Body        string `json:"body"`
		URL         string `json:"url"`
		Source      string `json:"source"`
		PublishedOn int64  `json:"published_on"`
		Categories  string `json:"categories"`
	} `json:"Data"`
}

func FetchAndAnalyzeCryptoNews(symbol string) ([]model.NewsItem, error) {
	cat := strings.ToUpper(strings.TrimSpace(symbol))
	if cat == "" || cat == "ALL" || cat == "MARKET" {
		cat = "BTC,ETH,Regulation,Market"
	}

	client := &http.Client{Timeout: 12 * time.Second}
	apiURL := fmt.Sprintf("https://min-api.cryptocompare.com/data/v2/news/?lang=EN&categories=%s", cat)
	resp, err := client.Get(apiURL)

	items := make([]model.NewsItem, 0)
	if err == nil {
		defer resp.Body.Close()
		var cc cryptoCompareNewsResp
		if json.NewDecoder(resp.Body).Decode(&cc) == nil && len(cc.Data) > 0 {
			limit := len(cc.Data)
			if limit > 8 {
				limit = 8
			}
			for i := 0; i < limit; i++ {
				d := cc.Data[i]
				pubTime := time.Unix(d.PublishedOn, 0).In(time.FixedZone("ICT", 7*3600)).Format("02/01/2006 15:04")
				sentiment := inferNewsSentiment(d.Title + " " + d.Body)
				reason := "ข่าวมีความเคลื่อนไหวตามรอบตลาดปกติ รอดูแรงซื้อขายยืนยันทิศทาง"
				impact := 5.0
				switch sentiment {
				case "BULLISH":
					reason = "มีปัจจัยบวกหนุนความเชื่อมั่นและแรงซื้อในตลาด"
					impact = 7.5
				case "BEARISH":
					reason = "มีปัจจัยกดดันหรือความกังวลด้านแรงขาย/กฎระเบียบ ควรระวังความผันผวน"
					impact = 7.5
				}

				snippet := d.Body
				if len(snippet) > 220 {
					snippet = snippet[:220] + "..."
				}

				items = append(items, model.NewsItem{
					Title:           d.Title,
					SummaryThai:     snippet,
					Source:          d.Source,
					PublishedAt:     pubTime,
					URL:             d.URL,
					Symbol:          strings.ToUpper(strings.TrimSpace(symbol)),
					Sentiment:       sentiment,
					SentimentReason: reason,
					ImpactScore:     impact,
				})
			}
		}
	}

	if len(items) == 0 {
		items = getDefaultCryptoNews(cat)
	}

	// ให้ Gemini ช่วยแปลและสรุปเป็นภาษาไทย + วิเคราะห์ Sentiment
	enrichNewsWithGemini(items)
	return items, nil
}

func enrichNewsWithGemini(items []model.NewsItem) {
	if len(items) == 0 {
		return
	}
	var sb strings.Builder
	for i, it := range items {
		sb.WriteString(fmt.Sprintf("%d. Title: %s\nBody: %s\n", i, it.Title, it.SummaryThai))
	}

	prompt := fmt.Sprintf(`คุณคือนักวิเคราะห์ข่าวคริปโต ช่วยสรุปข่าวด้านล่างนี้เป็นภาษาไทยสั้นๆ เข้าใจง่าย พร้อมประเมินทิศทาง (BULLISH = ข่าวดี, BEARISH = ข่าวลบ, NEUTRAL = เป็นกลาง) และเหตุผลสั้นๆ
ตอบกลับเป็น JSON Array เท่านั้น ตามลำดับ index 0 ถึง %d:
[
  {"index": 0, "summary_thai": "สรุปข่าวภาษาไทย 1-2 ประโยค", "sentiment": "BULLISH", "sentiment_reason": "ผลกระทบต่อราคาสั้นๆ", "impact_score": 7.5}
]

รายการข่าว:
%s`, len(items)-1, sb.String())

	raw, err := callGemini(prompt, "", "")
	if err != nil {
		log.Printf("[Gemini News] ใช้สรุปเบื้องต้นแทน: %v", err)
		return
	}

	var enriched []struct {
		Index           int     `json:"index"`
		SummaryThai     string  `json:"summary_thai"`
		Sentiment       string  `json:"sentiment"`
		SentimentReason string  `json:"sentiment_reason"`
		ImpactScore     float64 `json:"impact_score"`
	}
	if err := json.Unmarshal([]byte(cleanJSONBlock(raw)), &enriched); err != nil {
		return
	}
	for _, e := range enriched {
		if e.Index >= 0 && e.Index < len(items) {
			if e.SummaryThai != "" {
				items[e.Index].SummaryThai = e.SummaryThai
			}
			if e.Sentiment != "" {
				items[e.Index].Sentiment = strings.ToUpper(e.Sentiment)
			}
			if e.SentimentReason != "" {
				items[e.Index].SentimentReason = e.SentimentReason
			}
			if e.ImpactScore > 0 {
				items[e.Index].ImpactScore = e.ImpactScore
			}
		}
	}
}

func inferNewsSentiment(text string) string {
	l := strings.ToLower(text)
	bullWords := []string{"surge", "rally", "etf", "inflow", "adoption", "bull", "soar", "approve", "record high", "partnership", "upgrade"}
	bearWords := []string{"crash", "drop", "hack", "lawsuit", "sec", "ban", "outflow", "liquidat", "bear", "plunge", "warn", "inflation"}

	bullCount, bearCount := 0, 0
	for _, w := range bullWords {
		if strings.Contains(l, w) {
			bullCount++
		}
	}
	for _, w := range bearWords {
		if strings.Contains(l, w) {
			bearCount++
		}
	}
	if bullCount > bearCount {
		return "BULLISH"
	}
	if bearCount > bullCount {
		return "BEARISH"
	}
	return "NEUTRAL"
}

func getDefaultCryptoNews(cat string) []model.NewsItem {
	nowStr := time.Now().Format("02/01/2006 15:04")
	return []model.NewsItem{
		{
			Title:           "ภาพรวมตลาดคริปโตและกระแสเงินทุนไหลเข้า Bitcoin ETF",
			SummaryThai:     "นักลงทุนสถาบันยังคงจับตาทิศทางดอกเบี้ยสหรัฐฯ และปริมาณเงินทุนไหลเข้ากองทุน ETF อย่างใกล้ชิด",
			Source:          "Market Digest",
			PublishedAt:     nowStr,
			URL:             "https://www.binance.com/en/news",
			Symbol:          cat,
			Sentiment:       "NEUTRAL",
			SentimentReason: "ตลาดกำลังรอความชัดเจนของตัวเลขเศรษฐกิจมหภาค",
			ImpactScore:     6.0,
		},
	}
}
