import type { ActionResult, ExecutionPayload, OptionDescriptor, PackAccessResult, RemoteContentResult } from './types';

export async function fetchMenuOptions(): Promise<OptionDescriptor[]> {
  return window.go.app.App.GetMenuOptions();
}

export async function fetchRemoteContent(): Promise<RemoteContentResult> {
  return window.go.app.App.GetRemoteContent();
}

export async function verifyPackAccess(packId: string, password: string): Promise<PackAccessResult> {
  return window.go.app.App.VerifyPackAccess(packId, password);
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
