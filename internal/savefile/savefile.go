package savefile

import "os"

type Metadata struct {
	FPS         float64
	StripLength int
}

// read the incoming data and save it to a file

func SaveFile(data [][][3]byte, fileName string, metadata Metadata) error {
	// create a new file
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	// write the metadata to the file as bytes
	_, err = f.Write([]byte{byte(metadata.FPS), byte(metadata.StripLength)})
	if err != nil {
		return err
	}

	// write the data to the file as bytes
	numberOfFrames := len(data)
	for i := 0; i < numberOfFrames; i++ {
		for j := 0; j < metadata.StripLength; j++ {
			_, err = f.Write([]byte{data[i][j][0], data[i][j][1], data[i][j][2]})
			if err != nil {
				return err
			}
		}
	}

	// return nil
	return nil

}

// read the file and return the data
func ReadFile(fileName string) ([][][3]byte, error) {
	// open the file
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// read the metadata
	metadata := make([]byte, 2)
	_, err = f.Read(metadata)
	if err != nil {
		return nil, err
	}

	// read the data
	data := make([]byte, 97*3)

	dataToRead := 97 * 3
	for dataToRead > 0 {
		n, err := f.Read(data)
		if err != nil {
			return nil, err
		}
		dataToRead -= n
	}

	// convert to [][][3]byte
	numberOfFrames := len(data) / 97 / 3
	leds := make([][][3]byte, numberOfFrames)
	for i := 0; i < numberOfFrames; i++ {
		leds[i] = make([][3]byte, 97)
		for j := 0; j < 97; j++ {
			leds[i][j] = [3]byte{data[i*97*3+j*3], data[i*97*3+j*3+1], data[i*97*3+j*3+2]}
		}
	}

	// return the data
	return leds, nil
}
