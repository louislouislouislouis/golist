package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/louislouislouislouis/repr8ducer/utils"
)

func writeStringFile(filePath, value string) error {
	return writeByteFile(filePath, []byte(value))
}

func writeByteFile(filePath string, value []byte) error {
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		utils.Log.Err(err).Msgf("Error creating directories for file '%s'", filePath)
		return fmt.Errorf("error creating directories for file '%s': %v", filePath, err)
	}
	err = os.WriteFile(filePath, []byte(value), 0644)
	if err != nil {
		utils.Log.
			Err(err).
			Msg(
				fmt.Sprintf(
					"Error during creation of file '%s'",
					filePath,
				),
			)
		return fmt.Errorf(
			"Error during creation of file '%s': %v",
			filePath,
			err,
		)
	}
	return nil
}
