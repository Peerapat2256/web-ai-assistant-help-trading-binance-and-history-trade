import type {
  ParsedWithdrawalData,
  WithdrawalCreateRequest,
  WithdrawalHistory,
} from "~/models/withdrawal/withdrawal_model";

export function useWithdrawals() {
  const { $api } = useNuxtApp();

  const fetchWithdrawals = async (year = 0): Promise<WithdrawalHistory[]> => {
    const res = await $api.get<{ success: boolean; data: WithdrawalHistory[] }>("/withdrawals", {
      params: year > 0 ? { year } : {},
    });
    return res.data?.data || [];
  };

  const parseWithdrawalScreenshot = async (file: File): Promise<ParsedWithdrawalData> => {
    const formData = new FormData();
    formData.append("file", file);
    const res = await $api.post<{ success: boolean; data: ParsedWithdrawalData }>(
      "/withdrawals/parse-screenshot",
      formData,
      { headers: { "Content-Type": "multipart/form-data" } }
    );
    return res.data.data;
  };

  const createWithdrawal = async (payload: WithdrawalCreateRequest): Promise<WithdrawalHistory> => {
    const res = await $api.post<{ success: boolean; data: WithdrawalHistory }>("/withdrawals", payload);
    return res.data.data;
  };

  const updateWithdrawal = async (
    id: number,
    payload: Partial<WithdrawalCreateRequest>
  ): Promise<WithdrawalHistory> => {
    const res = await $api.put<{ success: boolean; data: WithdrawalHistory }>(`/withdrawals/${id}`, payload);
    return res.data.data;
  };

  const deleteWithdrawal = async (id: number) => {
    await $api.delete(`/withdrawals/${id}`);
  };

  return {
    fetchWithdrawals,
    parseWithdrawalScreenshot,
    createWithdrawal,
    updateWithdrawal,
    deleteWithdrawal,
  };
}
