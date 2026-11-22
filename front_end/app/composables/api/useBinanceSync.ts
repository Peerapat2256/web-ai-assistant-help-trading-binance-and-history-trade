export interface BinanceConfig {
  connected: boolean;
  api_key_masked: string;
  is_auto_sync: boolean;
  last_synced_at: string | null;
  last_sync_error: string;
}

export interface BinanceSyncResult {
  fetched_count: number;
  created_count: number;
  skipped_count: number;
}

export function useBinanceSync() {
  const { $api } = useNuxtApp();

  const fetchBinanceConfig = async (): Promise<BinanceConfig> => {
    const res = await $api.get<{ success: boolean; data: BinanceConfig }>("/binance/config");
    return res.data.data;
  };

  const saveBinanceConfig = async (payload: {
    api_key?: string;
    api_secret?: string;
    is_auto_sync: boolean;
  }): Promise<BinanceConfig> => {
    const res = await $api.put<{ success: boolean; data: BinanceConfig }>("/binance/config", payload);
    return res.data.data;
  };

  const deleteBinanceConfig = async () => {
    await $api.delete("/binance/config");
  };

  const testBinanceConnection = async (payload?: { api_key?: string; api_secret?: string }) => {
    const res = await $api.post<{ success: boolean; message: string; balance: number }>("/binance/test", payload || {});
    return res.data;
  };

  const syncBinanceTrades = async (days = 7) => {
    const res = await $api.post<{ success: boolean; message: string; data: BinanceSyncResult }>(
      "/binance/sync",
      {},
      { params: { days } }
    );
    return res.data;
  };

  return {
    fetchBinanceConfig,
    saveBinanceConfig,
    deleteBinanceConfig,
    testBinanceConnection,
    syncBinanceTrades,
  };
}
