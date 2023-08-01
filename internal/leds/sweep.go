package leds

func GenerateSweep(offset int) [][3]byte {
	ledArray := make([][3]byte, 97)
	for i := 0; i <= 97; i++ {
		ledArray[offset] = [3]byte{255, 255, 255}
	}
	return ledArray
}
