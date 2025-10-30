<template lang="pug">
div(class='h-[35px] bg-kumi-dark flex items-center justify-between px-2')
  div(class='flex items-center gap-1 sm:gap-2')
    img(class='w-5 h-5 sm:w-6 sm:h-6' :src='logoSrc' alt='PolyForge Logo')
    span(class='text-kumi-gray font-bold text-base sm:text-lg') {{ title }}
  div(class='flex items-center gap-4')
    button(class='text-kumi-gray hover:text-white transition-colors')
      svg(width='26' height='2' viewBox='0 0 26 2' fill='none')
        path(d='M0 1H26' stroke='currentColor' stroke-width='2' stroke-linecap='round')
    button(class='text-kumi-gray hover:text-white transition-colors' @click='$emit(\'close\')')
      svg(width='14' height='14' viewBox='0 0 14 14' fill='none')
        path(d='M1 13L7 7M7 7L13 1M7 7L1 1M7 7L13 13' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round')
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{ title?: string; logoSrc?: string }>(),
  {
    title: 'PolyForge v5.0.0',
    logoSrc: 'https://api.builder.io/api/v1/image/assets/TEMP/b191cf027cd26eecff909f4ae5e957fa0a990810?width=50',
  },
)

defineEmits<{ close: [] }>()
</script>
