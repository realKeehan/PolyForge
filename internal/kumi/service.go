package kumi

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"time"

	"crypto/tls"

	"polyforge/internal/kumi/install"
)

const (
	// version is a fallback only (unit tests, bare `go run ./internal/...`).
	// The real version comes from the repo-root VERSION file, embedded and
	// injected into AppVersion by package main — bump VERSION, not this.
	version = "5.5.2"
	quiltLoaderZipURL = "https://cdn.discordapp.com/attachments/1174802415531327599/1174934629644509245/quilt-loader-0.22.0-beta.1-1.20.1.zip"
	vanillaZipURL     = "https://cdn.discordapp.com/attachments/1174802415531327599/1175988618469310556/TurtelVanilla.zip"
	curseforgeZipURL  = "https://cdn.discordapp.com/attachments/1174802415531327599/1175988721158455316/TurtelCurse.zip"
	multimcZipURL     = "https://cdn.discordapp.com/attachments/1174802415531327599/1175988687146860544/TurtelMulti.zip"
	modrinthZipURL    = "https://cdn.discordapp.com/attachments/1174802415531327599/1175988772614180955/TurtelModrinth.zip?ex=68c2df24&is=68c18da4&hm=18c86a730e583bc86886fc31797246285a679075c476147e363e0c5f990dbbb3&"
	customZipURL      = "https://cdn.discordapp.com/attachments/1174802415531327599/1175988791434018937/TurtelCustom.zip"
	manualZipURL      = "https://cdn.discordapp.com/attachments/1174802415531327599/1175988798690185257/TurtelManual.zip"
	emptyZipWarning   = "No ZIP URL configured for this launcher yet."
)

var (
	//go:embed assets/launcher_icon_base64.txt
	launcherIconRaw  string
	launcherIconData string
)

func init() {
	launcherIconData = "data:image/png;base64," + strings.TrimSpace(launcherIconRaw)
}

// userAgent identifies the app (and its current version) on HTTP requests.
func userAgent() string {
	return "KUMI-Installer/" + currentAppVersion() + " (+https://keehan.co)"
}

// Service encapsulates installer behaviour and keeps shared dependencies.
type Service struct {
	ctx    context.Context
	client *http.Client
	// emitFn streams live install events to the UI; nil = no streaming (tests,
	// bare `go run`). Wired by the app layer via SetEmitter. See stream.go.
	emitFn func(event string, data ...interface{})
}

func NewService() *Service {
	transport := &http.Transport{TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
	return &Service{client: &http.Client{Transport: transport}}
}

func (s *Service) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *Service) Options() []OptionDescriptor {
	options := []OptionDescriptor{
		{ID: "vanilla", Title: "Vanilla Install", Description: "Install the Turtel SMP5 instance for the default Minecraft launcher."},
		{ID: "multimc", Title: "MultiMC Install", Description: "Provision the MultiMC instance.", RequiresPath: true, PathLabel: "MultiMC Root"},
		{ID: "curseforge", Title: "CurseForge Install", Description: "Install into CurseForge instances."},
		{ID: "modrinth", Title: "Modrinth Install", Description: "Install into Modrinth's Theseus launcher."},
		{ID: "gdlauncher", Title: "GDLauncher Install", Description: "Install into GDLauncher.", RequiresPath: true, PathLabel: "GDLauncher Root"},
		{ID: "atlauncher", Title: "ATLauncher Install", Description: "Install into ATLauncher.", RequiresPath: true, PathLabel: "ATLauncher Root"},
		{ID: "prismlauncher", Title: "PrismLauncher Install", Description: "Install into PrismLauncher.", RequiresPath: true, PathLabel: "PrismLauncher Root"},
		{ID: "bakaxl", Title: "BakaXL Install", Description: "Install into BakaXL.", RequiresPath: true, PathLabel: "BakaXL Root"},
		{ID: "feather", Title: "Feather Install", Description: "Install into Feather client.", RequiresPath: true, PathLabel: "Feather Root"},
		{ID: "technic", Title: "Technic Install", Description: "Install into Technic.", RequiresPath: true, PathLabel: "Technic Root"},
		{ID: "polymc", Title: "PolyMC Install", Description: "Install into PolyMC.", RequiresPath: true, PathLabel: "PolyMC Root"},
		{ID: "sklauncher", Title: "SK Launcher Install", Description: "Install into SK Launcher.", RequiresPath: true, PathLabel: "SK Launcher Root"},
		{ID: "freesm", Title: "Freesm Launcher Install", Description: "Install into Freesm Launcher.", RequiresPath: true, PathLabel: "Freesm Root"},
		{ID: "elyprism", Title: "PineconeMC Install", Description: "Install into PineconeMC (formerly ElyPrism).", RequiresPath: true, PathLabel: "PineconeMC Root"},
		{ID: "shatteredprism", Title: "ShatteredPrism Install", Description: "Install into ShatteredPrism.", RequiresPath: true, PathLabel: "ShatteredPrism Root"},
		{ID: "qwertz", Title: "QWERTZ Install", Description: "Install into QWERTZ Launcher.", RequiresPath: true, PathLabel: "QWERTZ Root"},
		{ID: "fjord", Title: "Fjord Launcher Install", Description: "Install into Fjord Launcher.", RequiresPath: true, PathLabel: "Fjord Root"},
		{ID: "hmcl", Title: "HMCL Install", Description: "Install into HMCL.", RequiresPath: true, PathLabel: "HMCL Root"},
		{ID: "ultimmc", Title: "UltimMC Install", Description: "Install into UltimMC.", RequiresPath: true, PathLabel: "UltimMC Root"},
		{ID: "polymerium", Title: "Polymerium Install", Description: "Install into Polymerium.", RequiresPath: true, PathLabel: "Polymerium Root"},
		{ID: "xmcl", Title: "X Minecraft Launcher Install", Description: "Install into X Minecraft Launcher (XMCL).", RequiresPath: true, PathLabel: "XMCL Root"},
		{ID: "custom", Title: "Custom Install", Description: "Install mods into a custom mods folder.", RequiresPath: true, PathLabel: "Mods Folder"},
		{ID: "manual", Title: "Manual Install", Description: "Download the manual installation zip to the chosen location.", RequiresPath: true, PathLabel: "Target Folder"},
		{ID: "about", Title: "About", Description: "View information about PolyForge."},
		{ID: "cake", Title: "Cake?", Description: "Trigger the playful easter egg."},
	}

	// Auto-detect launcher paths: cheap data-dir candidates first, then an
	// exe-based fallback (cache → Start Menu shortcuts → bounded scan) so
	// detection is not constrained to the well-known folders.
	for i := range options {
		detected := s.detectLauncherPath(options[i].ID)
		if detected != "" {
			options[i].DetectedPath = detected
			options[i].Found = true
		}
		// Modrinth keeps profiles under the custom_dir set in its app.db —
		// surface the resolved location behind the row's info icon.
		if options[i].ID == "modrinth" {
			options[i].Info = modrinthProfilesInfo()
		}
	}
	s.detectByExecutable(options)

	return options
}

// launcherExeNames maps option IDs to the executable names used for
// exe-based detection when the data-dir candidates miss. Matching is
// case-insensitive.
var launcherExeNames = map[string][]string{
	"curseforge":     {"CurseForge.exe"},
	"modrinth":       {"Modrinth App.exe"},
	"multimc":        {"MultiMC.exe"},
	"gdlauncher":     {"GDLauncher.exe", "GDLauncher Carbon.exe"},
	"atlauncher":     {"ATLauncher.exe"},
	"prismlauncher":  {"prismlauncher.exe"},
	"bakaxl":         {"BakaXL.exe"},
	"feather":        {"Feather Launcher.exe", "Feather.exe"},
	"technic":        {"TechnicLauncher.exe", "technic-launcher.exe"},
	"polymc":         {"polymc.exe"},
	"sklauncher":     {"SKlauncher.exe"},
	"freesm":         {"freesmlauncher.exe"},
	"elyprism":       {"PineconeMC.exe", "ElyPrismLauncher.exe", "elyprism.exe"},
	"shatteredprism": {"shatteredprism.exe"},
	"qwertz":         {"QWERTZ Launcher.exe", "QWERTZLauncher.exe"},
	"fjord":          {"fjordlauncher.exe"},
	"hmcl":           {"HMCL.exe"},
	"ultimmc":        {"UltimMC.exe"},
	"polymerium":     {"Polymerium.exe"},
	"xmcl":           {"xmcl.exe", "X Minecraft Launcher.exe"},
}

// detectByExecutable fills in launchers the data-dir pass missed by finding
// their executables: validated cache entries first, then Start Menu /
// taskbar / Desktop shortcuts (cheap, any drive), then a depth-limited
// concurrent scan of common install roots. Scan hits are cached so later
// startups skip the expensive step, and the deep scan itself runs at most
// once every 12 hours.
func (s *Service) detectByExecutable(options []OptionDescriptor) {
	wanted := map[string]string{} // lowercase exe name -> option ID
	index := map[string]int{}     // option ID -> options slice index
	for i := range options {
		if options[i].Found {
			continue
		}
		names := launcherExeNames[options[i].ID]
		if len(names) == 0 {
			continue
		}
		index[options[i].ID] = i
		for _, n := range names {
			wanted[strings.ToLower(n)] = options[i].ID
		}
	}
	if len(wanted) == 0 {
		return
	}

	found := map[string]string{} // option ID -> exe path
	dropFound := func() {
		for name, id := range wanted {
			if _, ok := found[id]; ok {
				delete(wanted, name)
			}
		}
	}

	cache, _ := LoadCache()
	cacheDirty := false
	remember := func(id, path string, ev Evidence, confidence string) {
		UpsertCandidate(cache, &Candidate{
			Launcher:   LauncherID(id),
			Path:       path,
			Kind:       "exe",
			Evidence:   ev,
			Confidence: confidence,
			LastUsed:   time.Now(),
			LastOK:     time.Now(),
			HashHint:   PathHint(path),
		})
		cacheDirty = true
	}

	// 1) Previously validated cache entries
	for id := range index {
		if cand := BestValidCachedCandidate(cache, LauncherID(id), ValidateExeByName(launcherExeNames[id]...)); cand != nil {
			found[id] = cand.Path
		}
	}
	dropFound()

	// 2+3) Shortcut resolution and the bounded filesystem scan both read
	// thousands of files, so they share a 12h throttle. Hits land in the
	// cache, making later startups instant; the Browse button covers any
	// launcher installed inside the throttle window.
	if len(wanted) > 0 && shouldRunDeepScan() {
		for id, path := range resolveShortcutTargets(wanted) {
			found[id] = path
			remember(id, path, EvStartMenuLnk, "high")
		}
		dropFound()

		if len(wanted) > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			for id, path := range scanForExes(ctx, commonScanRoots(), wanted, 5, 8) {
				found[id] = path
				remember(id, path, EvScan, "low")
			}
		}
		markDeepScanDone()
	}

	if cacheDirty {
		_ = SaveCache(cache)
	}

	for id, path := range found {
		i := index[id]
		options[i].DetectedPath = filepath.Dir(path)
		options[i].Found = true
	}
}

// ── Deep-scan throttle ───────────────────────────
// New installs are still picked up immediately via shortcuts and the cache;
// the stamp only limits how often the recursive scan can run.

func deepScanStampPath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "PolyForge", "last-deep-scan"), nil
}

func shouldRunDeepScan() bool {
	path, err := deepScanStampPath()
	if err != nil {
		return true
	}
	info, err := os.Stat(path)
	if err != nil {
		return true
	}
	return time.Since(info.ModTime()) > 12*time.Hour
}

func markDeepScanDone() {
	path, err := deepScanStampPath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(path, []byte(time.Now().Format(time.RFC3339)), 0o644)
}

// detectLauncherPath returns the first detected installation path for the given launcher ID.
func (s *Service) detectLauncherPath(id string) string {
	switch id {
	case "vanilla":
		if mc := defaultMinecraftDir(); pathExists(mc) {
			return mc
		}
		return ""
	case "curseforge":
		target, err := curseForgeTarget()
		if err != nil {
			return ""
		}
		// Check parent (CurseForge root) rather than the pack-specific folder
		parent := filepath.Dir(filepath.Dir(target))
		if pathExists(parent) {
			return parent
		}
		return ""
	case "modrinth":
		target, err := modrinthTarget()
		if err != nil {
			return ""
		}
		parent := filepath.Dir(filepath.Dir(target))
		if pathExists(parent) {
			return parent
		}
		return ""
	case "multimc":
		// MultiMC has no fixed data dir — its candidates are exe-probe roots
		// (home, Program Files), so a bare directory check would always
		// "find" it. Detection happens via the exe pipeline instead.
		return ""
	case "gdlauncher":
		return firstExistingDirectory(gdLauncherCandidates(""))
	case "atlauncher":
		return firstExistingDirectory(atLauncherCandidates(""))
	case "prismlauncher":
		return firstExistingDirectory(prismLauncherCandidates(""))
	case "bakaxl":
		return firstExistingDirectory(bakaXLCandidates(""))
	case "feather":
		return firstExistingDirectory(featherCandidates(""))
	case "technic":
		return firstExistingDirectory(technicCandidates(""))
	case "polymc":
		return firstExistingDirectory(polyMCCandidates(""))
	case "sklauncher":
		return firstExistingDirectory(skLauncherCandidates(""))
	case "freesm":
		return firstExistingDirectory(freesmCandidates(""))
	case "elyprism":
		return firstExistingDirectory(elyPrismCandidates(""))
	case "shatteredprism":
		return firstExistingDirectory(shatteredPrismCandidates(""))
	case "qwertz":
		return firstExistingDirectory(qwertzCandidates(""))
	case "fjord":
		return firstExistingDirectory(fjordCandidates(""))
	case "hmcl":
		return firstExistingDirectory(hmclCandidates(""))
	case "ultimmc":
		return firstExistingDirectory(ultimMCCandidates(""))
	case "polymerium":
		return firstExistingDirectory(polymeriumCandidates(""))
	case "xmcl":
		return firstExistingDirectory(xmclCandidates(""))
	default:
		return ""
	}
}

// installFromLocalPack installs a user-provided .polypack (manual
// profile mode): extracts the pack's overrides into the chosen target
// directory and records the installed manifest for future update diffs.
// Per-launcher profile generation is a TODO (see packformat.go).
func (s *Service) installFromLocalPack(payload ExecutionPayload) (*ActionResult, error) {
	result := NewResult()
	packPath := strings.TrimSpace(payload.Extra["packPath"])
	target := strings.TrimSpace(payload.Path)
	if packPath == "" {
		result.Error("No pack file provided.")
		result.Success = false
		return result, nil
	}
	if target == "" {
		result.Error("No target directory selected.")
		result.Success = false
		return result, nil
	}
	if err := ensureDir(target); err != nil {
		result.Error(fmt.Sprintf("cannot create target directory: %v", err))
		result.Success = false
		return result, nil
	}

	s.logStep(result, "info", fmt.Sprintf("Installing local pack from %s", packPath))
	if !s.extractAndVerifyPack(result, packPath, target) {
		result.Success = false
		return result, nil
	}
	result.Success = true
	return result, nil
}

// defaultMinecraftDir returns the platform-specific default .minecraft location.
func defaultMinecraftDir() string {
	switch goruntime.GOOS {
	case "windows":
		if cfg, err := os.UserConfigDir(); err == nil {
			return filepath.Join(cfg, ".minecraft")
		}
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, "Library", "Application Support", "minecraft")
		}
	default:
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".minecraft")
		}
	}
	return ""
}

func (s *Service) Execute(optionID string, payload ExecutionPayload) (*ActionResult, error) {
	switch optionID {
	case "vanilla":
		return s.installVanilla()
	case "multimc":
		return s.installMultiMC(payload.Path)
	case "curseforge":
		return s.installCurseForge()
	case "modrinth":
		return s.installModrinth()
	case "gdlauncher":
		return s.installGDLauncher(payload.Path)
	case "atlauncher":
		return s.installATLauncher(payload.Path)
	case "prismlauncher":
		return s.installPrismLauncher(payload.Path)
	case "bakaxl":
		return s.installBakaXL(payload.Path)
	case "feather":
		return s.installFeather(payload.Path)
	case "technic":
		return s.installTechnic(payload.Path)
	case "polymc":
		return s.installPolyMC(payload.Path)
	case "sklauncher":
		return s.installSKLauncher(payload.Path)
	case "freesm":
		return s.installFreesm(payload.Path)
	case "elyprism":
		return s.installElyPrism(payload.Path)
	case "shatteredprism":
		return s.installShatteredPrism(payload.Path)
	case "qwertz":
		return s.installQWERTZ(payload.Path)
	case "fjord":
		return s.installFjord(payload.Path)
	case "hmcl":
		return s.installHMCL(payload.Path)
	case "ultimmc":
		return s.installUltimMC(payload.Path)
	case "polymerium":
		return s.installPolymerium(payload.Path)
	case "xmcl":
		return s.installXMCL(payload.Path)
	case "custom":
		return s.installCustomMods(payload.Path)
	case "manual":
		return s.installManual(payload.Path)
	case "localpack":
		return s.installFromLocalPack(payload)
	case "hostedpack":
		return s.installHostedPack(payload)
	case "about":
		return s.aboutMessage(), nil
	case "cake":
		return s.cakeMessage(), nil
	default:
		return nil, fmt.Errorf("unknown option '%s'", optionID)
	}
}

func (s *Service) aboutMessage() *ActionResult {
	result := NewResult()
	result.Success = true
	result.Info("Keehan's Universal Modpack Installer (PolyForge) " + version)
	result.Info("Turtel Forever")
	result.Info("Don't try the cake...")
	return result
}

func (s *Service) cakeMessage() *ActionResult {
	result := NewResult()
	result.Success = true
	result.Warning("Nice computer, can I have it?!")
	result.Warning("The easter egg video is only available in the Windows build.")
	return result
}

func (s *Service) installDependencies() install.Dependencies {
	return install.Dependencies{
		DownloadAndExtract:     s.downloadAndExtract,
		AddLauncherProfile:     addLauncherProfile,
		EnsureDir:              ensureDir,
		PathExists:             pathExists,
		FirstExisting:          firstExisting,
		FirstExistingDirectory: firstExistingDirectory,
	}
}

func (s *Service) requirePath(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("a destination path must be provided")
	}
	return nil
}
