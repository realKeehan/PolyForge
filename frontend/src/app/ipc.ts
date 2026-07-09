import type { ActionResult, ExecutionPayload, OptionDescriptor, PackAccessResult, PolyPackInfo, RemoteContentResult, UpdateSelfResult } from './types';

export async function fetchMenuOptions(): Promise<OptionDescriptor[]> {
  return window.go.app.App.GetMenuOptions();
}

export async function fetchRemoteContent(): Promise<RemoteContentResult> {
  return window.go.app.App.GetRemoteContent();
}

export async function verifyPackAccess(packId: string, password: string): Promise<PackAccessResult> {
  return window.go.app.App.VerifyPackAccess(packId, password);
}

export async function updateSelf(): Promise<UpdateSelfResult> {
  return window.go.app.App.UpdateSelf();
}

export async function selectPackFile(): Promise<string | undefined> {
  try {
    const path = await window.go.app.App.SelectPackFile();
    return path || undefined;
  } catch (error) {
    console.error('Failed to open pack file dialog', error);
    return undefined;
  }
}

export async function inspectPolyPack(path: string): Promise<PolyPackInfo> {
  return window.go.app.App.InspectPolyPack(path);
}

export async function launchedPackPath(): Promise<string | undefined> {
  try {
    const path = await window.go.app.App.LaunchedPackPath();
    return path || undefined;
  } catch {
    return undefined;
  }
}

export async function runInstaller(optionId: string, payload: ExecutionPayload): Promise<ActionResult> {
  return window.go.app.App.Execute(optionId, payload);
}

export async function browseForDirectory(title: string): Promise<string | undefined> {
  try {
    const path = await window.go.app.App.SelectDirectory(title);
    return path || undefined;
  } catch (error) {
    console.error('Failed to open directory dialog', error);
    return undefined;
  }
}
