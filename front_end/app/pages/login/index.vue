<template>
  <div class="card agy-card agy-fade-up stagger-1 w-full max-w-md bg-base-100/90 border border-base-300 shadow-2xl">
    <div class="card-body p-6 sm:p-8">
      <div class="flex items-center gap-3 mb-2">
        <div class="w-10 h-10 rounded-xl bg-primary text-primary-content flex items-center justify-center font-extrabold text-xl shadow-[0_0_20px_rgba(252,213,53,0.4)]">
          B
        </div>
        <div>
          <h1 class="text-xl font-bold">เข้าสู่ระบบ</h1>
          <p class="text-xs opacity-60">Binance Trading Journal & AI Coach</p>
        </div>
      </div>

      <div v-if="errorMessage" class="alert alert-error text-xs py-2.5 mt-2">
        <span>{{ errorMessage }}</span>
      </div>

      <form @submit.prevent="handleLogin" class="space-y-4 mt-4">
        <label class="form-control w-full">
          <span class="label-text text-xs font-medium mb-1">อีเมล</span>
          <input
            v-model="form.email"
            type="email"
            placeholder="name@example.com"
            class="input input-bordered w-full"
            required
          />
        </label>

        <label class="form-control w-full">
          <span class="label-text text-xs font-medium mb-1">รหัสผ่าน</span>
          <input
            v-model="form.password"
            type="password"
            placeholder="กรอกรหัสผ่าน"
            class="input input-bordered w-full"
            required
          />
        </label>

        <button type="submit" class="btn btn-primary w-full" :disabled="loading">
          {{ loading ? "กำลังเข้าสู่ระบบ..." : "เข้าสู่ระบบ" }}
        </button>
      </form>

      <div class="divider text-xs opacity-50 my-4">หรือ</div>

      <div class="text-center text-sm">
        ยังไม่มีบัญชีใช่ไหม?
        <NuxtLink to="/register" class="link link-primary font-semibold ml-1">
          สมัครสมาชิกใหม่
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useLogin } from "~/composables/api/useLogin";

definePageMeta({ layout: "blank" });

const { login, errorMessage } = useLogin();
const loading = ref(false);
const form = reactive({
  email: "",
  password: "",
});

const handleLogin = async () => {
  loading.value = true;
  try {
    await login({ ...form });
    navigateTo("/dashboard");
  } catch {
    // แสดงข้อความ error ผ่าน errorMessage
  } finally {
    loading.value = false;
  }
};
</script>
