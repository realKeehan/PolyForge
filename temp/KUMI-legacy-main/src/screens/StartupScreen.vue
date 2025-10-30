<template lang="pug">
div(class='absolute inset-0 flex items-center justify-center')
  img(class='w-[90%] max-w-[675px] h-auto' :src='imageSrc' alt='PolyForge Splash')
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{ imageSrc?: string }>(),
  {
    imageSrc: 'https://api.builder.io/api/v1/image/assets/TEMP/34f84eb8ad02be680cc97170a5f6814628582876?width=1350',
  },
)
</script>
