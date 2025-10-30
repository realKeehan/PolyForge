<template lang="pug">
div(class='flex gap-4 items-center')
  button(
    type='button'
    class='px-5 h-14 rounded-full bg-kumi-dark-secondary/90 text-white/90 flex items-center gap-2 hover:bg-kumi-dark-secondary transition-colors'
    @click='showAbout'
  )
    svg(class='w-7 h-7 shrink-0' viewBox='0 0 40 40' fill='none' aria-hidden='true')
      circle(cx='20' cy='20' r='17' stroke='currentColor' stroke-width='3')
      path(d='M20 28V18' stroke='currentColor' stroke-width='3' stroke-linecap='round')
      circle(cx='20' cy='13' r='2.5' fill='currentColor')
    span(class='text-sm font-semibold uppercase tracking-wide') About
  a(class='w-14 h-14 rounded-full bg-kumi-dark-secondary/90 text-white/90 flex items-center justify-center hover:bg-kumi-dark-secondary transition-colors' href='#' aria-label='Website')
    svg(class='w-8 h-8' viewBox='0 0 40 40' fill='none' aria-hidden='true')
      g(clip-path='url(#clip0)')
        path(d='M11 19.75C11 24.5824 14.9175 28.5 19.75 28.5C24.5824 28.5 28.5 24.5824 28.5 19.75C28.5 14.9175 24.5824 11 19.75 11C14.9175 11 11 14.9175 11 19.75Z' stroke='currentColor' stroke-width='3' stroke-linecap='round' stroke-linejoin='round')
        path(d='M20.625 11.043C20.625 11.043 23.25 14.4998 23.25 19.7498C23.25 24.9998 20.625 28.4567 20.625 28.4567' stroke='currentColor' stroke-width='3' stroke-linecap='round' stroke-linejoin='round')
        path(d='M18.875 28.4567C18.875 28.4567 16.25 24.9998 16.25 19.7498C16.25 14.4998 18.875 11.043 18.875 11.043' stroke='currentColor' stroke-width='3' stroke-linecap='round' stroke-linejoin='round')
        path(d='M11.5508 22.8126H27.9489' stroke='currentColor' stroke-width='3' stroke-linecap='round' stroke-linejoin='round')
        path(d='M11.5508 16.6874H27.9489' stroke='currentColor' stroke-width='3' stroke-linecap='round' stroke-linejoin='round')
      defs
        clipPath(id='clip0')
          rect(width='40' height='40' fill='white')
  a(class='w-14 h-14 rounded-full bg-kumi-dark-secondary/90 text-white/90 flex items-center justify-center hover:bg-kumi-dark-secondary transition-colors' href='#' aria-label='GitHub')
    svg(class='w-8 h-8' viewBox='0 0 40 40' fill='none' aria-hidden='true')
      g(clip-path='url(#clip1)')
        path(d='M23.5642 28.5V25.9921C23.5979 25.5755 23.5401 25.1566 23.3946 24.7633C23.2491 24.3701 23.0194 24.0115 22.7206 23.7114C25.5386 23.4056 28.5001 22.3657 28.5001 17.5946C28.4998 16.3746 28.0179 15.2014 27.154 14.3177C27.563 13.2504 27.5341 12.0706 27.0732 11.0234C27.0732 11.0234 26.0142 10.7176 23.5642 12.3167C21.5073 11.7739 19.3391 11.7739 17.2822 12.3167C14.8322 10.7176 13.7732 11.0234 13.7732 11.0234C13.3122 12.0706 13.2833 13.2504 13.6924 14.3177C12.822 15.2079 12.3395 16.3918 12.3463 17.6208C12.3463 22.357 15.3078 23.3968 18.1257 23.7376C17.8305 24.0347 17.6028 24.389 17.4574 24.7774C17.3121 25.1658 17.2524 25.5798 17.2822 25.9921V28.5' stroke='currentColor' stroke-width='3' stroke-linecap='round' stroke-linejoin='round')
        path(d='M17.2821 26.7523C14.5897 27.6027 12.3462 26.7523 11 24.1308' stroke='currentColor' stroke-width='3' stroke-linecap='round' stroke-linejoin='round')
      defs
        clipPath(id='clip1')
          rect(width='40' height='40' fill='white')
  a(class='w-14 h-14 rounded-full bg-kumi-dark-secondary/90 text-white/90 flex items-center justify-center hover:bg-kumi-dark-secondary transition-colors' href='#' aria-label='Support')
    svg(class='w-8 h-8' viewBox='0 0 40 40' fill='none' aria-hidden='true')
      g(clip-path='url(#clip2)')
        path(d='M28.5 16.6994C28.5 18.2029 27.9804 19.647 27.0526 20.7153C24.9168 23.1751 22.8452 25.7402 20.6296 28.1108C20.1218 28.6463 19.3162 28.6268 18.8302 28.0671L12.447 20.7153C10.5177 18.4931 10.5177 14.9056 12.447 12.6835C14.3954 10.4395 17.5694 10.4395 19.5178 12.6835L19.7498 12.9507L19.9817 12.6836C20.9158 11.6072 22.1881 11 23.5171 11C24.8462 11 26.1183 11.6071 27.0526 12.6835C27.9805 13.7518 28.5 15.1959 28.5 16.6994Z' stroke='currentColor' stroke-width='3' stroke-linejoin='round')
      defs
        clipPath(id='clip2')
          rect(width='40' height='40' fill='white')
</template>

<script setup lang="ts">
const showAbout = () => {
  alert('PolyForge Launcher\nBuild 5.0.0')
}
</script>
