<template>
  <div class="min-h-screen relative flex items-center justify-center bg-base-200 agy-bg-grid overflow-hidden p-4">
    <div class="agy-orb agy-orb-gold"></div>
    <div class="agy-orb agy-orb-emerald"></div>

    <div class="relative z-10 w-full flex items-center justify-center">
      <div class="relative w-full max-w-md">
        <div
          class="hidden sm:block absolute -inset-6 rounded-full border border-primary/10 agy-ring-orbit pointer-events-none"
        >
          <span class="w-2.5 h-2.5 rounded-full bg-primary/70 absolute -top-1 left-1/2 -translate-x-1/2 shadow-[0_0_12px_rgba(252,213,53,0.8)]"></span>
        </div>
        <slot />
      </div>
    </div>
  </div>
</template>
