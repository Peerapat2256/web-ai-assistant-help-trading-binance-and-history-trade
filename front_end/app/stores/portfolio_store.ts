import { defineStore } from "pinia";
import type { PortfolioSetting } from "~/models/sheet/sheet_model";

export const usePortfolioStore = defineStore("portfolio_store", {
  state: () => ({
    setting: null as PortfolioSetting | null,
  }),
  actions: {
    setSetting(val: PortfolioSetting) {
      this.setting = val;
    },
  },
});
