# FilmTag 🎞️

A CLI tool for managing film photography metadata using ExifTool. Strip unwanted scanner EXIF data and apply proper film camera metadata to your scanned photos.

## Why I Built This

After digitising my film photography with a DSLR setup, I was frustrated that my scanned photos contained misleading metadata from the scanning camera instead of the actual film camera used. Existing tools like LensTagger required Lightroom, but I wanted a fast, standalone CLI solution for batch processing entire rolls.

FilmTag solves this by:
- Stripping scanner metadata (f-stop, shutter speed, make/model)
- Adding correct film camera information (camera, lens, film stock)
- Processing entire directories in seconds
- Supporting both interactive and automated workflows

## Installation

### Requirements
- **[ExifTool](https://exiftool.org/)** must be installed and available in your system's PATH.
- **macOS (via Homebrew):**
  ```bash
  brew install exiftool
  ```

### Via Go (Recommended)
```bash
go install github.com/rohanpandula/filmtag@latest
```

### After installation, ensure `filmtag` is in your PATH:
```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc
```

## Usage

### Interactive Mode (Default)
The tool will guide you through selecting the camera, lens, and film stock.

To process all JPEGs and TIFFs in the current directory:
```bash
filmtag .
```

To process a specific directory:
```bash
filmtag ./roll-001-portra400/
```

### Flag Mode (For Automation)
Provide all the metadata via flags for automated scripts.

**Fixed lens camera:**
```bash
filmtag --camera "Contax T3" --film "Kodak Portra 400" ./photos/
```

**Interchangeable lens camera:**
```bash
filmtag --camera "Minolta CLE" --lens "Leica APO-Summicron-M 35mm f/2 ASPH." --film "Kodak Portra 800" ./photos/
```

**Single file:**
```bash
filmtag --file photo.jpg --camera "Contax T3" --film "Cinestill 800T"
```

### Clean Mode (Strip Only)
Remove scanner metadata without adding new film metadata.
```bash
filmtag --clean ./scanned-photos/
```

## Supported Gear

*Manual entry is always supported for custom cameras, lenses, and film stocks during the interactive flow.*

### Cameras & Lenses

#### Fixed Lens
- **Contax T3** (Carl Zeiss Sonnar T 35mm f/2.8)

#### Interchangeable Lens (35mm)
- **Minolta CLE**
  - Leica APO-Summicron-M 35mm f/2 ASPH.
  - Minolta 28mm 2.8 M-Rokkor
  - Canon 50mm f/1.4 LTM
- **Canon 7E**

#### Medium Format (120)
- **Mamiya 645E**
  - Mamiya Sekor C 80mm f/2.8
  - Mamiya Sekor C 55mm f/2.8
  - Mamiya Sekor C 150mm f/4

### Film Stocks (35mm & 120 Format)
- Cinestill 50D
- Cinestill 800T
- Kodak Portra 400
- Kodak Portra 800
- Kodak Gold 200
- Kodak Ultramax 400
- Kodak Vision3 250D
- Kodak Vision3 500T

## Example Workflow

1.  Scan your film roll into a new directory.
2.  Run `filmtag` interactively on that directory.
    ```bash
    filmtag ./wedding-roll-portra400/
    ```
3.  Select your camera: `Contax T3`
4.  Select your film stock: `Kodak Portra 400`
5.  Confirm the changes and let the tool process the files.

**Result:** All photos in the directory now have the proper film metadata.

## What It Does

**Before:** Scanner metadata (e.g., `Canon EOS R, f/2.8, 1/60s, ISO 100`)

**After:** Correct film metadata (e.g., `Contax T3, Carl Zeiss Sonnar T 35mm f/2.8, Kodak Portra 400`)

## Command Line Options

| Flag | Shorthand | Description |
| :--- | :---: | :--- |
| `--camera` | `-c` | Camera name (e.g., 'Contax T3'). |
| `--lens` | `-l` | Lens name (for interchangeable lens cameras). |
| `--film` | | Film stock name (e.g., 'Kodak Portra 400'). |
| `--file` | `-f` | Process a single file instead of a directory. |
| `--clean` | | Strip scanner EXIF data only (no film metadata). |
| `--verbose`| `-v` | Enable verbose output. |
| `--help` | `-h` | Display help information for filmtag. |


## Contributing
This was built for a personal workflow but is open to contributions. Feel free to:
- Add more cameras/lenses to the built-in database
- Suggest additional film stocks
- Report issues or suggest new features
- Improve the documentation

## License
[MIT License](https://opensource.org/licenses/MIT) - use it however you want for your film photography workflow.

