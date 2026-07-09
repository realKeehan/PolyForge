export interface OptionDescriptor {
  id: string;
  title: string;
  description: string;
  requiresPath?: boolean;
  pathLabel?: string;
  detectedPath?: string;
  found?: boolean;
  /** Reference text shown behind an ⓘ icon on the row (may be multi-line). */
  info?: string;
}

export interface ExecutionPayload {
  path?: string;
  extra?: Record<string, string>;
}

export type LogLevel = 'info' | 'warning' | 'error';

export interface LogEntry {
  level: LogLevel;
  message: string;
}

export interface ActionResult {
  success: boolean;
  messages: LogEntry[];
  timestamp?: string;
}

export enum Step {
  Loading = 'loading',
  Startup = 'startup',
  License = 'license',
  Mode = 'mode',
  Modpack = 'modpack',
  Installer = 'installer',
  Status = 'status',
}

export type Mode = 'install' | 'repair' | 'update' | 'uninstall';

// ── Remote content manifest (fetched from polyforge.dev) ──

export interface RemotePack {
  id: string;
  name: string;
  description?: string;
  requiresPassword?: boolean;
  passwordHash?: string;
  /** Mod filenames the app self-destructs from an existing install next launch. */
  removeMods?: string[];
}

export interface RemoteOptionOverride {
  id: string;
  title?: string;
  description?: string;
}

export interface RemoteAppInfo {
  latestVersion: string;
  minSupportedVersion?: string;
  downloadUrl?: string;
  notes?: string;
}

export interface RemoteManifest {
  schemaVersion: number;
  updated?: string;
  app: RemoteAppInfo;
  modpacks?: RemotePack[];
  optionOverrides?: RemoteOptionOverride[];
  disabledOptions?: string[];
}

export interface RemoteContentResult {
  manifest?: RemoteManifest | null;
  fromCache: boolean;
  updateAvailable: boolean;
  mandatory: boolean;
  currentVersion: string;
  error?: string;
}

export interface PackAccessResult {
  granted: boolean;
  url?: string;
  error?: string;
  /** True when the verification server was unreachable */
  offline: boolean;
}

/** Outcome of an in-app self-update (download + verify + swap the binary). */
export interface UpdateSelfResult {
  applied: boolean;
  version?: string;
  error?: string;
}

/** Summary of a user-selected local .polypack (manual profile mode). */
export interface PolyPackInfo {
  path: string;
  id: string;
  name: string;
  version: string;
  minecraft?: string;
  loaderType?: string;
  loaderVersion?: string;
  modCount: number;
}

/** Sentinel modpack id representing a user-loaded local pack. */
export const LOCAL_PACK_ID = '__localpack__';

export interface AppState {
  step: Step;
  options: OptionDescriptor[];
  modpacks?: RemotePack[];
  localPack?: PolyPackInfo;
  selectedMode?: Mode;
  selectedModpack?: string;
  /** Resolved download URL for the selected hosted pack (empty for legacy/launcher-ZIP packs). */
  selectedPackUrl?: string;
  /** Display name of the selected hosted pack, used in download progress labels. */
  selectedPackName?: string;
  selectedInstaller?: OptionDescriptor;
  selectedPath?: string;
  logs: LogEntry[];
  busy: boolean;
  lastResult?: ActionResult | null;
  error?: string | null;
  loadingMessages: string[];
  loadingStarted: boolean;
  loadingComplete: boolean;
}
