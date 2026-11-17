<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold">แดชบอร์ดภาพรวมพอร์ต</h1>
        <p class="text-sm opacity-60">วิเคราะห์สถิติการเทรด Binance แบบเรียลไทม์</p>
      </div>
      <div class="flex gap-2">
        <NuxtLink to="/trades" class="btn btn-primary btn-sm gap-2">
          <Plus class="w-4 h-4" />
          บันทึกการเทรด / สแกนรูป
        </NuxtLink>
      </div>
    </div>

    <div v-if="summary?.is_daily_loss_alert" class="alert alert-error shadow-sm">
      <AlertTriangle class="w-5 h-5" />
      <div>
        <h3 class="font-bold text-sm">แจ้งเตือนวินัยการเทรด!</h3>
        <div class="text-xs">
          วันนี้คุณขาดทุนสุทธิ {{ summary.today_pnl.toFixed(2) }} USDT เกินขีดจำกัดต่อวันที่ตั้งไว้ (-{{ summary.max_daily_loss_usd }} USDT) แนะนำให้หยุดพักหน้าจอก่อนครับ
        </div>
      </div>
    </div>

    <div v-if="loading" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="n in 4" :key="n" class="skeleton h-28 rounded-2xl"></div>
    </div>

    <div v-else-if="summary" class="stats stats-vertical lg:stats-horizontal w-full bg-base-100 border border-base-300 shadow-sm">
      <div class="stat">
        <div class="stat-title text-xs">ยอดพอร์ตปัจจุบัน</div>
        <div class="stat-value text-2xl text-primary">
          {{ summary.current_balance.toLocaleString() }} <span class="text-sm font-normal opacity-60">USDT</span>
        </div>
        <div class="stat-desc">
          ทุนเริ่มต้น {{ summary.initial_balance.toLocaleString() }} USDT
        </div>
      </div>

      <div class="stat">
        <div class="stat-title text-xs">กำไร/ขาดทุนสุทธิ (Net PnL)</div>
        <div :class="['stat-value text-2xl', summary.net_pnl >= 0 ? 'text-success' : 'text-error']">
          {{ summary.net_pnl >= 0 ? '+' : '' }}{{ summary.net_pnl.toFixed(2) }}
        </div>
        <div :class="['stat-desc font-medium', summary.net_pnl_pct >= 0 ? 'text-success' : 'text-error']">
          {{ summary.net_pnl_pct >= 0 ? '+' : '' }}{{ summary.net_pnl_pct.toFixed(2) }}% จากทุนตั้งต้น
        </div>
      </div>

      <div class="stat">
        <div class="stat-title text-xs">อัตราชนะ (Win Rate)</div>
        <div class="stat-value text-2xl">{{ summary.win_rate.toFixed(1) }}%</div>
        <div class="stat-desc">
          ชนะ {{ summary.win_trades }} / แพ้ {{ summary.loss_trades }} (รวม {{ summary.total_trades }} ไม้)
        </div>
      </div>

      <div class="stat">
        <div class="stat-title text-xs">Profit Factor & R:R</div>
        <div class="stat-value text-2xl">{{ summary.profit_factor.toFixed(2) }}</div>
        <div class="stat-desc">
          เฉลี่ย Risk:Reward = 1 : {{ summary.avg_risk_reward.toFixed(2) }}
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="lg:col-span-2">
        <DashboardEquityChart :points="summary?.equity_curve || []" :loading="loading" />
      </div>

      <div class="card bg-base-100 border border-base-300 shadow-sm">
        <div class="card-body p-5">
          <h3 class="font-bold text-base mb-3">สถิติสำคัญเพิ่มเติม</h3>
          <div v-if="loading" class="space-y-3">
            <div v-for="n in 4" :key="n" class="skeleton h-10 w-full rounded-lg"></div>
          </div>
          <div v-else-if="summary" class="space-y-3 text-sm">
            <div class="flex justify-between items-center p-2.5 rounded-lg bg-base-200/60">
              <span class="opacity-70">กำไรสูงสุดต่อไม้</span>
              <span class="font-bold text-success">+{{ summary.best_trade_pnl.toFixed(2) }} USDT</span>
            </div>
            <div class="flex justify-between items-center p-2.5 rounded-lg bg-base-200/60">
              <span class="opacity-70">ขาดทุนหนักสุดต่อไม้</span>
              <span class="font-bold text-error">{{ summary.worst_trade_pnl.toFixed(2) }} USDT</span>
            </div>
            <div class="flex justify-between items-center p-2.5 rounded-lg bg-base-200/60">
              <span class="opacity-70">สถิติต่อเนื่องล่าสุด</span>
              <span class="badge" :class="summary.streak_type === 'WIN' ? 'badge-success' : summary.streak_type === 'LOSS' ? 'badge-error' : 'badge-ghost'">
                {{ summary.streak_type === 'WIN' ? `ชนะติด ${summary.current_streak} ไม้` : summary.streak_type === 'LOSS' ? `แพ้ติด ${summary.current_streak} ไม้` : 'ยังไม่มีข้อมูล' }}
              </span>
            </div>

            <div class="divider my-2 text-xs opacity-60">เทียบฝั่ง Long / Short</div>
            <div v-for="s in summary.side_stats" :key="s.side" class="flex justify-between items-center text-xs">
              <span class="badge badge-sm" :class="s.side === 'LONG' ? 'badge-success' : s.side === 'SHORT' ? 'badge-error' : 'badge-info'">
                {{ s.side }} ({{ s.total_trades }} ไม้)
              </span>
              <span>Win {{ s.win_rate.toFixed(0) }}%</span>
              <span :class="s.total_pnl >= 0 ? 'text-success font-bold' : 'text-error font-bold'">
                {{ s.total_pnl >= 0 ? '+' : '' }}{{ s.total_pnl.toFixed(2) }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="lg:col-span-2">
        <DashboardPnlCalendar
          :days="calendarDays"
          :year="calYear"
          :month="calMonth"
          :loading="calLoading"
          @change-month="onChangeMonth"
        />
      </div>

      <div class="space-y-6">
        <div class="card bg-base-100 border border-base-300 shadow-sm">
          <div class="card-body p-5">
            <h3 class="font-bold text-base mb-3">สถิติแยกตามคู่เหรียญ</h3>
            <div v-if="loading" class="space-y-2">
              <div v-for="n in 3" :key="n" class="skeleton h-8 w-full"></div>
            </div>
            <div v-else-if="!summary?.symbol_stats?.length" class="text-xs opacity-60 py-4 text-center">
              ยังไม่มีข้อมูลคู่เหรียญ
            </div>
            <div v-else class="space-y-2">
              <div v-for="sym in summary.symbol_stats" :key="sym.symbol" class="flex justify-between items-center p-2 rounded-lg bg-base-200/50 text-xs">
                <div>
                  <div class="font-bold">{{ sym.symbol }}</div>
                  <div class="opacity-60">{{ sym.total_trades }} ไม้ (Win {{ sym.win_rate.toFixed(0) }}%)</div>
                </div>
                <div :class="['font-bold', sym.total_pnl >= 0 ? 'text-success' : 'text-error']">
                  {{ sym.total_pnl >= 0 ? '+' : '' }}{{ sym.total_pnl.toFixed(2) }} USDT
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="card bg-base-100 border border-base-300 shadow-sm">
          <div class="card-body p-5">
            <h3 class="font-bold text-base mb-3">สถิติตามช่วงเวลาเทรด</h3>
            <div v-if="loading" class="space-y-2">
              <div v-for="n in 3" :key="n" class="skeleton h-8 w-full"></div>
            </div>
            <div v-else-if="!summary?.session_stats?.length" class="text-xs opacity-60 py-4 text-center">
              ยังไม่มีข้อมูลช่วงเวลา
            </div>
            <div v-else class="space-y-2">
              <div v-for="sess in summary.session_stats" :key="sess.session" class="flex justify-between items-center p-2 rounded-lg bg-base-200/50 text-xs">
                <div>
                  <div class="font-medium">{{ sess.session }}</div>
                  <div class="opacity-60">{{ sess.total_trades }} ไม้ (Win {{ sess.win_rate.toFixed(0) }}%)</div>
                </div>
                <div :class="['font-bold', sess.total_pnl >= 0 ? 'text-success' : 'text-error']">
                  {{ sess.total_pnl >= 0 ? '+' : '' }}{{ sess.total_pnl.toFixed(2) }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Plus, AlertTriangle } from "lucide-vue-next";
import { useDashboard } from "~/composables/api/useDashboard";
import type { DashboardSummary, PnlCalendarDay } from "~/models/dashboard/dashboard_model";

definePageMeta({ middleware: "auth" });

const { fetchSummary, fetchCalendar } = useDashboard();

const loading = ref(true);
const calLoading = ref(true);
const summary = ref<DashboardSummary | null>(null);
const calendarDays = ref<PnlCalendarDay[]>([]);

const now = new Date();
const calYear = ref(now.getFullYear());
const calMonth = ref(now.getMonth() + 1);

const loadCalendar = async (y: number, m: number) => {
  calLoading.value = true;
  try {
    calendarDays.value = await fetchCalendar(y, m);
  } finally {
    calLoading.value = false;
  }
};

const onChangeMonth = (y: number, m: number) => {
  calYear.value = y;
  calMonth.value = m;
  loadCalendar(y, m);
};

onMounted(async () => {
  loading.value = true;
  try {
    summary.value = await fetchSummary();
  } finally {
    loading.value = false;
  }
  loadCalendar(calYear.value, calMonth.value);
});
</script>
