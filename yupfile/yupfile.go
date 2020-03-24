package yupfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ericm/yup/output"
	"github.com/ericm/yup/sync"
)

// Pack represents a yupack
type Pack struct {
	name,
	version string
	aur bool
}

// Parse the yupfile
func Parse(argc string) error {
	args := strings.Split(argc, " ")
	if len(args) > 0 {
		arg := args[0]
		var file *os.File
		var err error
		defer file.Close()
		if filepath.IsAbs(arg) {
			// Absolute path
			file, err = os.OpenFile(arg, os.O_RDONLY, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			// Relative path
			dir, _ := os.Getwd()
			os.Chdir(dir)
			file, err = os.OpenFile(arg, os.O_RDONLY, os.ModePerm)
			if err != nil {
				return err
			}
		}
		b, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		data := string(b)
		packsS := strings.Split(data, "\n")
		var packs []Pack
		for _, pack := range packsS {
			if strs := strings.Split(pack, " "); len(strs) > 2 && strs[0][:2] != "//" {
				packs = append(packs, Pack{strs[0], strs[1], strs[2] == "aur"})
			}
		}

		return Install(packs)
	}

	return fmt.Errorf("Error parsing yupfile path")
}

// Install packages from the yupfile
func Install(packs []Pack) error {
	output.Printf("Installing packages from yupfile")
	refresh := exec.Command("sudo", "pacman", "-Sy")
	output.SetStd(refresh)
	if err := refresh.Run(); err != nil {
		return err
	}
	for _, pack := range packs {
		err := sync.Sync([]string{pack.name}, pack.aur, true)
		if err != nil {
			return err
		}
	}
	return nil
}
