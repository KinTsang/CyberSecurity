package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"golang.org/x/sys/windows/registry"
)

var (
	boldWhite   = color.New(color.FgWhite).Add(color.Bold)
	boldBlue    = color.New(color.FgBlue).Add(color.Bold)
	boldRed     = color.New(color.FgRed).Add(color.Bold)
	boldGreen   = color.New(color.FgGreen).Add(color.Bold)
	boldYellow  = color.New(color.FgYellow).Add(color.Bold)
	boldCyan    = color.New(color.FgCyan).Add(color.Bold)
	boldMagenta = color.New(color.FgMagenta).Add(color.Bold)

	currentUser, _ = user.Current()
	username       = strings.Split(currentUser.Username, `\`)[1]
)

// ------------------------- Helper Functions --------------------------
func check(e error, errMsg string) bool {
	if e != nil {
		fmt.Println(errMsg)
		return false
	}
	return true
}

func find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func getNumberOfSubKeysAndValues(k registry.Key) (uint32, uint32) {
	keyInfo, err := k.Stat()
	_ = check(err, "Unable to fetch Stat info from registry object...")
	return keyInfo.SubKeyCount, keyInfo.ValueCount
}

func openKey(hive registry.Key, subkey string, access uint32) registry.Key {
	key, err := registry.OpenKey(hive, subkey, access)
	check(err, "Unable to open registry key...")
	return key
}

func toTime(t []byte) time.Time {
	ft := &syscall.Filetime{
		LowDateTime:  binary.LittleEndian.Uint32(t[:4]),
		HighDateTime: binary.LittleEndian.Uint32(t[4:]),
	}
	return time.Unix(0, ft.Nanoseconds())
}

// ------------------- Program Body ---------------------
func getComputerInfo() {
	key := openKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
		registry.ALL_ACCESS,
	)
	defer key.Close()

	boldBlue.Println("◎ ☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶ Computer Build Info ☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶ ◎")
	fmt.Println("")

	productName, _, err := key.GetStringValue("ProductName")
	if success := check(err, "ProductName value not found in registry..."); !success {
		return
	}
	fmt.Println("Product Name : " + productName)

	currentVersion, _, err := key.GetStringValue("CurrentVersion")
	if success := check(err, "CurrentVersion value not found in registry..."); !success {
		return
	}
	fmt.Println("Current Version : " + currentVersion)

	currentBuildNumber, _, err := key.GetStringValue("CurrentBuildNumber")
	if success := check(err, "CurrentBuildNumber Value not found in registry..."); !success {
		return
	}
	fmt.Println("Build Number : " + currentBuildNumber)

	registeredOwner, _, err := key.GetStringValue("RegisteredOwner")
	if success := check(err, "RegisteredOwner value not found in registry..."); !success {
		return
	}
	fmt.Println("Registered Owner : " + registeredOwner)
	fmt.Println("")
}

func getInstalledApps() {
	key := openKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		registry.ALL_ACCESS,
	)
	defer key.Close()

	numOfSubKeys, numOfValues := getNumberOfSubKeysAndValues(key)
	subkeys, err := key.ReadSubKeyNames(int(numOfSubKeys))
	check(err, "Unable to read subkeys...")

	boldRed.Println("◎ ☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶ Installed Applications ☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶ ◎")
	fmt.Println("")
	for _, skey := range subkeys {
		k := openKey(
			registry.LOCAL_MACHINE,
			`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`+"\\"+skey,
			registry.ALL_ACCESS,
		)
		values, err := k.ReadValueNames(int(numOfValues))
		check(err, "Unable to read values from registry key...")
		if exist := find(values, "DisplayName"); exist {
			val, _, err := k.GetStringValue("DisplayName")
			check(err, "Unable to retrieve data from value DisplayName...")
			fmt.Println("\u2022 " + val)
		} else {
			fmt.Println("\u2022 " + skey)
		}
	}
}

func getEnVars() {
	key := openKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.ALL_ACCESS,
	)
	defer key.Close()

	_, numOfValues := getNumberOfSubKeysAndValues(key)
	environmentVariables, err := key.ReadValueNames(int(numOfValues))
	check(err, "Unable to read values from registry key...")

	boldGreen.Println("\n◎ ☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶ Environment Variables ☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶ ◎")
	fmt.Println("")

	for _, envar := range environmentVariables {
		envarValue, _, err := key.GetStringValue(envar)
		check(err, "Unable to retrieve data from value in registry key...")
		fmt.Println(envar + " ☰☰ " + envarValue)
	}
	fmt.Println("")
}

func getStartUpApps() {
	/*
		Registry Run and RunOnce Keys (Run and RunOnce registry keys cause programs to run each time that a user logs on)
			HKEY_LOCAL_MACHINE\Software\Microsoft\Windows\CurrentVersion\Run
			HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run
			HKEY_LOCAL_MACHINE\Software\Microsoft\Windows\CurrentVersion\RunOnce
			HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\RunOnce
	*/

}

func getJumpLists() {
	jumpListPath := fmt.Sprintf(
		`C:\Users\%s\AppData\Roaming\Microsoft\Windows\Recent\AutomaticDestinations`,
		username,
	)
	jumpListFiles, err := ioutil.ReadDir(jumpListPath)
	check(err, "Unable to read files in jump list directory...")

	boldYellow.Println("◎ ☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶ Jump List Files ☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶ ◎")
	fmt.Println()
	for _, file := range jumpListFiles {
		fmt.Println(file.Name())
	}
	fmt.Println()
}

func getLNKFiles() {
	LNKPath := fmt.Sprintf(
		`C:\Users\%s\AppData\Roaming\Microsoft\Windows\Recent`,
		username,
	)

	LNKFiles, err := ioutil.ReadDir(LNKPath)
	check(err, "Unable to read files from LNK file path location...")

	boldMagenta.Println("◎ ☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶ LNK Files ☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶ ◎")
	fmt.Println()
	for _, file := range LNKFiles {
		fi, err := os.Open(LNKPath + `\` + file.Name())
		_ = check(err, "Unable to open LNK file...")
		o := make([]byte, 28)
		_, _ = fi.Read(o)
		creationTime := make([]byte, 8)
		_, _ = fi.Read(creationTime)
		accessTime := make([]byte, 8)
		_, _ = fi.Read(accessTime)
		writeTime := make([]byte, 8)
		_, _ = fi.Read(writeTime)
		fileSize := make([]byte, 4)
		_, _ = fi.Read(fileSize)

		fmt.Println()
		fmt.Print("LNK File Name: ")
		boldWhite.Println(file.Name())
		fmt.Print("File Created On: ")
		boldYellow.Println(toTime(creationTime).String())
		fmt.Print("File Last Accessed On: ")
		boldYellow.Println(toTime(accessTime).String())
		fmt.Print("File Last Modified On: ")
		boldYellow.Println(toTime(writeTime).String())
		fmt.Print("LNK Target File Size: ")
		boldBlue.Println(binary.LittleEndian.Uint32(fileSize))
	}
	fmt.Println()
}

func getShellBags() {

}

func getPrefetchFiles() {

}

func getRecycleBinFiles() {
	recycleBinPath := `C:\$Recycle.Bin`
	recycleBinFiles, err := ioutil.ReadDir(recycleBinPath)
	check(err, "Unable to open recycle bin folder...")
	currentUser, err := user.Current()
	check(err, "Unable to get user info...")
	userSID := currentUser.Uid

	boldCyan.Println("◎ ☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶ Recycle Bin Files ☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶☶ ◎")

	for _, recycleFolder := range recycleBinFiles {
		folderName := recycleFolder.Name()
		if folderName == userSID {
			userRecycleBinContents, err := ioutil.ReadDir(recycleBinPath + `\` + folderName)
			check(err, "Unable able to open user's recycle bin folder...")
			for _, recycledFile := range userRecycleBinContents {
				if strings.HasPrefix(recycledFile.Name(), "$I") {
					fi, err := os.Open(recycleBinPath + `\` + folderName + `\` + recycledFile.Name())
					header := make([]byte, 8)
					_, err = fi.Read(header)
					check(err, "Unable to read data from dollar I recycle file...")

					operatingSystem := "Prior to Windows 10"
					if binary.LittleEndian.Uint64(header) == 2 {
						operatingSystem = "Windows 10"
					}

					fileSize := make([]byte, 8)
					_, err = fi.Read(fileSize)
					check(err, "Unable to read data from dollar I recycle file...")

					deletedTimeStamp := make([]byte, 8)
					_, err = fi.Read(deletedTimeStamp)
					check(err, "Unable to read data from dollar I recycle file...")

					dateDeleted := toTime(deletedTimeStamp)
					//convertToDate := fmt.Sfmt.Printlnf(`[DateTime]::FromFileTimeutc("%d")`, timeStamp)
					//date, err := exec.Command("powershell.exe", `-c`, convertToDate).CombinedOutput()
					check(err, "Unable to retrieve time stamp from recycled file...")

					fileNameLength := make([]byte, 4)
					_, _ = fi.Read(fileNameLength)

					dollarIFileSize, _ := fi.Stat()
					fileName := make([]byte, (dollarIFileSize.Size() - 8 - 8 - 8 - 4))
					_, err = fi.Read(fileName)

					fmt.Println("")
					fmt.Print("File Name: ")
					boldWhite.Println(string(fileName))
					fmt.Println("OS: " + operatingSystem)
					fmt.Print("File Deleted On: ")
					boldRed.Println(dateDeleted)
					fmt.Println("File size: " + strconv.Itoa(int(binary.LittleEndian.Uint64(fileSize))))

				}
			}
		}
	}
}

func main() {
	if len(os.Args[:]) < 2 {
		boldRed.Println("Please use `-h` option to see full list of program options ...")
		os.Exit(1)
	}

	var showAll = flag.Bool("a", false, "Show all Available Information")
	var viewComputerInfo = flag.Bool("c", false, "Show Computer Set Up Information")
	var viewInstalledApps = flag.Bool("i", false, "Show Installed Applications")
	var viewEnvars = flag.Bool("e", false, "Show Environment Variables")
	var viewStartUpApps = flag.Bool("s", false, "Show Programs That Run On Start Up")
	var viewJumpLists = flag.Bool("j", false, "Show Jump List Files")
	var viewLNKFiles = flag.Bool("l", false, "Show LNK Files")
	var viewShellBags = flag.Bool("b", false, "Show Shell Bag Keys")
	var viewPrefetchFiles = flag.Bool("p", false, "Show Prefetch Files")
	var viewRecycledFiles = flag.Bool("r", false, "Show Recycled Files")
	flag.Parse()

	if *showAll {
		getComputerInfo()
		getInstalledApps()
		getEnVars()
		getStartUpApps()
		getJumpLists()
		getLNKFiles()
		getShellBags()
		getPrefetchFiles()
		getRecycleBinFiles()
		return
	}

	if *viewComputerInfo {
		getComputerInfo()
	}

	if *viewInstalledApps {
		getInstalledApps()
	}

	if *viewEnvars {
		getEnVars()
	}

	if *viewStartUpApps {
		getStartUpApps()
	}

	if *viewJumpLists {
		getJumpLists()
	}

	if *viewLNKFiles {
		getLNKFiles()
	}

	if *viewShellBags {
		getShellBags()
	}

	if *viewPrefetchFiles {
		getPrefetchFiles()
	}

	if *viewRecycledFiles {
		getRecycleBinFiles()
	}
}
