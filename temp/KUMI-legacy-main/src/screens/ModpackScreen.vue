<template lang="pug">
div(class='p-4 sm:p-6 h-full flex flex-col')
  div(class='flex items-center gap-2 sm:gap-3 mb-4')
    svg(class='w-8 h-8 sm:w-10 sm:h-10' viewBox='0 0 40 40' fill='none')
      path(d='M21.1345 11.0384L25.1729 7L32.24 14.0672L28.2016 18.1056M21.1345 11.0384L7.41819 24.7547C7.15043 25.0224 7 25.3856 7 25.7642V32.24H13.4758C13.8545 32.24 14.2176 32.0897 14.4854 31.8218L28.2016 18.1056M21.1345 11.0384L28.2016 18.1056' stroke='#8F00FF' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round')
    h2(class='text-white text-lg sm:text-xl font-bold') Choose Modpack
  div(class='space-y-6 mb-4')
    OptionButton(v-for='modpack in modpacks' :key='modpack.id' :selected='selectedModpack === modpack.id' @click='selectModpack(modpack.id)')
      template(#icon)
        ModpackIcon(:variant='modpack.icon')
      | {{ modpack.label }}
  label(class='flex items-center gap-3 cursor-pointer mb-8')
    div(:class="['w-6 h-6 rounded-[2.5px] border-2 border-kumi-dark-secondary flex items-center justify-center transition-colors', cleanInstall ? 'bg-kumi-purple' : '']" @click='toggleCleanInstall')
      svg(v-if='cleanInstall' width='14' height='10' viewBox='0 0 18 14' fill='none')
        path(d='M2 8.084L6.056 12.14L16.196 2' stroke='black' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round')
    span(class='text-white text-[15px]') Clean Install
  FooterButtons(class='mt-auto pt-4' @back='$emit(\'back\')' @next='$emit(\'next\')')
</template>

<script setup lang="ts">
import { computed } from 'vue'

import FooterButtons from '../components/buttons/FooterButtons.vue'
import OptionButton from '../components/buttons/OptionButton.vue'
import ModpackIcon from '../components/icons/ModpackIcon.vue'

const props = defineProps<{ selectedModpack: string; cleanInstall: boolean }>()
const emit = defineEmits<{
  'update:selectedModpack': [string]
  'update:cleanInstall': [boolean]
  back: []
  next: []
}>()

const modpacks = [
  { id: 'turtel', label: 'Turtel SMP', icon: 'turtel' },
  { id: 'event', label: 'Event Pack', icon: 'event' },
] as const

type ModpackId = (typeof modpacks)[number]['id']

const selectedModpack = computed(() => props.selectedModpack)
const cleanInstall = computed(() => props.cleanInstall)

const selectModpack = (id: ModpackId) => {
  emit('update:selectedModpack', id)
}

const toggleCleanInstall = () => {
  emit('update:cleanInstall', !props.cleanInstall)
}
</script>
