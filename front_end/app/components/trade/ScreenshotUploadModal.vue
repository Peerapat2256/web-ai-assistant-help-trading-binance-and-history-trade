<template>
  <dialog :open="open" class="modal modal-bottom sm:modal-middle bg-black/60">
    <div :class="['modal-box bg-base-100 border border-base-300', isMultiMode ? 'max-w-4xl' : 'max-w-2xl']">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center gap-2">
          <Sparkles class="w-5 h-5 text-primary" />
          <h3 class="font-bold text-lg">สแกนรูปแคปจอ Binance ด้วย AI อัจฉริยะ</h3>
        </div>
        <button class="btn btn-sm btn-circle btn-ghost" @click="closeModal">✕</button>
      </div>

      <div v-if="step === 'upload'" class="space-y-4">
        <label
          @dragover.prevent="isDragging = true"
          @dragleave.prevent="isDragging = false"
          @drop.prevent="onDropFile"
          :class="[
            'flex flex-col items-center justify-center border-2 border-dashed rounded-2xl p-8 cursor-pointer transition-all text-center',
            isDragging ? 'border-primary bg-primary/10 scale-[1.01]' : 'border-base-300 hover:border-primary bg-base-200/40',
          ]"
        >
          <UploadCloud class="w-10 h-10 text-primary mb-2" />
          <span class="font-semibold text-sm">คลิกเลือกรูป, ลากวาง, หรือกด Ctrl + V วางรูปที่แคปได้เลย</span>
          <span class="text-xs opacity-60 mt-1">
            รองรับทั้งรูปออเดอร์เดี่ยว (PnL Card) และรูปประวัติหลายออเดอร์พร้อมกัน (Position History) + เช็คซ้ำอัตโนมัติ
          </span>
          <input ref="fileInputRef" type="file" accept="image/*" class="hidden" @change="onFileSelect" />
        </label>

        <div v-if="previewUrl && scanning" class="rounded-xl border border-base-300 bg-base-200/50 p-2">
          <img :src="previewUrl" alt="ภาพที่กำลังสแกน" class="w-full max-h-40 object-contain rounded-lg" />
        </div>

        <div v-if="scanning" class="space-y-3 p-4 rounded-xl bg-base-200/60">
          <div class="text-xs font-medium text-primary">AI กำลังวิเคราะห์รูปภาพ ตรวจสอบจำนวนออเดอร์ และเช็ครายการซ้ำ...</div>
          <div class="skeleton h-5 w-2/3"></div>
          <div class="skeleton h-10 w-full"></div>
          <div class="skeleton h-10 w-full"></div>
        </div>
      </div>

      <!-- โหมดหลายออเดอร์ (Multi-Order Batch Preview) -->
      <div v-else-if="isMultiMode" class="space-y-4">
        <div class="flex flex-col sm:flex-row gap-3 items-start">
          <div v-if="previewUrl" class="w-full sm:w-32 shrink-0 rounded-xl border border-base-300 bg-base-200/50 p-1.5">
            <img :src="previewUrl" alt="ภาพแคปจอที่อัปโหลด" class="w-full max-h-36 object-contain rounded-lg" />
          </div>
          <div class="flex-1 space-y-2 w-full">
            <div class="alert alert-success text-xs py-2.5">
              <div class="flex flex-wrap items-center gap-2 w-full justify-between">
                <span>{{ parsedSummary }}</span>
                <div class="flex gap-1.5">
                  <span class="badge badge-sm badge-neutral">พบ {{ batchItems.length }} รายการ</span>
                  <span v-if="duplicateCount > 0" class="badge badge-sm badge-warning">ซ้ำ {{ duplicateCount }} รายการ</span>
                </div>
              </div>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
              <label class="form-control">
                <span class="label-text text-[11px] mb-0.5">กลยุทธ์รวมสำหรับชุดนี้ (ไม่บังคับ)</span>
                <select v-model="batchSetupTag" class="select select-bordered select-xs">
                  <option value="">-- ไม่ระบุ --</option>
                  <option value="Breakout">Breakout</option>
                  <option value="แนวรับ-แนวต้าน">แนวรับ-แนวต้าน</option>
                  <option value="SMC / Order Block">SMC / Order Block</option>
                  <option value="EMA Cross">EMA Cross</option>
                  <option value="เทรดตามข่าว">เทรดตามข่าว</option>
                </select>
              </label>
              <label class="form-control">
                <span class="label-text text-[11px] mb-0.5">อารมณ์รวมสำหรับชุดนี้ (ไม่บังคับ)</span>
                <select v-model="batchEmotionTag" class="select select-bordered select-xs">
                  <option value="">-- ไม่ระบุ --</option>
                  <option value="ทำตามแผนเป๊ะ">ทำตามแผนเป๊ะ</option>
                  <option value="มั่นใจ">มั่นใจ</option>
                  <option value="รีบเข้า (FOMO)">รีบเข้า (FOMO)</option>
                  <option value="หัวร้อนเอาคืน">หัวร้อนเอาคืน</option>
                  <option value="ปิดหมูก่อนเป้า">ปิดหมูก่อนเป้า</option>
                </select>
              </label>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-between text-xs px-1">
          <div class="flex items-center gap-2">
            <button type="button" class="btn btn-ghost btn-xs" @click="selectAllNew">เลือกเฉพาะรายการใหม่</button>
            <button type="button" class="btn btn-ghost btn-xs" @click="toggleAll(true)">เลือกทั้งหมด</button>
            <button type="button" class="btn btn-ghost btn-xs" @click="toggleAll(false)">ยกเลิกทั้งหมด</button>
          </div>
          <span class="opacity-70">เลือกอยู่ <strong class="text-primary">{{ selectedBatchCount }}</strong> / {{ batchItems.length }} รายการ</span>
        </div>

        <div class="max-h-80 overflow-y-auto space-y-2.5 pr-1">
          <div
            v-for="(item, idx) in batchItems"
            :key="idx"
            :class="[
              'p-3 rounded-xl border transition-all',
              item.is_duplicate
                ? 'border-warning/50 bg-warning/5'
                : item.selected
                ? 'border-primary/50 bg-base-200/40'
                : 'border-base-300 bg-base-200/20 opacity-60',
            ]"
          >
            <div class="flex flex-wrap items-center justify-between gap-2 mb-2">
              <label class="flex items-center gap-2 cursor-pointer">
                <input v-model="item.selected" type="checkbox" class="checkbox checkbox-primary checkbox-sm" />
                <span class="font-bold text-sm">#{{ idx + 1 }} {{ item.symbol }}</span>
                <span
                  :class="[
                    'badge badge-xs font-semibold',
                    item.position_side === 'LONG' ? 'badge-success' : item.position_side === 'SHORT' ? 'badge-error' : 'badge-info',
                  ]"
                >
                  {{ item.position_side }} {{ item.leverage }}x
                </span>
              </label>
              <div class="flex items-center gap-2">
                <span v-if="item.is_duplicate" class="badge badge-warning badge-xs gap-1">
                  ซ้ำ: {{ item.duplicate_reason || 'มีในระบบแล้ว' }}
                </span>
                <span v-else class="badge badge-success badge-outline badge-xs">รายการใหม่</span>
                <span :class="['font-bold text-sm font-mono', item.pnl >= 0 ? 'text-success' : 'text-error']">
                  {{ item.pnl >= 0 ? '+' : '' }}{{ Number(item.pnl).toFixed(2) }} USDT
                </span>
              </div>
            </div>

            <div class="grid grid-cols-2 sm:grid-cols-6 gap-2">
              <label class="form-control">
                <span class="text-[10px] opacity-60">คู่เหรียญ</span>
                <input v-model="item.symbol" type="text" class="input input-bordered input-xs" />
              </label>
              <label class="form-control">
                <span class="text-[10px] opacity-60">ฝั่ง</span>
                <select v-model="item.position_side" class="select select-bordered select-xs">
                  <option value="LONG">LONG</option>
                  <option value="SHORT">SHORT</option>
                  <option value="SPOT">SPOT</option>
                </select>
              </label>
              <label class="form-control">
                <span class="text-[10px] opacity-60">Entry</span>
                <input v-model.number="item.entry_price" type="number" step="any" class="input input-bordered input-xs" />
              </label>
              <label class="form-control">
                <span class="text-[10px] opacity-60">Exit</span>
                <input v-model.number="item.exit_price" type="number" step="any" class="input input-bordered input-xs" />
              </label>
              <label class="form-control">
                <span class="text-[10px] opacity-60">PnL / ROI%</span>
                <div class="flex gap-1">
                  <input v-model.number="item.pnl" type="number" step="any" class="input input-bordered input-xs w-1/2" />
                  <input v-model.number="item.pnl_percentage" type="number" step="any" class="input input-bordered input-xs w-1/2" />
                </div>
              </label>
              <label class="form-control">
                <span class="text-[10px] opacity-60">เวลาปิดออเดอร์</span>
                <input v-model="item.trade_time" type="text" class="input input-bordered input-xs" />
              </label>
            </div>
          </div>
        </div>

        <div class="modal-action">
          <button type="button" class="btn btn-ghost btn-sm" @click="resetToUpload">อัปโหลดรูปใหม่</button>
          <button
            type="button"
            class="btn btn-primary btn-sm"
            :disabled="saving || selectedBatchCount === 0"
            @click="submitBatch"
          >
            {{ saving ? 'กำลังบันทึกและซิงค์...' : `บันทึก ${selectedBatchCount} รายการที่เลือกและซิงค์ Google Sheets` }}
          </button>
        </div>
      </div>

      <!-- โหมดออเดอร์เดี่ยว (Single Trade Preview) -->
      <form v-else @submit.prevent="submitForm" class="space-y-4">
        <div v-if="previewUrl" class="rounded-xl border border-base-300 bg-base-200/50 p-2">
          <img :src="previewUrl" alt="ภาพแคปจอที่อัปโหลด" class="w-full max-h-36 object-contain rounded-lg" />
        </div>

        <div
          :class="[
            'alert text-xs',
            parsedNoteConfidence === 'ai_verified' ? 'alert-success' : 'alert-warning',
          ]"
        >
          <span>{{ parsedSummary }}</span>
        </div>

        <div v-if="singleDuplicateReason" class="alert alert-warning text-xs py-2">
          <span>คำเตือนรายการซ้ำ: {{ singleDuplicateReason }} (หากต้องการบันทึกซ้ำสามารถกดยืนยันได้)</span>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <label class="form-control">
            <span class="label-text text-xs mb-1">คู่เหรียญ</span>
            <input v-model="form.symbol" type="text" class="input input-bordered input-sm" required />
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">ฝั่งที่เทรด</span>
            <select v-model="form.position_side" class="select select-bordered select-sm">
              <option value="LONG">LONG</option>
              <option value="SHORT">SHORT</option>
              <option value="SPOT">SPOT</option>
            </select>
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">Leverage (x)</span>
            <input v-model.number="form.leverage" type="number" min="1" class="input input-bordered input-sm" />
          </label>
        </div>

        <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <label class="form-control">
            <span class="label-text text-xs mb-1">ราคาเข้า (Entry)</span>
            <input v-model.number="form.entry_price" type="number" step="any" class="input input-bordered input-sm" />
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">ราคาปิด (Exit)</span>
            <input v-model.number="form.exit_price" type="number" step="any" class="input input-bordered input-sm" />
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">กำไร/ขาดทุน (USDT)</span>
            <input v-model.number="form.pnl" type="number" step="any" class="input input-bordered input-sm" required />
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">ROI (%)</span>
            <input v-model.number="form.pnl_percentage" type="number" step="any" class="input input-bordered input-sm" />
          </label>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <label class="form-control">
            <span class="label-text text-xs mb-1">วัน-เวลาเทรด</span>
            <input v-model="form.trade_time" type="text" class="input input-bordered input-sm" />
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">กลยุทธ์ที่ใช้</span>
            <select v-model="form.setup_tag" class="select select-bordered select-sm">
              <option value="">-- เลือกกลยุทธ์ --</option>
              <option value="Breakout">Breakout</option>
              <option value="แนวรับ-แนวต้าน">แนวรับ-แนวต้าน</option>
              <option value="SMC / Order Block">SMC / Order Block</option>
              <option value="EMA Cross">EMA Cross</option>
              <option value="เทรดตามข่าว">เทรดตามข่าว</option>
            </select>
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">อารมณ์ตอนเทรด</span>
            <select v-model="form.emotion_tag" class="select select-bordered select-sm">
              <option value="">-- เลือกอารมณ์ --</option>
              <option value="ทำตามแผนเป๊ะ">ทำตามแผนเป๊ะ</option>
              <option value="มั่นใจ">มั่นใจ</option>
              <option value="รีบเข้า (FOMO)">รีบเข้า (FOMO)</option>
              <option value="หัวร้อนเอาคืน">หัวร้อนเอาคืน</option>
              <option value="ปิดหมูก่อนเป้า">ปิดหมูก่อนเป้า</option>
            </select>
          </label>
        </div>

        <label class="form-control">
          <span class="label-text text-xs mb-1">หมายเหตุเพิ่มเติม</span>
          <textarea v-model="form.notes" rows="2" class="textarea textarea-bordered textarea-sm"></textarea>
        </label>

        <div class="modal-action">
          <button type="button" class="btn btn-ghost btn-sm" @click="resetToUpload">อัปโหลดรูปใหม่</button>
          <button type="submit" class="btn btn-primary btn-sm" :disabled="saving">
            {{ saving ? 'กำลังบันทึกและซิงค์...' : 'ยืนยันบันทึกและซิงค์ Google Sheets' }}
          </button>
        </div>
      </form>
    </div>
  </dialog>
</template>

<script setup lang="ts">
import { Sparkles, UploadCloud } from "lucide-vue-next";
import Swal from "sweetalert2";
import { useTrades } from "~/composables/api/useTrades";
import type { TradeCreateRequest } from "~/models/trade/request/trade_req";

interface BatchSelectableTrade extends TradeCreateRequest {
  selected: boolean;
  is_duplicate: boolean;
  duplicate_reason: string;
}

const props = defineProps<{
  open: boolean;
  initialFile?: File | null;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "saved"): void;
}>();

const { parseScreenshot, createTrade, createTradesBatch } = useTrades();
const step = ref<"upload" | "preview">("upload");
const scanning = ref(false);
const saving = ref(false);
const isDragging = ref(false);
const parsedSummary = ref("");
const parsedNoteConfidence = ref("");
const singleDuplicateReason = ref("");
const previewUrl = ref<string | null>(null);
const fileInputRef = ref<HTMLInputElement | null>(null);

const batchItems = ref<BatchSelectableTrade[]>([]);
const batchSetupTag = ref("");
const batchEmotionTag = ref("");

const isMultiMode = computed(() => batchItems.value.length > 1);
const duplicateCount = computed(() => batchItems.value.filter((i) => i.is_duplicate).length);
const selectedBatchCount = computed(() => batchItems.value.filter((i) => i.selected).length);

const form = reactive<TradeCreateRequest>({
  symbol: "BTCUSDT",
  position_side: "LONG",
  leverage: 10,
  entry_price: 0,
  exit_price: 0,
  quantity: 0,
  margin_used: 0,
  pnl: 0,
  pnl_percentage: 0,
  fee: 0,
  trade_time: "",
  setup_tag: "",
  emotion_tag: "",
  notes: "",
  screenshot_url: "",
});

const clearPreview = () => {
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value);
    previewUrl.value = null;
  }
};

const selectAllNew = () => {
  batchItems.value.forEach((item) => {
    item.selected = !item.is_duplicate;
  });
};

const toggleAll = (val: boolean) => {
  batchItems.value.forEach((item) => {
    item.selected = val;
  });
};

const scanFile = async (file: File) => {
  clearPreview();
  previewUrl.value = URL.createObjectURL(file);
  scanning.value = true;
  singleDuplicateReason.value = "";
  batchItems.value = [];

  try {
    const res = await parseScreenshot(file);
    const trades = res.trades || [];
    parsedSummary.value = res.raw_summary || "ตรวจสอบตัวเลขด้านล่างก่อนกดบันทึก";
    parsedNoteConfidence.value = res.confidence || "";

    if (trades.length > 1) {
      batchItems.value = trades.map((t) => ({
        symbol: t.symbol || "BTCUSDT",
        position_side: t.position_side || "LONG",
        leverage: t.leverage || 10,
        entry_price: t.entry_price || 0,
        exit_price: t.exit_price || 0,
        quantity: t.quantity || 0,
        margin_used: t.margin_used || 0,
        pnl: t.pnl || 0,
        pnl_percentage: t.pnl_percentage || 0,
        fee: t.fee || 0,
        trade_time: t.trade_time || "",
        setup_tag: "",
        emotion_tag: "",
        notes: "",
        screenshot_url: t.screenshot_url || res.screenshot_url || "",
        is_duplicate: Boolean(t.is_duplicate),
        duplicate_reason: t.duplicate_reason || "",
        selected: !t.is_duplicate,
      }));
    } else {
      const first = trades[0];
      form.symbol = first?.symbol || "BTCUSDT";
      form.position_side = first?.position_side || "LONG";
      form.leverage = first?.leverage || 10;
      form.entry_price = first?.entry_price || 0;
      form.exit_price = first?.exit_price || 0;
      form.quantity = first?.quantity || 0;
      form.margin_used = first?.margin_used || 0;
      form.pnl = first?.pnl || 0;
      form.pnl_percentage = first?.pnl_percentage || 0;
      form.fee = first?.fee || 0;
      form.trade_time = first?.trade_time || "";
      form.screenshot_url = first?.screenshot_url || res.screenshot_url || "";
      singleDuplicateReason.value = first?.is_duplicate
        ? first.duplicate_reason || "พบรายการนี้ในประวัติเทรดแล้ว"
        : "";
    }

    step.value = "preview";
  } catch (err: any) {
    Swal.fire("เกิดข้อผิดพลาด", err.response?.data?.message || "สแกนรูปไม่สำเร็จ", "error");
  } finally {
    scanning.value = false;
  }
};

const onFileSelect = async (e: Event) => {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (file) {
    await scanFile(file);
  }
};

const onDropFile = async (e: DragEvent) => {
  isDragging.value = false;
  const file = e.dataTransfer?.files?.[0];
  if (file && file.type.startsWith("image/")) {
    await scanFile(file);
  }
};

const onModalPaste = async (e: ClipboardEvent) => {
  if (!props.open || step.value !== "upload") return;
  const items = e.clipboardData?.items;
  if (!items) return;
  for (const item of items) {
    if (item.type.startsWith("image/")) {
      const file = item.getAsFile();
      if (file) {
        e.preventDefault();
        await scanFile(file);
        break;
      }
    }
  }
};

watch(
  () => [props.open, props.initialFile] as const,
  ([isOpen, initFile]) => {
    if (isOpen && initFile) {
      step.value = "upload";
      scanFile(initFile);
    }
  },
  { immediate: true }
);

const resetToUpload = () => {
  step.value = "upload";
  batchItems.value = [];
  singleDuplicateReason.value = "";
  clearPreview();
  if (fileInputRef.value) fileInputRef.value.value = "";
};

const closeModal = () => {
  resetToUpload();
  emit("close");
};

const submitBatch = async () => {
  const chosen = batchItems.value
    .filter((i) => i.selected)
    .map((i) => ({
      symbol: i.symbol,
      position_side: i.position_side,
      leverage: i.leverage,
      entry_price: i.entry_price,
      exit_price: i.exit_price,
      quantity: i.quantity,
      margin_used: i.margin_used,
      pnl: i.pnl,
      pnl_percentage: i.pnl_percentage,
      fee: i.fee,
      trade_time: i.trade_time,
      setup_tag: i.setup_tag || batchSetupTag.value,
      emotion_tag: i.emotion_tag || batchEmotionTag.value,
      notes: i.notes,
      screenshot_url: i.screenshot_url,
    }));

  if (chosen.length === 0) return;

  saving.value = true;
  try {
    const res = await createTradesBatch(chosen, false);
    Swal.fire({
      icon: "success",
      title: res.message || `บันทึก ${chosen.length} รายการเรียบร้อย`,
      timer: 1800,
      showConfirmButton: false,
    });
    emit("saved");
    closeModal();
  } catch (err: any) {
    Swal.fire("บันทึกไม่สำเร็จ", err.response?.data?.message || "กรุณาลองใหม่อีกครั้ง", "error");
  } finally {
    saving.value = false;
  }
};

const submitForm = async () => {
  saving.value = true;
  try {
    await createTrade({ ...form });
    Swal.fire({
      icon: "success",
      title: "บันทึกและซิงค์เรียบร้อย",
      timer: 1400,
      showConfirmButton: false,
    });
    emit("saved");
    closeModal();
  } catch (err: any) {
    Swal.fire("บันทึกไม่สำเร็จ", err.response?.data?.message || "กรุณาลองใหม่อีกครั้ง", "error");
  } finally {
    saving.value = false;
  }
};

onMounted(() => {
  window.addEventListener("paste", onModalPaste);
});

onBeforeUnmount(() => {
  window.removeEventListener("paste", onModalPaste);
  clearPreview();
});
</script>
