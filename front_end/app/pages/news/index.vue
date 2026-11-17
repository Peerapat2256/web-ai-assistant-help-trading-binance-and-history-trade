<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold">ข่าวคริปโต & วิเคราะห์ผลกระทบด้วย AI</h1>
        <p class="text-sm opacity-60">เลือกเหรียญที่สนใจเพื่อให้ระบบดึงข่าวล่าสุดมาสรุปภาษาไทยและประเมินทิศทางตลาด</p>
      </div>
      <button class="btn btn-sm btn-outline gap-2" @click="loadNews">
        <RefreshCw class="w-4 h-4" />
        รีเฟรชข่าว
      </button>
    </div>

    <div class="flex flex-wrap gap-2">
      <button
        v-for="cat in categories"
        :key="cat.value"
        :class="['btn btn-sm', selectedCategory === cat.value ? 'btn-primary' : 'btn-ghost bg-base-100 border border-base-300']"
        @click="selectCategory(cat.value)"
      >
        {{ cat.label }}
      </button>
    </div>

    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div v-for="n in 4" :key="n" class="card bg-base-100 border border-base-300 p-5 space-y-3">
        <div class="skeleton h-5 w-3/4"></div>
        <div class="skeleton h-4 w-full"></div>
        <div class="skeleton h-4 w-5/6"></div>
        <div class="skeleton h-8 w-1/3"></div>
      </div>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div v-for="(item, idx) in newsList" :key="idx" class="card bg-base-100 border border-base-300 shadow-sm">
        <div class="card-body p-5 flex flex-col justify-between gap-3">
          <div class="space-y-2">
            <div class="flex items-center justify-between gap-2 text-xs">
              <span class="badge badge-outline badge-sm">{{ item.source }}</span>
              <span
                :class="[
                  'badge badge-sm font-semibold',
                  item.sentiment === 'BULLISH'
                    ? 'badge-success'
                    : item.sentiment === 'BEARISH'
                    ? 'badge-error'
                    : 'badge-ghost',
                ]"
              >
                {{
                  item.sentiment === 'BULLISH'
                    ? 'ข่าวดี / มีโอกาสขึ้น'
                    : item.sentiment === 'BEARISH'
                    ? 'ข่าวลบ / ระวังย่อตัว'
                    : 'เป็นกลาง'
                }}
              </span>
            </div>

            <h3 class="font-bold text-base leading-snug">{{ item.title }}</h3>
            <p class="text-sm opacity-85 bg-base-200/60 p-3 rounded-xl leading-relaxed">
              {{ item.summary_thai }}
            </p>
            <div class="text-xs opacity-75">
              <span class="font-semibold text-primary">มุมมอง AI:</span> {{ item.sentiment_reason }}
            </div>
          </div>

          <div class="flex items-center justify-between pt-2 border-t border-base-300 text-xs">
            <span class="opacity-60">{{ item.published_at }}</span>
            <a :href="item.url" target="_blank" rel="noopener noreferrer" class="link link-primary font-medium inline-flex items-center gap-1">
              อ่านต้นฉบับ
              <ExternalLink class="w-3.5 h-3.5" />
            </a>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RefreshCw, ExternalLink } from "lucide-vue-next";
import { useMarketNews } from "~/composables/api/useMarketNews";
import type { NewsItem } from "~/models/news/news_model";

definePageMeta({ middleware: "auth" });

const { fetchNews } = useMarketNews();

const categories = [
  { value: "BTC", label: "Bitcoin (BTC)" },
  { value: "ETH", label: "Ethereum (ETH)" },
  { value: "BNB", label: "BNB" },
  { value: "SOL", label: "Solana (SOL)" },
  { value: "MARKET", label: "ภาพรวมตลาด / Fed & กฎระเบียบ" },
];

const selectedCategory = ref("BTC");
const loading = ref(true);
const newsList = ref<NewsItem[]>([]);

const loadNews = async () => {
  loading.value = true;
  try {
    newsList.value = await fetchNews(selectedCategory.value);
  } finally {
    loading.value = false;
  }
};

const selectCategory = (val: string) => {
  selectedCategory.value = val;
  loadNews();
};

onMounted(() => {
  loadNews();
});
</script>
