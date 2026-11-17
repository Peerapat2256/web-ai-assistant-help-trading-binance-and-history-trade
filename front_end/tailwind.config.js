/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./app/components/**/*.{vue,js,ts}",
    "./app/layouts/**/*.vue",
    "./app/pages/**/*.vue",
    "./app/app.vue",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ["DM Sans", "Sarabun", "system-ui", "sans-serif"],
        display: ["DM Sans", "Sarabun", "sans-serif"],
      },
    },
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: [
      {
        binancedark: {
          "primary": "#FCD535",
          "primary-content": "#0B0E11",
          "secondary": "#2B3139",
          "secondary-content": "#EAECEF",
          "accent": "#0ECB81",
          "accent-content": "#0B0E11",
          "neutral": "#1E2329",
          "neutral-content": "#EAECEF",
          "base-100": "#181A20",
          "base-200": "#0B0E11",
          "base-300": "#2B3139",
          "base-content": "#EAECEF",
          "info": "#3B82F6",
          "success": "#0ECB81",
          "warning": "#FCD535",
          "error": "#F6465D",
        },
        binancelight: {
          "primary": "#F0B90B",
          "primary-content": "#1E2329",
          "secondary": "#EAECEF",
          "secondary-content": "#1E2329",
          "accent": "#0ECB81",
          "accent-content": "#FFFFFF",
          "neutral": "#1E2329",
          "neutral-content": "#FFFFFF",
          "base-100": "#FFFFFF",
          "base-200": "#F5F7FA",
          "base-300": "#E6E8EA",
          "base-content": "#1E2329",
          "info": "#3B82F6",
          "success": "#0ECB81",
          "warning": "#F0B90B",
          "error": "#F6465D",
        },
      },
    ],
    darkTheme: "binancedark",
  },
};
