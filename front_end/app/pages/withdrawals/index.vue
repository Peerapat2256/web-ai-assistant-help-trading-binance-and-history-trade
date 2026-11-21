<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="agy-fade-up stagger-1 flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold">ประวัติการถอนเงิน & คำนวณภาษี</h1>
        <p class="text-sm opacity-60">บันทึกสลิปการถอนเงินด้วย AI เช็คหมายเลขธุรกรรมซ้ำอัตโนมัติ และประเมินภาษีเงินได้รายปี</p>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <select v-model.number="selectedYear" class="select select-bordered select-sm" @change="loadData">
          <option :value="0">ทุกปีภาษี</option>
          <option v-for="y in yearOptions" :key="y" :value="y">ปี {{ y }} (พ.ศ. {{ y + 543 }})</option>
        </select>
        <button class="btn btn-outline btn-sm gap-1.5" @click="exportCSV" :disabled="withdrawals.length === 0">
          <Download class="w-4 h-4" />
          ส่งออก CSV ยื่นภาษี
        </button>
        <button class="btn btn-primary btn-sm gap-1.5" @click="openScanModal">
          <Sparkles class="w-4 h-4" />
          สแกนสลิปถอนเงินด้วย AI
        </button>
        <button class="btn btn-outline btn-sm gap-1.5" @click="openManualModal">
          <Plus class="w-4 h-4" />
          เพิ่มรายการเอง
        </button>
      </div>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 agy-fade-up stagger-2">
      <div class="card agy-card bg-base-100/90 border border-base-300 shadow-sm">
        <div class="card-body p-4">
          <div class="text-xs opacity-60">ยอดถอนรวม ({{ selectedYear === 0 ? 'ทั้งหมด' : `ปี ${selectedYear}` }})</div>
          <div class="text-2xl font-bold text-primary font-mono mt-1">
            {{ formatMoney(totalWithdrawnTHB) }} <span class="text-xs font-normal opacity-70">THB</span>
          </div>
          <div class="text-[11px] opacity-60 mt-1">นับเฉพาะรายการสถานะ "สำเร็จ"</div>
        </div>
      </div>

      <div class="card agy-card bg-base-100/90 border border-base-300 shadow-sm">
        <div class="card-body p-4">
          <div class="text-xs opacity-60">จำนวนครั้งที่ถอนสำเร็จ</div>
          <div class="text-2xl font-bold font-mono mt-1">
            {{ completedCount }} <span class="text-xs font-normal opacity-70">ครั้ง</span>
          </div>
          <div class="text-[11px] opacity-60 mt-1">เฉลี่ย {{ formatMoney(avgPerWithdrawal) }} บาท / ครั้ง</div>
        </div>
      </div>

      <div class="card agy-card bg-base-100/90 border border-base-300 shadow-sm">
        <div class="card-body p-4">
          <div class="text-xs opacity-60">เงินได้สุทธิประเมิน (หลังหักลดหย่อน)</div>
          <div class="text-2xl font-bold text-info font-mono mt-1">
            {{ formatMoney(taxableIncome) }} <span class="text-xs font-normal opacity-70">THB</span>
          </div>
          <div class="text-[11px] opacity-60 mt-1">หักต้นทุน + ลดหย่อน {{ formatMoney(totalDeductions) }} บาท</div>
        </div>
      </div>

      <div class="card agy-card bg-base-100/90 border border-base-300 shadow-sm">
        <div class="card-body p-4">
          <div class="text-xs opacity-60">ภาษีเงินได้ประเมินเบื้องต้น</div>
          <div class="text-2xl font-bold text-warning font-mono mt-1">
            {{ formatMoney(taxResult.totalTax) }} <span class="text-xs font-normal opacity-70">THB</span>
          </div>
          <div class="text-[11px] opacity-60 mt-1">
            อัตราภาษีเฉลี่ย {{ effectiveTaxRate.toFixed(2) }}% (ฐานสูงสุด {{ taxResult.maxRate }}%)
          </div>
        </div>
      </div>
    </div>

    <!-- Thai Tax Estimator Panel -->
    <div class="card agy-card agy-fade-up stagger-3 bg-base-100/90 border border-base-300 shadow-sm">
      <div class="card-body p-5 space-y-4">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <Calculator class="w-5 h-5 text-primary" />
            <h2 class="font-bold text-base">เครื่องคำนวณภาษีเงินได้บุคคลธรรมดา (แบบขั้นบันไดของไทย)</h2>
          </div>
          <span class="badge badge-primary badge-outline text-xs">
            ทิป: แคปหน้าจอรายละเอียดการถอนแล้วกด Ctrl + V บนหน้านี้ได้เลย
          </span>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <label class="form-control">
            <span class="label-text text-xs mb-1">เงินต้น/ยอดฝากที่หักได้ (THB)</span>
            <input v-model.number="costBasis" type="number" min="0" step="any" class="input input-bordered input-sm font-mono" />
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">ค่าลดหย่อนส่วนตัวและอื่นๆ (THB)</span>
            <input v-model.number="personalDeduction" type="number" min="0" step="any" class="input input-bordered input-sm font-mono" />
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">รายได้อื่นในปีภาษีเดียวกัน (THB)</span>
            <input v-model.number="otherIncome" type="number" min="0" step="any" class="input input-bordered input-sm font-mono" />
          </label>
        </div>

        <div class="overflow-x-auto">
          <table class="table table-xs w-full">
            <thead>
              <tr class="bg-base-200/70">
                <th>ช่วงเงินได้สุทธิ (บาท)</th>
                <th class="text-center">อัตราภาษี</th>
                <th class="text-right">เงินได้ในช่วงนี้</th>
                <th class="text-right">ภาษีที่ต้องชำระในช่วงนี้</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="b in taxResult.brackets"
                :key="b.label"
                :class="b.taxableInBracket > 0 ? 'bg-primary/5 font-medium' : 'opacity-50'"
              >
                <td>{{ b.label }}</td>
                <td class="text-center">
                  <span class="badge badge-xs badge-outline">{{ b.rate }}%</span>
                </td>
                <td class="text-right font-mono">{{ formatMoney(b.taxableInBracket) }}</td>
                <td class="text-right font-mono text-warning">{{ formatMoney(b.taxInBracket) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p class="text-[11px] opacity-60">
          * หมายเหตุ: ตัวเลขนี้เป็นการประเมินเบื้องต้นจากยอดถอนเงินสด (THB) ตามอัตราก้าวหน้ากรมสรรพากร เพื่อใช้วางแผนภาษีเท่านั้น ควรเก็บหลักฐานการโอนเข้า-ออกและตรวจสอบเงื่อนไขกฎหมายภาษีคริปโตประจำปีภาษีอีกครั้ง
        </p>
      </div>
    </div>

    <!-- Withdrawal History Table -->
    <div class="card agy-card agy-fade-up stagger-4 bg-base-100/90 border border-base-300 shadow-sm overflow-hidden">
      <div class="p-4 border-b border-base-300/70 flex items-center justify-between">
        <h2 class="font-bold text-base">ตารางประวัติการถอนเงิน ({{ withdrawals.length }} รายการ)</h2>
      </div>

      <div v-if="loading" class="p-6 space-y-3">
        <div v-for="n in 5" :key="n" class="skeleton h-12 w-full rounded-lg"></div>
      </div>

      <div v-else-if="withdrawals.length === 0" class="p-12 text-center space-y-3">
        <p class="text-sm opacity-60">ยังไม่มีประวัติการถอนเงินในปีที่เลือก</p>
        <button class="btn btn-primary btn-sm gap-2" @click="openScanModal">
          <Sparkles class="w-4 h-4" />
          สแกนสลิปถอนเงินใบแรก
        </button>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="table table-sm w-full">
          <thead>
            <tr class="bg-base-200/70">
              <th>วันที่</th>
              <th class="text-right">จำนวนเงิน</th>
              <th>สถานะ</th>
              <th>วิธีการชำระเงิน</th>
              <th>หมายเลขธุรกรรม</th>
              <th>หมายเหตุการทำธุรกรรม</th>
              <th class="text-right">จัดการ</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="w in withdrawals" :key="w.withdrawal_id" class="hover">
              <td class="font-mono text-xs whitespace-nowrap">{{ formatDateTime(w.withdrawal_time) }}</td>
              <td class="text-right font-mono whitespace-nowrap">
                <div class="font-bold text-sm text-primary">-{{ formatMoney(w.amount) }} {{ w.currency }}</div>
                <div v-if="w.amount_usdt && w.currency !== 'USDT'" class="text-[11px] opacity-60">
                  หักพอร์ต -{{ formatMoney(w.amount_usdt) }} USDT
                </div>
              </td>
              <td>
                <span
                  :class="[
                    'badge badge-xs gap-1',
                    w.status === 'สำเร็จ' || w.status === 'Completed'
                      ? 'badge-success badge-outline'
                      : w.status === 'รอดำเนินการ'
                      ? 'badge-warning badge-outline'
                      : 'badge-error badge-outline',
                  ]"
                >
                  {{ w.status }}
                </span>
              </td>
              <td>
                <span class="badge badge-ghost badge-xs font-mono">{{ w.payment_method }}</span>
              </td>
              <td>
                <div v-if="w.transaction_id" class="flex items-center gap-1.5 font-mono text-xs">
                  <span>{{ w.transaction_id }}</span>
                  <button class="btn btn-ghost btn-xs px-1" @click="copyTx(w.transaction_id)" title="คัดลอก">
                    <Copy class="w-3 h-3 opacity-70" />
                  </button>
                </div>
                <span v-else class="opacity-40 text-xs">--</span>
              </td>
              <td class="text-xs opacity-75 max-w-[180px] truncate">{{ w.notes || "--" }}</td>
              <td class="text-right">
                <div class="flex justify-end gap-1">
                  <button class="btn btn-ghost btn-xs" @click="openEditModal(w)" title="แก้ไข">
                    <Pencil class="w-3.5 h-3.5" />
                  </button>
                  <button class="btn btn-ghost btn-xs text-error" @click="confirmDelete(w)" title="ลบ">
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal สแกน / เพิ่ม / แก้ไขรายการถอน -->
    <dialog :open="modalOpen" class="modal modal-bottom sm:modal-middle bg-black/60">
      <div class="modal-box max-w-xl bg-base-100 border border-base-300">
        <div class="flex items-center justify-between mb-4">
          <div class="flex items-center gap-2">
            <Sparkles v-if="modalMode === 'scan'" class="w-5 h-5 text-primary" />
            <h3 class="font-bold text-lg">
              {{ editingId ? 'แก้ไขประวัติการถอนเงิน' : modalMode === 'scan' ? 'สแกนสลิปถอนเงินด้วย AI' : 'เพิ่มประวัติการถอนเงิน' }}
            </h3>
          </div>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeModal">✕</button>
        </div>

        <!-- โซนอัปโหลดสลิป -->
        <div v-if="modalMode === 'scan' && !scannedReady" class="space-y-4">
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
            <span class="font-semibold text-sm">คลิกเลือกรูปสลิปการถอน, ลากวาง, หรือกด Ctrl + V วางได้เลย</span>
            <span class="text-xs opacity-60 mt-1">รองรับหน้า "รายละเอียดการถอน" จาก Binance TH / Binance / Bitkub</span>
            <input ref="fileInputRef" type="file" accept="image/*" class="hidden" @change="onFileSelect" />
          </label>

          <div v-if="previewUrl && scanning" class="rounded-xl border border-base-300 bg-base-200/50 p-2">
            <img :src="previewUrl" alt="สลิปที่กำลังสแกน" class="w-full max-h-40 object-contain rounded-lg" />
          </div>

          <div v-if="scanning" class="space-y-2 p-4 rounded-xl bg-base-200/60">
            <div class="text-xs font-medium text-primary">AI กำลังอ่านยอดถอน หมายเลขธุรกรรม และตรวจสอบรายการซ้ำ...</div>
            <div class="skeleton h-5 w-2/3"></div>
            <div class="skeleton h-10 w-full"></div>
          </div>
        </div>

        <!-- ฟอร์มตรวจสอบ / กรอกข้อมูล -->
        <form v-else @submit.prevent="submitForm" class="space-y-4">
          <div v-if="previewUrl" class="rounded-xl border border-base-300 bg-base-200/50 p-2">
            <img :src="previewUrl" alt="ภาพสลิป" class="w-full max-h-36 object-contain rounded-lg" />
          </div>

          <div v-if="scanSummary" class="alert alert-success text-xs py-2">
            <span>{{ scanSummary }}</span>
          </div>

          <div v-if="duplicateWarning" class="alert alert-warning text-xs py-2">
            <span>คำเตือนรายการซ้ำ: {{ duplicateWarning }}</span>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <label class="form-control sm:col-span-2">
              <span class="label-text text-xs mb-1">จำนวนเงินที่ถอน</span>
              <input v-model.number="form.amount" type="number" step="any" min="0.01" class="input input-bordered input-sm font-mono" required />
            </label>
            <label class="form-control">
              <span class="label-text text-xs mb-1">สกุลเงิน</span>
              <select v-model="form.currency" class="select select-bordered select-sm">
                <option value="THB">THB</option>
                <option value="USDT">USDT</option>
                <option value="USD">USD</option>
              </select>
            </label>
          </div>

          <label v-if="form.currency === 'THB'" class="form-control">
            <span class="label-text text-xs mb-1">ยอดหักพอร์ต (USDT) — เว้นว่างหรือใส่ 0 เพื่อให้ระบบแปลงจากเรทสด USDT/THB อัตโนมัติ</span>
            <input v-model.number="form.amount_usdt" type="number" step="any" min="0" placeholder="แปลงอัตโนมัติตามเรท Binance USDT/THB" class="input input-bordered input-sm font-mono" />
          </label>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <label class="form-control">
              <span class="label-text text-xs mb-1">วันที่และเวลา</span>
              <input v-model="form.withdrawal_time" type="text" placeholder="YYYY-MM-DD HH:MM:SS" class="input input-bordered input-sm font-mono" required />
            </label>
            <label class="form-control">
              <span class="label-text text-xs mb-1">สถานะ</span>
              <select v-model="form.status" class="select select-bordered select-sm">
                <option value="สำเร็จ">สำเร็จ</option>
                <option value="รอดำเนินการ">รอดำเนินการ</option>
                <option value="ล้มเหลว">ล้มเหลว</option>
              </select>
            </label>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <label class="form-control">
              <span class="label-text text-xs mb-1">วิธีการชำระเงิน</span>
              <input v-model="form.payment_method" type="text" placeholder="BANK_TRANSFER" class="input input-bordered input-sm font-mono" />
            </label>
            <label class="form-control">
              <span class="label-text text-xs mb-1">หมายเลขธุรกรรม</span>
              <input v-model="form.transaction_id" type="text" placeholder="W1330793249660076034" class="input input-bordered input-sm font-mono" />
            </label>
          </div>

          <label class="form-control">
            <span class="label-text text-xs mb-1">หมายเหตุการทำธุรกรรม</span>
            <input v-model="form.notes" type="text" placeholder="เช่น ถอนกำไรเข้าธนาคารกสิกร..." class="input input-bordered input-sm" />
          </label>

          <div class="modal-action">
            <button v-if="modalMode === 'scan' && !editingId" type="button" class="btn btn-ghost btn-sm" @click="scannedReady = false">
              สแกนรูปใหม่
            </button>
            <button type="button" class="btn btn-ghost btn-sm" @click="closeModal">ยกเลิก</button>
            <button type="submit" class="btn btn-primary btn-sm" :disabled="saving">
              {{ saving ? 'กำลังบันทึก...' : editingId ? 'บันทึกการแก้ไข' : 'บันทึกประวัติการถอน' }}
            </button>
          </div>
        </form>
      </div>
    </dialog>
  </div>
</template>

<script setup lang="ts">
import { Sparkles, Plus, Download, Calculator, UploadCloud, Copy, Pencil, Trash2 } from "lucide-vue-next";
import Swal from "sweetalert2";
import { useWithdrawals } from "~/composables/api/useWithdrawals";
import type { WithdrawalCreateRequest, WithdrawalHistory } from "~/models/withdrawal/withdrawal_model";

definePageMeta({ middleware: "auth" });

const {
  fetchWithdrawals,
  parseWithdrawalScreenshot,
  createWithdrawal,
  updateWithdrawal,
  deleteWithdrawal,
} = useWithdrawals();

const currentYear = new Date().getFullYear();
const yearOptions = [currentYear + 1, currentYear, currentYear - 1, currentYear - 2];
const selectedYear = ref<number>(currentYear);

const loading = ref(false);
const withdrawals = ref<WithdrawalHistory[]>([]);

// ตัวแปรคำนวณภาษี
const costBasis = ref<number>(0);
const personalDeduction = ref<number>(60000);
const otherIncome = ref<number>(0);

// ตัวแปร Modal
const modalOpen = ref(false);
const modalMode = ref<"scan" | "manual">("scan");
const scannedReady = ref(false);
const scanning = ref(false);
const saving = ref(false);
const isDragging = ref(false);
const editingId = ref<number | null>(null);
const previewUrl = ref<string | null>(null);
const scanSummary = ref("");
const duplicateWarning = ref("");
const fileInputRef = ref<HTMLInputElement | null>(null);

const form = reactive<WithdrawalCreateRequest>({
  transaction_id: "",
  amount: 0,
  currency: "THB",
  amount_usdt: 0,
  status: "สำเร็จ",
  payment_method: "BANK_TRANSFER",
  withdrawal_time: "",
  notes: "",
  screenshot_url: "",
});

const completedTHBItems = computed(() =>
  withdrawals.value.filter(
    (w) => (w.status === "สำเร็จ" || w.status === "Completed") && w.currency === "THB"
  )
);

const totalWithdrawnTHB = computed(() =>
  completedTHBItems.value.reduce((acc, w) => acc + Number(w.amount || 0), 0)
);

const completedCount = computed(() => completedTHBItems.value.length);

const avgPerWithdrawal = computed(() =>
  completedCount.value > 0 ? totalWithdrawnTHB.value / completedCount.value : 0
);

const totalDeductions = computed(() => Math.max(0, Number(costBasis.value || 0)) + Math.max(0, Number(personalDeduction.value || 0)));

const taxableIncome = computed(() => {
  const gross = totalWithdrawnTHB.value + Math.max(0, Number(otherIncome.value || 0));
  return Math.max(0, gross - totalDeductions.value);
});

const taxResult = computed(() => {
  const steps = [
    { label: "0 – 150,000", min: 0, max: 150000, rate: 0 },
    { label: "150,001 – 300,000", min: 150000, max: 300000, rate: 5 },
    { label: "300,001 – 500,000", min: 300000, max: 500000, rate: 10 },
    { label: "500,001 – 750,000", min: 500000, max: 750000, rate: 15 },
    { label: "750,001 – 1,000,000", min: 750000, max: 1000000, rate: 20 },
    { label: "1,000,001 – 2,000,000", min: 1000000, max: 2000000, rate: 25 },
    { label: "2,000,001 – 5,000,000", min: 2000000, max: 5000000, rate: 30 },
    { label: "มากกว่า 5,000,000", min: 5000000, max: Infinity, rate: 35 },
  ];

  let totalTax = 0;
  let maxRate = 0;
  const income = taxableIncome.value;

  const brackets = steps.map((s) => {
    const inBracket = Math.max(0, Math.min(income, s.max) - s.min);
    const taxInBracket = (inBracket * s.rate) / 100;
    if (inBracket > 0 && s.rate > maxRate) maxRate = s.rate;
    totalTax += taxInBracket;
    return {
      label: s.label,
      rate: s.rate,
      taxableInBracket: inBracket,
      taxInBracket,
    };
  });

  return { totalTax, maxRate, brackets };
});

const effectiveTaxRate = computed(() => {
  const gross = totalWithdrawnTHB.value + Math.max(0, Number(otherIncome.value || 0));
  if (gross <= 0) return 0;
  return (taxResult.value.totalTax / gross) * 100;
});

const formatMoney = (val: number) =>
  Number(val || 0).toLocaleString("th-TH", { minimumFractionDigits: 2, maximumFractionDigits: 2 });

const formatDateTime = (iso: string) => {
  if (!iso) return "-";
  const d = new Date(iso);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
};

const nowStr = () => formatDateTime(new Date().toISOString());

const loadData = async () => {
  loading.value = true;
  try {
    withdrawals.value = await fetchWithdrawals(selectedYear.value);
  } finally {
    loading.value = false;
  }
};

const clearPreview = () => {
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value);
    previewUrl.value = null;
  }
};

const resetForm = () => {
  form.transaction_id = "";
  form.amount = 0;
  form.currency = "THB";
  form.amount_usdt = 0;
  form.status = "สำเร็จ";
  form.payment_method = "BANK_TRANSFER";
  form.withdrawal_time = nowStr();
  form.notes = "";
  form.screenshot_url = "";
  scanSummary.value = "";
  duplicateWarning.value = "";
  editingId.value = null;
  clearPreview();
};

const openScanModal = () => {
  resetForm();
  modalMode.value = "scan";
  scannedReady.value = false;
  modalOpen.value = true;
};

const openManualModal = () => {
  resetForm();
  modalMode.value = "manual";
  scannedReady.value = true;
  modalOpen.value = true;
};

const openEditModal = (w: WithdrawalHistory) => {
  resetForm();
  editingId.value = w.withdrawal_id;
  modalMode.value = "manual";
  scannedReady.value = true;
  form.transaction_id = w.transaction_id || "";
  form.amount = w.amount;
  form.currency = w.currency || "THB";
  form.amount_usdt = w.amount_usdt || 0;
  form.status = w.status || "สำเร็จ";
  form.payment_method = w.payment_method || "BANK_TRANSFER";
  form.withdrawal_time = formatDateTime(w.withdrawal_time);
  form.notes = w.notes || "";
  modalOpen.value = true;
};

const closeModal = () => {
  modalOpen.value = false;
  resetForm();
};

const scanFile = async (file: File) => {
  clearPreview();
  previewUrl.value = URL.createObjectURL(file);
  scanning.value = true;
  duplicateWarning.value = "";
  scanSummary.value = "";

  try {
    const res = await parseWithdrawalScreenshot(file);
    form.transaction_id = res.transaction_id || "";
    form.amount = res.amount || 0;
    form.currency = res.currency || "THB";
    form.status = res.status || "สำเร็จ";
    form.payment_method = res.payment_method || "BANK_TRANSFER";
    form.withdrawal_time = res.withdrawal_time || nowStr();
    form.notes = res.notes || "";
    form.screenshot_url = res.screenshot_url || "";
    scanSummary.value = res.raw_summary || "สแกนสลิปสำเร็จ กรุณาตรวจสอบข้อมูลก่อนบันทึก";
    duplicateWarning.value = res.is_duplicate ? res.duplicate_reason || "พบรายการนี้ในระบบแล้ว" : "";
    scannedReady.value = true;
  } catch (err: any) {
    Swal.fire("สแกนไม่สำเร็จ", err.response?.data?.message || "กรุณาลองใหม่อีกครั้ง", "error");
  } finally {
    scanning.value = false;
  }
};

const onFileSelect = async (e: Event) => {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (file) await scanFile(file);
};

const onDropFile = async (e: DragEvent) => {
  isDragging.value = false;
  const file = e.dataTransfer?.files?.[0];
  if (file && file.type.startsWith("image/")) await scanFile(file);
};

const submitForm = async () => {
  saving.value = true;
  try {
    if (editingId.value) {
      await updateWithdrawal(editingId.value, { ...form });
      Swal.fire({ icon: "success", title: "อัปเดตเรียบร้อยแล้ว", timer: 1300, showConfirmButton: false });
    } else {
      await createWithdrawal({ ...form });
      Swal.fire({ icon: "success", title: "บันทึกประวัติการถอนแล้ว", timer: 1300, showConfirmButton: false });
    }
    closeModal();
    await loadData();
  } catch (err: any) {
    Swal.fire("บันทึกไม่สำเร็จ", err.response?.data?.message || "กรุณาลองใหม่อีกครั้ง", "error");
  } finally {
    saving.value = false;
  }
};

const confirmDelete = async (w: WithdrawalHistory) => {
  const res = await Swal.fire({
    title: "ยืนยันการลบ?",
    text: `ต้องการลบรายการถอน ${formatMoney(w.amount)} ${w.currency} (${w.transaction_id || "ไม่ระบุรหัส"}) ใช่หรือไม่?`,
    icon: "warning",
    showCancelButton: true,
    confirmButtonText: "ใช่, ลบเลย",
    cancelButtonText: "ยกเลิก",
    confirmButtonColor: "#F6465D",
  });
  if (!res.isConfirmed) return;

  try {
    await deleteWithdrawal(w.withdrawal_id);
    Swal.fire({ icon: "success", title: "ลบเรียบร้อยแล้ว", timer: 1200, showConfirmButton: false });
    await loadData();
  } catch (err: any) {
    Swal.fire("ลบไม่สำเร็จ", err.response?.data?.message || "เกิดข้อผิดพลาด", "error");
  }
};

const copyTx = async (tx: string) => {
  await navigator.clipboard.writeText(tx);
  Swal.fire({ icon: "success", title: "คัดลอกหมายเลขธุรกรรมแล้ว", timer: 1000, showConfirmButton: false });
};

const exportCSV = () => {
  const headers = ["วันที่", "จำนวนเงิน", "สกุลเงิน", "สถานะ", "วิธีการชำระเงิน", "หมายเลขธุรกรรม", "หมายเหตุ"];
  const rows = withdrawals.value.map((w) => [
    formatDateTime(w.withdrawal_time),
    w.amount.toFixed(2),
    w.currency,
    w.status,
    w.payment_method,
    w.transaction_id,
    `"${(w.notes || "").replace(/"/g, '""')}"`,
  ]);
  const csvContent = "\uFEFF" + [headers.join(","), ...rows.map((r) => r.join(","))].join("\n");
  const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `withdrawal_tax_${selectedYear.value || "all"}.csv`;
  a.click();
  URL.revokeObjectURL(url);
};

const onPagePaste = async (e: ClipboardEvent) => {
  const items = e.clipboardData?.items;
  if (!items) return;
  for (const item of items) {
    if (item.type.startsWith("image/")) {
      const file = item.getAsFile();
      if (file) {
        e.preventDefault();
        if (!modalOpen.value) {
          openScanModal();
        }
        await scanFile(file);
        break;
      }
    }
  }
};

onMounted(() => {
  loadData();
  window.addEventListener("paste", onPagePaste);
});

onBeforeUnmount(() => {
  window.removeEventListener("paste", onPagePaste);
  clearPreview();
});
</script>
