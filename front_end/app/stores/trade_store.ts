import { defineStore } from "pinia";
import type { TradeHistory } from "~/models/trade/response/trade_res";

export const useTradeStore = defineStore("trade_store", {
  state: () => ({
    trades: [] as TradeHistory[],
    loading: false,
  }),
  actions: {
    setTrades(items: TradeHistory[]) {
      this.trades = items;
    },
    setLoading(val: boolean) {
      this.loading = val;
    },
  },
});
