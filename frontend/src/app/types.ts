export interface OptionDescriptor {
  id: string;
  title: string;
  description: string;
  requiresPath?: boolean;
  pathLabel?: string;
  detectedPath?: string;
  found?: boolean;
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

export interface AppState {
  step: Step;
  options: OptionDescriptor[];
  modpacks?: RemotePack[];
  selectedMode?: Mode;
  selectedModpack?: string;
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
