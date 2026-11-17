<template>
  <div class="space-y-6">
    <div class="agy-fade-up stagger-1 flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold">AI โค้ชสอนเทรด & วิเคราะห์กราฟ</h1>
        <p class="text-sm opacity-60">ดึงข้อมูลราคาและอินดิเคเตอร์สดจาก Binance หรืออัปโหลดรูปกราฟให้ AI ช่วยวิเคราะห์</p>
      </div>
    </div>

    <div role="tablist" class="agy-fade-up stagger-1 tabs tabs-boxed bg-base-100/90 border border-base-300 w-fit p-1">
      <button
        role="tab"
        :class="['tab text-sm', activeTab === 'chart' ? 'tab-active font-semibold' : '']"
        @click="activeTab = 'chart'"
      >
        วิเคราะห์กราฟ & สอนเทรด
      </button>
      <button
        role="tab"
        :class="['tab text-sm', activeTab === 'journal' ? 'tab-active font-semibold' : '']"
        @click="activeTab = 'journal'"
      >
        AI วิเคราะห์นิสัยพอร์ตของฉัน
      </button>
    </div>

    <div v-if="activeTab === 'chart'" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="lg:col-span-2 space-y-6 agy-fade-up stagger-2">
        <div class="card agy-card bg-base-100/90 border border-base-300 shadow-sm">
          <div class="card-body p-5 space-y-4">
            <div class="flex flex-wrap items-center gap-3">
              <select v-model="symbol" class="select select-bordered select-sm font-bold" @change="loadMarket">
                <option value="BTCUSDT">BTC / USDT</option>
                <option value="ETHUSDT">ETH / USDT</option>
                <option value="BNBUSDT">BNB / USDT</option>
                <option value="SOLUSDT">SOL / USDT</option>
                <option value="XRPUSDT">XRP / USDT</option>
                <option value="DOGEUSDT">DOGE / USDT</option>
              </select>

              <div class="join">
                <button
                  v-for="tf in timeframes"
                  :key="tf"
                  :class="['btn btn-sm join-item', timeframe === tf ? 'btn-primary' : 'btn-ghost']"
                  @click="selectTimeframe(tf)"
                >
                  {{ tf }}
                </button>
              </div>

              <button class="btn btn-sm btn-ghost ml-auto gap-1.5" @click="loadMarket">
                <RefreshCw class="w-3.5 h-3.5" />
                อัปเดตราคา
              </button>
            </div>

            <div v-if="marketLoading" class="space-y-3">
              <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
                <div v-for="n in 4" :key="n" class="skeleton h-16 rounded-xl"></div>
              </div>
              <div class="skeleton h-60 w-full rounded-xl"></div>
            </div>

            <div v-else-if="market" class="space-y-4">
              <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
                <div class="p-3 rounded-xl bg-base-200/70">
                  <div class="text-xs opacity-60">ราคาล่าสุด</div>
                  <div class="text-lg font-bold text-primary">{{ market.current_price.toLocaleString() }}</div>
                  <div :class="['text-xs font-medium', market.change_24h_pct >= 0 ? 'text-success' : 'text-error']">
                    24 ชม.: {{ market.change_24h_pct >= 0 ? '+' : '' }}{{ market.change_24h_pct }}%
                  </div>
                </div>
                <div class="p-3 rounded-xl bg-base-200/70">
                  <div class="text-xs opacity-60">RSI (14)</div>
                  <div class="text-lg font-bold">{{ market.rsi_14 }}</div>
                  <div class="text-xs opacity-70">
                    {{ market.rsi_14 >= 70 ? 'โซนซื้อมาก (Overbought)' : market.rsi_14 <= 30 ? 'โซนขายมาก (Oversold)' : 'โซนปกติ' }}
                  </div>
                </div>
                <div class="p-3 rounded-xl bg-base-200/70">
                  <div class="text-xs opacity-60">EMA 20 / EMA 50</div>
                  <div class="text-xs font-bold mt-1">EMA20: {{ market.ema_20.toLocaleString() }}</div>
                  <div class="text-xs font-bold">EMA50: {{ market.ema_50.toLocaleString() }}</div>
                </div>
                <div class="p-3 rounded-xl bg-base-200/70">
                  <div class="text-xs opacity-60">แนวรับ / แนวต้าน</div>
                  <div class="text-xs font-bold text-success mt-1">รับ: {{ market.support.toLocaleString() }}</div>
                  <div class="text-xs font-bold text-error">ต้าน: {{ market.resistance.toLocaleString() }}</div>
                </div>
              </div>

              <AiCandlestickChart :klines="market.klines" :symbol="symbol" />
            </div>
          </div>
        </div>

        <div v-if="analyzing" class="card bg-base-100 border border-base-300 shadow-sm p-6 space-y-3">
          <div class="text-sm font-semibold text-primary">AI กำลังวิเคราะห์กราฟและเตรียมบทเรียนสำหรับคุณ...</div>
          <div class="skeleton h-5 w-1/2"></div>
          <div class="skeleton h-4 w-full"></div>
          <div class="skeleton h-4 w-5/6"></div>
          <div class="skeleton h-4 w-4/5"></div>
        </div>

        <div v-else-if="chartResult" class="card agy-card agy-fade-up stagger-1 bg-base-100/90 border border-base-300 shadow-sm">
          <div class="card-body p-5 space-y-3">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <h3 class="font-bold text-base flex items-center gap-2">
                <Bot class="w-5 h-5 text-primary" />
                ผลการวิเคราะห์ {{ chartResult.symbol }} ({{ chartResult.timeframe }})
              </h3>
              <span
                :class="[
                  'badge font-semibold',
                  chartResult.signal_direction === 'BULLISH'
                    ? 'badge-success'
                    : chartResult.signal_direction === 'BEARISH'
                    ? 'badge-error'
                    : 'badge-ghost',
                ]"
              >
                {{
                  chartResult.signal_direction === 'BULLISH'
                    ? 'มุมมองขาขึ้น (Bullish)'
                    : chartResult.signal_direction === 'BEARISH'
                    ? 'มุมมองขาลง (Bearish)'
                    : 'รอความชัดเจน (Neutral)'
                }}
              </span>
            </div>
            <div class="whitespace-pre-wrap text-sm leading-relaxed bg-base-200/50 p-4 rounded-xl">
              {{ chartResult.ai_response }}
            </div>
          </div>
        </div>
      </div>

      <div class="space-y-6 agy-fade-up stagger-3">
        <div class="card agy-card bg-base-100/90 border border-base-300 shadow-sm">
          <div class="card-body p-5 space-y-4">
            <h3 class="font-bold text-base">ตั้งค่าคำถามให้ AI วิเคราะห์</h3>

            <div class="form-control">
              <div class="flex items-center justify-between mb-1.5">
                <span class="label-text text-xs">แนบรูปกราฟเพิ่ม (ไม่บังคับ)</span>
                <span class="badge badge-xs badge-primary badge-outline">รองรับ Ctrl + V วางรูป</span>
              </div>

              <div v-if="chartPreviewUrl" class="relative rounded-xl overflow-hidden border border-primary/40 bg-base-200/70 p-2">
                <img :src="chartPreviewUrl" alt="ภาพกราฟที่แนบ" class="w-full max-h-44 object-contain rounded-lg" />
                <div class="flex items-center justify-between mt-2 px-1 text-[11px]">
                  <span class="opacity-70 truncate max-w-[180px]">{{ chartFile?.name || 'ภาพจาก Clipboard' }}</span>
                  <button type="button" class="btn btn-xs btn-error btn-outline gap-1" @click="clearChartFile">
                    <X class="w-3 h-3" />
                    ลบรูป
                  </button>
                </div>
              </div>

              <label
                v-else
                @dragover.prevent="isDragging = true"
                @dragleave.prevent="isDragging = false"
                @drop.prevent="onDropChartFile"
                :class="[
                  'flex flex-col items-center justify-center border-2 border-dashed rounded-xl p-4 cursor-pointer transition-all text-center',
                  isDragging ? 'border-primary bg-primary/10 scale-[1.01]' : 'border-base-300 hover:border-primary/70 bg-base-200/40',
                ]"
              >
                <ImagePlus class="w-6 h-6 text-primary mb-1.5" />
                <span class="text-xs font-medium">กด Ctrl + V วางรูปที่แคปจอได้เลย</span>
                <span class="text-[11px] opacity-60 mt-0.5">หรือคลิกเลือกไฟล์ / ลากรูปมาวางที่นี่</span>
                <input ref="chartInputRef" type="file" accept="image/*" class="hidden" @change="onChartFileChange" />
              </label>
            </div>

            <label class="form-control">
              <span class="label-text text-xs mb-1">สิ่งที่อยากให้ AI ช่วยดู</span>
              <textarea
                v-model="question"
                rows="3"
                placeholder="เช่น จังหวะนี้ควรรอเข้า Long ตรงแนวรับไหน หรือควรวาง Stop Loss ตรงไหนดี?"
                class="textarea textarea-bordered text-sm"
              ></textarea>
            </label>

            <label class="flex items-center justify-between cursor-pointer p-3 rounded-xl bg-base-200/60">
              <div>
                <div class="font-semibold text-sm">โหมดครูสอนเทรด</div>
                <div class="text-xs opacity-60">อธิบายเหตุผลของ RSI, EMA และวิธีอ่านกราฟ</div>
              </div>
              <input v-model="teachMode" type="checkbox" class="toggle toggle-primary" />
            </label>

            <button class="btn btn-primary w-full gap-2" :disabled="analyzing" @click="runChartAnalysis">
              <Sparkles class="w-4 h-4" />
              {{ analyzing ? 'กำลังวิเคราะห์...' : 'ให้ AI วิเคราะห์กราฟตอนนี้' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="space-y-6">
      <div class="card agy-card agy-fade-up stagger-2 bg-base-100/90 border border-base-300 shadow-sm">
        <div class="card-body p-6 space-y-4">
          <div class="flex flex-wrap items-center justify-between gap-4">
            <div>
              <h3 class="font-bold text-lg">AI โค้ชวิเคราะห์พฤติกรรมการเทรดส่วนตัว</h3>
              <p class="text-sm opacity-60">ระบบจะดึงประวัติเทรดของคุณไปวิเคราะห์จุดแข็ง จุดอ่อน และนิสัยที่ทำให้พอร์ตโตช้า</p>
            </div>
            <button class="btn btn-primary gap-2" :disabled="reviewing" @click="runJournalReview">
              <Sparkles class="w-4 h-4" />
              {{ reviewing ? 'กำลังอ่านประวัติพอร์ต...' : 'ให้ AI วิเคราะห์พอร์ตของฉัน' }}
            </button>
          </div>

          <div v-if="reviewing" class="space-y-3 pt-2">
            <div class="skeleton h-5 w-1/3"></div>
            <div class="skeleton h-4 w-full"></div>
            <div class="skeleton h-4 w-4/5"></div>
            <div class="skeleton h-4 w-3/4"></div>
          </div>

          <div v-else-if="journalResult" class="whitespace-pre-wrap text-sm leading-relaxed bg-base-200/50 p-5 rounded-xl">
            {{ journalResult.ai_response }}
          </div>
        </div>
      </div>
    </div>

    <div class="card agy-card agy-fade-up stagger-4 bg-base-100/90 border border-base-300 shadow-sm overflow-hidden">
      <div class="card-body p-5">
        <div class="flex flex-wrap items-center justify-between gap-2 mb-3">
          <div>
            <h3 class="font-bold text-base">ตารางประวัติการวิเคราะห์ AI ทั้งหมด</h3>
            <p class="text-xs opacity-60">บันทึกเก็บไว้ในฐานข้อมูลอัตโนมัติ รีเฟรชหน้าไม่หาย และสามารถกดดูซ้ำหรือลบออกได้</p>
          </div>
          <span class="badge badge-outline badge-sm">{{ logs.length }} รายการ</span>
        </div>

        <div v-if="logsLoading" class="space-y-2">
          <div v-for="n in 4" :key="n" class="skeleton h-12 w-full rounded-lg"></div>
        </div>

        <div v-else-if="logs.length === 0" class="text-xs opacity-60 text-center py-8">
          ยังไม่มีประวัติการวิเคราะห์ AI
        </div>

        <div v-else class="overflow-x-auto">
          <table class="table table-sm w-full">
            <thead>
              <tr class="bg-base-200/70">
                <th>วัน-เวลา</th>
                <th>คู่เหรียญ / TF</th>
                <th>ประเภท</th>
                <th>คำถาม / หัวข้อ</th>
                <th>มุมมองสัญญาณ</th>
                <th class="text-right">จัดการ</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="l in logs"
                :key="l.id"
                :class="[
                  'hover',
                  (chartResult?.id === l.id || journalResult?.id === l.id) ? 'bg-primary/5' : '',
                ]"
              >
                <td class="text-xs opacity-75 whitespace-nowrap">{{ l.created_at }}</td>
                <td class="font-bold text-xs whitespace-nowrap">
                  {{ l.symbol }} <span class="opacity-60 font-normal">({{ l.timeframe }})</span>
                </td>
                <td>
                  <span :class="['badge badge-xs', l.analysis_type === 'Chart' ? 'badge-primary badge-outline' : 'badge-info badge-outline']">
                    {{ l.analysis_type === 'Chart' ? 'วิเคราะห์กราฟ' : 'วิเคราะห์พอร์ต' }}
                  </span>
                </td>
                <td class="text-xs max-w-[240px] truncate">{{ l.prompt_summary || '-' }}</td>
                <td>
                  <span
                    :class="[
                      'badge badge-xs font-semibold',
                      l.signal_direction === 'BULLISH'
                        ? 'badge-success'
                        : l.signal_direction === 'BEARISH'
                        ? 'badge-error'
                        : 'badge-ghost',
                    ]"
                  >
                    {{
                      l.signal_direction === 'BULLISH'
                        ? 'ขาขึ้น'
                        : l.signal_direction === 'BEARISH'
                        ? 'ขาลง'
                        : 'เป็นกลาง'
                    }}
                  </span>
                </td>
                <td class="text-right">
                  <div class="flex justify-end gap-1">
                    <button type="button" class="btn btn-ghost btn-xs gap-1" @click="viewLog(l)" title="ดูผลวิเคราะห์">
                      <Eye class="w-3.5 h-3.5 text-primary" />
                      ดูผล
                    </button>
                    <button type="button" class="btn btn-ghost btn-xs text-error" @click="removeLog(l)" title="ลบประวัติ">
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Bot, Sparkles, RefreshCw, ImagePlus, X, Eye, Trash2 } from "lucide-vue-next";
import Swal from "sweetalert2";
import { useAiCoach } from "~/composables/api/useAiCoach";
import type { AiAnalysisResponse, MarketDataSummary } from "~/models/ai/ai_model";

definePageMeta({ middleware: "auth" });

const { fetchMarketKlines, analyzeChart, reviewJournal, fetchAiLogs, deleteAiLog } = useAiCoach();

const activeTab = ref<"chart" | "journal">("chart");
const symbol = ref("BTCUSDT");
const timeframe = ref("1h");
const timeframes = ["15m", "1h", "4h", "1d"];
const question = ref("");
const teachMode = ref(true);
const chartFile = ref<File | null>(null);
const chartPreviewUrl = ref<string | null>(null);
const chartInputRef = ref<HTMLInputElement | null>(null);
const isDragging = ref(false);

const marketLoading = ref(true);
const analyzing = ref(false);
const reviewing = ref(false);
const logsLoading = ref(false);

const market = ref<MarketDataSummary | null>(null);
const chartResult = ref<AiAnalysisResponse | null>(null);
const journalResult = ref<AiAnalysisResponse | null>(null);
const logs = ref<AiAnalysisResponse[]>([]);

const loadMarket = async () => {
  marketLoading.value = true;
  try {
    market.value = await fetchMarketKlines(symbol.value, timeframe.value);
  } catch (err: any) {
    Swal.fire("โหลดกราฟไม่สำเร็จ", err.response?.data?.message || "ไม่สามารถติดต่อ Binance ได้", "error");
  } finally {
    marketLoading.value = false;
  }
};

const selectTimeframe = (tf: string) => {
  timeframe.value = tf;
  loadMarket();
};

const setChartImage = (file: File | null) => {
  if (chartPreviewUrl.value) {
    URL.revokeObjectURL(chartPreviewUrl.value);
    chartPreviewUrl.value = null;
  }
  chartFile.value = file;
  if (file) {
    chartPreviewUrl.value = URL.createObjectURL(file);
  }
};

const clearChartFile = () => {
  setChartImage(null);
  if (chartInputRef.value) {
    chartInputRef.value.value = "";
  }
};

const onChartFileChange = (e: Event) => {
  const input = e.target as HTMLInputElement;
  setChartImage(input.files?.[0] || null);
};

const onDropChartFile = (e: DragEvent) => {
  isDragging.value = false;
  const file = e.dataTransfer?.files?.[0];
  if (file && file.type.startsWith("image/")) {
    setChartImage(file);
  }
};

const onWindowPaste = (e: ClipboardEvent) => {
  if (activeTab.value !== "chart") return;
  const items = e.clipboardData?.items;
  if (!items) return;
  for (const item of items) {
    if (item.type.startsWith("image/")) {
      const file = item.getAsFile();
      if (file) {
        e.preventDefault();
        setChartImage(file);
        break;
      }
    }
  }
};

const runChartAnalysis = async () => {
  analyzing.value = true;
  try {
    chartResult.value = await analyzeChart(symbol.value, timeframe.value, question.value, teachMode.value, chartFile.value);
    loadLogs();
  } catch (err: any) {
    Swal.fire("เกิดข้อผิดพลาด", err.response?.data?.message || "วิเคราะห์ไม่สำเร็จ", "error");
  } finally {
    analyzing.value = false;
  }
};

const runJournalReview = async () => {
  reviewing.value = true;
  try {
    journalResult.value = await reviewJournal();
    loadLogs();
  } catch (err: any) {
    Swal.fire("เกิดข้อผิดพลาด", err.response?.data?.message || "วิเคราะห์พอร์ตไม่สำเร็จ", "error");
  } finally {
    reviewing.value = false;
  }
};

const loadLogs = async () => {
  logsLoading.value = true;
  try {
    logs.value = await fetchAiLogs();
    if (!chartResult.value) {
      chartResult.value = logs.value.find((l) => l.analysis_type === "Chart") || null;
    }
    if (!journalResult.value) {
      journalResult.value = logs.value.find((l) => l.analysis_type === "TradeJournalReview") || null;
    }
  } finally {
    logsLoading.value = false;
  }
};

const viewLog = (l: AiAnalysisResponse) => {
  if (l.analysis_type === "TradeJournalReview") {
    activeTab.value = "journal";
    journalResult.value = l;
  } else {
    activeTab.value = "chart";
    chartResult.value = l;
  }
  window.scrollTo({ top: 0, behavior: "smooth" });
};

const removeLog = async (l: AiAnalysisResponse) => {
  const confirm = await Swal.fire({
    title: "ลบประวัติการวิเคราะห์?",
    text: `ต้องการลบผลการวิเคราะห์ ${l.symbol} (${l.created_at}) ออกจากระบบใช่หรือไม่?`,
    icon: "warning",
    showCancelButton: true,
    confirmButtonText: "ใช่, ลบเลย",
    cancelButtonText: "ยกเลิก",
    confirmButtonColor: "#F6465D",
  });
  if (!confirm.isConfirmed) return;

  try {
    await deleteAiLog(l.id);
    logs.value = logs.value.filter((item) => item.id !== l.id);
    if (chartResult.value?.id === l.id) {
      chartResult.value = logs.value.find((item) => item.analysis_type === "Chart") || null;
    }
    if (journalResult.value?.id === l.id) {
      journalResult.value = logs.value.find((item) => item.analysis_type === "TradeJournalReview") || null;
    }
    Swal.fire({ icon: "success", title: "ลบเรียบร้อยแล้ว", timer: 1200, showConfirmButton: false });
  } catch (err: any) {
    Swal.fire("เกิดข้อผิดพลาด", err.response?.data?.message || "ลบไม่สำเร็จ", "error");
  }
};

onMounted(() => {
  loadMarket();
  loadLogs();
  window.addEventListener("paste", onWindowPaste);
});

onBeforeUnmount(() => {
  window.removeEventListener("paste", onWindowPaste);
  if (chartPreviewUrl.value) {
    URL.revokeObjectURL(chartPreviewUrl.value);
  }
});
</script>
