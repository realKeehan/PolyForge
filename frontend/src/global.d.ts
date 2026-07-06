import type { ActionResult, ExecutionPayload, OptionDescriptor, PackAccessResult, PolyPackInfo, RemoteContentResult } from './app/types';


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

declare module '*.png' {
  const src: string;
  export default src;
}

declare module '*.ico' {
  const src: string;
  export default src;
}

declare module '*.wav' {
  const src: string;
  export default src;
}

declare module '*.mp4' {
  const src: string;
  export default src;
}

declare global {
  interface Window {
    go: {
      app: {
        App: {
          GetMenuOptions(): Promise<OptionDescriptor[]>;
          GetRemoteContent(): Promise<RemoteContentResult>;
          VerifyPackAccess(packID: string, password: string): Promise<PackAccessResult>;
          SelectPackFile(): Promise<string>;
          InspectPolyPack(path: string): Promise<PolyPackInfo>;
          LaunchedPackPath(): Promise<string>;
          Execute(optionID: string, payload: ExecutionPayload): Promise<ActionResult>;
          SelectDirectory(title: string): Promise<string>;
          CloneModrinthProfile(request: ModrinthCloneRequest): Promise<ActionResult>;
          SearchExecutable(request: ExecutableSearchRequest): Promise<ActionResult>;
          EnumerateApplications(): Promise<ActionResult>;
        };
      };
    };
  }
}

export {};
