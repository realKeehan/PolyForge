package installer

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	osruntime "runtime"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	userAgent         = "KUMI-Installer/5.5.1 (+https://keehan.co)"
	blipSoundURL      = "http://keehan.co/KUMI_Files/srcassetsaudiostdout.wav"
	notifySoundURL    = "http://keehan.co/KUMI_Files/srcassetsaudionoti.wav"
	quiltLoaderZipURL = "https://cdn.discordapp.com/attachments/1174802415531327599/1174934629644509245/quilt-loader-0.22.0-beta.1-1.20.1.zip"
	vanillaZipURL     = "https://cdn.discordapp.com/attachments/1174802415531327599/1175988618469310556/TurtelVanilla.zip"
	curseforgeZipURL  = "https://cdn.discordapp.com/attachments/1174802415531327599/1175988721158455316/TurtelCurse.zip"
	multimcZipURL     = "https://cdn.discordapp.com/attachments/1174802415531327599/1175988687146860544/TurtelMulti.zip"
	modrinthZipURL    = "https://cdn.discordapp.com/attachments/1174802415531327599/1175988772614180955/TurtelModrinth.zip?ex=68c2df24&is=68c18da4&hm=18c86a730e583bc86886fc31797246285a679075c476147e363e0c5f990dbbb3&"
	customZipURL      = "https://cdn.discordapp.com/attachments/1174802415531327599/1175988791434018937/TurtelCustom.zip"
	manualZipURL      = "https://cdn.discordapp.com/attachments/1174802415531327599/1175988798690185257/TurtelManual.zip"
)

var launcherIconData = "data:image/png;base64,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"

var emptyZipWarning = "No ZIP URL configured for this launcher yet."

// Service encapsulates all installer behaviour.
type Service struct {
	ctx    context.Context
	client *http.Client
}

// OptionDescriptor provides metadata for UI rendering.
type OptionDescriptor struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	RequiresPath bool   `json:"requiresPath"`
	PathLabel    string `json:"pathLabel,omitempty"`
}

// ExecutionPayload transports user supplied values.
type ExecutionPayload struct {
	Path  string            `json:"path,omitempty"`
	Extra map[string]string `json:"extra,omitempty"`
}

// LogEntry is a structured log message.
type LogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

// ActionResult summarises an operation.
type ActionResult struct {
	Success  bool       `json:"success"`
	Messages []LogEntry `json:"messages"`
}

// ModrinthCloneRequest mirrors the PowerShell clone script.
type ModrinthCloneRequest struct {
	DBPath            string `json:"dbPath"`
	SourcePath        string `json:"sourcePath"`
	NewPath           string `json:"newPath"`
	NewName           string `json:"newName"`
	GameVersion       string `json:"gameVersion"`
	ModLoader         string `json:"modLoader"`
	ModLoaderVersion  string `json:"modLoaderVersion"`
	ResetLastPlayed   bool   `json:"resetLastPlayed"`
	ResetPlayCounters bool   `json:"resetPlayCounters"`
}

// ExecutableSearchRequest describes the FastFind configuration.
type ExecutableSearchRequest struct {
	Query           string `json:"query"`
	SearchAllDrives bool   `json:"searchAllDrives"`
}

// ApplicationInfo describes entries from EnumerateApplications.
type ApplicationInfo struct {
	Name            string `json:"name"`
	Kind            string `json:"kind"`
	TargetPath      string `json:"targetPath"`
	AppUserModelID  string `json:"appUserModelId"`
	PackageFullName string `json:"packageFullName"`
	LaunchCommand   string `json:"launchCommand"`
	Type            string `json:"type"`
}

// NewService creates the installer service.
func NewService() *Service {
	client := &http.Client{
		Timeout: 0,
	}
	return &Service{client: client}
}

// SetContext stores the runtime context for UI interactions.
func (s *Service) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// Options exposes all menu entries.
func (s *Service) Options() []OptionDescriptor {
	return []OptionDescriptor{
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
		{ID: "custom", Title: "Custom Install", Description: "Install mods into a custom mods folder.", RequiresPath: true, PathLabel: "Mods Folder"},
		{ID: "manual", Title: "Manual Install", Description: "Download the manual installation zip to the chosen location.", RequiresPath: true, PathLabel: "Target Folder"},
		{ID: "about", Title: "About", Description: "View information about PolyForge."},
		{ID: "cake", Title: "Cake?", Description: "Trigger the playful easter egg."},
	}
}

// Execute routes to the appropriate handler.
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
	case "custom":
		return s.installCustomMods(payload.Path)
	case "manual":
		return s.installManual(payload.Path)
	case "about":
		return s.aboutMessage(), nil
	case "cake":
		return s.cakeMessage(), nil
	default:
		return nil, fmt.Errorf("unknown option '%s'", optionID)
	}
}

// CloneModrinthProfile reproduces the SQL heavy lifting from PowerShell.
func (s *Service) CloneModrinthProfile(request ModrinthCloneRequest) (*ActionResult, error) {
	result := newResult()

	if request.DBPath == "" {
		return nil, errors.New("database path is required")
	}
	if request.SourcePath == "" || request.NewPath == "" {
		return nil, errors.New("source and target profile identifiers are required")
	}
	if request.NewName == "" {
		request.NewName = request.NewPath
	}

	db, err := sql.Open("sqlite3", request.DBPath)
	if err != nil {
		result.Error(fmt.Sprintf("failed to open database: %v", err))
		result.Success = false
		return result, nil
	}
	defer db.Close()

	var srcCount int
	if err := db.QueryRow("SELECT COUNT(1) FROM profiles WHERE path = ?", request.SourcePath).Scan(&srcCount); err != nil {
		result.Error(fmt.Sprintf("failed to query source profile: %v", err))
		result.Success = false
		return result, nil
	}
	if srcCount < 1 {
		result.Error(fmt.Sprintf("source profile '%s' not found", request.SourcePath))
		result.Success = false
		return result, nil
	}

	var newCount int
	if err := db.QueryRow("SELECT COUNT(1) FROM profiles WHERE path = ?", request.NewPath).Scan(&newCount); err != nil {
		result.Error(fmt.Sprintf("failed to query target profile: %v", err))
		result.Success = false
		return result, nil
	}
	if newCount > 0 {
		result.Error(fmt.Sprintf("target profile '%s' already exists", request.NewPath))
		result.Success = false
		return result, nil
	}

	cloneSQL := `INSERT INTO profiles (
  path, install_stage, name, icon_path,
  game_version, mod_loader, mod_loader_version,
  groups, linked_project_id, linked_version_id, locked,
  created, modified, last_played,
  submitted_time_played, recent_time_played,
  override_java_path, override_extra_launch_args, override_custom_env_vars,
  override_mc_memory_max, override_mc_force_fullscreen,
  override_mc_game_resolution_x, override_mc_game_resolution_y,
  override_hook_pre_launch, override_hook_wrapper, override_hook_post_exit,
  protocol_version, launcher_feature_version
)
SELECT
  ?, install_stage, ?, icon_path,
  game_version, mod_loader, mod_loader_version,
  groups, linked_project_id, linked_version_id, locked,
  created, modified, last_played,
  submitted_time_played, recent_time_played,
  override_java_path, override_extra_launch_args, override_custom_env_vars,
  override_mc_memory_max, override_mc_force_fullscreen,
  override_mc_game_resolution_x, override_mc_game_resolution_y,
  override_hook_pre_launch, override_hook_wrapper, override_hook_post_exit,
  protocol_version, launcher_feature_version
FROM profiles WHERE path = ?`

	if _, err := db.Exec(cloneSQL, request.NewPath, request.NewName, request.SourcePath); err != nil {
		result.Error(fmt.Sprintf("failed to clone profile: %v", err))
		result.Success = false
		return result, nil
	}

	nowUnix := time.Now().UTC().Unix()
	updateSQL := `UPDATE profiles
SET game_version = ?,
    mod_loader = ?,
    mod_loader_version = ?,
    modified = ?,
    last_played = ?,
    recent_time_played = ?
WHERE path = ?`

	lastPlayed := sql.NullInt64{Valid: false}
	if !request.ResetLastPlayed {
		// Keep the cloned value by reading it back.
		row := db.QueryRow("SELECT last_played FROM profiles WHERE path = ?", request.NewPath)
		var lp sql.NullInt64
		if err := row.Scan(&lp); err == nil && lp.Valid {
			lastPlayed = lp
		}
	}

	playCount := 0
	if !request.ResetPlayCounters {
		row := db.QueryRow("SELECT recent_time_played FROM profiles WHERE path = ?", request.NewPath)
		var pc sql.NullInt64
		if err := row.Scan(&pc); err == nil && pc.Valid {
			playCount = int(pc.Int64)
		}
	}

	_, err = db.Exec(updateSQL,
		request.GameVersion,
		request.ModLoader,
		request.ModLoaderVersion,
		nowUnix,
		lastPlayed,
		playCount,
		request.NewPath,
	)
	if err != nil {
		result.Error(fmt.Sprintf("failed to update profile overrides: %v", err))
		result.Success = false
		return result, nil
	}

	verifySQL := `SELECT path, name, game_version, mod_loader, mod_loader_version FROM profiles WHERE path = ?`
	row := db.QueryRow(verifySQL, request.NewPath)
	var path, name, gameVersion, modLoader, modLoaderVersion string
	if err := row.Scan(&path, &name, &gameVersion, &modLoader, &modLoaderVersion); err == nil {
		result.Info(fmt.Sprintf("Cloned '%s' to '%s' (version %s - %s %s)", request.SourcePath, request.NewPath, gameVersion, modLoader, modLoaderVersion))
	}

	result.Success = true
	return result, nil
}

// SearchExecutable re-implements the FastFind helper in Go.
func (s *Service) SearchExecutable(query ExecutableSearchRequest) (*ActionResult, error) {
	result := newResult()
	exeName, args := splitExecutableQuery(query.Query)
	if exeName == "" {
		result.Error("no executable name found in query")
		result.Success = false
		return result, nil
	}

	preferredRoots := s.collectPreferredRoots()
	matches := s.scanRoots(preferredRoots, exeName, args)

	if len(matches) == 0 && query.SearchAllDrives {
		drives := s.enumerateDrives()
		matches = s.scanRoots(drives, exeName, args)
	}

	if len(matches) == 0 {
		result.Warning(fmt.Sprintf("No match for '%s' (exe: '%s'%s)", query.Query, exeName, formatArgs(args)))
		result.Success = false
		return result, nil
	}

	sort.Strings(matches)
	for _, match := range matches {
		result.Info(match)
	}
	result.Success = true
	return result, nil
}

// EnumerateApplications surfaces installed applications in a lightweight manner.
func (s *Service) EnumerateApplications() (*ActionResult, error) {
	result := newResult()
	infos, err := enumerateApplications()
	if err != nil {
		result.Error(fmt.Sprintf("failed to enumerate applications: %v", err))
		result.Success = false
		return result, nil
	}
	if len(infos) == 0 {
		result.Warning("no applications detected")
		result.Success = true
		return result, nil
	}

	for _, info := range infos {
		result.Info(fmt.Sprintf("%s [%s] -> %s", info.Name, info.Kind, info.LaunchCommand))
	}
	result.Success = true
	return result, nil
}

func (s *Service) installVanilla() (*ActionResult, error) {
	result := newResult()
	home, err := os.UserHomeDir()
	if err != nil {
		result.Error(fmt.Sprintf("unable to resolve user home: %v", err))
		result.Success = false
		return result, nil
	}

	vanillaDir := filepath.Join(home, "AppData", "Roaming", "KUMIProfiles", "TurtelSMP5")
	if err := ensureDir(vanillaDir); err != nil {
		result.Error(fmt.Sprintf("failed to create vanilla directory: %v", err))
		result.Success = false
		return result, nil
	}

	if err := s.downloadAndExtract(vanillaZipURL, vanillaDir, ""); err != nil {
		result.Error(fmt.Sprintf("vanilla download failed: %v", err))
		result.Success = false
		return result, nil
	}
	result.Info(fmt.Sprintf("Installed vanilla files to %s", vanillaDir))

	versionsDir := filepath.Join(home, "AppData", "Roaming", ".minecraft", "versions")
	if pathExists(versionsDir) {
		if err := s.downloadAndExtract(quiltLoaderZipURL, versionsDir, ""); err != nil {
			result.Warning(fmt.Sprintf("Quilt loader download failed: %v", err))
		} else {
			result.Info("Installed Quilt Loader profile")
		}
	} else {
		result.Warning(fmt.Sprintf("Minecraft versions directory not found at %s", versionsDir))
	}

	launcherJSON := filepath.Join(home, "AppData", "Roaming", ".minecraft", "launcher_profiles.json")
	if err := addLauncherProfile(launcherJSON); err != nil {
		result.Warning(fmt.Sprintf("Failed to update launcher profiles: %v", err))
	} else {
		result.Info("Ensured launcher profile 'turtelsmp'")
	}

	result.Success = true
	return result, nil
}

func (s *Service) installMultiMC(explicitRoot string) (*ActionResult, error) {
	result := newResult()
	root := firstExisting([]string{
		explicitRoot,
		filepath.Join(os.Getenv("USERPROFILE"), "MultiMC"),
		filepath.Join(os.Getenv("ProgramFiles")),
		filepath.Join(os.Getenv("ProgramFiles(x86)")),
	}, "MultiMC.exe")

	if root == "" {
		result.Warning("Unable to locate MultiMC.exe. Please provide the MultiMC root directory.")
		result.Success = false
		return result, nil
	}

	instanceDir := filepath.Join(filepath.Dir(root), "instances", "TurtelSMP5")
	if pathExists(instanceDir) {
		result.Warning("MultiMC instance already exists - skipping download")
		result.Success = true
		return result, nil
	}

	if err := ensureDir(instanceDir); err != nil {
		result.Error(fmt.Sprintf("failed to create instance directory: %v", err))
		result.Success = false
		return result, nil
	}

	if err := s.downloadAndExtract(multimcZipURL, instanceDir, ""); err != nil {
		result.Error(fmt.Sprintf("failed to provision MultiMC instance: %v", err))
		result.Success = false
		return result, nil
	}
	result.Info(fmt.Sprintf("Installed MultiMC instance to %s", instanceDir))
	result.Success = true
	return result, nil
}

func (s *Service) installCurseForge() (*ActionResult, error) {
	result := newResult()
	home, err := os.UserHomeDir()
	if err != nil {
		result.Error(fmt.Sprintf("unable to resolve user home: %v", err))
		result.Success = false
		return result, nil
	}
	target := filepath.Join(home, "curseforge", "minecraft", "Instances", "TurtelSMP5")
	if pathExists(target) {
		result.Warning("CurseForge instance already present - skipping")
		result.Success = true
		return result, nil
	}
	if err := ensureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to create CurseForge directory: %v", err))
		result.Success = false
		return result, nil
	}
	if err := s.downloadAndExtract(curseforgeZipURL, target, ""); err != nil {
		result.Error(fmt.Sprintf("failed to install CurseForge instance: %v", err))
		result.Success = false
		return result, nil
	}
	result.Info(fmt.Sprintf("Installed CurseForge instance to %s", target))
	result.Success = true
	return result, nil
}

func (s *Service) installModrinth() (*ActionResult, error) {
	result := newResult()
	home, err := os.UserHomeDir()
	if err != nil {
		result.Error(fmt.Sprintf("unable to resolve user home: %v", err))
		result.Success = false
		return result, nil
	}
	target := filepath.Join(home, "AppData", "Roaming", "com.modrinth.theseus", "profiles", "TurtelSMP5")
	if pathExists(target) {
		result.Warning("Modrinth instance already present - skipping")
		result.Success = true
		return result, nil
	}
	if err := ensureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to create Modrinth directory: %v", err))
		result.Success = false
		return result, nil
	}
	if err := s.downloadAndExtract(modrinthZipURL, target, "TurtelModrinth.zip"); err != nil {
		result.Error(fmt.Sprintf("failed to install Modrinth instance: %v", err))
		result.Success = false
		return result, nil
	}
	result.Info(fmt.Sprintf("Installed Modrinth instance to %s", target))
	result.Success = true
	return result, nil
}

func (s *Service) installGDLauncher(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "gdlauncher_next"),
		filepath.Join(os.Getenv("APPDATA"), "gdlauncher"),
	}
	return s.installInstanceWithOptionalZip("GDLauncher", candidates, "instances", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installATLauncher(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "ATLauncher"),
		`C:\ATLauncher`,
	}
	return s.installInstanceWithOptionalZip("ATLauncher", candidates, "Instances", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installPrismLauncher(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "PrismLauncher"),
		filepath.Join(os.Getenv("APPDATA"), "PrismLauncher", "minecraft"),
	}
	return s.installInstanceWithOptionalZip("PrismLauncher", candidates, "instances", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installBakaXL(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "BakaXL"),
		`C:\BakaXL`,
	}
	return s.installInstanceWithOptionalZip("BakaXL", candidates, "instances", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installFeather(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "feather"),
		filepath.Join(os.Getenv("APPDATA"), "FeatherClient"),
	}
	return s.installInstanceWithOptionalZip("Feather", candidates, "profiles", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installTechnic(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), ".technic"),
		`C:\.technic`,
	}
	return s.installInstanceWithOptionalZip("Technic", candidates, "modpacks", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installPolyMC(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "PolyMC"),
		filepath.Join(os.Getenv("APPDATA"), "polymc"),
	}
	return s.installInstanceWithOptionalZip("PolyMC", candidates, "instances", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installCustomMods(modsDir string) (*ActionResult, error) {
	result := newResult()
	if modsDir == "" {
		result.Error("mods directory is required")
		result.Success = false
		return result, nil
	}
	if err := ensureDir(modsDir); err != nil {
		result.Error(fmt.Sprintf("failed to prepare mods directory: %v", err))
		result.Success = false
		return result, nil
	}

	bypass := filepath.Join(modsDir, "bypass.turtel")
	if pathExists(bypass) {
		result.Warning("bypass.turtel present — skipping mod installation")
		result.Success = true
		return result, nil
	}

	result.Info("Moving existing files to 'not-turtel'")
	notTurtel := filepath.Join(modsDir, "not-turtel")
	if err := ensureDir(notTurtel); err != nil {
		result.Error(fmt.Sprintf("failed to create not-turtel directory: %v", err))
		result.Success = false
		return result, nil
	}

	entries, err := os.ReadDir(modsDir)
	if err != nil {
		result.Error(fmt.Sprintf("failed to enumerate mods directory: %v", err))
		result.Success = false
		return result, nil
	}
	for _, entry := range entries {
		if entry.Name() == "not-turtel" {
			continue
		}
		source := filepath.Join(modsDir, entry.Name())
		dest := filepath.Join(notTurtel, entry.Name())
		if err := os.Rename(source, dest); err != nil {
			result.Warning(fmt.Sprintf("failed to move %s: %v", entry.Name(), err))
		}
	}

	if err := s.downloadAndExtract(customZipURL, modsDir, "TurtelCustom.zip"); err != nil {
		result.Error(fmt.Sprintf("failed to install custom mods: %v", err))
		result.Success = false
		return result, nil
	}

	result.Info("Installed custom mod files")
	result.Success = true
	return result, nil
}

func (s *Service) installManual(target string) (*ActionResult, error) {
	result := newResult()
	var err error
	if target == "" {
		target, err = os.Getwd()
		if err != nil {
			result.Error(fmt.Sprintf("failed to determine working directory: %v", err))
			result.Success = false
			return result, nil
		}
	}
	if err := ensureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to prepare target directory: %v", err))
		result.Success = false
		return result, nil
	}

	if err := s.downloadAndExtract(manualZipURL, target, "TurtelManual.zip"); err != nil {
		result.Error(fmt.Sprintf("manual install failed: %v", err))
		result.Success = false
		return result, nil
	}

	result.Info(fmt.Sprintf("Manual package extracted to %s", target))
	result.Success = true
	return result, nil
}

func (s *Service) installInstanceWithOptionalZip(label string, candidates []string, subDir string, instanceName string, zipURL string, warning string) (*ActionResult, error) {
	result := newResult()
	root := firstExistingDirectory(candidates)
	if root == "" {
		result.Error(fmt.Sprintf("Unable to locate %s root. Please provide a valid path.", label))
		result.Success = false
		return result, nil
	}

	target := filepath.Join(root, subDir, instanceName)
	if pathExists(target) {
		result.Warning(fmt.Sprintf("%s instance already exists at %s", label, target))
		result.Success = true
		return result, nil
	}

	if err := ensureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to create %s target directory: %v", label, err))
		result.Success = false
		return result, nil
	}

	if zipURL == "" {
		result.Warning(warning)
		result.Success = true
		return result, nil
	}

	if err := s.downloadAndExtract(zipURL, target, ""); err != nil {
		result.Error(fmt.Sprintf("failed to provision %s instance: %v", label, err))
		result.Success = false
		return result, nil
	}

	result.Info(fmt.Sprintf("Installed %s instance to %s", label, target))
	result.Success = true
	return result, nil
}

func (s *Service) aboutMessage() *ActionResult {
	result := newResult()
	result.Info("Keehan's Universal Modpack Installer (KUMI) simplifies installing and updating the Turtel SMP5 modpack across multiple launchers.")
	result.Info("Created by Keehan 2023 – Turtel Forever – Version 5.5.1")
	result.Info("Don't try the cake…")
	result.Success = true
	return result
}

func (s *Service) cakeMessage() *ActionResult {
	result := newResult()
	result.Info("Nice Computer, Can I have it?! (In the Wails edition the easter egg displays a message instead of hijacking your screens.)")
	result.Success = true
	return result
}

func (s *Service) downloadAndExtract(url, destination, explicitName string) error {
	if strings.TrimSpace(url) == "" {
		return errors.New("no download URL configured")
	}

	tmp, err := os.CreateTemp("", "polyforge-*.zip")
	if err != nil {
		return err
	}

	tempPath := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tempPath)
	}()

	if err := s.downloadFile(url, tmp); err != nil {
		return err
	}

	if err := tmp.Sync(); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if explicitName != "" {
		namedPath := filepath.Join(filepath.Dir(tempPath), explicitName)
		if err := os.Rename(tempPath, namedPath); err == nil {
			tempPath = namedPath
		}
	}

	return extractZip(tempPath, destination)
}

func (s *Service) downloadFile(url string, file *os.File) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("download failed with status %s", resp.Status)
	}

	if _, err := io.Copy(file, resp.Body); err != nil {
		return err
	}
	return nil
}

func extractZip(zipPath, destination string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		targetPath := filepath.Join(destination, file.Name)
		if !strings.HasPrefix(targetPath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return fmt.Errorf("zip entry %s is outside the destination", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := ensureDir(targetPath); err != nil {
				return err
			}
			continue
		}

		if err := ensureDir(filepath.Dir(targetPath)); err != nil {
			return err
		}

		src, err := file.Open()
		if err != nil {
			return err
		}

		dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
		if err != nil {
			src.Close()
			return err
		}

		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			dst.Close()
			return err
		}
		src.Close()
		dst.Close()
	}
	return nil
}

func addLauncherProfile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return err
	}

	profiles, ok := data["profiles"].(map[string]interface{})
	if !ok {
		profiles = make(map[string]interface{})
		data["profiles"] = profiles
	}

	if _, exists := profiles["turtelsmp"]; exists {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	gameDir := filepath.Join(home, "AppData", "Roaming", "KUMIProfiles", "TurtelSMP5")
	profile := map[string]interface{}{
		"gameDir":       gameDir,
		"icon":          launcherIconData,
		"javaArgs":      "-Xmx4G -XX:+UnlockExperimentalVMOptions -XX:+UseG1GC -XX:G1NewSizePercent=20 -XX:G1ReservePercent=20 -XX:MaxGCPauseMillis=50 -XX:G1HeapRegionSize=32M",
		"lastUsed":      time.Now().UTC().Format(time.RFC3339Nano),
		"lastVersionId": "quilt-loader-0.22.0-beta.1-1.20.1",
		"name":          "TurtelSMP5",
		"type":          "",
	}

	profiles["turtelsmp"] = profile

	updated, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, updated, 0644)
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func pathExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func firstExisting(candidates []string, exeName string) string {
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		probe := candidate
		if filepath.Ext(candidate) == "" {
			probe = filepath.Join(candidate, exeName)
		}
		if pathExists(probe) {
			return probe
		}
	}
	return ""
}

func firstExistingDirectory(candidates []string) string {
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if info.IsDir() {
			return candidate
		}
	}
	return ""
}

func newResult() *ActionResult {
	return &ActionResult{Success: false, Messages: []LogEntry{}}
}

func (r *ActionResult) Info(message string) {
	r.Messages = append(r.Messages, LogEntry{Level: "info", Message: message})
}

func (r *ActionResult) Warning(message string) {
	r.Messages = append(r.Messages, LogEntry{Level: "warning", Message: message})
}

func (r *ActionResult) Error(message string) {
	r.Messages = append(r.Messages, LogEntry{Level: "error", Message: message})
}

func splitExecutableQuery(query string) (string, string) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return "", ""
	}
	if strings.HasPrefix(trimmed, "\"") {
		parts := strings.SplitN(trimmed, "\"", 3)
		if len(parts) >= 3 {
			return parts[1], strings.TrimSpace(parts[2])
		}
	}
	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return "", ""
	}
	exe := fields[0]
	args := strings.Join(fields[1:], " ")
	return exe, args
}

func formatArgs(args string) string {
	if strings.TrimSpace(args) == "" {
		return ""
	}
	return fmt.Sprintf(", args contain: '%s'", args)
}

func (s *Service) collectPreferredRoots() []string {
	roots := []string{}
	add := func(path string) {
		if path != "" && pathExists(path) {
			roots = append(roots, path)
		}
	}

	user := os.Getenv("USERPROFILE")
	appData := os.Getenv("APPDATA")
	localAppData := os.Getenv("LOCALAPPDATA")
	programData := os.Getenv("ProgramData")

	add(filepath.Join(user, "AppData", "Roaming", "Microsoft", "Internet Explorer", "Quick Launch", "User Pinned"))
	add(filepath.Join(user, "AppData", "Roaming", "Microsoft", "Internet Explorer", "Quick Launch"))
	add(filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs"))
	add(filepath.Join(programData, "Microsoft", "Windows", "Start Menu", "Programs"))
	add(filepath.Join(user, "curseforge", "minecraft", "Install"))
	add(filepath.Join(user, "Desktop"))
	add(os.Getenv("ProgramFiles"))
	add(os.Getenv("ProgramFiles(x86)"))
	add(programData)
	add("D:\\Program Files")
	add("D:\\Program Files (x86)")
	add("D:\\Programs")
	add(localAppData)
	if localAppData != "" {
		add(filepath.Join(filepath.Dir(localAppData), "LocalLow"))
	}
	add(filepath.Join(user, "Downloads"))
	add(appData)
	return roots
}

func (s *Service) scanRoots(roots []string, exeName string, args string) []string {
	var matches []string
	for _, root := range roots {
		if root == "" {
			continue
		}
		if s.ctx != nil {
			runtime.EventsEmit(s.ctx, "search:progress", fmt.Sprintf("Scanning %s", root))
		}
		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if !strings.EqualFold(filepath.Base(path), exeName) {
				if strings.HasSuffix(strings.ToLower(path), ".lnk") {
					if resolved := resolveShortcut(path, exeName, args); resolved != "" {
						matches = append(matches, resolved)
						return filepath.SkipDir
					}
				}
				return nil
			}
			matches = append(matches, path)
			return filepath.SkipDir
		})
		if len(matches) > 0 {
			break
		}
	}
	return unique(matches)
}

func (s *Service) enumerateDrives() []string {
	drives := []string{}
	if osruntime.GOOS != "windows" {
		return drives
	}
	for _, letter := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		root := fmt.Sprintf("%c:\\", letter)
		if pathExists(root) {
			drives = append(drives, root)
		}
	}
	return drives
}

func resolveShortcut(path, exeName, args string) string {
	// Shortcut resolution is not implemented cross-platform.
	// We simply return empty to avoid Windows specific COM usage here.
	return ""
}

func unique(values []string) []string {
	seen := map[string]struct{}{}
	var result []string
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func enumerateApplications() ([]ApplicationInfo, error) {
	// Full COM based enumeration is outside the scope for this example.
	// We provide a placeholder that returns an empty slice on non-Windows systems.
	if osruntime.GOOS != "windows" {
		return []ApplicationInfo{}, nil
	}
	// Implementing Windows shell enumeration requires COM initialisation which is
	// not trivial without external dependencies. The placeholder communicates that
	// the feature is not yet implemented.
	return []ApplicationInfo{}, errors.New("application enumeration is not implemented on this platform build")
}
