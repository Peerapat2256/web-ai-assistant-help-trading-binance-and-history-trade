<template>
  <div class="drawer lg:drawer-open min-h-screen bg-base-200 agy-bg-grid relative">
    <div class="agy-orb agy-orb-gold"></div>
    <div class="agy-orb agy-orb-emerald"></div>

    <input id="main-drawer" type="checkbox" class="drawer-toggle" v-model="drawerOpen" />

    <div class="drawer-content flex flex-col min-w-0 relative z-10">
      <header class="navbar bg-base-100/80 backdrop-blur-xl border-b border-base-300/70 px-4 lg:px-6 sticky top-0 z-30">
        <div class="flex-none lg:hidden">
          <label for="main-drawer" class="btn btn-square btn-ghost">
            <Menu class="w-5 h-5" />
          </label>
        </div>

        <div class="flex-1 flex items-center gap-3">
          <span class="font-bold text-base lg:text-lg tracking-tight hidden sm:inline">
            {{ pageTitle }}
          </span>
          <NuxtLink
            to="/settings"
            :class="[
              'badge badge-outline gap-1.5 text-xs hidden md:inline-flex transition-all hover:scale-105',
              sheetState?.is_auto_sync && sheetState?.spreadsheet_id ? 'badge-success' : 'badge-warning',
            ]"
          >
            <span
              :class="[
                'w-2 h-2 rounded-full',
                sheetState?.is_auto_sync && sheetState?.spreadsheet_id ? 'bg-success agy-pulse-dot' : 'bg-warning',
              ]"
            ></span>
            {{
              sheetState?.is_auto_sync && sheetState?.spreadsheet_id
                ? `ซิงค์ชีต: ${sheetState.sheet_name || 'Trades'}`
                : 'คลิกเพื่อเชื่อม Google Sheet ส่วนตัว'
            }}
          </NuxtLink>
        </div>

        <div class="flex-none flex items-center gap-2">
          <button
            class="btn btn-ghost btn-circle btn-sm hover:bg-base-200"
            @click="toggleTheme"
            :title="theme === 'binancedark' ? 'สลับเป็นโหมดสว่าง' : 'สลับเป็นโหมดมืด'"
          >
            <Sun v-if="theme === 'binancedark'" class="w-4 h-4 text-primary" />
            <Moon v-else class="w-4 h-4" />
          </button>

          <div class="dropdown dropdown-end">
            <div tabindex="0" role="button" class="btn btn-ghost btn-sm gap-2 normal-case">
              <div class="avatar placeholder">
                <div class="bg-primary text-primary-content rounded-full w-7 h-7 flex items-center justify-center font-bold text-xs shadow-[0_0_12px_rgba(252,213,53,0.35)]">
                  <span>{{ userInitial }}</span>
                </div>
              </div>
              <span class="hidden sm:inline font-medium text-sm">
                {{ authStore.user?.fullname || "ผู้ใช้งาน" }}
              </span>
            </div>
            <ul tabindex="0" class="dropdown-content z-40 menu p-2 shadow-xl bg-base-100/95 backdrop-blur-xl border border-base-300 rounded-box w-60 mt-2">
              <li class="menu-title text-xs opacity-70 px-3 py-1 truncate">
                {{ authStore.user?.email }}
              </li>
              <li>
                <NuxtLink to="/settings" @click="drawerOpen = false">
                  <Settings class="w-4 h-4" />
                  ตั้งค่าพอร์ต & Google Sheets
                </NuxtLink>
              </li>
              <li>
                <button @click="handleLogout" class="text-error">
                  <LogOut class="w-4 h-4" />
                  ออกจากระบบ
                </button>
              </li>
            </ul>
          </div>
        </div>
      </header>

      <main class="flex-1 p-4 lg:p-6 max-w-7xl w-full mx-auto">
        <slot />
      </main>
    </div>

    <div class="drawer-side z-40">
      <label for="main-drawer" aria-label="ปิดเมนู" class="drawer-overlay"></label>
      <aside class="w-64 min-h-full bg-base-100/85 backdrop-blur-xl border-r border-base-300/70 flex flex-col">
        <div class="p-5 border-b border-base-300/70 flex items-center gap-3">
          <div class="w-9 h-9 rounded-xl bg-primary text-primary-content flex items-center justify-center font-extrabold text-lg shadow-[0_0_18px_rgba(252,213,53,0.35)]">
            B
          </div>
          <div>
            <div class="font-bold text-base leading-tight">Binance Journal</div>
            <div class="text-xs opacity-60">บันทึกเทรด & AI โค้ช</div>
          </div>
        </div>

        <ul class="menu p-3 gap-1.5 flex-1 text-sm font-medium">
          <li v-for="item in navItems" :key="item.to">
            <NuxtLink
              :to="item.to"
              @click="drawerOpen = false"
              :class="[
                'rounded-xl py-2.5 px-3 flex items-center gap-3 transition-all duration-200',
                route.path === item.to
                  ? 'bg-primary text-primary-content font-semibold shadow-[0_6px_18px_-6px_rgba(252,213,53,0.55)]'
                  : 'hover:bg-base-200/80 hover:translate-x-1',
              ]"
            >
              <component :is="item.icon" class="w-4 h-4 shrink-0" />
              <span>{{ item.label }}</span>
            </NuxtLink>
          </li>
        </ul>

        <div class="p-4 border-t border-base-300/70">
          <button
            @click="handleLogout"
            class="btn btn-outline btn-error btn-sm w-full justify-start gap-2"
          >
            <LogOut class="w-4 h-4" />
            ออกจากระบบ
          </button>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  LayoutDashboard,
  History,
  Wallet,
  Bot,
  Newspaper,
  Calculator,
  Settings,
  Menu,
  Sun,
  Moon,
  LogOut,
} from "lucide-vue-next";
import { useAuthStore } from "~/stores/auth";
import { useLogin } from "~/composables/api/useLogin";
import { useSheetSync } from "~/composables/api/useSheetSync";
import type { SheetConfig } from "~/models/sheet/sheet_model";

const route = useRoute();
const authStore = useAuthStore();
const { logout } = useLogin();
const { fetchSheetConfig } = useSheetSync();

const drawerOpen = ref(false);
const theme = useState<string>("app-theme", () => "binancedark");
const sheetState = useState<SheetConfig | null>("user-sheet-config", () => null);

const navItems = [
  { to: "/dashboard", label: "แดชบอร์ดภาพรวม", icon: LayoutDashboard },
  { to: "/trades", label: "ประวัติเทรด & สแกนรูป", icon: History },
  { to: "/withdrawals", label: "ประวัติการถอน & ภาษี", icon: Wallet },
  { to: "/ai-coach", label: "AI โค้ช & วิเคราะห์กราฟ", icon: Bot },
  { to: "/news", label: "ข่าวคริปโต & วิเคราะห์", icon: Newspaper },
  { to: "/calculator", label: "คำนวณความเสี่ยง", icon: Calculator },
  { to: "/settings", label: "ตั้งค่าพอร์ต & Google Sheets", icon: Settings },
];

const pageTitle = computed(() => {
  const found = navItems.find((i) => i.to === route.path);
  return found ? found.label : "Binance Trading Journal";
});

const userInitial = computed(() => {
  const name = authStore.user?.fullname || "U";
  return name.charAt(0).toUpperCase();
});

const toggleTheme = () => {
  theme.value = theme.value === "binancedark" ? "binancelight" : "binancedark";
  localStorage.setItem("app-theme", theme.value);
};

const handleLogout = () => {
  sheetState.value = null;
  logout();
};

onMounted(async () => {
  if (!sheetState.value && useCookie("token").value) {
    try {
      sheetState.value = await fetchSheetConfig();
    } catch {
      // ข้ามหากโหลดสถานะไม่สำเร็จ
    }
  }
});
</script>
