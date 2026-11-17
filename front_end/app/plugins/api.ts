import axios from "axios";

export default defineNuxtPlugin(() => {
  const config = useRuntimeConfig();
  const api = axios.create({
    baseURL: config.public.apiBase || "http://localhost:8080",
    timeout: 120000,
  });

  api.interceptors.request.use((req) => {
    const token = useCookie("token").value;
    if (token) {
      req.headers.Authorization = `Bearer ${token}`;
    }
    return req;
  });

  api.interceptors.response.use(
    (res) => res,
    async (err) => {
      if (err.response?.status === 401 && import.meta.client) {
        const token = useCookie("token");
        const refreshToken = useCookie("refresh_token");
        if (refreshToken.value && !err.config?._retry) {
          err.config._retry = true;
          try {
            const refreshRes = await axios.post(
              `${config.public.apiBase || "http://localhost:8080"}/auth/refresh-token`,
              { refresh_token: refreshToken.value }
            );
            if (refreshRes.data?.token) {
              token.value = refreshRes.data.token;
              refreshToken.value = refreshRes.data.refresh_token;
              err.config.headers.Authorization = `Bearer ${refreshRes.data.token}`;
              return api(err.config);
            }
          } catch {
            token.value = null;
            refreshToken.value = null;
            navigateTo("/login");
          }
        }
      }
      return Promise.reject(err);
    }
  );

  return {
    provide: {
      api,
    },
  };
});
