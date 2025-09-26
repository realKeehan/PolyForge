package kumi

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"polyforge/internal/kumi/install"
)

const (
	version           = "5.5.1"
	userAgent         = "KUMI-Installer/5.5.1 (+https://keehan.co)"
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

// Service encapsulates installer behaviour and keeps shared dependencies.
type Service struct {
	ctx    context.Context
	client *http.Client
}

func NewService() *Service {
	return &Service{client: &http.Client{}}
}

func (s *Service) SetContext(ctx context.Context) {
	s.ctx = ctx
}

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
		DownloadAndExtract: s.downloadAndExtract,
		AddLauncherProfile: addLauncherProfile,
	}
}

func (s *Service) requirePath(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("a destination path must be provided")
	}
	return nil
}
