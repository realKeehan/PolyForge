<template lang="pug">
button(:class="['flex items-center gap-4 min-h-[56px] w-full rounded-[10px] px-3 transition-all', selected ? 'bg-kumi-purple-bg border-[1.5px] border-kumi-purple' : 'bg-kumi-dark-secondary hover:bg-kumi-dark-secondary/80']" @click="$emit('click')")
  div(class='w-[56px] h-[56px] flex items-center justify-center flex-shrink-0 ml-0')
    svg(class='w-10 h-10 text-white' v-if='icon' viewBox='0 0 48 48' fill='none')
      template(v-if="icon === 'vanilla'")
        path(d='M12 18L24 12L36 18V30L24 36L12 30V18Z' stroke='currentColor' stroke-width='2.5' stroke-linejoin='round')
        path(d='M12 18L24 24L36 18' stroke='currentColor' stroke-width='2.5' stroke-linejoin='round')
        path(d='M24 24V36' stroke='currentColor' stroke-width='2.5' stroke-linecap='round')
      template(v-else-if="icon === 'multimc'")
        rect(x='10' y='10' width='10' height='10' rx='2' stroke='currentColor' stroke-width='2.5')
        rect(x='18' y='18' width='14' height='14' rx='3' stroke='currentColor' stroke-width='2.5')
        circle(cx='34' cy='14' r='4' stroke='currentColor' stroke-width='2.5')
      template(v-else-if="icon === 'curseforge'")
        path(d='M14 30H20L22 26H30L34 20H18L14 14' stroke='currentColor' stroke-width='2.5' stroke-linejoin='round' stroke-linecap='round')
        path(d='M22 12L22 18' stroke='currentColor' stroke-width='2.5' stroke-linecap='round')
      template(v-else-if="icon === 'modrinth'")
        path(d='M14 32V16H19L24 24L29 16H34V32' stroke='currentColor' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round')
        path(d='M34 32L29 24L24 32L19 24L14 32' stroke='currentColor' stroke-width='2.5' stroke-linecap='round')
      template(v-else-if="icon === 'prism'")
        path(d='M24 10L38 34H10L24 10Z' stroke='currentColor' stroke-width='2.5' stroke-linejoin='round')
        path(d='M24 16L32 30H16L24 16Z' stroke='currentColor' stroke-width='2.5' stroke-linejoin='round')
      template(v-else-if="icon === 'gdlauncher'")
        path(d='M18 34L22 24L16 22L24 12L32 20L26 22L30 30L18 34Z' stroke='currentColor' stroke-width='2.5' stroke-linejoin='round' stroke-linecap='round')
        circle(cx='24' cy='24' r='3' stroke='currentColor' stroke-width='2.5')
      template(v-else-if="icon === 'atlauncher'")
        path(d='M24 12L36 30H28L24 36L20 30H12L24 12Z' stroke='currentColor' stroke-width='2.5' stroke-linejoin='round')
        path(d='M24 36V30' stroke='currentColor' stroke-width='2.5' stroke-linecap='round')
      template(v-else-if="icon === 'bakaxl'")
        path(d='M16 12H26C31 12 34 16 34 20C34 24 31 28 26 28H16V12Z' stroke='currentColor' stroke-width='2.5' stroke-linejoin='round')
        path(d='M16 28H28C32 28 36 31 36 36H20' stroke='currentColor' stroke-width='2.5' stroke-linecap='round')
      template(v-else-if="icon === 'feather'")
        path(d='M32 12C25 14 16 22 16 32C22 30 28 24 28 24C26 30 22 32 20 34C24 36 30 34 34 28C36 24 36 16 32 12Z' stroke='currentColor' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round')
      template(v-else-if="icon === 'technic'")
        circle(cx='20' cy='20' r='7' stroke='currentColor' stroke-width='2.5')
        path(d='M20 12L22 16L26 18L22 20L20 24L18 20L14 18L18 16L20 12Z' stroke='currentColor' stroke-width='2.5' stroke-linejoin='round')
        rect(x='26' y='22' width='8' height='12' rx='2' stroke='currentColor' stroke-width='2.5')
      template(v-else-if="icon === 'polymc'")
        path(d='M24 10L36 18V32L24 40L12 32V18L24 10Z' stroke='currentColor' stroke-width='2.5' stroke-linejoin='round')
        path(d='M24 16L30 20V26L24 30L18 26V20L24 16Z' stroke='currentColor' stroke-width='2.5' stroke-linejoin='round')
      template(v-else-if="icon === 'custom'")
        path(d='M14 16H28L32 20V34H14V16Z' stroke='currentColor' stroke-width='2.5' stroke-linejoin='round')
        path(d='M28 16V22H34' stroke='currentColor' stroke-width='2.5' stroke-linecap='round')
      template(v-else-if="icon === 'manual'")
        path(d='M16 12H32V36H16C13.7909 36 12 34.2091 12 32V16C12 13.7909 13.7909 12 16 12Z' stroke='currentColor' stroke-width='2.5' stroke-linejoin='round')
        path(d='M16 16H30M16 24H30M16 32H26' stroke='currentColor' stroke-width='2.5' stroke-linecap='round')
      template(v-else)
        rect(x='12' y='14' width='24' height='20' rx='4' stroke='currentColor' stroke-width='2.5')
  span(class='text-white text-xl font-mono flex-1 text-left')
    slot
</template>

<script setup lang="ts">
defineProps<{
  selected?: boolean
  icon?: string
}>()

defineEmits<{
  click: []
}>()
</script>
