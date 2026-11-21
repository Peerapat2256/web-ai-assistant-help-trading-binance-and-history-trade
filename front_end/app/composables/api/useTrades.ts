import type { TradeCreateRequest } from "~/models/trade/request/trade_req";
import type {
  BatchTradeCreateResponse,
  ParsedScreenshotBatchResponse,
  TradeHistory,
} from "~/models/trade/response/trade_res";
import { useTradeStore } from "~/stores/trade_store";

export function useTrades() {
  const { $api } = useNuxtApp();
  const tradeStore = useTradeStore();

  const fetchTrades = async (symbol = "", side = "ALL") => {
    tradeStore.setLoading(true);
    try {
      const res = await $api.get<{ success: boolean; data: TradeHistory[] }>("/trades", {
        params: { symbol, side },
      });
      const list = res.data?.data || [];
      tradeStore.setTrades(list);
      return list;
    } finally {
      tradeStore.setLoading(false);
    }
  };

  const parseScreenshot = async (file: File): Promise<ParsedScreenshotBatchResponse> => {
    const formData = new FormData();
    formData.append("file", file);
    const res = await $api.post<{ success: boolean; data: ParsedScreenshotBatchResponse }>(
      "/trades/parse-screenshot",
      formData,
      { headers: { "Content-Type": "multipart/form-data" } }
    );
    return res.data.data;
  };

  const createTrade = async (payload: TradeCreateRequest): Promise<TradeHistory> => {
    const res = await $api.post<{ success: boolean; data: TradeHistory }>("/trades", payload);
    return res.data.data;
  };

  const createTradesBatch = async (
    trades: TradeCreateRequest[],
    skipDuplicates = true
  ): Promise<{ message: string; data: BatchTradeCreateResponse }> => {
    const res = await $api.post<{ success: boolean; message: string; data: BatchTradeCreateResponse }>(
      "/trades/batch",
      { trades, skip_duplicates: skipDuplicates }
    );
    return { message: res.data.message, data: res.data.data };
  };

  const updateTrade = async (id: number, payload: Partial<TradeCreateRequest>): Promise<TradeHistory> => {
    const res = await $api.put<{ success: boolean; data: TradeHistory }>(`/trades/${id}`, payload);
    return res.data.data;
  };

  const deleteTrade = async (id: number) => {
    await $api.delete(`/trades/${id}`);
  };

  return {
    fetchTrades,
    parseScreenshot,
    createTrade,
    createTradesBatch,
    updateTrade,
    deleteTrade,
  };
}
