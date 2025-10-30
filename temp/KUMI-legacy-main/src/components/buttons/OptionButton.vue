<template lang="pug">
button(:class="['flex items-center gap-4 min-h-[56px] w-full rounded-[10px] transition-all px-3', selected ? 'bg-kumi-purple-bg border-[1.5px] border-kumi-purple' : 'bg-kumi-dark-secondary hover:bg-kumi-dark-secondary/80']" @click="$emit('click')")
  div(class='w-[56px] h-[56px] flex items-center justify-center flex-shrink-0')
    slot(name='icon')
  span(class='text-white text-xl font-mono flex-1 text-left')
    slot
</template>

<script setup lang="ts">
defineProps<{
  selected?: boolean
}>()

defineEmits<{
  click: []
}>()
</script>
