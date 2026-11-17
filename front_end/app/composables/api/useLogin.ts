import type { LoginRequest } from "~/models/auth/request/login_post_req";
import type { RegisterRequest } from "~/models/auth/request/regis_post_req";
import type { LoginResponse } from "~/models/auth/response/login_get_res";
import { useAuthStore } from "~/stores/auth";

export function useLogin() {
  const { $api } = useNuxtApp();
  const authStore = useAuthStore();

  const token = useCookie<string | null>("token", { path: "/", maxAge: 60 * 60 * 24 });
  const refreshToken = useCookie<string | null>("refresh_token", { path: "/", maxAge: 60 * 60 * 24 * 30 });
  const errorMessage = ref<string | null>(null);

  const login = async (request: LoginRequest): Promise<LoginResponse> => {
    errorMessage.value = null;
    try {
      const res = await $api.post<LoginResponse>("/auth/login", request);
      if (!res.data?.token) {
        throw new Error("ไม่พบโทเคนเข้าสู่ระบบ");
      }
      token.value = res.data.token;
      refreshToken.value = res.data.refresh_token;
      authStore.setUser(res.data.user, res.data.token);
      return res.data;
    } catch (err: any) {
      const msg = err.response?.data?.message || err.message || "เข้าสู่ระบบไม่สำเร็จ";
      errorMessage.value = msg;
      throw new Error(msg);
    }
  };

  const register = async (request: RegisterRequest): Promise<LoginResponse> => {
    errorMessage.value = null;
    try {
      const res = await $api.post<LoginResponse>("/auth/register", request);
      if (!res.data?.token) {
        throw new Error("สมัครสมาชิกไม่สำเร็จ");
      }
      token.value = res.data.token;
      refreshToken.value = res.data.refresh_token;
      authStore.setUser(res.data.user, res.data.token);
      return res.data;
    } catch (err: any) {
      const msg = err.response?.data?.message || err.message || "สมัครสมาชิกไม่สำเร็จ";
      errorMessage.value = msg;
      throw new Error(msg);
    }
  };

  const logout = () => {
    token.value = null;
    refreshToken.value = null;
    authStore.logout();
    navigateTo("/login");
  };

  return { login, register, logout, errorMessage };
}
