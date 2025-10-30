<template lang="pug">
div(class='app-root min-h-screen bg-kumi-dark flex items-center justify-center p-2 sm:p-4 font-mono')
  div(class='app-window w-full max-w-[750px] rounded-[10px] bg-kumi-dark shadow-2xl overflow-hidden scale-75 sm:scale-90 md:scale-100 origin-center')
    AppHeader(@close='handleClose')
    div(class='app-content relative min-h-[400px] sm:h-[450px] bg-topo bg-cover bg-center')
      StartupScreen(v-if="currentScreen === 'startup'")
      LicenseScreen(
        v-else-if="currentScreen === 'license'"
        :accepted='acceptedLicense'
        @update:accepted='acceptedLicense = $event'
        @proceed="goToScreen('options')"
      )
      OptionsScreen(
        v-else-if="currentScreen === 'options'"
        :selected='selectedOption'
        @update:selected='selectedOption = $event'
        @back="goToScreen('license')"
        @next="goToScreen('modpack')"
      )
      ModpackScreen(
        v-else-if="currentScreen === 'modpack'"
        :selected-modpack='selectedModpack'
        :clean-install='cleanInstall'
        @update:selectedModpack='selectedModpack = $event'
        @update:cleanInstall='cleanInstall = $event'
        @back="goToScreen('options')"
        @next="goToScreen('launcher')"
      )
      LauncherScreen(
        v-else-if="currentScreen === 'launcher'"
        :selected-launcher='selectedLauncher'
        :launchers='launchers'
        @update:selectedLauncher='selectedLauncher = $event'
        @back="goToScreen('modpack')"
        @install='startInstallation'
      )
      StatusScreen(
        v-else-if="currentScreen === 'status'"
        :install-log='installLog'
        :install-complete='installComplete'
        @back="goToScreen('launcher')"
        @close='handleClose'
      )
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

import AppHeader from './components/layout/AppHeader.vue'
import { launcherChoices, type LauncherChoice } from './data/launcherOptions'
import LauncherScreen from './screens/LauncherScreen.vue'
import LicenseScreen from './screens/LicenseScreen.vue'
import ModpackScreen from './screens/ModpackScreen.vue'
import OptionsScreen from './screens/OptionsScreen.vue'
import StartupScreen from './screens/StartupScreen.vue'
import StatusScreen from './screens/StatusScreen.vue'

type Screen = 'startup' | 'license' | 'options' | 'modpack' | 'launcher' | 'status'
type OptionId = 'install' | 'update' | 'uninstall' | 'repair'
type ModpackId = 'turtel' | 'event'
type LauncherId = LauncherChoice['id']

type WailsRuntime = {
  Quit?: () => void
  EventsEmit?: (event: string, ...args: unknown[]) => void
}

type WailsWindow = typeof window & { runtime?: WailsRuntime }

const currentScreen = ref<Screen>('startup')
const acceptedLicense = ref(false)
const selectedOption = ref<OptionId>('install')
const selectedModpack = ref<ModpackId>('turtel')
const selectedLauncher = ref<LauncherId>('vanilla')
const cleanInstall = ref(true)
const installLog = ref('')
const installComplete = ref(false)
const launchers = launcherChoices

let startupTimer: ReturnType<typeof setTimeout> | undefined
let installTimer: ReturnType<typeof setInterval> | undefined

const getRuntime = (): WailsRuntime | undefined => {
  if (typeof window === 'undefined') {
    return undefined
  }

  return (window as WailsWindow).runtime
}

const stopInstallationTimer = () => {
  if (installTimer !== undefined) {
    clearInterval(installTimer)
    installTimer = undefined
  }
}

const goToScreen = (screen: Screen) => {
  if (screen !== 'status') {
    stopInstallationTimer()
    installComplete.value = false
  }

  currentScreen.value = screen
  getRuntime()?.EventsEmit?.('screen-change', screen)
}

const handleClose = () => {
  const runtime = getRuntime()

  if (runtime?.Quit) {
    runtime.Quit()
    return
  }

  if (typeof window !== 'undefined' && typeof window.close === 'function') {
    window.close()
  } else {
    console.info('Close requested but no runtime was available.')
  }
}

const startInstallation = () => {
  goToScreen('status')
  installLog.value = 'Starting Install...\n\n'
  installComplete.value = false
  stopInstallationTimer()

  const selectedLauncherName =
    launchers.find((launcher) => launcher.id === selectedLauncher.value)?.name ?? selectedLauncher.value

  const logLines = [
    `Operation: ${selectedOption.value}`,
    `Modpack: ${selectedModpack.value}${cleanInstall.value ? ' (clean install)' : ''}`,
    `Launcher: ${selectedLauncherName}`,
    '',
    'Creating required directories...',
    '✅ Directory exists: C:\\Users\\USER\\AppData\\Roaming\\BetterMinecraft',
    '✅ Directory exists: C:\\Users\\USER\\AppData\\Roaming\\BetterMinecraft\\data',
    '✅ Directory exists: C:\\Users\\USER\\AppData\\Roaming\\BetterMinecraft\\themes',
    '✅ Directory exists: C:\\Users\\USER\\AppData\\Roaming\\BetterMinecraft\\plugins',
    '✅ Directories verified',
    '',
    'Downloading asar file',
    '✅ Downloaded BetterMinecraft package from the official source',
    '✅ Package checksum verified',
    '',
    'Injecting shims...',
    'Injecting into: C:\\Users\\USER\\AppData\\Local\\Minecraft\\app-1.0.9211\\modules\\Minecraft_desktop_core-1\\Minecraft_desktop_core',
    '✅ Injection successful',
    '',
    'Restarting Minecraft...',
    'Attempting to close running instances',
    '✅ Minecraft not running',
    '✅ Restart command queued',
    '',
    'Install completed!',
  ]

  let index = 0

  getRuntime()?.EventsEmit?.('install-started', {
    option: selectedOption.value,
    modpack: selectedModpack.value,
    launcher: selectedLauncher.value,
    cleanInstall: cleanInstall.value,
  })

  installTimer = setInterval(() => {
    if (index < logLines.length) {
      installLog.value += `${logLines[index]}\n`
      index += 1
      return
    }

    stopInstallationTimer()
    installComplete.value = true
    getRuntime()?.EventsEmit?.('install-complete', {
      option: selectedOption.value,
      modpack: selectedModpack.value,
      launcher: selectedLauncher.value,
      cleanInstall: cleanInstall.value,
    })
  }, 300)
}

onMounted(() => {
  startupTimer = setTimeout(() => {
    goToScreen('license')
  }, 2000)
})

onBeforeUnmount(() => {
  if (startupTimer !== undefined) {
    clearTimeout(startupTimer)
  }

  stopInstallationTimer()
})
</script>

<style scoped lang="scss">
.font-rounded {
  font-family: 'M PLUS Rounded 1c', sans-serif;
}

.app-root {
  -webkit-font-smoothing: antialiased;
}

.app-window {
  transition: transform 0.2s ease;
}
</style>
