<template lang="pug">
svg(v-if="variant === 'turtel'" class='w-12 h-12' viewBox='0 0 52 52' fill='none')
  ellipse(cx='26' cy='26' rx='14' ry='11' fill='#10b981')
  path(d='M18 26C18 21 22 18 26 18C30 18 34 21 34 26C34 31 30 34 26 34C22 34 18 31 18 26Z' fill='#059669')
  circle(cx='36' cy='24' r='4' fill='#10b981')
  circle(cx='14' cy='24' r='4' fill='#10b981')
  circle(cx='26' cy='24' r='2' fill='white')
  circle(cx='26' cy='24' r='1' fill='#0f172a')
  rect(x='21' y='28' width='10' height='3' rx='1.5' fill='#ecfeff' opacity='0.75')
svg(v-else-if="variant === 'event'" class='w-12 h-12' viewBox='0 0 52 52' fill='none')
  path(d='M26 12L29.6 21.2L39.5 21.9L32 28.1L34.6 37.7L26 32.6L17.4 37.7L20 28.1L12.5 21.9L22.4 21.2L26 12Z' fill='#fbbf24')
  circle(cx='16' cy='14' r='3' fill='#f97316')
  circle(cx='36' cy='16' r='2.5' fill='#60a5fa')
  rect(x='22' y='38' width='8' height='3' rx='1.5' fill='#fbbf24' opacity='0.7')
</template>

<script setup lang="ts">
withDefaults(defineProps<{ variant?: 'turtel' | 'event' }>(), {
  variant: 'turtel',
})
</script>
