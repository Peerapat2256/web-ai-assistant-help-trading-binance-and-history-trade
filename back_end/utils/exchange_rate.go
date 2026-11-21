package utils

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	usdtThbLock      sync.RWMutex
	usdtThbRate      float64 = 34.50
	usdtThbFetchedAt time.Time
)

// GetUSDTTHBRate ดึงอัตราแลกเปลี่ยน USDT/THB ล่าสุดจาก Binance (แคชไว้ 15 นาที)
func GetUSDTTHBRate() float64 {
	usdtThbLock.RLock()
	if time.Since(usdtThbFetchedAt) < 15*time.Minute && usdtThbRate > 0 {
		r := usdtThbRate
		usdtThbLock.RUnlock()
		return r
	}
	usdtThbLock.RUnlock()

	client := &http.Client{Timeout: 5 * time.Second}
	rate := 0.0

	// 1) ลองดึงจาก Binance Ticker USDTTHB
	if resp, err := client.Get("https://api.binance.com/api/v3/ticker/price?symbol=USDTTHB"); err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var payload struct {
				Price string `json:"price"`
			}
			if json.NewDecoder(resp.Body).Decode(&payload) == nil {
				if p, err := strconv.ParseFloat(payload.Price, 64); err == nil && p > 20 && p < 60 {
					rate = p
				}
			}
		}
	}

	// 2) Fallback ไปยัง open.er-api.com (USD -> THB)
	if rate <= 0 {
		if resp, err := client.Get("https://open.er-api.com/v6/latest/USD"); err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var payload struct {
					Rates map[string]float64 `json:"rates"`
				}
				if json.NewDecoder(resp.Body).Decode(&payload) == nil {
					if thb, ok := payload.Rates["THB"]; ok && thb > 20 && thb < 60 {
						rate = thb
					}
				}
			}
		}
	}

	usdtThbLock.Lock()
	defer usdtThbLock.Unlock()
	if rate > 0 {
		usdtThbRate = rate
		usdtThbFetchedAt = time.Now()
	}
	return usdtThbRate
}
