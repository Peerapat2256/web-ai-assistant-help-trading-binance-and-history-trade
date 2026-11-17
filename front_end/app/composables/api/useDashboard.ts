import type { DashboardSummary, PnlCalendarDay } from "~/models/dashboard/dashboard_model";

export function useDashboard() {
  const { $api } = useNuxtApp();

  const fetchSummary = async (): Promise<DashboardSummary> => {
    const res = await $api.get<DashboardSummary>("/dashboard/summary");
    return res.data;
  };

  const fetchCalendar = async (year: number, month: number): Promise<PnlCalendarDay[]> => {
    const res = await $api.get<{ success: boolean; data: PnlCalendarDay[] }>("/dashboard/calendar", {
      params: { year, month },
    });
    return res.data?.data || [];
  };

  return { fetchSummary, fetchCalendar };
}
