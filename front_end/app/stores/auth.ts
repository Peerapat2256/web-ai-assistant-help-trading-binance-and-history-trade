import { defineStore } from "pinia";
import type { LoginUser } from "~/models/auth/response/login_get_res";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    user: null as LoginUser | null,
    token: "",
  }),
  actions: {
    setUser(user: LoginUser, token: string) {
      this.user = user;
      this.token = token;
    },
    logout() {
      this.user = null;
      this.token = "";
    },
  },
  persist: true,
});
