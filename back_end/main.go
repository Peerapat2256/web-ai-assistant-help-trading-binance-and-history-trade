package main

import (
	"backend/model/query"
	"backend/repository"
	"backend/system"
	"backend/worker"
)

func main() {
	db, err := repository.NewDbConnection()
	if err != nil {
		panic(err)
	}
	query.SetDefault(db)
	worker.StartSheetRetryWorker()
	worker.StartBinanceAutoSyncWorker()
	system.StartServer()
}
