// Library for reading and parsing config files.
package conf

import (
	"encoding/gob"
	"encoding/json"
	"os"

	"github.com/golang/glog"
)

// Loads data from the gob file.
func LoadFromFile(filePath string, object interface{}) error {
	glog.V(1).Infof("Opening file %q.", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(object)
	if err != nil {
		return err
	}

	return file.Close()
}

// Saves data to the gob file.
func SaveToFile(filePath string, object interface{}) error {
	glog.V(1).Infof("Opening file %q.", filePath)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	encoder := gob.NewEncoder(file)
	encoder.Encode(object)
	return file.Close()
}

// Loads data from the JSON file.
func LoadFromJson(filePath string, object interface{}) error {
	glog.V(1).Infof("Opening file %q.", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(object)
	if err != nil {
		return err
	}

	return file.Close()
}

// Loads data from the gob file in the config directory:
// $HOME/.config/{appName}.gob
func LoadConfigFromFile(appName string, object interface{}) error {
	return LoadFromFile(getPath(appName, "gob"), object)
}

// Saves data from the gob file in the config directory:
// $HOME/.config/{appName}.gob
func SaveConfigToFile(appName string, object interface{}) error {
	return SaveToFile(getPath(appName, "gob"), object)
}

// Loads data from the JSON file in the config directory:
// $HOME/.config/{appName}.json
func LoadConfigFromJson(appName string, object interface{}) error {
	return LoadFromJson(getPath(appName, "json"), object)
}

func getPath(appName string, extension string) string {
	return os.Getenv("HOME") + "/.config/" + appName + "." + extension
}
