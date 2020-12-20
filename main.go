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
	"bluedevil-git",
	"bluez-qt-git",
	"breeze-git",
	"breeze-icons-git",
	"buho-git",
	"calindori-git",
	"discover-git",
	"frameworkintegration-git",
	"index-git",
	"kaccounts-integration-git",
	"kaccounts-providers-git",
	"kactivities-git",
	"kactivities-stats-git",
	"kactivitymanagerd-git",
	"kalk-git",
	"karchive-git",
	"kauth-git",
	"kbookmarks-git",
	"kcalendarcore-git",
	"kclock-git",
	"kcmutils-git",
	"kcodecs-git",
	"kcompletion-git",
	"kconfig-git",
	"kconfigwidgets-git",
	"kcontacts-git",
	"kcoreaddons-git",
	"kcrash-git",
	"kdbusaddons-git",
	"kde-cli-tools-git",
	"kdeclarative-git",
	"kdecoration-git",
	"kded-git",
	"kdelibs4support-git",
	"kdesignerplugin-git",
	"kdesu-git",
	"kdnssd-git",
	"kemoticons-git",
	"keysmith-git",
	"kglobalaccel-git",
	"kguiaddons-git",
	"kholidays-git",
	"ki18n-git",
	"kiconthemes-git",
	"kidletime-git",
	"kinit-git",
	"kio-extras-git",
	"kio-git",
	"kirigami-addons-git",
	"kirigami2-git",
	"kitemmodels-git",
	"kitemviews-git",
	"kjobwidgets-git",
	"kjs-git",
	"knewstuff-git",
	"knotifications-git",
	"knotifyconfig-git",
	"kongress-git",
	"kpackage-git",
	"kparts-git",
	"kpeople-git",
	"kpeoplesink-git",
	"kpeoplevcard-git",
	"kplotting-git",
	"kpty-git",
	"kquickcharts-git",
	"kquicksyntaxhighlighter-git",
	"krecorder-git",
	"krunner-git",
	"kscreen-git",
	"kscreenlocker-git",
	"kservice-git",
	"ktexteditor-git",
	"ktextwidgets-git",
	"kunitconversion-git",
	"kuserfeedback-git",
	"kwallet-git",
	"kwallet-pam-git",
	"kwayland-git",
	"kwayland-integration-git",
	"kwayland-server-git",
	"kweather-git",
	"kwidgetsaddons-git",
	"kwin-git",
	"kwindowsystem-git",
	"kxmlgui-git",
	"libkgapi-git",
	"libkscreen-git",
	"libksysguard-git",
	"libofono-qt",
	"libqofono-qt5",
	"manjaro-arm-keyring",
	"manjaro-keyring",
	"maliit-framework-git",
	"maliit-keyboard-git",
	"mauikit-git",
	"milou-git",
	"modemmanager-qt-git",
	"mplus-font",
	"networkmanager-qt-git",
	"nota-git",
	"ofonoctl",
	"okular-mobile-git",
	"oxygen-git",
	"plasma-angelfish-git",
	"plasma-dialer-git",
	"plasma-framework-git",
	"plasma-integration-git",
	"plasma-mobile-nm-git",
	"plasma-mobile-settings",
	"plasma-nano-git",
	"plasma-pa-git",
	"plasma-phone-components-git",
	"plasma-phonebook-git",
	"plasma-pix-git",
	"plasma-settings-git",
	"plasma-wayland-protocols-git",
	"plasma-wayland-session-git",
	"plasma-workspace-git",
	"polkit-kde-agent-git",
	"powerdevil-git",
	"presage-git",
	"prison-git",
	"purpose-git",
	"qmlkonsole-git",
	"qqc2-desktop-style-git",
	"qt5-3d",
	"qt5-base",
	"qt5-charts",
	"qt5-connectivity",
	"qt5-datavis3d",
	"qt5-declarative",
	"qt5-doc",
	"qt5-es2-base",
	"qt5-es2-declarative",
	"qt5-es2-multimedia",
	"qt5-es2-wayland",
	"qt5-es2-xcb-private-headers",
	"qt5-examples",
	"qt5-feedback",
	"qt5-gamepad",
	"qt5-graphicaleffects",
	"qt5-imageformats",
	"qt5-location",
	"qt5-lottie",
	"qt5-mqtt",
	"qt5-multimedia",
	"qt5-networkauth",
	"qt5-pim-git",
	"qt5-purchasing",
	"qt5-quick3d",
	"qt5-quickcontrols",
	"qt5-quickcontrols2",
	"qt5-quicktimeline",
	"qt5-remoteobjects",
	"qt5-script",
	"qt5-scxml",
	"qt5-sensors",
	"qt5-serialbus",
	"qt5-serialport",
	"qt5-speech",
	"qt5-svg",
	"qt5-tools",
	"qt5-translations",
	"qt5-virtualkeyboard",
	"qt5-wayland",
	"qt5-webchannel",
	"qt5-webengine",
	"qt5-webglplugin",
	"qt5-webkit",
	"qt5-websockets",
	"qt5-webview",
	"qt5-x11extras",
	"qt5-xcb-private-headers",
	"qt5-xmlpatterns",
	"signon-kwallet-extension-git",
	"solid-git",
	"sonnet-git",
	"spacebar-git",
	"syntax-highlighting-git",
	"telepathy-ofono",
	"threadweaver-git",
	"vvave-git",
	"xdg-desktop-portal-kde-git",
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
		"ofono-qrtr",
	} {
		fileName := ""
		pkgPath := filepath.Join("rootfs", "pkgs", pkgName)
		err = filepath.Walk(pkgPath, func(p string, info os.FileInfo, err error) error {
			if strings.HasSuffix(p, ".pkg.tar.xz") {
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
	baseRepoURL := "https://mirror.alpix.eu/manjaro/arm-unstable/%s/aarch64"
	for _, repo := range []string{"core", "extra", "community"} {
		dbFile := fmt.Sprintf("%s.tar.gz", repo)
		dirPath := filepath.Join("repo", "plasma-mobile", "aarch64")
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			err = os.MkdirAll(dirPath, 0755)
			if err != nil {
				fmt.Printf("Failed to create %s: %v\n", dirPath, err)
				os.Exit(1)
			}
		}
		err := downloadFile(dbFile, fmt.Sprintf("%s/%s", fmt.Sprintf(baseRepoURL, repo), fmt.Sprintf("%s.db", repo)))
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
					fileURL := fmt.Sprintf("%s/%s", fmt.Sprintf(baseRepoURL, repo), fileName)
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
