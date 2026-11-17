<template>
  <div class="card agy-card bg-base-100/90 border border-base-300 shadow-sm">
    <div class="card-body p-5">
      <div class="flex items-center justify-between mb-3">
        <div>
          <h3 class="font-bold text-base">กราฟการเติบโตของพอร์ต (Equity Curve)</h3>
          <p class="text-xs opacity-60">แสดงยอดเงินคงเหลือสะสมตามวันที่มีการเทรด</p>
        </div>
      </div>

      <div v-if="loading" class="skeleton h-64 w-full rounded-xl"></div>

      <div v-else-if="points.length === 0" class="h-64 flex items-center justify-center text-sm opacity-60">
        ยังไม่มีข้อมูลประวัติการเทรดสำหรับสร้างกราฟ
      </div>

      <div v-else class="h-64">
        <Line :data="chartData" :options="chartOptions" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Filler,
  Legend,
} from "chart.js";
import { Line } from "vue-chartjs";
import type { EquityCurvePoint } from "~/models/dashboard/dashboard_model";

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler, Legend);

const props = defineProps<{
  points: EquityCurvePoint[];
  loading: boolean;
}>();

const chartData = computed(() => ({
  labels: props.points.map((p) => p.date),
  datasets: [
    {
      label: "มูลค่าพอร์ต (USDT)",
      data: props.points.map((p) => p.balance),
      borderColor: "#FCD535",
      backgroundColor: "rgba(252, 213, 53, 0.14)",
      borderWidth: 2.5,
      fill: true,
      tension: 0.3,
      pointRadius: 3,
      pointBackgroundColor: "#FCD535",
    },
  ],
}));

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false,
    },
  },
  scales: {
    x: {
      grid: { color: "rgba(132, 142, 156, 0.12)" },
      ticks: { color: "#848E9C", font: { size: 11 } },
    },
    y: {
      grid: { color: "rgba(132, 142, 156, 0.12)" },
      ticks: { color: "#848E9C", font: { size: 11 } },
    },
  },
};
</script>
