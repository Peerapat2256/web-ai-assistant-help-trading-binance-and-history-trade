<template>
  <div :data-theme="theme" class="min-h-screen bg-base-200 text-base-content">
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>

<script setup lang="ts">
const theme = useState<string>("app-theme", () => "binancedark");

onMounted(() => {
  const saved = localStorage.getItem("app-theme");
  if (saved === "binancedark" || saved === "binancelight") {
    theme.value = saved;
  }
});
</script>
