<template>
  <div class="card agy-card bg-base-100/90 border border-base-300 shadow-sm">
    <div class="card-body p-5">
      <div class="flex flex-wrap items-center justify-between gap-2 mb-4">
        <div>
          <h3 class="font-bold text-base">ปฏิทินกำไร-ขาดทุนรายวัน (PnL Calendar)</h3>
          <p class="text-xs opacity-60">ดูภาพรวมวันที่ได้กำไรและขาดทุนในแต่ละเดือน</p>
        </div>
        <div class="flex items-center gap-2">
          <button class="btn btn-sm btn-ghost btn-square" @click="changeMonth(-1)">
            <ChevronLeft class="w-4 h-4" />
          </button>
          <span class="font-semibold text-sm min-w-[130px] text-center">
            {{ monthLabel }}
          </span>
          <button class="btn btn-sm btn-ghost btn-square" @click="changeMonth(1)">
            <ChevronRight class="w-4 h-4" />
          </button>
        </div>
      </div>

      <div v-if="loading" class="grid grid-cols-7 gap-2">
        <div v-for="n in 35" :key="n" class="skeleton h-16 w-full rounded-lg"></div>
      </div>

      <div v-else>
        <div class="grid grid-cols-7 gap-1.5 text-center text-xs font-semibold opacity-60 mb-2">
          <div>อา.</div>
          <div>จ.</div>
          <div>อ.</div>
          <div>พ.</div>
          <div>พฤ.</div>
          <div>ศ.</div>
          <div>ส.</div>
        </div>

        <div class="grid grid-cols-7 gap-1.5">
          <div
            v-for="blank in firstDayOffset"
            :key="'b-' + blank"
            class="h-16 rounded-lg bg-base-200/40"
          ></div>

          <div
            v-for="cell in calendarCells"
            :key="cell.date"
            :class="[
              'agy-calendar-cell h-16 rounded-lg p-1.5 flex flex-col justify-between border transition-colors',
              cell.data
                ? cell.data.pnl >= 0
                  ? 'bg-success/15 border-success/40 text-success'
                  : 'bg-error/15 border-error/40 text-error'
                : 'bg-base-200/60 border-base-300/60',
            ]"
          >
            <div class="flex justify-between items-center text-[11px] font-medium">
              <span class="text-base-content opacity-75">{{ cell.day }}</span>
              <span v-if="cell.data" class="text-[10px] opacity-80">
                {{ cell.data.trade_count }} ไม้
              </span>
            </div>

            <div v-if="cell.data" class="text-right font-bold text-xs truncate">
              {{ cell.data.pnl >= 0 ? '+' : '' }}{{ cell.data.pnl.toFixed(2) }}
            </div>
            <div v-else class="text-right text-[10px] opacity-30">-</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ChevronLeft, ChevronRight } from "lucide-vue-next";
import type { PnlCalendarDay } from "~/models/dashboard/dashboard_model";

const props = defineProps<{
  days: PnlCalendarDay[];
  year: number;
  month: number;
  loading: boolean;
}>();

const emit = defineEmits<{
  (e: "change-month", year: number, month: number): void;
}>();

const thaiMonths = [
  "มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
  "กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม",
];

const monthLabel = computed(() => {
  return `${thaiMonths[props.month - 1] || ""} ${props.year}`;
});

const firstDayOffset = computed(() => {
  return new Date(props.year, props.month - 1, 1).getDay();
});

const calendarCells = computed(() => {
  const totalDays = new Date(props.year, props.month, 0).getDate();
  const map = new Map<string, PnlCalendarDay>();
  for (const d of props.days) {
    map.set(d.date, d);
  }
  const cells = [];
  for (let d = 1; d <= totalDays; d++) {
    const mm = String(props.month).padStart(2, "0");
    const dd = String(d).padStart(2, "0");
    const key = `${props.year}-${mm}-${dd}`;
    cells.push({
      day: d,
      date: key,
      data: map.get(key) || null,
    });
  }
  return cells;
});

const changeMonth = (delta: number) => {
  let newMonth = props.month + delta;
  let newYear = props.year;
  if (newMonth < 1) {
    newMonth = 12;
    newYear--;
  } else if (newMonth > 12) {
    newMonth = 1;
    newYear++;
  }
  emit("change-month", newYear, newMonth);
};
</script>
