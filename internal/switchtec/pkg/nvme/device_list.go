package nvme

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"regexp"
)

// Return the devices that match the provided regexp in Model Number, Serial Number,
// or Node Qualifying Name (NQN). Returned paths are of the form /dev/nvme[0-9]+.
func DeviceList(r string) ([]string, error) {

	deviceRegexp, err := regexp.Compile(r)
	if err != nil {
		return nil, err
	}

	// Perform an initial discovery of the /dev/nvme*... devices
	devicePath := "/dev"
	rawDeviceRegexp := regexp.MustCompile(`^nvme\d+$`)

	files, err := ioutil.ReadDir(devicePath)
	if err != nil {
		return nil, err
	}

	devices := make([]string, 0)
	for _, file := range files {
		if rawDeviceRegexp.MatchString(file.Name()) {

			// Check Serial Number or SUBNQN for the matching device regexp
			cmd := fmt.Sprintf("nvme id-ctrl %s | grep -e sn -e subnqn -e mn", path.Join(devicePath, file.Name()))
			rsp, err := exec.Command("bash", "-c", cmd).Output()
			if err != nil {
				continue
			}

			if deviceRegexp.Match(rsp) {
				devices = append(devices, path.Join(devicePath, file.Name()))
			}
		}
	}

	return devices, nil
}
