package cpu

import (
	"reflect"

	"github.com/pqkallio/nes-emulator/bus"
)

type cpuFlag uint8

// Status register flags.
const (
	carryFlag             cpuFlag = 0b0000_0001
	zeroFlag              cpuFlag = 0b0000_0010
	disableInterruptsFlag cpuFlag = 0b0000_0100
	decimalModeFlag       cpuFlag = 0b0000_1000
	breakFlag             cpuFlag = 0b0001_0000
	unusedFlag            cpuFlag = 0b0010_0000
	overflowFlag          cpuFlag = 0b0100_0000
	negativeFlag          cpuFlag = 0b1000_0000
)

const (
	stackBase uint16 = 0x0100
	signMask  uint8  = 0x80
)

type Cpu struct {
	aReg         uint8
	xReg         uint8
	yReg         uint8
	sp           uint8
	pc           uint16
	status       uint8
	nCycles      uint8
	absoluteAddr uint16
	relativeAddr uint16
	fetchedData  uint8
	opCode       uint8
	opCodeLookup [256]instruction
	bus          *bus.Bus
}

func NewCpu(bus *bus.Bus) *Cpu {
	cpu := &Cpu{bus: bus}

	setOpCodeLookups(cpu)

	return cpu
}

func (c *Cpu) setFlag(flag cpuFlag, value bool) {
	if value {
		c.status |= uint8(flag)
	} else {
		c.status &= ^uint8(flag)
	}
}

func (c *Cpu) getFlag(flag cpuFlag) bool {
	return c.status&uint8(flag) != 0
}

// Tick is called by the emulator to advance the CPU by one cycle.
func (c *Cpu) Tick() {
	if c.nCycles == 0 {
		opCode := c.bus.ReadData(c.pc)
		c.opCode = opCode
		c.pc++

		instruction := c.opCodeLookup[opCode]
		c.nCycles = instruction.nCycles

		additionalAddrCycles := instruction.addr()
		additionalOpCycles := instruction.op()

		if additionalAddrCycles != 0 && additionalOpCycles != 0 {
			c.nCycles += additionalAddrCycles + additionalOpCycles
		}

	}

	c.nCycles--
}

// Reset resets the CPU.
func (c *Cpu) Reset() {
	c.aReg = 0
	c.xReg = 0
	c.yReg = 0

	c.sp = 0xfd
	c.status = 0x00 | uint8(unusedFlag)

	// Fetch the address of the first instruction from memory location 0xfffc
	c.absoluteAddr = 0xfffc
	addr := uint16(c.bus.ReadData(c.absoluteAddr))
	addr |= uint16(c.bus.ReadData(c.absoluteAddr+1)) << 8
	c.pc = addr

	c.relativeAddr = 0
	c.absoluteAddr = 0
	c.fetchedData = 0

	c.nCycles = 8
}

// Nmi is called when the cpu receives a non-maskable interrupt.
// A nmi is always handled.
func (c *Cpu) Nmi() {
	panic("not implemented")
}

// Irq is called to request an interrupt.
// If the interrupts are disabled, the interrupt is ignored.
func (c *Cpu) Irq() {
	panic("not implemented")
}

func (c *Cpu) fetchData() {
	switch reflect.ValueOf(c.opCodeLookup[c.opCode].addr).Pointer() {
	case reflect.ValueOf(c.accumulator).Pointer():
		c.fetchedData = c.aReg
	default:
		c.fetchedData = c.bus.ReadData(c.absoluteAddr)
	}
}
