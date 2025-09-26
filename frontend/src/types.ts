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

export interface LogEntry {
  level: 'info' | 'warning' | 'error';
  message: string;
}

export interface ActionResult {
  success: boolean;
  messages: LogEntry[];
}
