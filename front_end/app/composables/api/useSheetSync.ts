import type { PortfolioSetting, SheetConfig } from "~/models/sheet/sheet_model";

export function useSheetSync() {
  const { $api } = useNuxtApp();

  const fetchPortfolioSetting = async (): Promise<PortfolioSetting> => {
    const res = await $api.get<{ success: boolean; data: PortfolioSetting }>("/portfolio/settings");
    return res.data.data;
  };

  const savePortfolioSetting = async (payload: Partial<PortfolioSetting>): Promise<PortfolioSetting> => {
    const res = await $api.put<{ success: boolean; data: PortfolioSetting }>("/portfolio/settings", payload);
    return res.data.data;
  };

  const fetchSheetConfig = async (): Promise<SheetConfig> => {
    const res = await $api.get<{ success: boolean; data: SheetConfig }>("/sheets/config");
    return res.data.data;
  };

  const saveSheetConfig = async (payload: {
    spreadsheet_id: string;
    sheet_name: string;
    is_auto_sync: boolean;
  }): Promise<SheetConfig> => {
    const res = await $api.put<{ success: boolean; data: SheetConfig }>("/sheets/config", payload);
    return res.data.data;
  };

  const testSheetConnection = async (payload: { spreadsheet_id: string; sheet_name: string }) => {
    const res = await $api.post<{ success: boolean; message: string; title: string }>("/sheets/test", payload);
    return res.data;
  };

  const resyncAllTrades = async () => {
    const res = await $api.post<{ success: boolean; message: string; count: number }>("/sheets/resync");
    return res.data;
  };

  return {
    fetchPortfolioSetting,
    savePortfolioSetting,
    fetchSheetConfig,
    saveSheetConfig,
    testSheetConnection,
    resyncAllTrades,
  };
}
