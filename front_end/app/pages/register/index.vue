<template>
  <div class="card agy-card agy-fade-up stagger-1 w-full max-w-md bg-base-100/90 border border-base-300 shadow-2xl">
    <div class="card-body p-6 sm:p-8">
      <div class="flex items-center gap-3 mb-2">
        <div class="w-10 h-10 rounded-xl bg-primary text-primary-content flex items-center justify-center font-extrabold text-xl shadow-[0_0_20px_rgba(252,213,53,0.4)]">
          B
        </div>
        <div>
          <h1 class="text-xl font-bold">สมัครสมาชิกใหม่</h1>
          <p class="text-xs opacity-60">เริ่มต้นบันทึกสถิติการเทรดของคุณฟรี</p>
        </div>
      </div>

      <div v-if="errorMessage" class="alert alert-error text-xs py-2.5 mt-2">
        <span>{{ errorMessage }}</span>
      </div>

      <form @submit.prevent="handleRegister" class="space-y-3.5 mt-4">
        <label class="form-control w-full">
          <span class="label-text text-xs font-medium mb-1">ชื่อ-นามสกุล / ชื่อเล่น</span>
          <input
            v-model="form.fullname"
            type="text"
            placeholder="เช่น สมชาย นักเทรด"
            class="input input-bordered w-full"
            required
          />
        </label>

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
          <span class="label-text text-xs font-medium mb-1">เบอร์โทรศัพท์ (ไม่บังคับ)</span>
          <input
            v-model="form.phone_number"
            type="text"
            placeholder="08xxxxxxxx"
            class="input input-bordered w-full"
          />
        </label>

        <label class="form-control w-full">
          <span class="label-text text-xs font-medium mb-1">รหัสผ่าน (อย่างน้อย 6 ตัวอักษร)</span>
          <input
            v-model="form.password"
            type="password"
            minlength="6"
            placeholder="ตั้งรหัสผ่าน"
            class="input input-bordered w-full"
            required
          />
        </label>

        <button type="submit" class="btn btn-primary w-full mt-2" :disabled="loading">
          {{ loading ? "กำลังสร้างบัญชี..." : "สมัครสมาชิก" }}
        </button>
      </form>

      <div class="divider text-xs opacity-50 my-4">มีบัญชีอยู่แล้ว?</div>

      <div class="text-center text-sm">
        <NuxtLink to="/login" class="link link-primary font-semibold">
          กลับไปหน้าเข้าสู่ระบบ
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useLogin } from "~/composables/api/useLogin";

definePageMeta({ layout: "blank" });

const { register, errorMessage } = useLogin();
const loading = ref(false);
const form = reactive({
  fullname: "",
  email: "",
  phone_number: "",
  password: "",
});

const handleRegister = async () => {
  loading.value = true;
  try {
    await register({ ...form });
    navigateTo("/dashboard");
  } catch {
    // แสดงข้อความ error ผ่าน errorMessage
  } finally {
    loading.value = false;
  }
};
</script>
