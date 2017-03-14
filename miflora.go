package miflora

import (
	"bytes"
	"encoding/hex"
	"errors"
	"os/exec"
	"strings"
)

type Miflora struct {
	mac     string
	adapter string
}

func NewMiflora(mac string, adapter string) *Miflora {
	return &Miflora{
		mac:     mac,
		adapter: adapter,
	}
}

func gattCharRead(mac string, handle string, adapter string) ([]byte, error) {
	cmd := exec.Command("gatttool", "-b", mac, "--char-read", "-a", handle, "-i", adapter)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	// Characteristic value/descriptor: 64 10 32 2e 36 2e 32
	s := out.String()
	if !strings.HasPrefix(s, "Characteristic value/descriptor: ") {
		return nil, errors.New("Unexpected response")
	}

	// Decode the hex bytes
	r := strings.NewReplacer(" ", "", "\n", "")
	s = r.Replace(s[33:])
	h, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return h, nil
}

type Firmware struct {
	Version string
	Battery byte
}

func (m *Miflora) ReadFirmware() (Firmware, error) {
	data, err := gattCharRead(m.mac, "0x38", m.adapter)
	if err != nil {
		return Firmware{}, err
	}
	f := Firmware{
		Version: string(data[2:]),
		Battery: data[0],
	}
	return f, nil
}

type Sensors struct {
	Temperature  float64
	Moisture     byte
	Light        uint16
	Conductivity uint16
}

func (m *Miflora) ReadSensors() (Sensors, error) {
	data, err := gattCharRead(m.mac, "0x35", m.adapter)
	if err != nil {
		return Sensors{}, err
	}
	s := Sensors{
		Temperature:  float64(uint16(data[1])*256+uint16(data[0])) / 10,
		Moisture:     data[7],
		Light:        uint16(data[4])*256 + uint16(data[3]),
		Conductivity: uint16(data[9])*256 + uint16(data[8]),
	}
	return s, nil
}
