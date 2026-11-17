<template>
  <div class="space-y-3">
    <div class="flex flex-wrap items-center justify-between gap-2 text-xs">
      <div class="flex flex-wrap items-center gap-3 font-mono bg-base-200/70 px-3 py-1.5 rounded-xl border border-base-300/70">
        <span class="opacity-70 font-sans">{{ activeCandle?.time_str || '-' }}</span>
        <span>O: <b class="text-base-content">{{ formatPrice(activeCandle?.open) }}</b></span>
        <span>H: <b class="text-success">{{ formatPrice(activeCandle?.high) }}</b></span>
        <span>L: <b class="text-error">{{ formatPrice(activeCandle?.low) }}</b></span>
        <span>C: <b :class="isUp(activeCandle) ? 'text-success' : 'text-error'">{{ formatPrice(activeCandle?.close) }}</b></span>
        <span
          v-if="activeCandle"
          :class="['badge badge-xs font-semibold', isUp(activeCandle) ? 'badge-success' : 'badge-error']"
        >
          {{ candleChangePct(activeCandle) }}%
        </span>
      </div>

      <div class="flex items-center gap-2">
        <div class="join bg-base-200/70 p-0.5 rounded-lg">
          <button
            v-for="cnt in [30, 50, 80]"
            :key="cnt"
            type="button"
            :class="['btn btn-xs join-item', visibleCount === cnt ? 'btn-primary' : 'btn-ghost']"
            @click="visibleCount = cnt"
          >
            {{ cnt }} แท่ง
          </button>
        </div>

        <div class="join bg-base-200/70 p-0.5 rounded-lg">
          <button
            type="button"
            :class="['btn btn-xs join-item', chartMode === 'candle' ? 'btn-primary' : 'btn-ghost']"
            @click="chartMode = 'candle'"
          >
            แท่งเทียน
          </button>
          <button
            type="button"
            :class="['btn btn-xs join-item', chartMode === 'line' ? 'btn-primary' : 'btn-ghost']"
            @click="chartMode = 'line'"
          >
            กราฟเส้น
          </button>
        </div>
      </div>
    </div>

    <div class="relative rounded-xl bg-base-200/40 border border-base-300/60 p-2 select-none">
      <div class="flex items-center gap-4 text-[11px] px-2 pb-1 opacity-75">
        <span class="inline-flex items-center gap-1.5">
          <span class="w-2.5 h-0.5 bg-[#FCD535] inline-block"></span>
          EMA 20
        </span>
        <span class="inline-flex items-center gap-1.5">
          <span class="w-2.5 h-0.5 bg-[#3B82F6] inline-block"></span>
          EMA 50
        </span>
      </div>

      <svg
        viewBox="0 0 820 290"
        class="w-full h-64 sm:h-72 overflow-visible"
        @mouseleave="hoverIndex = null"
      >
        <defs>
          <linearGradient id="lineAreaGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#0ECB81" stop-opacity="0.28" />
            <stop offset="100%" stop-color="#0ECB81" stop-opacity="0.0" />
          </linearGradient>
        </defs>

        <!-- เส้นกริดแนวนอนและป้ายราคา -->
        <g v-for="(tick, idx) in yTicks" :key="'y-' + idx">
          <line
            :x1="padLeft"
            :x2="padLeft + plotW"
            :y1="tick.y"
            :y2="tick.y"
            stroke="rgba(132, 142, 156, 0.14)"
            stroke-dasharray="3 3"
          />
          <text
            :x="padLeft + plotW + 8"
            :y="tick.y + 3"
            fill="#848E9C"
            font-size="10"
            font-family="monospace"
          >
            {{ formatPrice(tick.price) }}
          </text>
        </g>

        <!-- ไฮไลท์แท่งที่กำลังชี้ -->
        <g v-if="hoverIndex !== null && visibleCandles[hoverIndex]">
          <rect
            :x="candleX(hoverIndex) - stepW / 2"
            :y="padTop"
            :width="stepW"
            :height="plotH"
            fill="rgba(252, 213, 53, 0.07)"
          />
          <line
            :x1="candleX(hoverIndex)"
            :x2="candleX(hoverIndex)"
            :y1="padTop"
            :y2="padTop + plotH"
            stroke="rgba(252, 213, 53, 0.45)"
            stroke-dasharray="3 3"
          />
          <line
            :x1="padLeft"
            :x2="padLeft + plotW"
            :y1="priceToY(visibleCandles[hoverIndex]!.close)"
            :y2="priceToY(visibleCandles[hoverIndex]!.close)"
            stroke="rgba(252, 213, 53, 0.45)"
            stroke-dasharray="3 3"
          />
        </g>

        <!-- เส้น EMA 20 และ EMA 50 -->
        <polyline
          v-if="ema20Points"
          :points="ema20Points"
          fill="none"
          stroke="#FCD535"
          stroke-width="1.3"
          opacity="0.85"
        />
        <polyline
          v-if="ema50Points"
          :points="ema50Points"
          fill="none"
          stroke="#3B82F6"
          stroke-width="1.3"
          opacity="0.85"
        />

        <!-- โหมดกราฟแท่งเทียน -->
        <g v-if="chartMode === 'candle'">
          <g v-for="(c, i) in visibleCandles" :key="c.open_time">
            <line
              :x1="candleX(i)"
              :x2="candleX(i)"
              :y1="priceToY(c.high)"
              :y2="priceToY(c.low)"
              :stroke="isUp(c) ? '#0ECB81' : '#F6465D'"
              stroke-width="1.4"
            />
            <rect
              :x="candleX(i) - bodyW / 2"
              :y="Math.min(priceToY(c.open), priceToY(c.close))"
              :width="bodyW"
              :height="Math.max(2, Math.abs(priceToY(c.close) - priceToY(c.open)))"
              :fill="isUp(c) ? '#0ECB81' : '#F6465D'"
              rx="1"
            />
          </g>
        </g>

        <!-- โหมดกราฟเส้น -->
        <g v-else>
          <polygon v-if="lineAreaPoints" :points="lineAreaPoints" fill="url(#lineAreaGrad)" />
          <polyline
            v-if="closeLinePoints"
            :points="closeLinePoints"
            fill="none"
            stroke="#0ECB81"
            stroke-width="2.2"
          />
        </g>

        <!-- พื้นที่จับการชี้เมาส์แยกทีละแท่ง -->
        <rect
          v-for="(c, i) in visibleCandles"
          :key="'hit-' + c.open_time"
          :x="candleX(i) - stepW / 2"
          :y="padTop"
          :width="stepW"
          :height="plotH"
          fill="transparent"
          class="cursor-crosshair"
          @mouseenter="hoverIndex = i"
        />

        <!-- ป้ายเวลาด้านล่าง -->
        <g v-for="xt in xTicks" :key="'x-' + xt.x">
          <text
            :x="xt.x"
            :y="padTop + plotH + 18"
            fill="#848E9C"
            font-size="10"
            text-anchor="middle"
          >
            {{ xt.label }}
          </text>
        </g>
      </svg>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { MarketKline } from "~/models/ai/ai_model";

const props = defineProps<{
  klines: MarketKline[];
  symbol: string;
}>();

const chartMode = ref<"candle" | "line">("candle");
const visibleCount = ref(50);
const hoverIndex = ref<number | null>(null);

const padLeft = 10;
const padRight = 76;
const padTop = 14;
const padBottom = 26;
const svgW = 820;
const svgH = 290;
const plotW = svgW - padLeft - padRight;
const plotH = svgH - padTop - padBottom;

const calcEmaSeries = (closes: number[], period: number): number[] => {
  if (closes.length === 0) return [];
  const k = 2 / (period + 1);
  const res: number[] = [];
  let ema = closes[0]!;
  for (let i = 0; i < closes.length; i++) {
    if (i === 0) {
      ema = closes[0]!;
    } else {
      ema = closes[i]! * k + ema * (1 - k);
    }
    res.push(ema);
  }
  return res;
};

const sliceOffset = computed(() => Math.max(0, props.klines.length - visibleCount.value));

const visibleCandles = computed(() => props.klines.slice(sliceOffset.value));

const visibleEma20 = computed(() => {
  const all = calcEmaSeries(props.klines.map((k) => k.close), 20);
  return all.slice(sliceOffset.value);
});

const visibleEma50 = computed(() => {
  const all = calcEmaSeries(props.klines.map((k) => k.close), 50);
  return all.slice(sliceOffset.value);
});

const priceBounds = computed(() => {
  if (visibleCandles.value.length === 0) return { min: 0, max: 1 };
  let min = Infinity;
  let max = -Infinity;
  for (const c of visibleCandles.value) {
    if (c.low < min) min = c.low;
    if (c.high > max) max = c.high;
  }
  for (const v of visibleEma20.value) {
    if (v < min) min = v;
    if (v > max) max = v;
  }
  for (const v of visibleEma50.value) {
    if (v < min) min = v;
    if (v > max) max = v;
  }
  const span = max - min || max * 0.01 || 1;
  return {
    min: min - span * 0.06,
    max: max + span * 0.06,
  };
});

const stepW = computed(() => {
  const len = visibleCandles.value.length || 1;
  return plotW / len;
});

const bodyW = computed(() => Math.max(3, Math.min(14, stepW.value * 0.65)));

const candleX = (i: number) => padLeft + stepW.value * (i + 0.5);

const priceToY = (price: number) => {
  const { min, max } = priceBounds.value;
  const ratio = (price - min) / (max - min || 1);
  return padTop + plotH - ratio * plotH;
};

const yTicks = computed(() => {
  const { min, max } = priceBounds.value;
  const ticks = [];
  for (let i = 0; i <= 4; i++) {
    const ratio = i / 4;
    const price = min + (max - min) * ratio;
    ticks.push({ price, y: priceToY(price) });
  }
  return ticks;
});

const xTicks = computed(() => {
  const list = visibleCandles.value;
  if (list.length === 0) return [];
  const step = Math.max(1, Math.floor(list.length / 5));
  const out = [];
  for (let i = 0; i < list.length; i += step) {
    out.push({ x: candleX(i), label: list[i]!.time_str });
  }
  return out;
});

const ema20Points = computed(() =>
  visibleEma20.value.map((v, i) => `${candleX(i)},${priceToY(v)}`).join(" ")
);

const ema50Points = computed(() =>
  visibleEma50.value.map((v, i) => `${candleX(i)},${priceToY(v)}`).join(" ")
);

const closeLinePoints = computed(() =>
  visibleCandles.value.map((c, i) => `${candleX(i)},${priceToY(c.close)}`).join(" ")
);

const lineAreaPoints = computed(() => {
  const list = visibleCandles.value;
  if (list.length === 0) return "";
  const bottomY = padTop + plotH;
  const firstX = candleX(0);
  const lastX = candleX(list.length - 1);
  return `${firstX},${bottomY} ${closeLinePoints.value} ${lastX},${bottomY}`;
});

const activeCandle = computed(() => {
  const list = visibleCandles.value;
  if (list.length === 0) return null;
  if (hoverIndex.value !== null && list[hoverIndex.value]) {
    return list[hoverIndex.value]!;
  }
  return list[list.length - 1]!;
});

const isUp = (c: MarketKline | null | undefined) => {
  if (!c) return true;
  return c.close >= c.open;
};

const candleChangePct = (c: MarketKline) => {
  if (!c.open) return "0.00";
  const pct = ((c.close - c.open) / c.open) * 100;
  return (pct >= 0 ? "+" : "") + pct.toFixed(2);
};

const formatPrice = (val: number | undefined) => {
  if (val === undefined || val === null) return "-";
  if (val >= 1000) return val.toLocaleString(undefined, { maximumFractionDigits: 2 });
  if (val >= 1) return val.toFixed(4);
  return val.toFixed(6);
};
</script>
