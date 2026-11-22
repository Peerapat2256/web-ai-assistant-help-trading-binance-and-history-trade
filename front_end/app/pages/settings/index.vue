<template>
  <div class="space-y-6">
    <div class="agy-fade-up stagger-1">
      <h1 class="text-2xl font-bold">ตั้งค่าพอร์ต & การเชื่อมต่อ Google Sheets</h1>
      <p class="text-sm opacity-60">แยกไฟล์ Google Sheet ส่วนตัวตามบัญชีผู้ใช้ พร้อมซิงค์เพิ่ม/แก้ไข/ลบอัตโนมัติ</p>
    </div>

    <div v-if="loading" class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="skeleton h-96 rounded-2xl"></div>
      <div class="skeleton h-96 rounded-2xl"></div>
    </div>

    <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="card agy-card agy-fade-up stagger-2 bg-base-100/90 border border-base-300 shadow-sm">
        <div class="card-body p-6 space-y-4">
          <h3 class="font-bold text-lg flex items-center gap-2">
            <Wallet class="w-5 h-5 text-primary" />
            ตั้งค่าพอร์ตการเทรด
          </h3>

          <form @submit.prevent="savePortfolio" class="space-y-4">
            <label class="form-control">
              <span class="label-text text-xs mb-1">ทุนเริ่มต้นทั้งหมด (USDT)</span>
              <input v-model.number="portfolioForm.initial_balance" type="number" step="any" class="input input-bordered input-sm" required />
            </label>

            <label class="form-control">
              <span class="label-text text-xs mb-1">เป้าหมายกำไรต่อเดือน (%)</span>
              <input v-model.number="portfolioForm.monthly_target_pct" type="number" step="any" class="input input-bordered input-sm" />
            </label>

            <label class="form-control">
              <span class="label-text text-xs mb-1">เพดานขาดทุนสูงสุดต่อวัน (USDT) — แจ้งเตือนเมื่อเกิน</span>
              <input v-model.number="portfolioForm.max_daily_loss_usd" type="number" step="any" class="input input-bordered input-sm" />
            </label>

            <button type="submit" class="btn btn-primary btn-sm mt-2" :disabled="savingPortfolio">
              {{ savingPortfolio ? "กำลังบันทึก..." : "บันทึกการตั้งค่าพอร์ต" }}
            </button>
          </form>
        </div>
      </div>

      <div class="card agy-card agy-fade-up stagger-3 bg-base-100/90 border border-base-300 shadow-sm">
        <div class="card-body p-6 space-y-4">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <h3 class="font-bold text-lg flex items-center gap-2">
              <FileSpreadsheet class="w-5 h-5 text-success" />
              การเชื่อมต่อ Google Sheets ส่วนตัว
            </h3>
            <a
              v-if="sheetUrl"
              :href="sheetUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-xs btn-outline btn-success gap-1"
            >
              เปิดดู Google Sheet ของฉัน
              <ExternalLink class="w-3.5 h-3.5" />
            </a>
          </div>

          <div class="p-3.5 rounded-xl bg-base-200/70 border border-base-300/80 space-y-2 text-xs">
            <div class="font-semibold text-primary">
              คู่มือเชื่อมต่อ Google Sheet ส่วนตัว (แยกข้อมูลของใครของมัน 100%)
            </div>
            <ol class="list-decimal list-inside space-y-1 opacity-80">
              <li>สร้างไฟล์ Google Sheet เปล่าของตัวเองบน Google Drive</li>
              <li>กดปุ่ม <b>แชร์ (Share)</b> แล้วเพิ่มอีเมลบอทด้านล่างโดยตั้งสิทธิ์เป็น <b>Editor (ผู้แก้ไข)</b></li>
              <li>วางลิงก์ Google Sheet ของคุณในช่องด้านล่าง แล้วกด <b>บันทึก</b> และ <b>ทดสอบการเชื่อมต่อ</b></li>
            </ol>

            <div class="flex flex-wrap items-center gap-2 pt-1">
              <code class="px-2.5 py-1.5 rounded-lg bg-base-100 border border-base-300 text-[11px] flex-1 truncate select-all">
                {{ botEmail || 'inance-sheet-bot@round-music-482118-b7.iam.gserviceaccount.com' }}
              </code>
              <button type="button" class="btn btn-xs btn-primary gap-1" @click="copyBotEmail">
                <Check v-if="copied" class="w-3.5 h-3.5" />
                <Copy v-else class="w-3.5 h-3.5" />
                {{ copied ? 'คัดลอกแล้ว' : 'คัดลอกอีเมลบอท' }}
              </button>
            </div>
          </div>

          <form @submit.prevent="saveSheets" class="space-y-3.5">
            <label class="form-control">
              <span class="label-text text-xs mb-1">Spreadsheet ID หรือลิงก์ Google Sheet ของคุณ</span>
              <input
                v-model="sheetForm.spreadsheet_id"
                type="text"
                placeholder="วางลิงก์ https://docs.google.com/spreadsheets/d/... หรือรหัส ID"
                class="input input-bordered input-sm"
              />
            </label>

            <div class="form-control">
              <div class="flex items-center justify-between mb-1">
                <span class="label-text text-xs">ชื่อแท็บแผ่นงาน (Sheet Tab)</span>
                <button
                  type="button"
                  class="link link-primary text-[11px]"
                  @click="setPersonalTabName"
                >
                  ตั้งชื่อแท็บแยกตามชื่อฉันอัตโนมัติ
                </button>
              </div>
              <input v-model="sheetForm.sheet_name" type="text" placeholder="Trades" class="input input-bordered input-sm" />
            </div>

            <label class="flex items-center justify-between cursor-pointer p-3 rounded-xl bg-base-200/60">
              <div>
                <div class="font-semibold text-sm">ซิงค์อัตโนมัติ (Auto Sync)</div>
                <div class="text-xs opacity-60">เพิ่ม แก้ไข หรือลบข้อมูลใน Sheet ทันทีเมื่อทำรายการบนเว็บ</div>
              </div>
              <input v-model="sheetForm.is_auto_sync" type="checkbox" class="toggle toggle-success" />
            </label>

            <div v-if="sheetLastSynced" class="text-xs opacity-60">
              ซิงค์ล่าสุด: {{ new Date(sheetLastSynced).toLocaleString("th-TH") }}
            </div>

            <div class="flex flex-wrap gap-2 pt-2">
              <button type="submit" class="btn btn-primary btn-sm" :disabled="savingSheet">
                {{ savingSheet ? "กำลังบันทึก..." : "บันทึกการตั้งค่าชีต" }}
              </button>
              <button type="button" class="btn btn-outline btn-success btn-sm" :disabled="testingSheet" @click="testConnection">
                {{ testingSheet ? "กำลังทดสอบ..." : "ทดสอบการเชื่อมต่อ & สร้างหัวตาราง" }}
              </button>
              <button type="button" class="btn btn-outline btn-sm" :disabled="resyncing" @click="resyncAll">
                {{ resyncing ? "กำลังซิงค์..." : "ซิงค์ข้อมูลทั้งหมดลงชีตใหม่" }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- เชื่อมต่อ Binance Futures API (Auto-Sync) -->
      <div class="card agy-card agy-fade-up stagger-4 bg-base-100/90 border border-base-300 shadow-sm lg:col-span-2">
        <div class="card-body p-6 space-y-4">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <h3 class="font-bold text-lg flex items-center gap-2">
              <RefreshCw class="w-5 h-5 text-primary" />
              เชื่อมต่อ Binance Futures API (ดึงประวัติเทรดอัตโนมัติ)
            </h3>
            <span
              :class="[
                'badge badge-sm gap-1.5',
                binanceConfig?.connected ? 'badge-success badge-outline' : 'badge-ghost',
              ]"
            >
              <span
                :class="[
                  'w-2 h-2 rounded-full',
                  binanceConfig?.connected ? 'bg-success agy-pulse-dot' : 'bg-base-content/40',
                ]"
              ></span>
              {{ binanceConfig?.connected ? `เชื่อมต่อแล้ว (${binanceConfig.api_key_masked})` : 'ยังไม่ได้เชื่อมต่อ' }}
            </span>
          </div>

          <div class="p-3.5 rounded-xl bg-base-200/70 border border-base-300/80 space-y-1.5 text-xs">
            <div class="font-semibold text-primary">วิธีสร้าง API Key แบบปลอดภัย 100% (อ่านได้อย่างเดียว — Read-Only)</div>
            <ol class="list-decimal list-inside space-y-1 opacity-80">
              <li>เข้า Binance → <b>API Management</b> → กด <b>Create API</b> (เลือก System generated)</li>
              <li>เปิดสิทธิ์แค่ <b>Enable Reading</b> เท่านั้น (<span class="text-error font-semibold">ห้ามติ๊ก</span> Enable Futures / Enable Withdrawals)</li>
              <li>คัดลอก <b>API Key</b> และ <b>Secret Key</b> มาวางด้านล่าง (ระบบเข้ารหัส AES-256 ก่อนเก็บลงฐานข้อมูลแยกตามบัญชีของคุณ)</li>
            </ol>
          </div>

          <div v-if="binanceConfig?.last_sync_error" class="alert alert-warning text-xs py-2">
            <span>หมายเหตุการซิงค์ล่าสุด: {{ binanceConfig.last_sync_error }}</span>
          </div>

          <form @submit.prevent="saveBinance" class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <label class="form-control">
                <span class="label-text text-xs mb-1">
                  Binance API Key {{ binanceConfig?.connected ? '(เว้นว่างหากใช้ Key เดิม)' : '' }}
                </span>
                <input
                  v-model="binanceForm.api_key"
                  type="text"
                  :placeholder="binanceConfig?.connected ? binanceConfig.api_key_masked : 'วาง Binance API Key'"
                  class="input input-bordered input-sm font-mono"
                />
              </label>

              <label class="form-control">
                <span class="label-text text-xs mb-1">
                  Binance Secret Key {{ binanceConfig?.connected ? '(เว้นว่างหากใช้ Secret เดิม)' : '' }}
                </span>
                <input
                  v-model="binanceForm.api_secret"
                  type="password"
                  :placeholder="binanceConfig?.connected ? '••••••••••••••••••••' : 'วาง Binance Secret Key'"
                  class="input input-bordered input-sm font-mono"
                />
              </label>
            </div>

            <label class="flex items-center justify-between cursor-pointer p-3 rounded-xl bg-base-200/60">
              <div>
                <div class="font-semibold text-sm">ดึงประวัติเทรดอัตโนมัติ (Auto-Sync ทุก 1 นาที)</div>
                <div class="text-xs opacity-60">
                  เมื่อปิด Position บน Binance Futures ระบบจะดึงเข้าประวัติเทรด เช็คซ้ำ และซิงค์เข้า Google Sheets ให้อัตโนมัติ
                </div>
              </div>
              <input v-model="binanceForm.is_auto_sync" type="checkbox" class="toggle toggle-primary" />
            </label>

            <div v-if="binanceConfig?.last_synced_at" class="text-xs opacity-60">
              ซิงค์จาก Binance ล่าสุด: {{ new Date(binanceConfig.last_synced_at).toLocaleString("th-TH") }}
            </div>

            <div class="flex flex-wrap gap-2 pt-1">
              <button type="submit" class="btn btn-primary btn-sm" :disabled="savingBinance">
                {{ savingBinance ? "กำลังบันทึก..." : "บันทึกการเชื่อมต่อ Binance" }}
              </button>
              <button type="button" class="btn btn-outline btn-success btn-sm" :disabled="testingBinance" @click="testBinance">
                {{ testingBinance ? "กำลังทดสอบ..." : "ทดสอบการเชื่อมต่อ" }}
              </button>
              <button
                v-if="binanceConfig?.connected"
                type="button"
                class="btn btn-outline btn-info btn-sm"
                :disabled="syncingBinance"
                @click="syncBinance(7)"
              >
                {{ syncingBinance ? "กำลังดึงข้อมูล..." : "ดึงประวัติ 7 วันล่าสุด" }}
              </button>
              <button
                v-if="binanceConfig?.connected"
                type="button"
                class="btn btn-outline btn-sm"
                :disabled="syncingBinance"
                @click="syncBinance(30)"
              >
                ดึงย้อนหลัง 30 วัน
              </button>
              <button
                v-if="binanceConfig?.connected"
                type="button"
                class="btn btn-ghost btn-sm text-error ml-auto"
                @click="disconnectBinance"
              >
                ยกเลิกการเชื่อมต่อ
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Wallet, FileSpreadsheet, ExternalLink, Copy, Check, RefreshCw } from "lucide-vue-next";
import Swal from "sweetalert2";
import { useAuthStore } from "~/stores/auth";
import { useSheetSync } from "~/composables/api/useSheetSync";
import { useBinanceSync, type BinanceConfig } from "~/composables/api/useBinanceSync";
import type { SheetConfig } from "~/models/sheet/sheet_model";

definePageMeta({ middleware: "auth" });

const authStore = useAuthStore();
const sheetState = useState<SheetConfig | null>("user-sheet-config", () => null);

const {
  fetchPortfolioSetting,
  savePortfolioSetting,
  fetchSheetConfig,
  saveSheetConfig,
  testSheetConnection,
  resyncAllTrades,
} = useSheetSync();

const {
  fetchBinanceConfig,
  saveBinanceConfig,
  deleteBinanceConfig,
  testBinanceConnection,
  syncBinanceTrades,
} = useBinanceSync();

const loading = ref(true);
const savingPortfolio = ref(false);
const savingSheet = ref(false);
const testingSheet = ref(false);
const resyncing = ref(false);
const copied = ref(false);
const sheetLastSynced = ref<string | null>(null);
const botEmail = ref("");
const sheetUrl = ref("");

const binanceConfig = ref<BinanceConfig | null>(null);
const savingBinance = ref(false);
const testingBinance = ref(false);
const syncingBinance = ref(false);
const binanceForm = reactive({
  api_key: "",
  api_secret: "",
  is_auto_sync: true,
});

const portfolioForm = reactive({
  initial_balance: 1000,
  currency: "USDT",
  monthly_target_pct: 10,
  max_daily_loss_usd: 50,
});

const sheetForm = reactive({
  spreadsheet_id: "",
  sheet_name: "Trades",
  is_auto_sync: true,
});

const syncLocalSheetState = (s: SheetConfig) => {
  sheetForm.spreadsheet_id = s.spreadsheet_id;
  sheetForm.sheet_name = s.sheet_name || "Trades";
  sheetForm.is_auto_sync = s.is_auto_sync;
  sheetLastSynced.value = s.last_synced_at;
  botEmail.value = s.bot_email || "";
  sheetUrl.value = s.sheet_url || "";
  sheetState.value = s;
};

onMounted(async () => {
  loading.value = true;
  try {
    const [p, s, b] = await Promise.all([
      fetchPortfolioSetting(),
      fetchSheetConfig(),
      fetchBinanceConfig().catch(() => null),
    ]);
    if (p) {
      portfolioForm.initial_balance = p.initial_balance;
      portfolioForm.currency = p.currency;
      portfolioForm.monthly_target_pct = p.monthly_target_pct;
      portfolioForm.max_daily_loss_usd = p.max_daily_loss_usd;
    }
    if (s) {
      syncLocalSheetState(s);
    }
    if (b) {
      binanceConfig.value = b;
      binanceForm.is_auto_sync = b.is_auto_sync;
    }
  } finally {
    loading.value = false;
  }
});

const copyBotEmail = async () => {
  const email = botEmail.value || "inance-sheet-bot@round-music-482118-b7.iam.gserviceaccount.com";
  await navigator.clipboard.writeText(email);
  copied.value = true;
  setTimeout(() => {
    copied.value = false;
  }, 2000);
};

const setPersonalTabName = () => {
  const rawName = (authStore.user?.fullname || "").replace(/[^a-zA-Z0-9ก-๙_]/g, "").slice(0, 18);
  const uid = authStore.user?.id || 1;
  sheetForm.sheet_name = rawName ? `Trades_${rawName}_${uid}` : `Trades_ID${uid}`;
};

const savePortfolio = async () => {
  savingPortfolio.value = true;
  try {
    await savePortfolioSetting({ ...portfolioForm });
    Swal.fire({ icon: "success", title: "บันทึกการตั้งค่าพอร์ตสำเร็จ!", timer: 1400, showConfirmButton: false });
  } catch (err: any) {
    Swal.fire("เกิดข้อผิดพลาด", err.response?.data?.message || "บันทึกไม่สำเร็จ", "error");
  } finally {
    savingPortfolio.value = false;
  }
};

const saveSheets = async () => {
  savingSheet.value = true;
  try {
    const updated = await saveSheetConfig({ ...sheetForm });
    syncLocalSheetState(updated);
    Swal.fire({ icon: "success", title: "บันทึกการตั้งค่า Google Sheets สำเร็จ!", timer: 1400, showConfirmButton: false });
  } catch (err: any) {
    Swal.fire("เกิดข้อผิดพลาด", err.response?.data?.message || "บันทึกไม่สำเร็จ", "error");
  } finally {
    savingSheet.value = false;
  }
};

const testConnection = async () => {
  testingSheet.value = true;
  try {
    const res = await testSheetConnection({
      spreadsheet_id: sheetForm.spreadsheet_id,
      sheet_name: sheetForm.sheet_name,
    });
    Swal.fire("เชื่อมต่อสำเร็จ!", res.message, "success");
  } catch (err: any) {
    Swal.fire("เชื่อมต่อไม่สำเร็จ", err.response?.data?.message || "กรุณาตรวจสอบการแชร์สิทธิ์ให้บอท", "error");
  } finally {
    testingSheet.value = false;
  }
};

const resyncAll = async () => {
  resyncing.value = true;
  try {
    const res = await resyncAllTrades();
    Swal.fire("ซิงค์สำเร็จ!", res.message, "success");
  } catch (err: any) {
    Swal.fire("ซิงค์ไม่สำเร็จ", err.response?.data?.message || "เกิดข้อผิดพลาด", "error");
  } finally {
    resyncing.value = false;
  }
};

const saveBinance = async () => {
  savingBinance.value = true;
  try {
    const updated = await saveBinanceConfig({
      api_key: binanceForm.api_key || undefined,
      api_secret: binanceForm.api_secret || undefined,
      is_auto_sync: binanceForm.is_auto_sync,
    });
    binanceConfig.value = updated;
    binanceForm.api_key = "";
    binanceForm.api_secret = "";
    Swal.fire({ icon: "success", title: "บันทึกการเชื่อมต่อ Binance สำเร็จ!", timer: 1500, showConfirmButton: false });
  } catch (err: any) {
    Swal.fire("บันทึกไม่สำเร็จ", err.response?.data?.message || "กรุณาตรวจสอบ API Key และ Secret Key", "error");
  } finally {
    savingBinance.value = false;
  }
};

const testBinance = async () => {
  testingBinance.value = true;
  try {
    const res = await testBinanceConnection({
      api_key: binanceForm.api_key || undefined,
      api_secret: binanceForm.api_secret || undefined,
    });
    Swal.fire("เชื่อมต่อ Binance สำเร็จ!", res.message, "success");
  } catch (err: any) {
    Swal.fire("เชื่อมต่อไม่สำเร็จ", err.response?.data?.message || "กรุณาตรวจสอบสิทธิ์ Enable Reading ของ API Key", "error");
  } finally {
    testingBinance.value = false;
  }
};

const syncBinance = async (days: number) => {
  syncingBinance.value = true;
  try {
    const res = await syncBinanceTrades(days);
    Swal.fire("ซิงค์ประวัติเทรดสำเร็จ!", res.message, "success");
    binanceConfig.value = await fetchBinanceConfig();
  } catch (err: any) {
    Swal.fire("ซิงค์ไม่สำเร็จ", err.response?.data?.message || "เกิดข้อผิดพลาดในการดึงข้อมูลจาก Binance", "error");
  } finally {
    syncingBinance.value = false;
  }
};

const disconnectBinance = async () => {
  const res = await Swal.fire({
    title: "ยกเลิกการเชื่อมต่อ Binance?",
    text: "ระบบจะลบ API Key และ Secret Key ของคุณออกจากฐานข้อมูลทันที",
    icon: "warning",
    showCancelButton: true,
    confirmButtonText: "ใช่, ลบเลย",
    cancelButtonText: "ยกเลิก",
    confirmButtonColor: "#F6465D",
  });
  if (!res.isConfirmed) return;

  try {
    await deleteBinanceConfig();
    binanceConfig.value = { connected: false, api_key_masked: "", is_auto_sync: true, last_synced_at: null, last_sync_error: "" };
    binanceForm.api_key = "";
    binanceForm.api_secret = "";
    Swal.fire({ icon: "success", title: "ยกเลิกการเชื่อมต่อเรียบร้อยแล้ว", timer: 1300, showConfirmButton: false });
  } catch (err: any) {
    Swal.fire("เกิดข้อผิดพลาด", err.response?.data?.message || "ลบไม่สำเร็จ", "error");
  }
};
</script>
