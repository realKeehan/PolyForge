import type { ActionResult, ExecutionPayload, OptionDescriptor } from './types';

export async function fetchMenuOptions(): Promise<OptionDescriptor[]> {
  return window.go.app.App.GetMenuOptions();
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
