<template lang="pug">
div(class='p-4 sm:p-6 h-full flex flex-col')
  div(class='flex items-center gap-2 sm:gap-3 mb-4')
    svg(class='w-8 h-8 sm:w-10 sm:h-10' viewBox='0 0 40 40' fill='none')
      path(d='M11 27.4875V12.1859C11 10.9787 11.9787 10 13.1859 10H27.8318C28.194 10 28.4875 10.2936 28.4875 10.6558V24.9893' stroke='#8F00FF' stroke-width='1.5' stroke-linecap='round')
      path(d='M15.3721 10V18.7438L18.1045 16.995L20.8369 18.7438V10' stroke='#8F00FF' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round')
      path(d='M13.1855 25.3016H28.4871' stroke='#8F00FF' stroke-width='1.5' stroke-linecap='round')
      path(d='M13.1855 29.6735H28.4871' stroke='#8F00FF' stroke-width='1.5' stroke-linecap='round')
      path(d='M13.1859 29.6735C11.9787 29.6735 11 28.6948 11 27.4875C11 26.2802 11.9787 25.3016 13.1859 25.3016' stroke='#8F00FF' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round')
    h2(class='text-white text-lg sm:text-xl font-bold') Status Log
  div(class='bg-kumi-dark-secondary rounded-[5px] p-4 flex-1 overflow-y-auto relative mb-4 pb-12')
    pre(class='text-white text-[15px] font-mono leading-normal whitespace-pre-wrap break-words') {{ installLog }}
    div(class='absolute bottom-3 right-3')
      button(class='px-2 py-1 bg-kumi-dark-tertiary/50 text-white text-[15px] rounded-[2.5px]' @click='copyLog') {{ copied ? 'Copied!' : 'Copy' }}
  div(class='mt-auto flex items-center justify-between pt-4')
    SocialButtons
    div(class='flex gap-2')
      button(class='px-4 py-2 rounded-[10px] bg-kumi-dark-tertiary text-white font-rounded text-xl font-bold hover:bg-kumi-dark-tertiary/80 transition-colors' @click='$emit(\'back\')') Back
      button(class='px-4 py-2 rounded-[10px] border-[1.5px] border-kumi-purple bg-kumi-purple-dark text-white font-rounded text-xl font-bold hover:bg-kumi-purple-bg transition-colors' @click='$emit(\'close\')') {{ installComplete ? 'Close' : 'Installing...' }}
</template>

<script setup lang="ts">
import { computed } from 'vue'

import SocialButtons from '../components/layout/SocialButtons.vue'
import { useClipboard } from '../composables/useClipboard'

const props = defineProps<{ installLog: string; installComplete: boolean }>()

const installLog = computed(() => props.installLog)
const installComplete = computed(() => props.installComplete)
const { copy, copied } = useClipboard()

const copyLog = async () => {
  await copy(installLog.value)
}
</script>
