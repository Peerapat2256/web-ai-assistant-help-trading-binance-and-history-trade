<template>
  <div class="space-y-6">
    <div class="agy-fade-up stagger-1">
      <h1 class="text-2xl font-bold">เครื่องมือคำนวณความเสี่ยงก่อนเข้าเทรด</h1>
      <p class="text-sm opacity-60">คำนวณขนาดไม้, จุดล้างพอร์ต, ราคาถัวเฉลี่ย (DCA) และเช็คลิสต์เตือนสติก่อนเปิดออเดอร์</p>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="card agy-card agy-fade-up stagger-2 bg-base-100/90 border border-base-300 shadow-sm">
        <div class="card-body p-5 space-y-4">
          <h3 class="font-bold text-base">1. คำนวณขนาดไม้และความเสี่ยง (Position Size)</h3>

          <div class="grid grid-cols-2 gap-3">
            <label class="form-control">
              <span class="label-text text-xs mb-1">ทุนทั้งหมด (USDT)</span>
              <input v-model.number="pos.balance" type="number" class="input input-bordered input-sm" />
            </label>
            <label class="form-control">
              <span class="label-text text-xs mb-1">ยอมขาดทุนได้กี่ % ต่อไม้</span>
              <input v-model.number="pos.riskPct" type="number" step="0.1" class="input input-bordered input-sm" />
            </label>
            <label class="form-control">
              <span class="label-text text-xs mb-1">ราคาเข้า (Entry Price)</span>
              <input v-model.number="pos.entry" type="number" step="any" class="input input-bordered input-sm" />
            </label>
            <label class="form-control">
              <span class="label-text text-xs mb-1">ราคา Stop Loss</span>
              <input v-model.number="pos.stopLoss" type="number" step="any" class="input input-bordered input-sm" />
            </label>
            <label class="form-control col-span-2">
              <span class="label-text text-xs mb-1">Leverage ที่ใช้ (x)</span>
              <input v-model.number="pos.leverage" type="number" min="1" class="input input-bordered input-sm" />
            </label>
          </div>

          <div class="p-4 rounded-xl bg-base-200/70 space-y-2 text-sm">
            <div class="flex justify-between">
              <span class="opacity-70">เงินที่ยอมเสียได้สูงสุด:</span>
              <span class="font-bold text-error">{{ posResult.riskAmount.toFixed(2) }} USDT</span>
            </div>
            <div class="flex justify-between">
              <span class="opacity-70">ระยะห่าง Stop Loss:</span>
              <span class="font-bold">{{ posResult.slDistancePct.toFixed(2) }}%</span>
            </div>
            <div class="flex justify-between">
              <span class="opacity-70">ขนาดมูลค่าออเดอร์ (Position Value):</span>
              <span class="font-bold text-primary">{{ posResult.positionSizeUSD.toFixed(2) }} USDT</span>
            </div>
            <div class="flex justify-between">
              <span class="opacity-70">เงินต้นมาร์จิ้นที่ต้องใช้จริง:</span>
              <span class="font-bold text-success">{{ posResult.requiredMargin.toFixed(2) }} USDT</span>
            </div>
          </div>
        </div>
      </div>

      <div class="card agy-card agy-fade-up stagger-3 bg-base-100/90 border border-base-300 shadow-sm">
        <div class="card-body p-5 space-y-4">
          <h3 class="font-bold text-base">2. คำนวณจุดล้างพอร์ต & เป้า ROE%</h3>

          <div class="grid grid-cols-2 gap-3">
            <label class="form-control">
              <span class="label-text text-xs mb-1">ฝั่งที่เทรด</span>
              <select v-model="liq.side" class="select select-bordered select-sm">
                <option value="LONG">LONG</option>
                <option value="SHORT">SHORT</option>
              </select>
            </label>
            <label class="form-control">
              <span class="label-text text-xs mb-1">Leverage (x)</span>
              <input v-model.number="liq.leverage" type="number" min="1" max="125" class="input input-bordered input-sm" />
            </label>
            <label class="form-control">
              <span class="label-text text-xs mb-1">ราคาเข้า (Entry)</span>
              <input v-model.number="liq.entry" type="number" step="any" class="input input-bordered input-sm" />
            </label>
            <label class="form-control">
              <span class="label-text text-xs mb-1">ราคาเป้าหมาย (Take Profit)</span>
              <input v-model.number="liq.tp" type="number" step="any" class="input input-bordered input-sm" />
            </label>
          </div>

          <div class="p-4 rounded-xl bg-base-200/70 space-y-2 text-sm">
            <div class="flex justify-between">
              <span class="opacity-70">จุดล้างพอร์ตโดยประมาณ (Isolated):</span>
              <span class="font-bold text-error">{{ liqResult.liqPrice.toFixed(4) }}</span>
            </div>
            <div class="flex justify-between">
              <span class="opacity-70">กำไรที่เป้าหมาย (ROE %):</span>
              <span :class="['font-bold', liqResult.roePct >= 0 ? 'text-success' : 'text-error']">
                {{ liqResult.roePct >= 0 ? '+' : '' }}{{ liqResult.roePct.toFixed(2) }}%
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="card agy-card agy-fade-up stagger-4 bg-base-100/90 border border-base-300 shadow-sm">
        <div class="card-body p-5 space-y-4">
          <h3 class="font-bold text-base">3. คำนวณราคาถัวเฉลี่ย (DCA Calculator)</h3>

          <div class="grid grid-cols-2 gap-3">
            <label class="form-control">
              <span class="label-text text-xs mb-1">ราคาไม้ที่ 1</span>
              <input v-model.number="dca.p1" type="number" step="any" class="input input-bordered input-sm" />
            </label>
            <label class="form-control">
              <span class="label-text text-xs mb-1">จำนวนเงินไม้ที่ 1 (USDT)</span>
              <input v-model.number="dca.u1" type="number" step="any" class="input input-bordered input-sm" />
            </label>
            <label class="form-control">
              <span class="label-text text-xs mb-1">ราคาไม้ที่ 2</span>
              <input v-model.number="dca.p2" type="number" step="any" class="input input-bordered input-sm" />
            </label>
            <label class="form-control">
              <span class="label-text text-xs mb-1">จำนวนเงินไม้ที่ 2 (USDT)</span>
              <input v-model.number="dca.u2" type="number" step="any" class="input input-bordered input-sm" />
            </label>
          </div>

          <div class="p-4 rounded-xl bg-base-200/70 space-y-2 text-sm">
            <div class="flex justify-between">
              <span class="opacity-70">ราคาต้นทุนเฉลี่ยใหม่:</span>
              <span class="font-bold text-primary">{{ dcaResult.avgPrice.toFixed(4) }}</span>
            </div>
            <div class="flex justify-between">
              <span class="opacity-70">จำนวนเงินรวมทั้งหมด:</span>
              <span class="font-bold">{{ dcaResult.totalUSD.toFixed(2) }} USDT</span>
            </div>
          </div>
        </div>
      </div>

      <div class="card agy-card agy-fade-up stagger-5 bg-base-100/90 border border-base-300 shadow-sm">
        <div class="card-body p-5 space-y-3">
          <div class="flex items-center justify-between">
            <h3 class="font-bold text-base">4. เช็คลิสต์เตือนสติก่อนเข้าออเดอร์</h3>
            <span class="badge" :class="allChecked ? 'badge-success' : 'badge-warning'">
              {{ checkedCount }} / {{ checklist.length }} ข้อ
            </span>
          </div>

          <div class="space-y-2">
            <label
              v-for="(item, idx) in checklist"
              :key="idx"
              class="flex items-center gap-3 p-2.5 rounded-xl bg-base-200/50 cursor-pointer hover:bg-base-200"
            >
              <input v-model="item.done" type="checkbox" class="checkbox checkbox-primary checkbox-sm" />
              <span :class="['text-sm', item.done ? 'line-through opacity-50' : '']">{{ item.text }}</span>
            </label>
          </div>

          <div v-if="allChecked" class="alert alert-success text-xs py-2">
            <span>แผนครบถ้วน! สามารถเข้าออเดอร์ตามวินัยที่วางไว้ได้เลยครับ</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: "auth" });

const pos = reactive({
  balance: 1000,
  riskPct: 2,
  entry: 90000,
  stopLoss: 88500,
  leverage: 10,
});

const posResult = computed(() => {
  const riskAmount = (pos.balance * pos.riskPct) / 100;
  const diff = Math.abs(pos.entry - pos.stopLoss);
  const slDistancePct = pos.entry > 0 ? (diff / pos.entry) * 100 : 0;
  const positionSizeUSD = slDistancePct > 0 ? riskAmount / (slDistancePct / 100) : 0;
  const requiredMargin = pos.leverage > 0 ? positionSizeUSD / pos.leverage : positionSizeUSD;
  return { riskAmount, slDistancePct, positionSizeUSD, requiredMargin };
});

const liq = reactive({
  side: "LONG",
  leverage: 10,
  entry: 90000,
  tp: 94500,
});

const liqResult = computed(() => {
  const factor = liq.leverage > 0 ? 0.99 / liq.leverage : 0.99;
  const liqPrice = liq.side === "LONG" ? liq.entry * (1 - factor) : liq.entry * (1 + factor);
  let roePct = 0;
  if (liq.entry > 0) {
    const rawPct = liq.side === "LONG" ? (liq.tp - liq.entry) / liq.entry : (liq.entry - liq.tp) / liq.entry;
    roePct = rawPct * liq.leverage * 100;
  }
  return { liqPrice: Math.max(0, liqPrice), roePct };
});

const dca = reactive({
  p1: 90000,
  u1: 500,
  p2: 85000,
  u2: 500,
});

const dcaResult = computed(() => {
  const q1 = dca.p1 > 0 ? dca.u1 / dca.p1 : 0;
  const q2 = dca.p2 > 0 ? dca.u2 / dca.p2 : 0;
  const totalUSD = dca.u1 + dca.u2;
  const totalQty = q1 + q2;
  const avgPrice = totalQty > 0 ? totalUSD / totalQty : 0;
  return { avgPrice, totalUSD };
});

const checklist = reactive([
  { text: "ตรวจสอบแนวโน้มหลักใน Timeframe ใหญ่ (4H / 1D) แล้ว", done: false },
  { text: "กำหนดจุด Stop Loss ชัดเจนและยอมรับการขาดทุนได้", done: false },
  { text: "สัดส่วน Risk : Reward คุ้มค่า (อย่างน้อย 1 : 2)", done: false },
  { text: "เช็คข่าวแรง (เช่น CPI, Fed) ที่กำลังจะประกาศแล้ว", done: false },
  { text: "ไม่ได้เข้าเทรดเพราะอารมณ์ FOMO หรือต้องการเอาคืน", done: false },
]);

const checkedCount = computed(() => checklist.filter((i) => i.done).length);
const allChecked = computed(() => checkedCount.value === checklist.length);
</script>
