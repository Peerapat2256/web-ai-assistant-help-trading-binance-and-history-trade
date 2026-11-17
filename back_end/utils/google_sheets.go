package utils

import (
	"backend/model/model"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func GetServiceAccountEmail() string {
	credFile := strings.TrimSpace(os.Getenv("GOOGLE_CREDENTIALS_FILE"))
	if credFile == "" {
		credFile = "./service-account.json"
	}
	data, err := os.ReadFile(credFile)
	if err == nil {
		var parsed struct {
			ClientEmail string `json:"client_email"`
		}
		if json.Unmarshal(data, &parsed) == nil && parsed.ClientEmail != "" {
			return parsed.ClientEmail
		}
	}
	return "inance-sheet-bot@round-music-482118-b7.iam.gserviceaccount.com"
}

var sheetHeaders = []interface{}{
	"รหัสอ้างอิง",
	"วันที่-เวลา",
	"คู่เหรียญ",
	"ฝั่ง",
	"Leverage",
	"ราคาเข้า",
	"ราคาปิด",
	"จำนวน",
	"มาร์จิ้น (USDT)",
	"กำไร/ขาดทุน (USDT)",
	"ROI (%)",
	"ค่าธรรมเนียม",
	"กลยุทธ์ / อารมณ์ / หมายเหตุ",
}

func ExtractSpreadsheetID(input string) string {
	trimmed := strings.TrimSpace(input)
	if strings.Contains(trimmed, "/spreadsheets/d/") {
		parts := strings.Split(trimmed, "/spreadsheets/d/")
		if len(parts) > 1 {
			idPart := strings.Split(parts[1], "/")[0]
			return strings.TrimSpace(idPart)
		}
	}
	return trimmed
}

func getSheetsService(ctx context.Context) (*sheets.Service, error) {
	credFile := strings.TrimSpace(os.Getenv("GOOGLE_CREDENTIALS_FILE"))
	if credFile == "" {
		credFile = "./service-account.json"
	}
	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("ไม่พบไฟล์กุญแจ Google Service Account (%s)", credFile)
	}
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credFile), option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return nil, fmt.Errorf("เชื่อมต่อ Google Sheets API ไม่สำเร็จ: %v", err)
	}
	return srv, nil
}

func EnsureSheetAndHeaders(ctx context.Context, srv *sheets.Service, spreadsheetID, sheetName string) (int64, string, error) {
	if sheetName == "" {
		sheetName = "Trades"
	}

	ss, err := srv.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return 0, "", fmt.Errorf("ไม่สามารถเข้าถึงไฟล์ Google Sheet ได้ (ตรวจสอบการแชร์สิทธิ์ให้บอท): %v", err)
	}

	var targetSheetID int64 = -1
	for _, s := range ss.Sheets {
		if s.Properties != nil && s.Properties.Title == sheetName {
			targetSheetID = s.Properties.SheetId
			break
		}
	}

	if targetSheetID == -1 {
		req := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				{
					AddSheet: &sheets.AddSheetRequest{
						Properties: &sheets.SheetProperties{
							Title: sheetName,
						},
					},
				},
			},
		}
		resp, err := srv.Spreadsheets.BatchUpdate(spreadsheetID, req).Context(ctx).Do()
		if err != nil {
			return 0, "", fmt.Errorf("สร้างแท็บ '%s' ใน Google Sheet ไม่สำเร็จ: %v", sheetName, err)
		}
		if len(resp.Replies) > 0 && resp.Replies[0].AddSheet != nil {
			targetSheetID = resp.Replies[0].AddSheet.Properties.SheetId
		}
	}

	headerRange := fmt.Sprintf("'%s'!A1:M1", sheetName)
	existing, err := srv.Spreadsheets.Values.Get(spreadsheetID, headerRange).Context(ctx).Do()
	needHeader := err != nil || len(existing.Values) == 0 || len(existing.Values[0]) == 0 || fmt.Sprintf("%v", existing.Values[0][0]) != "รหัสอ้างอิง"

	if needHeader {
		vr := &sheets.ValueRange{
			Values: [][]interface{}{sheetHeaders},
		}
		_, err = srv.Spreadsheets.Values.Update(spreadsheetID, headerRange, vr).ValueInputOption("USER_ENTERED").Context(ctx).Do()
		if err != nil {
			return 0, "", fmt.Errorf("เขียนหัวตาราง Google Sheet ไม่สำเร็จ: %v", err)
		}
	}

	title := ""
	if ss.Properties != nil {
		title = ss.Properties.Title
	}
	return targetSheetID, title, nil
}

func formatTradeRow(trade *model.TradeHistory) []interface{} {
	detailParts := []string{}
	if trade.SetupTag != "" {
		detailParts = append(detailParts, "กลยุทธ์: "+trade.SetupTag)
	}
	if trade.EmotionTag != "" {
		detailParts = append(detailParts, "อารมณ์: "+trade.EmotionTag)
	}
	if trade.Notes != "" {
		detailParts = append(detailParts, trade.Notes)
	}
	detailStr := strings.Join(detailParts, " | ")

	return []interface{}{
		trade.TradeCode,
		trade.TradeTime.Format("2006-01-02 15:04"),
		trade.Symbol,
		trade.PositionSide,
		fmt.Sprintf("%dx", trade.Leverage),
		trade.EntryPrice,
		trade.ExitPrice,
		trade.Quantity,
		trade.MarginUsed,
		trade.Pnl,
		trade.PnlPercentage,
		trade.Fee,
		detailStr,
	}
}

func TestSheetConnection(spreadsheetID, sheetName string) (string, error) {
	ctx := context.Background()
	id := ExtractSpreadsheetID(spreadsheetID)
	if id == "" {
		return "", fmt.Errorf("กรุณาระบุ Spreadsheet ID")
	}
	srv, err := getSheetsService(ctx)
	if err != nil {
		return "", err
	}
	_, title, err := EnsureSheetAndHeaders(ctx, srv, id, sheetName)
	if err != nil {
		return "", err
	}
	return title, nil
}

func AppendTradeToSheet(spreadsheetID, sheetName string, trade *model.TradeHistory) error {
	ctx := context.Background()
	id := ExtractSpreadsheetID(spreadsheetID)
	if id == "" {
		return fmt.Errorf("ยังไม่ได้ตั้งค่า Spreadsheet ID")
	}
	if sheetName == "" {
		sheetName = "Trades"
	}

	srv, err := getSheetsService(ctx)
	if err != nil {
		return err
	}
	if _, _, err := EnsureSheetAndHeaders(ctx, srv, id, sheetName); err != nil {
		return err
	}

	vr := &sheets.ValueRange{
		Values: [][]interface{}{formatTradeRow(trade)},
	}
	appendRange := fmt.Sprintf("'%s'!A:M", sheetName)
	_, err = srv.Spreadsheets.Values.Append(id, appendRange, vr).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Context(ctx).Do()
	return err
}

func UpdateTradeInSheet(spreadsheetID, sheetName string, trade *model.TradeHistory) error {
	ctx := context.Background()
	id := ExtractSpreadsheetID(spreadsheetID)
	if id == "" {
		return fmt.Errorf("ยังไม่ได้ตั้งค่า Spreadsheet ID")
	}
	if sheetName == "" {
		sheetName = "Trades"
	}

	srv, err := getSheetsService(ctx)
	if err != nil {
		return err
	}
	if _, _, err := EnsureSheetAndHeaders(ctx, srv, id, sheetName); err != nil {
		return err
	}

	colARange := fmt.Sprintf("'%s'!A:A", sheetName)
	resp, err := srv.Spreadsheets.Values.Get(id, colARange).Context(ctx).Do()
	if err != nil {
		return err
	}

	foundRow := -1
	for idx, row := range resp.Values {
		if len(row) > 0 && fmt.Sprintf("%v", row[0]) == trade.TradeCode {
			foundRow = idx + 1
			break
		}
	}

	if foundRow == -1 {
		return AppendTradeToSheet(id, sheetName, trade)
	}

	updateRange := fmt.Sprintf("'%s'!A%d:M%d", sheetName, foundRow, foundRow)
	vr := &sheets.ValueRange{
		Values: [][]interface{}{formatTradeRow(trade)},
	}
	_, err = srv.Spreadsheets.Values.Update(id, updateRange, vr).ValueInputOption("USER_ENTERED").Context(ctx).Do()
	return err
}

func DeleteTradeFromSheet(spreadsheetID, sheetName string, tradeCode string) error {
	ctx := context.Background()
	id := ExtractSpreadsheetID(spreadsheetID)
	if id == "" {
		return fmt.Errorf("ยังไม่ได้ตั้งค่า Spreadsheet ID")
	}
	if sheetName == "" {
		sheetName = "Trades"
	}

	srv, err := getSheetsService(ctx)
	if err != nil {
		return err
	}
	sheetID, _, err := EnsureSheetAndHeaders(ctx, srv, id, sheetName)
	if err != nil {
		return err
	}

	colARange := fmt.Sprintf("'%s'!A:A", sheetName)
	resp, err := srv.Spreadsheets.Values.Get(id, colARange).Context(ctx).Do()
	if err != nil {
		return err
	}

	foundIdx := -1
	for idx, row := range resp.Values {
		if idx == 0 {
			continue
		}
		if len(row) > 0 && fmt.Sprintf("%v", row[0]) == tradeCode {
			foundIdx = idx
			break
		}
	}

	if foundIdx == -1 {
		return nil
	}

	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteDimension: &sheets.DeleteDimensionRequest{
					Range: &sheets.DimensionRange{
						SheetId:    sheetID,
						Dimension:  "ROWS",
						StartIndex: int64(foundIdx),
						EndIndex:   int64(foundIdx + 1),
					},
				},
			},
		},
	}
	_, err = srv.Spreadsheets.BatchUpdate(id, req).Context(ctx).Do()
	return err
}

func FullSyncTradesToSheet(spreadsheetID, sheetName string, trades []model.TradeHistory) error {
	ctx := context.Background()
	id := ExtractSpreadsheetID(spreadsheetID)
	if id == "" {
		return fmt.Errorf("ยังไม่ได้ตั้งค่า Spreadsheet ID")
	}
	if sheetName == "" {
		sheetName = "Trades"
	}

	srv, err := getSheetsService(ctx)
	if err != nil {
		return err
	}
	if _, _, err := EnsureSheetAndHeaders(ctx, srv, id, sheetName); err != nil {
		return err
	}

	clearRange := fmt.Sprintf("'%s'!A2:M", sheetName)
	_, err = srv.Spreadsheets.Values.Clear(id, clearRange, &sheets.ClearValuesRequest{}).Context(ctx).Do()
	if err != nil {
		return err
	}

	if len(trades) == 0 {
		return nil
	}

	rows := make([][]interface{}, 0, len(trades))
	for i := range trades {
		rows = append(rows, formatTradeRow(&trades[i]))
	}

	vr := &sheets.ValueRange{Values: rows}
	writeRange := fmt.Sprintf("'%s'!A2", sheetName)
	_, err = srv.Spreadsheets.Values.Update(id, writeRange, vr).ValueInputOption("USER_ENTERED").Context(ctx).Do()
	return err
}
