<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold">ตั้งค่าพอร์ต & การเชื่อมต่อ Google Sheets</h1>
      <p class="text-sm opacity-60">จัดการทุนตั้งต้น เป้าหมายรายวัน และเชื่อมไฟล์ Google Sheets สำหรับบันทึกและลบข้อมูลอัตโนมัติ</p>
    </div>

    <div v-if="loading" class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="skeleton h-80 rounded-2xl"></div>
      <div class="skeleton h-80 rounded-2xl"></div>
    </div>

    <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="card bg-base-100 border border-base-300 shadow-sm">
        <div class="card-body p-6 space-y-4">
          <h3 class="font-bold text-lg flex items-center gap-2">
            <Wallet class="w-5 h-5 text-primary" />
            ตั้งค่าพอร์ตการเทรด
          </h3>

          <form @submit.prevent="savePortfolio" class="space-y-3.5">
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

      <div class="card bg-base-100 border border-base-300 shadow-sm">
        <div class="card-body p-6 space-y-4">
          <h3 class="font-bold text-lg flex items-center gap-2">
            <FileSpreadsheet class="w-5 h-5 text-success" />
            การเชื่อมต่อ Google Sheets
          </h3>

          <form @submit.prevent="saveSheets" class="space-y-3.5">
            <label class="form-control">
              <span class="label-text text-xs mb-1">Spreadsheet ID หรือลิงก์ Google Sheet</span>
              <input
                v-model="sheetForm.spreadsheet_id"
                type="text"
                placeholder="วางลิงก์ Google Sheet หรือรหัส ID"
                class="input input-bordered input-sm"
                required
              />
            </label>

            <label class="form-control">
              <span class="label-text text-xs mb-1">ชื่อแท็บแผ่นงาน (Sheet Tab)</span>
              <input v-model="sheetForm.sheet_name" type="text" placeholder="Trades" class="input input-bordered input-sm" />
            </label>

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
    </div>
  </div>
</template>

<script setup lang="ts">
import { Wallet, FileSpreadsheet } from "lucide-vue-next";
import Swal from "sweetalert2";
import { useSheetSync } from "~/composables/api/useSheetSync";

definePageMeta({ middleware: "auth" });

const {
  fetchPortfolioSetting,
  savePortfolioSetting,
  fetchSheetConfig,
  saveSheetConfig,
  testSheetConnection,
  resyncAllTrades,
} = useSheetSync();

const loading = ref(true);
const savingPortfolio = ref(false);
const savingSheet = ref(false);
const testingSheet = ref(false);
const resyncing = ref(false);
const sheetLastSynced = ref<string | null>(null);

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

onMounted(async () => {
  loading.value = true;
  try {
    const [p, s] = await Promise.all([fetchPortfolioSetting(), fetchSheetConfig()]);
    if (p) {
      portfolioForm.initial_balance = p.initial_balance;
      portfolioForm.currency = p.currency;
      portfolioForm.monthly_target_pct = p.monthly_target_pct;
      portfolioForm.max_daily_loss_usd = p.max_daily_loss_usd;
    }
    if (s) {
      sheetForm.spreadsheet_id = s.spreadsheet_id;
      sheetForm.sheet_name = s.sheet_name || "Trades";
      sheetForm.is_auto_sync = s.is_auto_sync;
      sheetLastSynced.value = s.last_synced_at;
    }
  } finally {
    loading.value = false;
  }
});

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
    sheetForm.spreadsheet_id = updated.spreadsheet_id;
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
</script>
