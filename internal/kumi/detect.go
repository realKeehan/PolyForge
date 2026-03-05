package kumi

import (
	"os"
	"path/filepath"
)

func multiMCCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("USERPROFILE")),
		filepath.Join(os.Getenv("ProgramFiles")),
		filepath.Join(os.Getenv("ProgramFiles(x86)")),
	}
}

func curseForgeTarget() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "curseforge", "minecraft", "Instances", "TurtelSMP5"), nil
}

func modrinthTarget() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "AppData", "Roaming", "com.modrinth.theseus", "profiles", "TurtelSMP5"), nil
}

func gdLauncherCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "gdlauncher_next"),
		filepath.Join(os.Getenv("APPDATA"), "gdlauncher"),
	}
}

func atLauncherCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "ATLauncher"),
		"C:\\ATLauncher",
	}
}

func prismLauncherCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "PrismLauncher"),
		filepath.Join(os.Getenv("APPDATA"), "PrismLauncher", "minecraft"),
	}
}

func bakaXLCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "BakaXL"),
		"C:\\BakaXL",
	}
}

func featherCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "feather"),
		filepath.Join(os.Getenv("APPDATA"), "FeatherClient"),
	}
}

func technicCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), ".technic"),
		"C:\\.technic",
	}
}

func polyMCCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "PolyMC"),
		filepath.Join(os.Getenv("APPDATA"), "polymc"),
	}
}

func skLauncherCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "SKLauncher"),
		filepath.Join(os.Getenv("APPDATA"), ".sklauncher"),
	}
}

func freesmCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "FreesmLauncher"),
		filepath.Join(os.Getenv("APPDATA"), "freesmlauncher"),
	}
}

func elyPrismCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "ElyPrism"),
		filepath.Join(os.Getenv("APPDATA"), "ElyPrismLauncher"),
	}
}

func shatteredPrismCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "ShatteredPrism"),
	}
}

func qwertzCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "QWERTZ"),
		filepath.Join(os.Getenv("APPDATA"), "qwertz"),
	}
}

func fjordCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "FjordLauncher"),
	}
}

func hmclCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), ".hmcl"),
		filepath.Join(os.Getenv("USERPROFILE"), ".hmcl"),
	}
}

func ultimMCCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "UltimMC"),
	}
}

func polymeriumCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "Polymerium"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Polymerium"),
	}
}

func xmclCandidates(explicit string) []string {
	return []string{
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "xmcl"),
		filepath.Join(os.Getenv("APPDATA"), "X Minecraft Launcher"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "xmcl"),
	}
}
