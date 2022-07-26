package cpu

import (
	"reflect"

	"github.com/pqkallio/nes-emulator/bus"
)

type CpuFlag uint8

// Status register flags.
const (
	Carry             CpuFlag = 0b0000_0001
	Zero              CpuFlag = 0b0000_0010
	DisableInterrupts CpuFlag = 0b0000_0100
	DecimalMode       CpuFlag = 0b0000_1000
	Break             CpuFlag = 0b0001_0000
	Unused            CpuFlag = 0b0010_0000
	Overflow          CpuFlag = 0b0100_0000
	Negative          CpuFlag = 0b1000_0000
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

func (c *Cpu) SetFlag(flag CpuFlag, value bool) {
	if value {
		c.status |= uint8(flag)
	} else {
		c.status &= ^uint8(flag)
	}
}

func (c *Cpu) GetFlag(flag CpuFlag) bool {
	return c.status&uint8(flag) != 0
}

// Addressing modes

// Accum accumulator addressing mode is represented with a one byte instruction,
// implying an operation on the accumulator.
func (c *Cpu) Accum() uint8 {
	return 0
}

// ImmAddr the second byte of the instruction contains the operand.
func (c *Cpu) ImmAddr() uint8 {
	c.pc++
	c.absoluteAddr = c.pc

	return 0
}

// AbsAddr the second byte of the instruction contains the eight lower bits of
// the effective address and the third byte contains the eight higher bits.
func (c *Cpu) AbsAddr() uint8 {
	lo := uint16(c.bus.ReadData(c.pc))
	c.pc++
	hi := uint16(c.bus.ReadData(c.pc)) << 8
	c.pc++

	c.absoluteAddr = lo | hi

	return 0
}

// ZeroPageAddr the second byte contains the eight lower bits of the effective
// address. The higher order bits are assumed to be zero.
func (c *Cpu) ZeroPageAddr() uint8 {
	return c.indexedZeroPageAddr(0)
}

func (c *Cpu) IndexedZeroPageAddrX() uint8 {
	return c.indexedZeroPageAddr(c.xReg)
}

func (c *Cpu) IndexedZeroPageAddrY() uint8 {
	return c.indexedZeroPageAddr(c.yReg)
}

func (c *Cpu) indexedZeroPageAddr(idx uint8) uint8 {
	c.absoluteAddr = uint16(c.bus.ReadData(c.pc) + idx)
	c.pc++
	c.absoluteAddr &= 0x00FF

	return 0
}

func (c *Cpu) IndexedAbsAddrX() uint8 {
	return c.indexedAbsAddr(c.xReg)
}

func (c *Cpu) IndexedAbsAddrY() uint8 {
	return c.indexedAbsAddr(c.yReg)
}

func (c *Cpu) indexedAbsAddr(idx uint8) uint8 {
	lo := uint16(c.bus.ReadData(c.pc))
	c.pc++
	hi := uint16(c.bus.ReadData(c.pc)) << 8
	c.pc++

	c.absoluteAddr = (lo | hi) + uint16(idx)

	switch {
	case hi != (c.absoluteAddr & 0xff00):
		return 1
	default:
		return 0
	}
}

func (c *Cpu) ImpliedAddr() uint8 {
	return 0
}

func (c *Cpu) RelativeAddr() uint8 {
	jump := c.bus.ReadData(c.pc)

	relativeAddr := uint16(jump)
	if jump&signMask != 0 {
		// If the 8-bit value is to be interpreted as a negative number,
		// the high byte should be all ones (2's complement).
		relativeAddr = 0xff00 & relativeAddr
	}

	c.relativeAddr = relativeAddr

	return 0
}

func (c *Cpu) IndexedIndirectAddr() uint8 {
	ptr := uint16(c.bus.ReadData(c.pc))
	c.pc++

	loPtr := (ptr + uint16(c.xReg)) & 0x00ff
	hiPtr := (ptr + uint16(c.xReg) + 1) & 0x00ff

	addr := uint16(loPtr) | uint16(hiPtr)<<8

	c.absoluteAddr = addr

	return 0
}

func (c *Cpu) IndirectIndexedAddr() uint8 {
	ptr := uint16(c.bus.ReadData(c.pc))
	c.pc++

	ptrLo := uint16(c.bus.ReadData(ptr & 0x00ff))
	ptrHi := uint16(c.bus.ReadData((ptr+1)&0x00ff)) << 8

	c.absoluteAddr = (ptrLo | ptrHi) + uint16(c.yReg)

	if c.absoluteAddr&0xff00 != ptrHi {
		return 1
	}

	return 0
}

func (c *Cpu) AbsIndirectAddr() uint8 {
	ptrLo := uint16(c.bus.ReadData(c.pc))
	c.pc++
	ptrHi := uint16(c.bus.ReadData(c.pc)) << 8
	c.pc++

	ptr := ptrLo | ptrHi

	var addrLo, addrHi uint16

	switch ptrLo {
	case 0x00ff:
		// Simulate a hardware bug when crossing page boundaries.
		addrLo = uint16(c.bus.ReadData(ptr))
		addrHi = uint16(c.bus.ReadData(ptr&0xff00)) << 8
	default:
		addrLo = uint16(c.bus.ReadData(ptr))
		addrHi = uint16(c.bus.ReadData(ptr+1)) << 8
	}

	c.absoluteAddr = addrHi | addrLo

	return 0
}

// Opcodes

func (c *Cpu) Adc() uint8 {
	c.fetchData()
	data := c.fetchedData

	carry := uint16(0)
	if c.GetFlag(Carry) {
		carry = 1
	}

	result := uint16(c.aReg) + uint16(data) + carry

	// Set status flags
	c.SetFlag(Carry, result > 0xff)                      // if the result didn't fit in one byte, set the carry flag
	c.SetFlag(Zero, result&0x00ff == 0)                  // if the result is zero, set the zero flag
	c.SetFlag(Negative, result&uint16(signMask) != 0x00) // if the leftmost bit in the result is set, set the negative flag
	// if the accumulator and operand are negative and the result is positive,
	// or if the accumulator and operand are positive and the result is negative, set the overflow flag
	c.SetFlag(
		Overflow,
		(c.aReg&signMask != 0 && data&signMask != 0 && result&uint16(signMask) == 0) ||
			(c.aReg&signMask == 0 && data&signMask == 0 && result&uint16(signMask) != 0),
	)

	c.aReg = uint8(result & 0x00ff)

	return 1
}

func (c *Cpu) And() uint8 {
	c.fetchData()
	data := c.fetchedData

	c.aReg = c.aReg & data
	c.SetFlag(Zero, c.aReg == 0x00)
	c.SetFlag(Negative, c.aReg&signMask != 0x00)

	return 1
}

func (c *Cpu) Asl() uint8 {
	c.fetchData()
	data := c.fetchedData

	c.SetFlag(Carry, data&signMask != 0)
	data <<= 1
	c.SetFlag(Zero, data == 0)
	c.SetFlag(Negative, data&signMask != 0)

	c.aReg = data

	return 0
}

func (c *Cpu) Bit() uint8 {
	c.fetchData()
	data := c.fetchedData

	c.SetFlag(Zero, c.aReg&data == 0)
	c.SetFlag(Negative, data&uint8(Negative) != 0)
	c.SetFlag(Overflow, data&uint8(Overflow) != 0)

	return 0
}

func (c *Cpu) Bmi() uint8 {
	return c.branchOn(c.GetFlag(Negative))
}

func (c *Cpu) Bne() uint8 {
	return c.branchOn(!c.GetFlag(Zero))
}

func (c *Cpu) Bpl() uint8 {
	return c.branchOn(!c.GetFlag(Negative))
}

func (c *Cpu) Brk() uint8 {
	c.SetFlag(Break, true)
	c.pc++
	return 0
}

func (c *Cpu) Bvc() uint8 {
	return c.branchOn(!c.GetFlag(Overflow))
}

func (c *Cpu) Bvs() uint8 {
	return c.branchOn(c.GetFlag(Overflow))
}

func (c *Cpu) Clc() uint8 {
	c.SetFlag(Carry, false)
	return 0
}

func (c *Cpu) Cld() uint8 {
	c.SetFlag(DecimalMode, false)
	return 0
}

func (c *Cpu) Cli() uint8 {
	c.SetFlag(DisableInterrupts, false)
	return 0
}

func (c *Cpu) Clv() uint8 {
	c.SetFlag(Overflow, false)
	return 0
}

func (c *Cpu) Ora() uint8 {
	data := c.bus.ReadData(c.absoluteAddr)
	c.aReg = c.aReg | data
	return 0
}

func (c *Cpu) Php() uint8 {
	panic("not implemented")
}

func (c *Cpu) Jsr() uint8 {
	panic("not implemented")
}

func (c *Cpu) Rol() uint8 {
	panic("not implemented")
}

func (c *Cpu) Plp() uint8 {
	panic("not implemented")
}

func (c *Cpu) Sec() uint8 {
	c.SetFlag(Carry, true)
	return 0
}

func (c *Cpu) Rti() uint8 {
	panic("not implemented")
}

func (c *Cpu) Eor() uint8 {
	panic("not implemented")
}

func (c *Cpu) Lsr() uint8 {
	panic("not implemented")
}

func (c *Cpu) Pha() uint8 {
	c.bus.WriteData(stackBase+uint16(c.sp), c.aReg)
	c.sp--
	return 0
}

func (c *Cpu) Pla() uint8 {
	c.sp++
	c.aReg = c.bus.ReadData(stackBase + uint16(c.sp))

	c.SetFlag(Zero, c.aReg == 0x00)
	c.SetFlag(Negative, c.aReg&signMask != 0x00)

	return 0
}

func (c *Cpu) Jmp() uint8 {
	panic("not implemented")
}

func (c *Cpu) Rts() uint8 {
	panic("not implemented")
}

func (c *Cpu) Sbc() uint8 {
	// The substraction for unsigned numbers can be achieved with the following formula:
	// accum + ~memory + (1 - carry)
	m := ^uint16(c.bus.ReadData(c.absoluteAddr))

	carry := uint16(1)
	if c.GetFlag(Carry) {
		// When substracting, if the carry flag has been set, it means that a borrow has occured.
		carry = 0
	}

	result := uint16(c.aReg) + m + carry

	// Set status flags
	c.SetFlag(Carry, result > 0xff)                      // if the result didn't fit in one byte, set the carry flag
	c.SetFlag(Zero, result&0x00ff == 0)                  // if the result is zero, set the zero flag
	c.SetFlag(Negative, result&uint16(signMask) != 0x00) // if the leftmost bit in the result is set, set the negative flag
	// if the accumulator and operand are negative and the result is positive,
	// or if the accumulator and operand are positive and the result is negative, set the overflow flag
	c.SetFlag(
		Overflow,
		(c.aReg&signMask != 0 && m&uint16(signMask) != 0 && result&uint16(signMask) == 0) ||
			(c.aReg&signMask == 0 && m&uint16(signMask) == 0 && result&uint16(signMask) != 0),
	)

	c.aReg = uint8(result & 0x00ff)

	return 1
}

func (c *Cpu) Ror() uint8 {
	panic("not implemented")
}

func (c *Cpu) Sei() uint8 {
	c.SetFlag(DisableInterrupts, true)
	return 0
}

func (c *Cpu) Sta() uint8 {
	panic("not implemented")
}

func (c *Cpu) Stx() uint8 {
	panic("not implemented")
}

func (c *Cpu) Sty() uint8 {
	panic("not implemented")
}

func (c *Cpu) Dey() uint8 {
	panic("not implemented")
}

func (c *Cpu) Txa() uint8 {
	panic("not implemented")
}

func (c *Cpu) Bcc() uint8 {
	return c.branchOn(!c.GetFlag(Carry))
}

func (c *Cpu) Tya() uint8 {
	panic("not implemented")
}

func (c *Cpu) Txs() uint8 {
	panic("not implemented")
}

func (c *Cpu) Ldy() uint8 {
	panic("not implemented")
}

func (c *Cpu) Lda() uint8 {
	panic("not implemented")
}

func (c *Cpu) Ldx() uint8 {
	panic("not implemented")
}

func (c *Cpu) Tay() uint8 {
	panic("not implemented")
}

func (c *Cpu) Tax() uint8 {
	panic("not implemented")
}

func (c *Cpu) Bcs() uint8 {
	return c.branchOn(c.GetFlag(Carry))
}

func (c *Cpu) Tsx() uint8 {
	panic("not implemented")
}

func (c *Cpu) Cpy() uint8 {
	panic("not implemented")
}

func (c *Cpu) Cmp() uint8 {
	panic("not implemented")
}

func (c *Cpu) Dec() uint8 {
	panic("not implemented")
}

func (c *Cpu) Iny() uint8 {
	panic("not implemented")
}

func (c *Cpu) Dex() uint8 {
	panic("not implemented")
}

func (c *Cpu) Cpx() uint8 {
	panic("not implemented")
}

func (c *Cpu) Inc() uint8 {
	panic("not implemented")
}

func (c *Cpu) Inx() uint8 {
	panic("not implemented")
}

func (c *Cpu) Beq() uint8 {
	return c.branchOn(c.GetFlag(Zero))
}

func (c *Cpu) Sed() uint8 {
	c.SetFlag(DecimalMode, true)
	return 0
}

func (c *Cpu) Nop() uint8 {
	panic("not implemented")
}

func (c *Cpu) Illegal() uint8 {
	panic("not implemented")
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
	panic("not implemented")
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

func (c *Cpu) branchOn(cond bool) uint8 {
	if !cond {
		return 0
	}

	c.nCycles++

	c.absoluteAddr = c.pc + c.relativeAddr

	if c.absoluteAddr&0xff00 != c.pc&0xff00 {
		c.nCycles++
	}

	c.pc = c.absoluteAddr

	return 0
}

func (c *Cpu) fetchData() {
	switch reflect.ValueOf(c.opCodeLookup[c.opCode].addr).Pointer() {
	case reflect.ValueOf(c.Accum).Pointer():
		c.fetchedData = c.aReg
	default:
		c.fetchedData = c.bus.ReadData(c.absoluteAddr)
	}
}
