import type { NewsItem } from "~/models/news/news_model";

export function useMarketNews() {
  const { $api } = useNuxtApp();

  const fetchNews = async (symbol = "BTC"): Promise<NewsItem[]> => {
    const res = await $api.get<{ success: boolean; data: NewsItem[] }>("/news", {
      params: { symbol },
    });
    return res.data?.data || [];
  };

  return { fetchNews };
}
