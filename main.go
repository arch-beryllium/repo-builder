package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mholt/archiver/v3"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
)

var wantedManjaroPackages = []string{
	"attica-git",
	"bluez-qt-git",
	"breeze-git",
	"breeze-icons-git",
	"frameworkintegration-git",
	"kactivities-git",
	"kactivities-stats-git",
	"kactivitymanagerd-git",
	"karchive-git",
	"kauth-git",
	"kbookmarks-git",
	"kcmutils-git",
	"kcodecs-git",
	"kcompletion-git",
	"kconfig-git",
	"kconfigwidgets-git",
	"kcoreaddons-git",
	"kcrash-git",
	"kdbusaddons-git",
	"kde-cli-tools-git",
	"kdeclarative-git",
	"kdecoration-git",
	"kded-git",
	"kdelibs4support-git",
	"kdesu-git",
	"kemoticons-git",
	"kglobalaccel-git",
	"kguiaddons-git",
	"kholidays-git",
	"ki18n-git",
	"kiconthemes-git",
	"kidletime-git",
	"kinit-git",
	"kio-git",
	"kirigami2-git",
	"kitemmodels-git",
	"kitemviews-git",
	"kjobwidgets-git",
	"knewstuff-git",
	"knotifications-git",
	"knotifyconfig-git",
	"kpackage-git",
	"kparts-git",
	"kpeople-git",
	"kpty-git",
	"kquickcharts-git",
	"krunner-git",
	"kscreenlocker-git",
	"kservice-git",
	"ktexteditor-git",
	"ktextwidgets-git",
	"kunitconversion-git",
	"kuserfeedback-git",
	"kwallet-git",
	"kwayland-git",
	"kwayland-integration-git",
	"kwayland-server-git",
	"kwidgetsaddons-git",
	"kwin-git",
	"kwindowsystem-git",
	"kxmlgui-git",
	"libkscreen-git",
	"libksysguard-git",
	"libofono-qt",
	"libqofono-qt5",
	"milou-git",
	"modemmanager-qt-git",
	"networkmanager-qt-git",
	"packagekit-qt5",
	"plasma-framework-git",
	"plasma-integration-git",
	"plasma-mobile-settings",
	"plasma-nano-git",
	"plasma-phone-components-git",
	"plasma-wayland-session-git",
	"plasma-workspace-git",
	"prison-git",
	"qqc2-desktop-style-git",
	"qt5-pim-git",
	"qt5-feedback",
	"solid-git",
	"sonnet-git",
	"syntax-highlighting-git",
	"threadweaver-git",
}

func main() {
	checkRoot()
	downloadManjaroPackages()
	buildCustomPackages()
}

func checkRoot() {
	stdout, err := exec.Command("ps", "-o", "user=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if strings.ReplaceAll(string(stdout), "\n", "") != "root" {
		fmt.Println("This program must run as root")
		os.Exit(1)
	}
}

func addPackage(repo, fileName string) {
	repoAddCmd := exec.Command("repo-add", "-R", "-n", "-p", fmt.Sprintf("%s.db.tar.xz", repo), fileName)
	repoAddCmd.Stdout = os.Stdout
	repoAddCmd.Stderr = os.Stderr
	repoAddCmd.Dir = filepath.Join("repo", repo, "aarch64")
	err := repoAddCmd.Run()
	if err != nil {
		fmt.Printf("Failed to run repo-add: %v\n", err)
		os.Exit(1)
	}
}

func buildCustomPackages() {
	rootfsURL := "http://de3.mirror.archlinuxarm.org/os/ArchLinuxARM-aarch64-latest.tar.gz"
	dirPath := filepath.Join("repo", "beryllium", "aarch64")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			fmt.Printf("Failed to create %s: %v\n", dirPath, err)
			os.Exit(1)
		}
	}

	if _, err := os.Stat(path.Base(rootfsURL)); os.IsNotExist(err) {
		err = downloadFile(path.Base(rootfsURL), rootfsURL)
		if err != nil {
			fmt.Printf("Failed to download rootfs: %v\n", err)
			os.Exit(1)
		}
		if _, err = os.Stat("rootfs"); err == nil {
			err = os.RemoveAll("rootfs")
			if err != nil {
				fmt.Printf("Failed to remove rootfs: %v\n", err)
				os.Exit(1)
			}
		}
		err = archiver.Unarchive(path.Base(rootfsURL), "rootfs")
		if err != nil {
			fmt.Printf("Failed to read %s: %v\n", path.Base(rootfsURL), err)
			os.Exit(1)
		}
		err = ioutil.WriteFile(filepath.Join("rootfs", "etc", "pacman.d", "mirrorlist"), []byte("Server = http://localhost:8080/$repo/$arch"), 0755)
		if err != nil {
			fmt.Printf("Failed to write mirrorlist: %v\n", err)
			os.Exit(1)
		}
		err = copy.Copy("initial_setup", filepath.Join("rootfs", "initial_setup"))
		if err != nil {
			fmt.Printf("Failed to copy initial_setup: %v\n", err)
			os.Exit(1)
		}
		err = os.Chmod(filepath.Join("rootfs", "initial_setup"), 0755)
		if err != nil {
			fmt.Printf("Failed to chmod initial_setup: %v\n", err)
			os.Exit(1)
		}
		chroot("/initial_setup")
	}

	err := copy.Copy("build", filepath.Join("rootfs", "build"))
	if err != nil {
		fmt.Printf("Failed to copy build: %v\n", err)
		os.Exit(1)
	}
	err = os.Chmod(filepath.Join("rootfs", "build"), 0755)
	if err != nil {
		fmt.Printf("Failed to chmod build: %v\n", err)
		os.Exit(1)
	}
	chroot("/build")
	for _, pkgName := range []string{
		"alsa-ucm-beryllium",
		"pd-mapper-git",
		"qrtr-git",
		"rmtfs-git",
		"tqftpserv-git",
	} {
		fileName := ""
		pkgPath := filepath.Join("rootfs", "pkgs", pkgName)
		err = filepath.Walk(pkgPath, func(p string, info os.FileInfo, err error) error {
			if strings.HasPrefix(p, filepath.Join(pkgPath, pkgName)) && strings.HasSuffix(p, ".pkg.tar.xz") {
				fileName = path.Base(p)
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Failed to list files in %s: %v\n", pkgPath, err)
			os.Exit(1)
		}
		err = os.Rename(filepath.Join(pkgPath, fileName), filepath.Join("repo", "beryllium", "aarch64", fileName))
		if err != nil {
			fmt.Printf("Failed to move %s: %v\n", fileName, err)
			os.Exit(1)
		}
		addPackage("beryllium", fileName)
	}
}

func chroot(cmd string) {
	chrootCmd := exec.Command("bash", "-c", fmt.Sprintf("./do_chroot %s", cmd))
	chrootCmd.Stdout = os.Stdout
	chrootCmd.Stderr = os.Stderr
	err := chrootCmd.Run()
	if err != nil {
		fmt.Printf("Failed to chroot: %v\n", err)
		os.Exit(1)
	}
}

func downloadManjaroPackages() {
	baseRepoURL := "https://ftp.halifax.rwth-aachen.de/manjaro/arm-unstable/community/aarch64"
	dbFile := "community.tar.gz"
	dirPath := filepath.Join("repo", "plasma-mobile", "aarch64")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			fmt.Printf("Failed to create %s: %v\n", dirPath, err)
			os.Exit(1)
		}
	}
	err := downloadFile(dbFile, fmt.Sprintf("%s/%s", baseRepoURL, "community.db"))
	if err != nil {
		fmt.Printf("Failed to download repo db: %v", err)
		os.Exit(1)
	}
	var tmpDir string
	tmpDir, err = ioutil.TempDir("", "arch-repo-builder-*")
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	err = archiver.Unarchive(dbFile, tmpDir)
	if err != nil {
		fmt.Printf("Failed to read %s: %v\n", dbFile, err)
		os.Exit(1)
	}
	var dirs []os.FileInfo
	dirs, err = ioutil.ReadDir(tmpDir)
	for _, dir := range dirs {
		descFilePath := filepath.Join(tmpDir, dir.Name(), "desc")
		var content []byte
		content, err = ioutil.ReadFile(descFilePath)
		if err != nil {
			fmt.Printf("Failed to read %s: %v\n", descFilePath, err)
			os.Exit(1)
		}
		fileName := ""
		pkgName := ""
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if line == "%FILENAME%" {
				fileName = lines[i+1]
			}
			if line == "%NAME%" {
				pkgName = lines[i+1]
			}
			if len(fileName) > 0 && len(pkgName) > 0 {
				break
			}
		}
		for _, pkg := range wantedManjaroPackages {
			if pkg == pkgName {
				filePath := filepath.Join(dirPath, fileName)
				fileURL := fmt.Sprintf("%s/%s", baseRepoURL, fileName)
				if _, err = os.Stat(filePath); os.IsNotExist(err) {
					err = downloadFile(filePath, fileURL)
					if err != nil {
						fmt.Printf("Failed to download %s: %v\n", fileURL, err)
						os.Exit(1)
					}
					addPackage("plasma-mobile", fileName)
				}
				break
			}
		}
	}
	err = os.RemoveAll(tmpDir)
	if err != nil {
		fmt.Printf("Failed to remove %s: %v\n", tmpDir, err)
		os.Exit(1)
	}
}

func printDownloadPercent(done chan chan struct{}, path string, expectedSize int64) {
	var completedCh chan struct{}
	for {
		fi, err := os.Stat(path)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		size := fi.Size()

		if size == 0 {
			size = 1
		}

		var percent = float64(size) / float64(expectedSize) * 100

		fmt.Printf("\033[2K\r %.0f %% / 100 %%", percent)

		if completedCh != nil {
			close(completedCh)
			return
		}

		select {
		case completedCh = <-done:
		case <-time.After(time.Second / 60):
		}
	}
}

func downloadFile(filepath string, url string) error {
	fmt.Println(url)

	start := time.Now()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	expectedSize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return errors.Wrap(err, "failed to get Content-Length header")
	}

	doneCh := make(chan chan struct{})
	go printDownloadPercent(doneCh, filepath, int64(expectedSize))

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	doneCompletedCh := make(chan struct{})
	doneCh <- doneCompletedCh
	<-doneCompletedCh

	elapsed := time.Since(start)
	fmt.Printf("\033[2K\rDownload completed in %.2fs\n", elapsed.Seconds())
	return nil
}
