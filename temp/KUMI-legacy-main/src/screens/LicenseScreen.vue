<template lang="pug">
div(class='p-4 sm:p-6 h-full flex flex-col')
  div(class='flex items-center gap-2 sm:gap-3 mb-4')
    svg(class='w-8 h-8 sm:w-10 sm:h-10' viewBox='0 0 40 40' fill='none')
      path(d='M11 27.4875V12.1859C11 10.9787 11.9787 10 13.1859 10H27.8318C28.194 10 28.4875 10.2936 28.4875 10.6558V24.9893' stroke='#8F00FF' stroke-width='1.5' stroke-linecap='round')
      path(d='M15.3721 10V18.7438L18.1045 16.995L20.8369 18.7438V10' stroke='#8F00FF' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round')
      path(d='M13.186 25.3016H28.4876' stroke='#8F00FF' stroke-width='1.5' stroke-linecap='round')
      path(d='M13.186 29.6735H28.4876' stroke='#8F00FF' stroke-width='1.5' stroke-linecap='round')
      path(d='M13.1859 29.6735C11.9787 29.6735 11 28.6948 11 27.4875C11 26.2802 11.9787 25.3016 13.1859 25.3016' stroke='#8F00FF' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round')
    h2(class='text-white text-lg sm:text-xl font-bold') License Agreement
  div(class='bg-kumi-dark-secondary rounded-[5px] p-3 sm:p-4 h-[250px] sm:h-[275px] overflow-y-auto relative mb-4 pb-12')
    div(class='text-white text-sm sm:text-[15px] leading-normal' ref='licenseContentRef')
      p(class='mb-4') End-User License Agreement for PolyForge.
      p(class='mb-4')
        | PLEASE READ THIS AGREEMENT CAREFULLY. IT CONTAINS IMPORTANT TERMS THAT AFFECT YOU AND YOUR USE OF THE SOFTWARE. BY INSTALLING, COPYING OR USING THE SOFTWARE, YOU AGREE TO BE BOUND BY THE TERMS OF THIS AGREEMENT. IF YOU DO NOT AGREE TO THESE TERMS, DO NOT INSTALL, COPY, OR USE THE SOFTWARE.
      p(class='mb-4')
        | This End-User License Agreement (EULA) is a legal agreement between you--either an individual or a single entity--and the author(s) of this Software for the product identified above, which includes computer software and may include associated media, and online or electronic documentation ("Software").
      p(class='mb-6')
        | By installing, copying, or otherwise using the Software, you agree to be bounded by the terms of this EULA. If you do not agree to the terms of this EULA, do not install or use the Software.
      p(class='mb-2 font-bold') 1. GRANT OF LICENSE.
      p(class='mb-4')
        | This EULA grants you a non-exclusive, non-sublicensable, non-transferable license to install and use the Software. You may install and use an unlimited number of copies of the Software.
      p(class='mb-2 font-bold') 2. COPYRIGHT.
      p(class='mb-4')
        | All title and copyrights in and to the Software (including but not limited to any images, libraries, and examples incorporated into the Software), the accompanying documentation, and any copies of the Software are owned by the author(s) of this Software.
      p(class='mb-2 font-bold') 3. NO WARRANTIES.
      p(class='mb-4')
        | The author(s) of this Software expressly disclaims any warranty for the Software. The Software and any related documentation is provided "as is" without warranty of any kind, either express or implied, including, without limitation, the implied warranties or merchantability, fitness for a particular purpose, or noninfringement. The entire risk arising out of use or performance of the Software remains with you.
      p(class='mb-2 font-bold') 4. NO LIABILITY FOR DAMAGES.
      p(class='mb-4')
        | In no event shall the author(s) of this Software be liable for any special, consequential, incidental or indirect damages whatsoever (including, without limitation, damages for loss of business profits, business interruption, loss of business information, or any other pecuniary loss) arising out of the use of or inability to use this product, even if the author(s) of this Software is aware of the possibility of such damages and known defects.
    div(class='absolute bottom-3 right-3')
      button(class='px-2 py-1 bg-kumi-dark-tertiary/50 text-white text-[15px] rounded-[2.5px]' @click='copyLicense') {{ copied ? 'Copied!' : 'Copy' }}
  label(class='flex items-center gap-3 cursor-pointer mb-8')
    div(:class="['w-6 h-6 rounded-[2.5px] border-2 border-kumi-dark-secondary flex items-center justify-center transition-colors', accepted ? 'bg-kumi-purple' : '']" @click='toggleAccepted')
      svg(v-if='accepted' width='14' height='10' viewBox='0 0 18 14' fill='none')
        path(d='M2 8.084L6.056 12.14L16.196 2' stroke='black' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round')
    span(class='text-white text-[15px]') I accept the license agreement.
  FooterButtons(class='mt-auto pt-4' :show-back='false' :next-enabled='accepted' @next='$emit(\'proceed\')')
</template>

<script setup lang="ts">
import { ref } from 'vue'

import FooterButtons from '../components/buttons/FooterButtons.vue'
import { useClipboard } from '../composables/useClipboard'

const props = defineProps<{ accepted: boolean }>()
const emit = defineEmits<{ 'update:accepted': [boolean]; proceed: [] }>()

const licenseContentRef = ref<HTMLElement | null>(null)
const { copy, copied } = useClipboard()

const toggleAccepted = () => {
  emit('update:accepted', !props.accepted)
}

const copyLicense = async () => {
  const text = licenseContentRef.value?.innerText ?? ''
  await copy(text)
}
</script>
