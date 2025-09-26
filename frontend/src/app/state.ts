import type { ActionResult, AppState, LogEntry, Mode, OptionDescriptor } from './types';
import { Step } from './types';

type Listener = (state: AppState) => void;

function createInitialState(): AppState {
  return {
    step: Step.Loading,
    options: [],
    logs: [],
    busy: false,
    lastResult: null,
    error: null,
  };
}

export class Store {
  private state: AppState = createInitialState();
  private listeners = new Set<Listener>();

  subscribe(listener: Listener): () => void {
    this.listeners.add(listener);
    listener(this.state);
    return () => {
      this.listeners.delete(listener);
    };
  }

  private emit() {
    for (const listener of this.listeners) {
      listener(this.state);
    }
  }

  reset() {
    this.state = createInitialState();
    this.emit();
  }

  setOptions(options: OptionDescriptor[]) {
    this.state = { ...this.state, options };
    this.emit();
  }

  setStep(step: Step) {
    if (this.state.step === step) return;
    this.state = { ...this.state, step };
    this.emit();
  }

  setMode(mode: Mode) {
    this.state = { ...this.state, selectedMode: mode };
    this.emit();
  }

  setModpack(modpack: string) {
    this.state = { ...this.state, selectedModpack: modpack };
    this.emit();
  }

  setInstaller(option: OptionDescriptor | undefined) {
    this.state = { ...this.state, selectedInstaller: option };
    this.emit();
  }

  setPath(path: string | undefined) {
    this.state = { ...this.state, selectedPath: path };
    this.emit();
  }

  setBusy(busy: boolean) {
    if (this.state.busy === busy) return;
    this.state = { ...this.state, busy };
    this.emit();
  }

  clearLogs() {
    this.state = { ...this.state, logs: [] };
    this.emit();
  }

  appendLogs(entries: LogEntry[]) {
    if (entries.length === 0) return;
    this.state = { ...this.state, logs: [...this.state.logs, ...entries] };
    this.emit();
  }

  setResult(result: ActionResult | null) {
    this.state = { ...this.state, lastResult: result };
    this.emit();
  }

  setError(message: string | null) {
    this.state = { ...this.state, error: message };
    this.emit();
  }

  getState(): AppState {
    return this.state;
  }
}

export function createStore(): Store {
  return new Store();
}
