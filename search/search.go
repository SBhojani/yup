package search

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/ericm/yup/output"
)

func setColor(repo *string) {
	// Set colour for repo
	switch *repo {
	case "core":
		// Purple
		*repo = "\033[95mcore\033[0m"
		break
	case "extra":
		// Green
		*repo = "\033[32mextra\033[0m"
		break
	case "community":
		// Cyan
		*repo = "\033[36mcommunity\033[0m"
		break
	case "multilib":
		// Yellow
		*repo = "\033[33mmultilib\033[0m"
		break
	}
}

// Pacman returns []Package parsed from pacman
func Pacman(query string, print bool, installed bool) ([]output.Package, error) {

	search := exec.Command("pacman", "-Ss", query)
	run, err := search.Output()
	if err != nil {
		return []output.Package{}, err
	}

	// Find Package vals
	searchOutput := string(run)
	pacOut := []string{}
	last := ""
	for i, pac := range strings.Split(searchOutput, "\n") {
		if i%2 == 0 {
			last = pac
		} else {
			pacOut = append(pacOut, fmt.Sprintf("%s\n%s", last, pac))
		}
	}

	// Regex definitions
	nameRe := regexp.MustCompile("(?:/)+(\\S+)")
	repoRe := regexp.MustCompile("^([A-z]+)")
	versionRe := regexp.MustCompile("^(?:\\S+ ){1}(\\S+)")
	installedRe := regexp.MustCompile("\\[(.+)\\]")
	siRe := regexp.MustCompile("(?:\\:)(.+)")

	packs := []output.Package{}
	for _, pac := range pacOut {
		pack := output.Package{
			Name:        nameRe.FindString(pac)[1:],
			Repo:        repoRe.FindString(pac),
			Version:     strings.Split(versionRe.FindString(pac), " ")[1],
			Installed:   len(installedRe.FindString(pac)) != 0,
			Description: strings.Split(pac, "\n")[1][4:],
		}

		setColor(&pack.Repo)

		if installed {
			query = "="
		}
		if pack.Installed && len(query) > 0 {
			// Add extra install info
			// Get info from pacman -Sii package
			// Add extra install info
			pacmanSi := exec.Command("pacman", "-Sii", pack.Name)
			siOut, err := pacmanSi.Output()
			if err != nil {
				return []output.Package{}, output.Errorf("%s", err)
			}

			// Sets the other vals
			info := siRe.FindAllString(string(siOut), -1)
			pack.InstalledVersion = info[2][2:]
			pack.InstalledSize = info[16][2:]
			pack.DownloadSize = info[15][2:]

			// Checks if index is off and fixes it using a search
			if pack.InstalledSize == "None" {
				index := -1
				spl := strings.Split(string(siOut), "\n")
				for i, s := range spl {
					if strings.Contains(s, "Download") {
						index = len(spl) - i - 2
						break
					}
				}
				pack.InstalledSize = info[len(info)-index+1][2:]
				pack.DownloadSize = info[len(info)-index][2:]
			}

		}

		// Print
		if print {
			output.PrintPackage(pack)
		}

		if installed && pack.Installed {
			packs = append(packs, pack)
		} else if !installed {
			packs = append(packs, pack)
		}

	}

	return packs, nil

}

// PacmanSi parses Installed only from pacman -Si
func PacmanSi() ([]output.Package, error) {
	out := []output.Package{}

	pacmanSi := exec.Command("pacman", "-Si")
	siOut, err := pacmanSi.Output()
	if err != nil {
		return []output.Package{}, err
	}

	siRe := regexp.MustCompile("(?:\\:)(.+)")

	// Get each pack
	packs := strings.Split(string(siOut), "\n\n")
	for _, pack := range packs {
		parts := siRe.FindAllString(string(pack), -1)
		if len(parts) > 0 {
			// Package it into the object
			newPack := output.Package{
				Name:          parts[1][2:],
				Version:       parts[2][2:],
				Repo:          parts[0][2:],
				Description:   parts[3][2:],
				DownloadSize:  parts[len(parts)-5][2:],
				InstalledSize: parts[len(parts)-4][2:],
				Installed:     true,
			}

			setColor(&newPack.Repo)

			newPack.InstalledSizeInt = ToBytes(newPack.InstalledSize)
			out = append(out, newPack)
		}

	}

	return out, nil
}

// ToBytes Turns 1 KiB into 1024
func ToBytes(data string) int {
	valF, err := strconv.ParseFloat(data[:len(data)-4], 32)
	if err != nil {
		fmt.Fprint(os.Stderr, output.Errorf("%s", err))
	}
	val := int(valF)
	switch data[len(data)-3:] {
	case "KiB":
		return val * 1000
	case "MiB":
		return val * 1000000
	case "GiB":
		return val * 1000000000
	default:
		return -1
	}
}
