<template lang="pug">
svg(v-if="variant === 'install'" class='w-12 h-12 text-white' viewBox='0 0 52 52' fill='none')
  rect(x='14' y='36' width='24' height='6' rx='3' fill='currentColor' fill-opacity='0.35')
  path(d='M26 12V30' stroke='currentColor' stroke-width='3.5' stroke-linecap='round')
  path(d='M20 24L26 30L32 24' stroke='currentColor' stroke-width='3.5' stroke-linecap='round' stroke-linejoin='round')
svg(v-else-if="variant === 'update'" class='w-12 h-12 text-white' viewBox='0 0 52 52' fill='none')
  path(d='M35 19C32.5 15 28.2 12.5 23.4 13C16.6 13.7 11 19.7 11 26.6C11 34.2 17.4 40.2 25 40C29.7 39.8 33.8 37.2 36 33' stroke='currentColor' stroke-width='3' stroke-linecap='round' stroke-linejoin='round')
  path(d='M31 18H36C36.5523 18 37 17.5523 37 17V12' stroke='currentColor' stroke-width='3' stroke-linecap='round' stroke-linejoin='round')
  path(d='M21 34H16C15.4477 34 15 34.4477 15 35V40' stroke='currentColor' stroke-width='3' stroke-linecap='round' stroke-linejoin='round')
svg(v-else-if="variant === 'uninstall'" class='w-12 h-12 text-white' viewBox='0 0 52 52' fill='none')
  path(d='M18 18H34' stroke='currentColor' stroke-width='3' stroke-linecap='round')
  path(d='M32 18L30.5 38C30.2 41.4 27.4 44 24 44C20.6 44 17.8 41.4 17.5 38L16 18' stroke='currentColor' stroke-width='3' stroke-linecap='round' stroke-linejoin='round')
  path(d='M21 18V14C21 12.8954 21.8954 12 23 12H27C28.1046 12 29 12.8954 29 14V18' stroke='currentColor' stroke-width='3' stroke-linecap='round' stroke-linejoin='round')
  path(d='M22 24V34' stroke='currentColor' stroke-width='3' stroke-linecap='round')
  path(d='M28 24V34' stroke='currentColor' stroke-width='3' stroke-linecap='round')
svg(v-else-if="variant === 'repair'" class='w-12 h-12 text-white' viewBox='0 0 52 52' fill='none')
  path(d='M20 14L24 18L21 21L25 25L22 28L18 24L15 27L12 24L18 18C18 18 16 14 20 14Z' fill='currentColor' fill-opacity='0.45')
  path(d='M32 18C34.8 18 37 20.2 37 23C37 25 35.8 26.7 34 27.5L31 30.5L28 27.5L31 24.5V22' stroke='currentColor' stroke-width='3' stroke-linecap='round' stroke-linejoin='round')
  path(d='M28 32L24 36L20 32L24 28L28 32Z' fill='currentColor')
</template>

<script setup lang="ts">
withDefaults(defineProps<{ variant?: 'install' | 'update' | 'uninstall' | 'repair' }>(), {
  variant: 'install',
})
</script>
