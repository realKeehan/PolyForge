export type LauncherIconKey =
  | 'vanilla'
  | 'multimc'
  | 'curseforge'
  | 'modrinth'
  | 'gdlauncher'
  | 'atlauncher'
  | 'prism'
  | 'bakaxl'
  | 'feather'
  | 'technic'
  | 'polymc'
  | 'custom'
  | 'manual'

export interface LauncherChoice {
  id: string
  name: string
  icon: LauncherIconKey
}

export const launcherChoices: LauncherChoice[] = [
  { id: 'vanilla', name: 'Vanilla Launcher', icon: 'vanilla' },
  { id: 'multimc', name: 'MultiMC', icon: 'multimc' },
  { id: 'curseforge', name: 'CurseForge', icon: 'curseforge' },
  { id: 'modrinth', name: 'Modrinth', icon: 'modrinth' },
  { id: 'gdlauncher', name: 'GD Launcher', icon: 'gdlauncher' },
  { id: 'atlauncher', name: 'AT Launcher', icon: 'atlauncher' },
  { id: 'prism', name: 'Prism Launcher', icon: 'prism' },
  { id: 'bakaxl', name: 'BakaXL', icon: 'bakaxl' },
  { id: 'feather', name: 'Feather', icon: 'feather' },
  { id: 'technic', name: 'Technic Launcher', icon: 'technic' },
  { id: 'polymc', name: 'PolyMC', icon: 'polymc' },
  { id: 'custom', name: 'Custom Path', icon: 'custom' },
  { id: 'manual', name: 'Manual Install', icon: 'manual' },
]
