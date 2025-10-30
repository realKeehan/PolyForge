<template lang="pug">
div(class='p-4 sm:p-6 h-full flex flex-col')
  div(class='flex items-center gap-2 sm:gap-3 mb-4')
    svg(class='w-8 h-8 sm:w-10 sm:h-10' viewBox='0 0 40 40' fill='none')
      path(d='M19.6962 23.2867C21.2546 23.2867 22.5178 22.0234 22.5178 20.465C22.5178 18.9066 21.2546 17.6433 19.6962 17.6433C18.1378 17.6433 16.8745 18.9066 16.8745 20.465C16.8745 22.0234 18.1378 23.2867 19.6962 23.2867Z' stroke='#8F00FF' stroke-width='2' stroke-linecap='round' stroke-linejoin='round')
      path(d='M32.395 20.465C29.7305 24.6848 24.9434 28.93 19.6975 28.93C14.4516 28.93 9.66444 24.6848 7 20.465C10.2429 16.4558 14.0424 12 19.6975 12C25.3527 12 29.1522 16.4557 32.395 20.465Z' stroke='#8F00FF' stroke-width='2' stroke-linecap='round' stroke-linejoin='round')
    h2(class='text-white text-lg sm:text-xl font-bold') Choose an Option
  div(class='space-y-3 sm:space-y-6 mb-8')
    OptionButton(v-for='option in options' :key='option.id' :selected='selected === option.id' @click='selectOption(option.id)')
      template(#icon)
        OperationIcon(:variant='option.icon')
      | {{ option.label }}
  FooterButtons(class='mt-auto pt-4' @back='$emit(\'back\')' @next='$emit(\'next\')')
</template>

<script setup lang="ts">
import { computed } from 'vue'

import OptionButton from '../components/buttons/OptionButton.vue'
import FooterButtons from '../components/buttons/FooterButtons.vue'
import OperationIcon from '../components/icons/OperationIcon.vue'

const props = defineProps<{ selected: string }>()
const emit = defineEmits<{ 'update:selected': [string]; back: []; next: [] }>()

const options = [
  { id: 'install', label: 'Install Modpack', icon: 'install' },
  { id: 'update', label: 'Update Modpack', icon: 'update' },
  { id: 'uninstall', label: 'Uninstall Modpack', icon: 'uninstall' },
  { id: 'repair', label: 'Repair Modpack', icon: 'repair' },
] as const

type OptionId = (typeof options)[number]['id']

const selected = computed(() => props.selected)

const selectOption = (option: OptionId) => {
  emit('update:selected', option)
}
</script>
