package output

import "os"

func WriteFile(path, content string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	_, err = f.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}
