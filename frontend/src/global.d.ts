import type { ActionResult, ExecutionPayload, OptionDescriptor } from './app/types';

export interface ModrinthCloneRequest {
  dbPath: string;
  sourcePath: string;
  newPath: string;
  newName: string;
  gameVersion: string;
  modLoader: string;
  modLoaderVersion: string;
  resetLastPlayed?: boolean;
  resetPlayCounters?: boolean;
}

export interface ExecutableSearchRequest {
  query: string;
  searchAllDrives?: boolean;
}

export interface BackendBindings {
  GetMenuOptions(): Promise<OptionDescriptor[]>;
  Execute(optionID: string, payload: ExecutionPayload): Promise<ActionResult>;
  SelectDirectory(title: string): Promise<string>;
  CloneModrinthProfile(request: ModrinthCloneRequest): Promise<ActionResult>;
  SearchExecutable(request: ExecutableSearchRequest): Promise<ActionResult>;
  EnumerateApplications(): Promise<ActionResult>;
}

declare global {
  interface Window {
    go: {
      app: {
        App: BackendBindings;
      };
    };
  }
}

export type BackendAPI = BackendBindings;

export {};
