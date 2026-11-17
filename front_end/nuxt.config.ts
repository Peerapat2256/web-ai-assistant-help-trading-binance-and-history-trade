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
    tsConfig: { compilerOptions: { module: "ESNext" } },
    nodeTsConfig: { compilerOptions: { module: "ESNext" } },
    sharedTsConfig: { compilerOptions: { module: "ESNext" } },
  },
  hooks: {
    "prepare:types": (ctx) => {
      for (const cfg of [ctx.tsConfig, ctx.nodeTsConfig, ctx.sharedTsConfig]) {
        if (cfg?.compilerOptions) {
          cfg.compilerOptions.module = "ESNext";
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
