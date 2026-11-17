<template>
  <div class="space-y-6">
    <div class="agy-fade-up stagger-1 flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold">ประวัติการเทรด & สแกนรูปแคปจอ</h1>
        <p class="text-sm opacity-60">บันทึกข้อมูลอัตโนมัติและซิงค์กับ Google Sheets ส่วนตัวของคุณทันทีเมื่อเพิ่ม/แก้ไข/ลบ</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <a
          v-if="sheetState?.sheet_url"
          :href="sheetState.sheet_url"
          target="_blank"
          rel="noopener noreferrer"
          class="btn btn-outline btn-success btn-sm gap-1.5"
        >
          <ExternalLink class="w-4 h-4" />
          เปิด Google Sheet
        </a>
        <button class="btn btn-primary btn-sm gap-2" @click="openScanModal = true">
          <Sparkles class="w-4 h-4" />
          อัปโหลดรูปแคปจอ Binance
        </button>
        <button class="btn btn-outline btn-sm gap-2" @click="openManualCreate">
          <Plus class="w-4 h-4" />
          เพิ่มรายการเอง
        </button>
      </div>
    </div>

    <div class="card agy-card agy-fade-up stagger-2 bg-base-100/90 border border-base-300 shadow-sm">
      <div class="card-body p-4 flex-row flex-wrap items-center gap-3">
        <input
          v-model="filterSymbol"
          type="text"
          placeholder="ค้นหาคู่เหรียญ เช่น BTC, ETH..."
          class="input input-bordered input-sm w-full sm:w-56"
          @keyup.enter="loadTrades"
        />
        <select v-model="filterSide" class="select select-bordered select-sm" @change="loadTrades">
          <option value="ALL">ทุกฝั่ง (Long / Short / Spot)</option>
          <option value="LONG">เฉพาะ LONG</option>
          <option value="SHORT">เฉพาะ SHORT</option>
          <option value="SPOT">เฉพาะ SPOT</option>
        </select>
        <button class="btn btn-sm btn-ghost gap-1.5" @click="loadTrades">
          <Search class="w-4 h-4" />
          ค้นหา
        </button>
      </div>
    </div>

    <div class="card agy-card agy-fade-up stagger-3 bg-base-100/90 border border-base-300 shadow-sm overflow-hidden">
      <div v-if="tradeStore.loading" class="p-6 space-y-3">
        <div v-for="n in 6" :key="n" class="skeleton h-12 w-full rounded-lg"></div>
      </div>

      <div v-else-if="tradeStore.trades.length === 0" class="p-12 text-center space-y-3">
        <p class="text-sm opacity-60">ยังไม่มีรายการบันทึกการเทรดในระบบ</p>
        <button class="btn btn-primary btn-sm gap-2" @click="openScanModal = true">
          <Sparkles class="w-4 h-4" />
          เริ่มอัปโหลดรูปแคปจอแรกของคุณ
        </button>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="table table-sm w-full">
          <thead>
            <tr class="bg-base-200/70">
              <th>รหัส / เวลา</th>
              <th>คู่เหรียญ</th>
              <th>ฝั่ง / Lev</th>
              <th class="text-right">ราคาเข้า → ออก</th>
              <th class="text-right">กำไร/ขาดทุน</th>
              <th>กลยุทธ์ / อารมณ์</th>
              <th>Google Sheets</th>
              <th class="text-right">จัดการ</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="t in tradeStore.trades" :key="t.trade_id" class="hover">
              <td>
                <div class="font-mono text-xs font-semibold">{{ t.trade_code }}</div>
                <div class="text-[11px] opacity-60">{{ formatTime(t.trade_time) }}</div>
              </td>
              <td class="font-bold">{{ t.symbol }}</td>
              <td>
                <span
                  :class="[
                    'badge badge-sm font-semibold',
                    t.position_side === 'LONG' ? 'badge-success' : t.position_side === 'SHORT' ? 'badge-error' : 'badge-info',
                  ]"
                >
                  {{ t.position_side }} {{ t.leverage }}x
                </span>
              </td>
              <td class="text-right font-mono text-xs">
                {{ t.entry_price }} → {{ t.exit_price }}
              </td>
              <td class="text-right">
                <div :class="['font-bold text-sm', t.pnl >= 0 ? 'text-success' : 'text-error']">
                  {{ t.pnl >= 0 ? '+' : '' }}{{ t.pnl.toFixed(2) }} USDT
                </div>
                <div class="text-[11px] opacity-70">
                  ROI: {{ t.pnl_percentage >= 0 ? '+' : '' }}{{ t.pnl_percentage.toFixed(2) }}%
                </div>
              </td>
              <td>
                <div class="flex flex-wrap gap-1">
                  <span v-if="t.setup_tag" class="badge badge-outline badge-xs">{{ t.setup_tag }}</span>
                  <span v-if="t.emotion_tag" class="badge badge-ghost badge-xs">{{ t.emotion_tag }}</span>
                </div>
                <div v-if="t.notes" class="text-[11px] opacity-60 truncate max-w-[160px] mt-0.5">{{ t.notes }}</div>
              </td>
              <td>
                <span
                  :class="[
                    'badge badge-xs',
                    t.sync_sheet_status === 'Synced'
                      ? 'badge-success badge-outline'
                      : t.sync_sheet_status === 'Failed'
                      ? 'badge-error badge-outline'
                      : 'badge-warning badge-outline',
                  ]"
                >
                  {{ t.sync_sheet_status === 'Synced' ? 'ซิงค์แล้ว' : t.sync_sheet_status === 'Failed' ? 'ซิงค์ไม่สำเร็จ' : 'รอซิงค์' }}
                </span>
              </td>
              <td class="text-right">
                <div class="flex justify-end gap-1">
                  <button class="btn btn-ghost btn-xs" @click="openEdit(t)" title="แก้ไข">
                    <Pencil class="w-3.5 h-3.5" />
                  </button>
                  <button class="btn btn-ghost btn-xs text-error" @click="confirmDelete(t)" title="ลบ">
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <TradeScreenshotUploadModal :open="openScanModal" @close="openScanModal = false" @saved="loadTrades" />
    <TradeFormModal
      :open="openFormModal"
      :edit-trade="selectedTrade"
      @close="openFormModal = false"
      @saved="loadTrades"
    />
  </div>
</template>

<script setup lang="ts">
import { Sparkles, Plus, Search, Pencil, Trash2, ExternalLink } from "lucide-vue-next";
import Swal from "sweetalert2";
import { useTrades } from "~/composables/api/useTrades";
import { useTradeStore } from "~/stores/trade_store";
import type { TradeHistory } from "~/models/trade/response/trade_res";
import type { SheetConfig } from "~/models/sheet/sheet_model";

definePageMeta({ middleware: "auth" });

const sheetState = useState<SheetConfig | null>("user-sheet-config", () => null);
const tradeStore = useTradeStore();
const { fetchTrades, deleteTrade } = useTrades();

const filterSymbol = ref("");
const filterSide = ref("ALL");
const openScanModal = ref(false);
const openFormModal = ref(false);
const selectedTrade = ref<TradeHistory | null>(null);

const loadTrades = async () => {
  await fetchTrades(filterSymbol.value, filterSide.value);
};

const openManualCreate = () => {
  selectedTrade.value = null;
  openFormModal.value = true;
};

const openEdit = (t: TradeHistory) => {
  selectedTrade.value = t;
  openFormModal.value = true;
};

const confirmDelete = async (t: TradeHistory) => {
  const result = await Swal.fire({
    title: "ยืนยันการลบ?",
    text: `ต้องการลบรายการ ${t.trade_code} (${t.symbol}) และลบออกจาก Google Sheets ด้วยใช่หรือไม่?`,
    icon: "warning",
    showCancelButton: true,
    confirmButtonText: "ใช่, ลบเลย",
    cancelButtonText: "ยกเลิก",
    confirmButtonColor: "#F6465D",
  });
  if (!result.isConfirmed) return;

  try {
    await deleteTrade(t.trade_id);
    Swal.fire({ icon: "success", title: "ลบเรียบร้อยแล้ว", timer: 1300, showConfirmButton: false });
    await loadTrades();
  } catch (err: any) {
    Swal.fire("เกิดข้อผิดพลาด", err.response?.data?.message || "ลบไม่สำเร็จ", "error");
  }
};

const formatTime = (iso: string) => {
  if (!iso) return "-";
  return new Date(iso).toLocaleString("th-TH", {
    day: "2-digit",
    month: "2-digit",
    year: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
};

onMounted(() => {
  loadTrades();
});
</script>
