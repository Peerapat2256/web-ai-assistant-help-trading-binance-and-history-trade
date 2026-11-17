<template>
  <div class="min-h-screen flex items-center justify-center">
    <div class="skeleton w-48 h-8 rounded-lg"></div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: "blank" });

onMounted(() => {
  const token = useCookie("token");
  if (token.value) {
    navigateTo("/dashboard");
  } else {
    navigateTo("/login");
  }
});
</script>
