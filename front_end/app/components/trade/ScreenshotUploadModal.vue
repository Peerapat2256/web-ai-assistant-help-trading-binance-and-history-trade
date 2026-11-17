<template>
  <dialog :open="open" class="modal modal-bottom sm:modal-middle bg-black/60">
    <div class="modal-box max-w-2xl bg-base-100 border border-base-300">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center gap-2">
          <Sparkles class="w-5 h-5 text-primary" />
          <h3 class="font-bold text-lg">สแกนรูปแคปจอ Binance ด้วย AI</h3>
        </div>
        <button class="btn btn-sm btn-circle btn-ghost" @click="closeModal">✕</button>
      </div>

      <div v-if="step === 'upload'" class="space-y-4">
        <label
          class="flex flex-col items-center justify-center border-2 border-dashed border-base-300 hover:border-primary rounded-2xl p-8 cursor-pointer bg-base-200/40 transition-colors"
        >
          <UploadCloud class="w-10 h-10 text-primary mb-2" />
          <span class="font-semibold text-sm">คลิกเพื่อเลือกรูปแคปจอ Binance (PnL Card / History)</span>
          <span class="text-xs opacity-60 mt-1">รองรับไฟล์ PNG, JPG, WEBP</span>
          <input type="file" accept="image/*" class="hidden" @change="onFileSelect" />
        </label>

        <div v-if="scanning" class="space-y-3 p-4 rounded-xl bg-base-200/60">
          <div class="text-xs font-medium text-primary">AI กำลังอ่านตัวเลขและข้อมูลจากรูปภาพ...</div>
          <div class="skeleton h-5 w-2/3"></div>
          <div class="skeleton h-10 w-full"></div>
          <div class="skeleton h-10 w-full"></div>
        </div>
      </div>

      <form v-else @submit.prevent="submitForm" class="space-y-4">
        <div
          :class="[
            'alert text-xs',
            parsedNoteConfidence === 'ai_verified' ? 'alert-success' : 'alert-warning',
          ]"
        >
          <span>{{ parsedSummary }}</span>
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
          <button type="button" class="btn btn-ghost btn-sm" @click="step = 'upload'">อัปโหลดรูปใหม่</button>
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

const props = defineProps<{ open: boolean }>();
const emit = defineEmits<{
  (e: "close"): void;
  (e: "saved"): void;
}>();

const { parseScreenshot, createTrade } = useTrades();
const step = ref<"upload" | "preview">("upload");
const scanning = ref(false);
const saving = ref(false);
const parsedSummary = ref("");
const parsedNoteConfidence = ref("");

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

const onFileSelect = async (e: Event) => {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;

  scanning.value = true;
  try {
    const res = await parseScreenshot(file);
    form.symbol = res.symbol || "BTCUSDT";
    form.position_side = res.position_side || "LONG";
    form.leverage = res.leverage || 10;
    form.entry_price = res.entry_price || 0;
    form.exit_price = res.exit_price || 0;
    form.quantity = res.quantity || 0;
    form.margin_used = res.margin_used || 0;
    form.pnl = res.pnl || 0;
    form.pnl_percentage = res.pnl_percentage || 0;
    form.fee = res.fee || 0;
    form.trade_time = res.trade_time || "";
    form.screenshot_url = res.screenshot_url || "";
    parsedSummary.value = res.raw_summary || "ตรวจสอบตัวเลขด้านล่างก่อนกดบันทึก";
    parsedNoteConfidence.value = res.confidence || "";
    step.value = "preview";
  } catch (err: any) {
    Swal.fire("เกิดข้อผิดพลาด", err.response?.data?.message || "สแกนรูปไม่สำเร็จ", "error");
  } finally {
    scanning.value = false;
  }
};

const submitForm = async () => {
  saving.value = true;
  try {
    await createTrade({ ...form });
    Swal.fire({
      icon: "success",
      title: "บันทึกสำเร็จ!",
      text: "บันทึกประวัติเทรดและส่งข้อมูลไปที่ Google Sheets แล้ว",
      timer: 1800,
      showConfirmButton: false,
    });
    emit("saved");
    closeModal();
  } catch (err: any) {
    Swal.fire("เกิดข้อผิดพลาด", err.response?.data?.message || "บันทึกไม่สำเร็จ", "error");
  } finally {
    saving.value = false;
  }
};

const closeModal = () => {
  step.value = "upload";
  emit("close");
};
</script>
