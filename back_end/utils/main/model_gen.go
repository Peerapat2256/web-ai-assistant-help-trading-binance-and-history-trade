package main

import (
	"backend/repository"
	"log"

	"gorm.io/gen"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:      "./model/query",
		ModelPkgPath: "./model",
		Mode:         gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	db, err := repository.NewDbConnection()
	if err != nil {
		log.Fatalf("เชื่อมต่อฐานข้อมูลไม่สำเร็จ: %v", err)
	}

	g.UseDB(db)
	g.ApplyBasic(
		g.GenerateModel("member"),
		g.GenerateModel("portfolio_setting"),
		g.GenerateModel("trade_history"),
		g.GenerateModel("ai_analysis_log"),
		g.GenerateModel("sheet_config"),
	)
	g.Execute()
}
