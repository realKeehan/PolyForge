<template lang="pug">
div(class='flex items-center justify-between')
  SocialButtons
  div(class='flex gap-2')
    button(class='px-4 py-2 rounded-[10px] bg-kumi-dark-tertiary text-white font-rounded text-xl font-bold hover:bg-kumi-dark-tertiary/80 transition-colors' v-if='showBack' @click='$emit(\'back\')') Back
    button(:disabled='!nextEnabled' @click='$emit(\'next\')' :class="['px-4 py-2 rounded-[10px] border-[1.5px] border-kumi-purple text-white font-rounded text-xl font-bold transition-colors', nextEnabled ? 'bg-kumi-purple-dark hover:bg-kumi-purple-bg' : 'bg-kumi-purple-dark/50 cursor-not-allowed']") {{ nextLabel }}
</template>

<script setup lang="ts">
import SocialButtons from '../layout/SocialButtons.vue'

withDefaults(
  defineProps<{
    showBack?: boolean
    nextEnabled?: boolean
    nextLabel?: string
  }>(),
  {
    showBack: true,
    nextEnabled: true,
    nextLabel: 'Next',
  },
)

defineEmits<{
  back: []
  next: []
}>()
</script>

<style scoped lang="scss">
.font-rounded {
  font-family: 'M PLUS Rounded 1c', sans-serif;
}
</style>
