export interface OptionDescriptor {
  id: string;
  title: string;
  description: string;
  requiresPath?: boolean;
  pathLabel?: string;
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
  License = 'license',
  Mode = 'mode',
  Modpack = 'modpack',
  Installer = 'installer',
  Status = 'status',
}

export type Mode = 'install' | 'repair' | 'update' | 'uninstall';

export interface AppState {
  step: Step;
  options: OptionDescriptor[];
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
