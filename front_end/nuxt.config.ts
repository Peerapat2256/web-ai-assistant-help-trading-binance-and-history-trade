export default defineNuxtConfig({
  srcDir: "app",
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  app: {
    pageTransition: { name: "page", mode: "out-in" },
  },
  modules: [
    "@nuxtjs/tailwindcss",
    "@pinia/nuxt",
    "pinia-plugin-persistedstate/nuxt",
  ],
  css: ["~/assets/css/main.css"],
  typescript: {
    tsConfig: { compilerOptions: { module: "esnext" } },
    nodeTsConfig: { compilerOptions: { module: "esnext" } },
    sharedTsConfig: { compilerOptions: { module: "esnext" } },
  },
  hooks: {
    "prepare:types": (ctx) => {
      for (const cfg of [ctx.tsConfig, ctx.nodeTsConfig, ctx.sharedTsConfig]) {
        if (cfg?.compilerOptions) {
          cfg.compilerOptions.module = "esnext";
          delete (cfg.compilerOptions as Record<string, unknown>).libReplacement;
        }
      }
    },
  },
  runtimeConfig: {
    public: {
      apiBase: "http://localhost:8080",
      googleClientId: "",
    },
  },
});
