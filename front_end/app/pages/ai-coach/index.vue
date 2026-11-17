<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold">AI โค้ชสอนเทรด & วิเคราะห์กราฟ</h1>
        <p class="text-sm opacity-60">ดึงข้อมูลราคาและอินดิเคเตอร์สดจาก Binance หรืออัปโหลดรูปกราฟให้ AI ช่วยวิเคราะห์</p>
      </div>
    </div>

    <div role="tablist" class="tabs tabs-boxed bg-base-100 border border-base-300 w-fit p-1">
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
      <div class="lg:col-span-2 space-y-6">
        <div class="card bg-base-100 border border-base-300 shadow-sm">
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

              <div class="h-60">
                <Line :data="priceChartData" :options="priceChartOptions" />
              </div>
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

        <div v-else-if="chartResult" class="card bg-base-100 border border-base-300 shadow-sm">
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

      <div class="space-y-6">
        <div class="card bg-base-100 border border-base-300 shadow-sm">
          <div class="card-body p-5 space-y-4">
            <h3 class="font-bold text-base">ตั้งค่าคำถามให้ AI วิเคราะห์</h3>

            <label class="form-control">
              <span class="label-text text-xs mb-1">แนบรูปกราฟเพิ่ม (ไม่บังคับ)</span>
              <input type="file" accept="image/*" class="file-input file-input-bordered file-input-sm w-full" @change="onChartFileChange" />
            </label>

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
      <div class="card bg-base-100 border border-base-300 shadow-sm">
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

      <div class="card bg-base-100 border border-base-300 shadow-sm">
        <div class="card-body p-5">
          <h3 class="font-bold text-base mb-3">ประวัติการวิเคราะห์ล่าสุด</h3>
          <div v-if="logsLoading" class="space-y-2">
            <div v-for="n in 3" :key="n" class="skeleton h-14 w-full rounded-lg"></div>
          </div>
          <div v-else-if="logs.length === 0" class="text-xs opacity-60 text-center py-4">
            ยังไม่มีประวัติการวิเคราะห์
          </div>
          <div v-else class="space-y-3">
            <div v-for="l in logs" :key="l.id" class="p-3.5 rounded-xl bg-base-200/50 border border-base-300">
              <div class="flex justify-between items-center text-xs mb-1.5">
                <span class="font-bold text-primary">{{ l.symbol }} ({{ l.analysis_type }})</span>
                <span class="opacity-60">{{ l.created_at }}</span>
              </div>
              <p class="text-xs whitespace-pre-wrap line-clamp-4">{{ l.ai_response }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Bot, Sparkles, RefreshCw } from "lucide-vue-next";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
} from "chart.js";
import { Line } from "vue-chartjs";
import Swal from "sweetalert2";
import { useAiCoach } from "~/composables/api/useAiCoach";
import type { AiAnalysisResponse, MarketDataSummary } from "~/models/ai/ai_model";

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend);

definePageMeta({ middleware: "auth" });

const { fetchMarketKlines, analyzeChart, reviewJournal, fetchAiLogs } = useAiCoach();

const activeTab = ref<"chart" | "journal">("chart");
const symbol = ref("BTCUSDT");
const timeframe = ref("1h");
const timeframes = ["15m", "1h", "4h", "1d"];
const question = ref("");
const teachMode = ref(true);
const chartFile = ref<File | null>(null);

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

const onChartFileChange = (e: Event) => {
  const input = e.target as HTMLInputElement;
  chartFile.value = input.files?.[0] || null;
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
  } finally {
    logsLoading.value = false;
  }
};

const priceChartData = computed(() => {
  const klines = market.value?.klines || [];
  return {
    labels: klines.map((k) => k.time_str),
    datasets: [
      {
        label: `ราคาปิด ${symbol.value}`,
        data: klines.map((k) => k.close),
        borderColor: "#0ECB81",
        borderWidth: 2,
        pointRadius: 0,
        tension: 0.2,
      },
    ],
  };
});

const priceChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    x: { grid: { display: false }, ticks: { maxTicksLimit: 8, color: "#848E9C", font: { size: 10 } } },
    y: { grid: { color: "rgba(132, 142, 156, 0.12)" }, ticks: { color: "#848E9C", font: { size: 10 } } },
  },
};

onMounted(() => {
  loadMarket();
  loadLogs();
});
</script>
