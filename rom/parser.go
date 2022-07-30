package rom

import (
	"fmt"
	"math"
	"os"
)

type Header int

const (
	prgRomSizeLSB = iota + 4
	chrRomSizeLSB
	flags6
	flags7
	mapperFlags
	romSizeMSBs
	prgRamFlags
	chrRamFlags
	timingFlags
	systemTypeFlags
	miscellaneousRomFlags
	defaultExpansionDeviceFlags
)

func ParseNesFile(filepath string) (*ROM, error) {
	rom := &ROM{}

	contents, err := readFile(filepath)
	if err != nil {
		return nil, err
	}

	prgRomEndIdx, err := parsePrgRom(rom, contents)
	if err != nil {
		return nil, err
	}

	chrRomEndIdx, err := parseChrRom(rom, contents, prgRomEndIdx)
	if err != nil {
		return nil, err
	}

	rom.miscellaneousROM = contents[chrRomEndIdx:]

	rom.flags6 = contents[flags6]
	rom.flags7 = contents[flags7]
	rom.mapperFlags = contents[mapperFlags]
	rom.prgRamFlags = contents[prgRamFlags]
	rom.chrRamFlags = contents[chrRamFlags]
	rom.timingFlags = contents[timingFlags]
	rom.systemTypeFlags = contents[systemTypeFlags]
	rom.miscellaneousRomFlags = contents[miscellaneousRomFlags]
	rom.defaultExpansionDeviceFlags = contents[defaultExpansionDeviceFlags]

	return rom, nil
}

func parsePrgRom(rom *ROM, contents []byte) (int, error) {
	prgRomSizeLSB := uint16(contents[prgRomSizeLSB])
	prgRomSizeMSB := uint16(contents[romSizeMSBs] & 0x0f)

	prgRomSize := int((prgRomSizeLSB) | prgRomSizeMSB<<8)
	prgRomSizeLarge := 0.0

	if prgRomSizeMSB == 0x0f {
		exponent := float64(prgRomSizeLSB >> 2)
		multiplier := float64(prgRomSizeLSB & 0x03)

		prgRomSizeLarge = math.Pow(2, exponent+1) * (multiplier*2 + 1)
	}

	if prgRomSizeLarge != 0.0 {
		return 0, fmt.Errorf("PRG ROM size is too large, not supported: %f", prgRomSizeLarge)
	}

	prgRomOffset := uint16(16)
	trainerAreaPresent := contents[flags6]&0x04 != 0

	if trainerAreaPresent {
		prgRomOffset += 512
	}

	prgRomEnd := int(prgRomOffset) + prgRomSize*16384

	prgRom := contents[prgRomOffset:prgRomEnd]

	rom.prgROM = prgRom

	return prgRomEnd, nil
}

func parseChrRom(rom *ROM, contents []byte, prgRomEndIdx int) (int, error) {
	chrRomSizeLSB := uint16(contents[chrRomSizeLSB])
	chrRomSizeMSB := uint16(contents[romSizeMSBs] & 0xf0)

	chrRomSize := int(chrRomSizeLSB | chrRomSizeMSB<<4)
	chrRomSizeLarge := 0.0

	if chrRomSizeMSB == 0xf0 {
		exponent := float64(chrRomSizeLSB >> 2)
		multiplier := float64(chrRomSizeLSB & 0x03)

		chrRomSizeLarge = math.Pow(2, exponent+1) * (multiplier*2 + 1)
	}

	if chrRomSizeLarge != 0.0 {
		return 0, fmt.Errorf("CHR ROM size is too large, not supported: %f", chrRomSizeLarge)
	}

	chrRomEnd := prgRomEndIdx + chrRomSize*8192

	chrRom := contents[prgRomEndIdx:chrRomEnd]

	rom.chrROM = chrRom

	return chrRomEnd, nil
}

func readFile(filepath string) ([]byte, error) {
	romFile, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	if len(romFile) < 20 {
		return nil, fmt.Errorf("ROM file is too small: %d bytes", len(romFile))
	}

	// Check that the files contents start with an ASCII the string "NES" + <EOF>.
	if string(romFile[:4]) != "NES\x1a" {
		return nil, fmt.Errorf("ROM file is not a valid NES file")
	}

	if (romFile[flags7]>>2)&0x03 != 0x10 {
		return nil, fmt.Errorf("only ROM files of type NES 2.0 are supported at the moment")
	}

	return romFile, nil
}
