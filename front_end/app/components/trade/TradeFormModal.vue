<template>
  <dialog :open="open" class="modal modal-bottom sm:modal-middle bg-black/60">
    <div class="modal-box max-w-2xl bg-base-100 border border-base-300">
      <div class="flex items-center justify-between mb-4">
        <h3 class="font-bold text-lg">
          {{ editTrade ? 'แก้ไขข้อมูลการเทรด' : 'เพิ่มบันทึกการเทรดใหม่' }}
        </h3>
        <button class="btn btn-sm btn-circle btn-ghost" @click="$emit('close')">✕</button>
      </div>

      <form @submit.prevent="submitForm" class="space-y-4">
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <label class="form-control">
            <span class="label-text text-xs mb-1">คู่เหรียญ</span>
            <input v-model="form.symbol" type="text" placeholder="เช่น BTCUSDT" class="input input-bordered input-sm" required />
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
            <span class="label-text text-xs mb-1">ราคาเข้า</span>
            <input v-model.number="form.entry_price" type="number" step="any" class="input input-bordered input-sm" />
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">ราคาปิด</span>
            <input v-model.number="form.exit_price" type="number" step="any" class="input input-bordered input-sm" />
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">มาร์จิ้น (USDT)</span>
            <input v-model.number="form.margin_used" type="number" step="any" class="input input-bordered input-sm" />
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">ค่าธรรมเนียม</span>
            <input v-model.number="form.fee" type="number" step="any" class="input input-bordered input-sm" />
          </label>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <label class="form-control">
            <span class="label-text text-xs mb-1">กำไร/ขาดทุนสุทธิ (USDT)</span>
            <input v-model.number="form.pnl" type="number" step="any" class="input input-bordered input-sm" required />
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">ROI (%)</span>
            <input v-model.number="form.pnl_percentage" type="number" step="any" class="input input-bordered input-sm" />
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">วัน-เวลาเทรด</span>
            <input v-model="form.trade_time" type="datetime-local" class="input input-bordered input-sm" />
          </label>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <label class="form-control">
            <span class="label-text text-xs mb-1">กลยุทธ์ (Setup Tag)</span>
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
            <span class="label-text text-xs mb-1">อารมณ์ตอนเข้าเทรด</span>
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
          <span class="label-text text-xs mb-1">บันทึกเพิ่มเติม</span>
          <textarea v-model="form.notes" rows="2" class="textarea textarea-bordered textarea-sm"></textarea>
        </label>

        <div class="modal-action">
          <button type="button" class="btn btn-ghost btn-sm" @click="$emit('close')">ยกเลิก</button>
          <button type="submit" class="btn btn-primary btn-sm" :disabled="saving">
            {{ saving ? 'กำลังบันทึก...' : 'บันทึกข้อมูล' }}
          </button>
        </div>
      </form>
    </div>
  </dialog>
</template>

<script setup lang="ts">
import Swal from "sweetalert2";
import { useTrades } from "~/composables/api/useTrades";
import type { TradeCreateRequest } from "~/models/trade/request/trade_req";
import type { TradeHistory } from "~/models/trade/response/trade_res";

const props = defineProps<{
  open: boolean;
  editTrade: TradeHistory | null;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "saved"): void;
}>();

const { createTrade, updateTrade } = useTrades();
const saving = ref(false);

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
});

watch(
  () => [props.open, props.editTrade],
  () => {
    if (props.editTrade) {
      form.symbol = props.editTrade.symbol;
      form.position_side = props.editTrade.position_side;
      form.leverage = props.editTrade.leverage;
      form.entry_price = props.editTrade.entry_price;
      form.exit_price = props.editTrade.exit_price;
      form.quantity = props.editTrade.quantity;
      form.margin_used = props.editTrade.margin_used;
      form.pnl = props.editTrade.pnl;
      form.pnl_percentage = props.editTrade.pnl_percentage;
      form.fee = props.editTrade.fee;
      form.trade_time = props.editTrade.trade_time?.slice(0, 16) || "";
      form.setup_tag = props.editTrade.setup_tag;
      form.emotion_tag = props.editTrade.emotion_tag;
      form.notes = props.editTrade.notes;
    } else {
      form.symbol = "BTCUSDT";
      form.position_side = "LONG";
      form.leverage = 10;
      form.entry_price = 0;
      form.exit_price = 0;
      form.quantity = 0;
      form.margin_used = 0;
      form.pnl = 0;
      form.pnl_percentage = 0;
      form.fee = 0;
      form.trade_time = new Date().toISOString().slice(0, 16);
      form.setup_tag = "";
      form.emotion_tag = "";
      form.notes = "";
    }
  },
  { immediate: true }
);

const submitForm = async () => {
  saving.value = true;
  try {
    if (props.editTrade) {
      await updateTrade(props.editTrade.trade_id, { ...form });
      Swal.fire({ icon: "success", title: "อัปเดตสำเร็จ!", timer: 1400, showConfirmButton: false });
    } else {
      await createTrade({ ...form });
      Swal.fire({ icon: "success", title: "บันทึกสำเร็จ!", timer: 1400, showConfirmButton: false });
    }
    emit("saved");
    emit("close");
  } catch (err: any) {
    Swal.fire("เกิดข้อผิดพลาด", err.response?.data?.message || "บันทึกไม่สำเร็จ", "error");
  } finally {
    saving.value = false;
  }
};
</script>
