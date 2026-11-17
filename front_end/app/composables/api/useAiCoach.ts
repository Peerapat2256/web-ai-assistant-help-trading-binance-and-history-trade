import type { AiAnalysisResponse, MarketDataSummary } from "~/models/ai/ai_model";

export function useAiCoach() {
  const { $api } = useNuxtApp();

  const fetchMarketKlines = async (symbol: string, timeframe: string): Promise<MarketDataSummary> => {
    const res = await $api.get<{ success: boolean; data: MarketDataSummary }>("/market/klines", {
      params: { symbol, timeframe },
    });
    return res.data.data;
  };

  const analyzeChart = async (
    symbol: string,
    timeframe: string,
    question: string,
    teachMode: boolean,
    file?: File | null
  ): Promise<AiAnalysisResponse> => {
    const formData = new FormData();
    formData.append("symbol", symbol);
    formData.append("timeframe", timeframe);
    formData.append("question", question);
    formData.append("teach_mode", teachMode ? "true" : "false");
    if (file) {
      formData.append("file", file);
    }
    const res = await $api.post<{ success: boolean; data: AiAnalysisResponse }>(
      "/ai/analyze-chart",
      formData,
      { headers: { "Content-Type": "multipart/form-data" } }
    );
    return res.data.data;
  };

  const reviewJournal = async (): Promise<AiAnalysisResponse> => {
    const res = await $api.post<{ success: boolean; data: AiAnalysisResponse }>("/ai/review-journal");
    return res.data.data;
  };

  const fetchAiLogs = async (): Promise<AiAnalysisResponse[]> => {
    const res = await $api.get<{ success: boolean; data: AiAnalysisResponse[] }>("/ai/logs");
    return res.data?.data || [];
  };

  const deleteAiLog = async (id: number) => {
    await $api.delete(`/ai/logs/${id}`);
  };

  return {
    fetchMarketKlines,
    analyzeChart,
    reviewJournal,
    fetchAiLogs,
    deleteAiLog,
  };
}
