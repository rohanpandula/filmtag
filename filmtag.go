package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

// Data structures
type CameraType int

const (
	Fixed CameraType = iota
	Interchangeable
	MediumFormat
)

type Lens struct {
	Name        string
	FocalLength int
	MaxAperture float64
}

type Camera struct {
	Make             string
	Model            string
	Type             CameraType
	FixedLens        *Lens
	CompatibleLenses []Lens
}

type FilmStock struct {
	Name   string
	ISO    int
	Format string
}

type CLIFlags struct {
	Camera    string
	Lens      string
	Film      string
	FilePath  string
	Clean     bool
	Verbose   bool
}

// Predefined camera database
var cameras = map[string]Camera{
	"Contax T3": {
		Make:  "Contax",
		Model: "T3",
		Type:  Fixed,
		FixedLens: &Lens{
			Name:        "Carl Zeiss Sonnar T 35mm f/2.8",
			FocalLength: 35,
			MaxAperture: 2.8,
		},
	},
	"Minolta CLE": {
		Make:  "Minolta",
		Model: "CLE",
		Type:  Interchangeable,
		CompatibleLenses: []Lens{
			{Name: "Leica APO-Summicron-M 35mm f/2 ASPH.", FocalLength: 35, MaxAperture: 2.0},
			{Name: "Minolta 28mm 2.8 M-Rokkor", FocalLength: 28, MaxAperture: 2.8},
			{Name: "Canon 50mm f/1.4 LTM", FocalLength: 50, MaxAperture: 1.4},
		},
	},
	"Canon 7E": {
		Make:  "Canon",
		Model: "7E",
		Type:  Interchangeable,
		CompatibleLenses: []Lens{
			{Name: "Canon 50mm f/1.8 LTM", FocalLength: 50, MaxAperture: 1.8},
			{Name: "Canon 35mm f/2.8 LTM", FocalLength: 35, MaxAperture: 2.8},
		},
	},
	"Mamiya 645E": {
		Make:  "Mamiya",
		Model: "645E",
		Type:  MediumFormat,
		CompatibleLenses: []Lens{
			{Name: "Mamiya Sekor C 80mm f/2.8", FocalLength: 80, MaxAperture: 2.8},
			{Name: "Mamiya Sekor C 55mm f/2.8", FocalLength: 55, MaxAperture: 2.8},
			{Name: "Mamiya Sekor C 150mm f/4", FocalLength: 150, MaxAperture: 4.0},
		},
	},
}

var filmStocks = []FilmStock{
	{Name: "Cinestill 50D", ISO: 50, Format: "35mm"},
	{Name: "Cinestill 800T", ISO: 800, Format: "35mm"},
	{Name: "Kodak Portra 400", ISO: 400, Format: "35mm"},
	{Name: "Kodak Portra 800", ISO: 800, Format: "35mm"},
	{Name: "Kodak Gold 200", ISO: 200, Format: "35mm"},
	{Name: "Kodak Ultramax 400", ISO: 400, Format: "35mm"},
	{Name: "Kodak Vision3 250D", ISO: 250, Format: "35mm"},
	{Name: "Kodak Vision3 500T", ISO: 500, Format: "35mm"},
	// 120 format versions
	{Name: "Cinestill 50D", ISO: 50, Format: "120"},
	{Name: "Cinestill 800T", ISO: 800, Format: "120"},
	{Name: "Kodak Portra 400", ISO: 400, Format: "120"},
	{Name: "Kodak Portra 800", ISO: 800, Format: "120"},
	{Name: "Kodak Gold 200", ISO: 200, Format: "120"},
	{Name: "Kodak Ultramax 400", ISO: 400, Format: "120"},
	{Name: "Kodak Vision3 250D", ISO: 250, Format: "120"},
	{Name: "Kodak Vision3 500T", ISO: 500, Format: "120"},
}

// Validation functions
func validateEnvironment() error {
	if _, err := exec.LookPath("exiftool"); err != nil {
		return fmt.Errorf("exiftool not found in PATH. Install from https://exiftool.org")
	}
	
	cmd := exec.Command("exiftool", "-ver")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("exiftool test failed: %w", err)
	}
	
	return nil
}

func isJPEG(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg"
}

func scanDirectory(path string) ([]string, error) {
	var jpegFiles []string
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isJPEG(filePath) {
			jpegFiles = append(jpegFiles, filePath)
		}
		return nil
	})
	return jpegFiles, err
}

func validateFiles(files []string) error {
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", file)
		}
		if !isJPEG(file) {
			return fmt.Errorf("unsupported file type: %s (only JPEG supported)", file)
		}
	}
	return nil
}

// ExifTool integration
func executeExifTool(args ...string) error {
	cmd := exec.Command("exiftool", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func stripScannerMetadata(files []string) error {
	args := []string{
		"-fnumber=",
		"-shutterspeed=",
		"-iso=",
		"-make=",
		"-model=",
		"-overwrite_original",
		"-P",
	}
	args = append(args, files...)
	return executeExifTool(args...)
}

func applyFilmMetadata(camera Camera, lens Lens, film FilmStock, files []string) error {
	args := []string{
		fmt.Sprintf("-make=%s", camera.Make),
		fmt.Sprintf("-model=%s", camera.Model),
		fmt.Sprintf("-lensmodel=%s", lens.Name),
		fmt.Sprintf("-focallength=%d", lens.FocalLength),
		fmt.Sprintf("-maxaperturevalue=%.1f", lens.MaxAperture),
		fmt.Sprintf("-iso=%d", film.ISO),
		fmt.Sprintf("-usercomment=%s", film.Name),
		"-overwrite_original",
		"-P",
	}
	args = append(args, files...)
	return executeExifTool(args...)
}

// Interactive flow functions
func selectCamera() (Camera, error) {
	fmt.Println("🎞️  Film Metadata Tool")
	
	options := []string{}
	keys := []string{}
	for k, v := range cameras {
		if v.Type == Fixed {
			options = append(options, fmt.Sprintf("%s (%s)", k, v.FixedLens.Name))
		} else if v.Type == MediumFormat {
			options = append(options, fmt.Sprintf("%s (interchangeable - 120 film)", k))
		} else {
			options = append(options, fmt.Sprintf("%s (interchangeable)", k))
		}
		keys = append(keys, k)
	}
	options = append(options, "Other (manual entry)")
	keys = append(keys, "")

	var selectedIndex int
	prompt := &survey.Select{
		Message: "📷 Select Camera:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &selectedIndex); err != nil {
		return Camera{}, err
	}
	
	if selectedIndex < len(cameras) {
		return cameras[keys[selectedIndex]], nil
	}
	
	// Manual camera entry
	var make, model string
	
	makePrompt := &survey.Input{
		Message: "📝 Enter camera make:",
	}
	if err := survey.AskOne(makePrompt, &make); err != nil {
		return Camera{}, err
	}
	
	modelPrompt := &survey.Input{
		Message: "📝 Enter camera model:",
	}
	if err := survey.AskOne(modelPrompt, &model); err != nil {
		return Camera{}, err
	}
	
	return Camera{
		Make:  make,
		Model: model,
		Type:  Fixed, // Assume fixed lens for manual entry
	}, nil
}

func selectLens(camera Camera) (Lens, error) {
	if camera.Type == Fixed {
		return *camera.FixedLens, nil
	}
	
	options := []string{}
	for _, lens := range camera.CompatibleLenses {
		options = append(options, lens.Name)
	}
	options = append(options, "Add new lens", "Manual entry")
	
	var selectedIndex int
	prompt := &survey.Select{
		Message: fmt.Sprintf("🔍 Select Lens for %s:", camera.Model),
		Options: options,
	}
	if err := survey.AskOne(prompt, &selectedIndex); err != nil {
		return Lens{}, err
	}
	
	if selectedIndex < len(camera.CompatibleLenses) {
		return camera.CompatibleLenses[selectedIndex], nil
	}
	
	// Manual lens entry
	var lensName string
	
	namePrompt := &survey.Input{
		Message: "📝 Enter lens name:",
	}
	if err := survey.AskOne(namePrompt, &lensName); err != nil {
		return Lens{}, err
	}
	
	focalPrompt := &survey.Input{
		Message: "📝 Enter focal length (mm):",
	}
	var focalStr string
	if err := survey.AskOne(focalPrompt, &focalStr); err != nil {
		return Lens{}, err
	}
	
	focalLength, err := strconv.Atoi(focalStr)
	if err != nil {
		return Lens{}, fmt.Errorf("invalid focal length: %s", focalStr)
	}
	
	aperturePrompt := &survey.Input{
		Message: "📝 Enter max aperture (f-number):",
	}
	var apertureStr string
	if err := survey.AskOne(aperturePrompt, &apertureStr); err != nil {
		return Lens{}, err
	}
	
	maxAperture, err := strconv.ParseFloat(apertureStr, 64)
	if err != nil {
		return Lens{}, fmt.Errorf("invalid aperture: %s", apertureStr)
	}
	
	return Lens{
		Name:        lensName,
		FocalLength: focalLength,
		MaxAperture: maxAperture,
	}, nil
}

func selectFilmStock(camera Camera) (FilmStock, error) {
	format := "35mm"
	if camera.Type == MediumFormat {
		format = "120"
	}
	
	options := []string{}
	filtered := []FilmStock{}
	for _, film := range filmStocks {
		if film.Format == format {
			if format == "120" {
				options = append(options, fmt.Sprintf("%s (%s)", film.Name, format))
			} else {
				options = append(options, film.Name)
			}
			filtered = append(filtered, film)
		}
	}
	options = append(options, "Other (manual entry)")
	
	var selectedIndex int
	prompt := &survey.Select{
		Message: "🎬 Select Film Stock:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &selectedIndex); err != nil {
		return FilmStock{}, err
	}
	
	if selectedIndex < len(filtered) {
		return filtered[selectedIndex], nil
	}
	
	// Manual film entry
	var filmName string
	
	namePrompt := &survey.Input{
		Message: "📝 Enter film name:",
	}
	if err := survey.AskOne(namePrompt, &filmName); err != nil {
		return FilmStock{}, err
	}
	
	isoPrompt := &survey.Input{
		Message: "📝 Enter ISO speed:",
	}
	var isoStr string
	if err := survey.AskOne(isoPrompt, &isoStr); err != nil {
		return FilmStock{}, err
	}
	
	filmISO, err := strconv.Atoi(isoStr)
	if err != nil {
		return FilmStock{}, fmt.Errorf("invalid ISO value: %s", isoStr)
	}
	
	return FilmStock{
		Name:   filmName,
		ISO:    filmISO,
		Format: format,
	}, nil
}

func confirmConfiguration(camera Camera, lens Lens, film FilmStock, fileCount int) (bool, error) {
	filmDisplay := film.Name
	if film.Format == "120" {
		filmDisplay = fmt.Sprintf("%s (%s)", film.Name, film.Format)
	}
	
	fmt.Printf("\n✅ Configuration:\n")
	fmt.Printf("   Camera: %s %s\n", camera.Make, camera.Model)
	fmt.Printf("   Lens: %s (%dmm f/%.1f)\n", lens.Name, lens.FocalLength, lens.MaxAperture)
	fmt.Printf("   Film: %s (ISO %d)\n", filmDisplay, film.ISO)
	fmt.Printf("   Files: %d JPEGs\n\n", fileCount)
	
	var confirm bool
	prompt := &survey.Confirm{
		Message: "⚠️  This will strip scanner EXIF data and add film camera metadata. Continue?",
		Default: false,
	}
	if err := survey.AskOne(prompt, &confirm); err != nil {
		return false, err
	}
	return confirm, nil
}

// Flag-based functions
func findCameraByName(name string) (Camera, error) {
	if camera, exists := cameras[name]; exists {
		return camera, nil
	}
	return Camera{}, fmt.Errorf("camera not found: %s", name)
}

func findLensByName(camera Camera, name string) (Lens, error) {
	if camera.Type == Fixed {
		return *camera.FixedLens, nil
	}
	
	for _, lens := range camera.CompatibleLenses {
		if lens.Name == name {
			return lens, nil
		}
	}
	return Lens{}, fmt.Errorf("lens not found: %s", name)
}

func findFilmByName(name string, format string) (FilmStock, error) {
	for _, film := range filmStocks {
		if film.Name == name && film.Format == format {
			return film, nil
		}
	}
	return FilmStock{}, fmt.Errorf("film not found: %s (%s)", name, format)
}

// Main command functions
func runInteractiveMode(path string) error {
	files, err := scanDirectory(path)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no JPEG files found in %s", path)
	}
	
	fmt.Printf("📁 Found %d JPEG files in %s\n\n", len(files), path)
	
	camera, err := selectCamera()
	if err != nil {
		return err
	}
	
	lens, err := selectLens(camera)
	if err != nil {
		return err
	}
	
	film, err := selectFilmStock(camera)
	if err != nil {
		return err
	}
	
	confirmed, err := confirmConfiguration(camera, lens, film, len(files))
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Operation cancelled.")
		return nil
	}
	
	return processFiles(files, camera, lens, film)
}

func runFlagMode(flags CLIFlags, path string) error {
	files := []string{}
	var err error
	
	if flags.FilePath != "" {
		files = []string{flags.FilePath}
	} else {
		files, err = scanDirectory(path)
		if err != nil {
			return err
		}
	}
	
	if len(files) == 0 {
		return fmt.Errorf("no JPEG files found")
	}
	
	camera, err := findCameraByName(flags.Camera)
	if err != nil {
		return err
	}
	
	lens, err := findLensByName(camera, flags.Lens)
	if err != nil {
		return err
	}
	
	format := "35mm"
	if camera.Type == MediumFormat {
		format = "120"
	}
	
	film, err := findFilmByName(flags.Film, format)
	if err != nil {
		return err
	}
	
	fmt.Printf("🎞️  Film Metadata Tool\n")
	fmt.Printf("📁 Found %d JPEG files\n\n", len(files))
	
	confirmed, err := confirmConfiguration(camera, lens, film, len(files))
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Operation cancelled.")
		return nil
	}
	
	return processFiles(files, camera, lens, film)
}

func runCleanMode(path string) error {
	files, err := scanDirectory(path)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no JPEG files found in %s", path)
	}
	
	fmt.Printf("🧹 Clean Mode: Strip scanner EXIF data only\n")
	fmt.Printf("📁 Found %d JPEG files in %s\n\n", len(files), path)
	
	var confirm bool
	prompt := &survey.Confirm{
		Message: "⚠️  This will remove scanner make/model, f-stop, shutter speed, and ISO. Continue?",
		Default: false,
	}
	if err := survey.AskOne(prompt, &confirm); err != nil {
		return err
	}
	if !confirm {
		fmt.Println("Operation cancelled.")
		return nil
	}
	
	fmt.Println("🔄 Processing...")
	if err := stripScannerMetadata(files); err != nil {
		return err
	}
	fmt.Printf("✅ Complete! Cleaned %d files.\n", len(files))
	return nil
}

func processFiles(files []string, camera Camera, lens Lens, film FilmStock) error {
	if err := validateFiles(files); err != nil {
		return err
	}
	
	fmt.Println("🔄 Processing...")
	if err := stripScannerMetadata(files); err != nil {
		return fmt.Errorf("failed to strip metadata: %w", err)
	}
	if err := applyFilmMetadata(camera, lens, film, files); err != nil {
		return fmt.Errorf("failed to apply metadata: %w", err)
	}
	fmt.Printf("✅ Complete! Updated %d files.\n", len(files))
	return nil
}

func main() {
	if err := validateEnvironment(); err != nil {
		fmt.Printf("❌ Environment check failed: %v\n", err)
		os.Exit(1)
	}
	
	var flags CLIFlags
	
	var rootCmd = &cobra.Command{
		Use:   "filmtag [directory]",
		Short: "CLI tool to manage film photography metadata using ExifTool",
		Long: `filmtag strips scanner EXIF data and applies correct film camera metadata to JPEG files.

Examples:
  filmtag ./roll-001/                    # Interactive mode
  filmtag --clean ./scanned-negs/        # Strip scanner data only
  filmtag -camera "Contax T3" -film "Portra 400" ./roll-001/
  filmtag -f photo.jpg -camera "Minolta CLE" -lens "Leica APO-Summicron-M 35mm f/2 ASPH." -film "Portra 800"`,
		Args: func(cmd *cobra.Command, args []string) error {
			if flags.FilePath != "" {
				return nil // Single file mode
			}
			if len(args) < 1 {
				return fmt.Errorf("directory path required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			
			if flags.Clean {
				return runCleanMode(path)
			}
			
			if flags.Camera != "" {
				return runFlagMode(flags, path)
			}
			
			return runInteractiveMode(path)
		},
	}
	
	rootCmd.Flags().StringVarP(&flags.Camera, "camera", "c", "", "Camera name (e.g., 'Contax T3')")
	rootCmd.Flags().StringVarP(&flags.Lens, "lens", "l", "", "Lens name (for interchangeable lens cameras)")
	rootCmd.Flags().StringVar(&flags.Film, "film", "", "Film stock name (e.g., 'Kodak Portra 400')")
	rootCmd.Flags().StringVarP(&flags.FilePath, "file", "f", "", "Process single file instead of directory")
	rootCmd.Flags().BoolVar(&flags.Clean, "clean", false, "Strip scanner EXIF data only (no film metadata)")
	rootCmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Verbose output")
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}
}
