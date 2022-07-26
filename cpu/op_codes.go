package cpu

func (c *Cpu) adc() uint8 {
	c.fetchData()
	data := c.fetchedData

	carry := uint16(0)
	if c.getFlag(carryFlag) {
		carry = 1
	}

	result := uint16(c.aReg) + uint16(data) + carry

	// Set status flags
	c.setFlag(carryFlag, result > 0xff)                      // if the result didn't fit in one byte, set the carry flag
	c.setFlag(zeroFlag, result&0x00ff == 0)                  // if the result is zero, set the zero flag
	c.setFlag(negativeFlag, result&uint16(signMask) != 0x00) // if the leftmost bit in the result is set, set the negative flag
	// if the accumulator and operand are negative and the result is positive,
	// or if the accumulator and operand are positive and the result is negative, set the overflow flag
	c.setFlag(
		overflowFlag,
		(c.aReg&signMask != 0 && data&signMask != 0 && result&uint16(signMask) == 0) ||
			(c.aReg&signMask == 0 && data&signMask == 0 && result&uint16(signMask) != 0),
	)

	c.aReg = uint8(result & 0x00ff)

	return 1
}

func (c *Cpu) and() uint8 {
	c.fetchData()
	data := c.fetchedData

	c.aReg = c.aReg & data
	c.setFlag(zeroFlag, c.aReg == 0x00)
	c.setFlag(negativeFlag, c.aReg&signMask != 0x00)

	return 1
}

func (c *Cpu) asl() uint8 {
	c.fetchData()
	data := c.fetchedData

	c.setFlag(carryFlag, data&signMask != 0)
	data <<= 1
	c.setFlag(zeroFlag, data == 0)
	c.setFlag(negativeFlag, data&signMask != 0)

	c.aReg = data

	return 0
}

func (c *Cpu) bit() uint8 {
	c.fetchData()
	data := c.fetchedData

	c.setFlag(zeroFlag, c.aReg&data == 0)
	c.setFlag(negativeFlag, data&uint8(negativeFlag) != 0)
	c.setFlag(overflowFlag, data&uint8(overflowFlag) != 0)

	return 0
}

func (c *Cpu) bmi() uint8 {
	return c.branchOn(c.getFlag(negativeFlag))
}

func (c *Cpu) bne() uint8 {
	return c.branchOn(!c.getFlag(zeroFlag))
}

func (c *Cpu) bpl() uint8 {
	return c.branchOn(!c.getFlag(negativeFlag))
}

func (c *Cpu) brk() uint8 {
	c.setFlag(breakFlag, true)
	c.pc++
	return 0
}

func (c *Cpu) bvc() uint8 {
	return c.branchOn(!c.getFlag(overflowFlag))
}

func (c *Cpu) bvs() uint8 {
	return c.branchOn(c.getFlag(overflowFlag))
}

func (c *Cpu) clc() uint8 {
	c.setFlag(carryFlag, false)
	return 0
}

func (c *Cpu) cld() uint8 {
	c.setFlag(decimalModeFlag, false)
	return 0
}

func (c *Cpu) cli() uint8 {
	c.setFlag(disableInterruptsFlag, false)
	return 0
}

func (c *Cpu) clv() uint8 {
	c.setFlag(overflowFlag, false)
	return 0
}

func (c *Cpu) ora() uint8 {
	data := c.bus.ReadData(c.absoluteAddr)
	c.aReg = c.aReg | data
	return 0
}

func (c *Cpu) php() uint8 {
	panic("not implemented")
}

func (c *Cpu) jsr() uint8 {
	panic("not implemented")
}

func (c *Cpu) rol() uint8 {
	panic("not implemented")
}

func (c *Cpu) plp() uint8 {
	panic("not implemented")
}

func (c *Cpu) sec() uint8 {
	c.setFlag(carryFlag, true)
	return 0
}

func (c *Cpu) rti() uint8 {
	panic("not implemented")
}

func (c *Cpu) eor() uint8 {
	panic("not implemented")
}

func (c *Cpu) lsr() uint8 {
	panic("not implemented")
}

func (c *Cpu) pha() uint8 {
	c.bus.WriteData(stackBase+uint16(c.sp), c.aReg)
	c.sp--
	return 0
}

func (c *Cpu) pla() uint8 {
	c.sp++
	c.aReg = c.bus.ReadData(stackBase + uint16(c.sp))

	c.setFlag(zeroFlag, c.aReg == 0x00)
	c.setFlag(negativeFlag, c.aReg&signMask != 0x00)

	return 0
}

func (c *Cpu) jmp() uint8 {
	panic("not implemented")
}

func (c *Cpu) rts() uint8 {
	panic("not implemented")
}

func (c *Cpu) sbc() uint8 {
	// The substraction for unsigned numbers can be achieved with the following formula:
	// accum + ~memory + (1 - carry)
	m := ^uint16(c.bus.ReadData(c.absoluteAddr))

	carry := uint16(1)
	if c.getFlag(carryFlag) {
		// When substracting, if the carry flag has been set, it means that a borrow has occured.
		carry = 0
	}

	result := uint16(c.aReg) + m + carry

	// Set status flags
	c.setFlag(carryFlag, result > 0xff)                      // if the result didn't fit in one byte, set the carry flag
	c.setFlag(zeroFlag, result&0x00ff == 0)                  // if the result is zero, set the zero flag
	c.setFlag(negativeFlag, result&uint16(signMask) != 0x00) // if the leftmost bit in the result is set, set the negative flag
	// if the accumulator and operand are negative and the result is positive,
	// or if the accumulator and operand are positive and the result is negative, set the overflow flag
	c.setFlag(
		overflowFlag,
		(c.aReg&signMask != 0 && m&uint16(signMask) != 0 && result&uint16(signMask) == 0) ||
			(c.aReg&signMask == 0 && m&uint16(signMask) == 0 && result&uint16(signMask) != 0),
	)

	c.aReg = uint8(result & 0x00ff)

	return 1
}

func (c *Cpu) ror() uint8 {
	panic("not implemented")
}

func (c *Cpu) sei() uint8 {
	c.setFlag(disableInterruptsFlag, true)
	return 0
}

func (c *Cpu) sta() uint8 {
	panic("not implemented")
}

func (c *Cpu) stx() uint8 {
	panic("not implemented")
}

func (c *Cpu) sty() uint8 {
	panic("not implemented")
}

func (c *Cpu) dey() uint8 {
	panic("not implemented")
}

func (c *Cpu) txa() uint8 {
	panic("not implemented")
}

func (c *Cpu) bcc() uint8 {
	return c.branchOn(!c.getFlag(carryFlag))
}

func (c *Cpu) tya() uint8 {
	panic("not implemented")
}

func (c *Cpu) txs() uint8 {
	panic("not implemented")
}

func (c *Cpu) ldy() uint8 {
	panic("not implemented")
}

func (c *Cpu) lda() uint8 {
	panic("not implemented")
}

func (c *Cpu) ldx() uint8 {
	panic("not implemented")
}

func (c *Cpu) tay() uint8 {
	panic("not implemented")
}

func (c *Cpu) tax() uint8 {
	panic("not implemented")
}

func (c *Cpu) bcs() uint8 {
	return c.branchOn(c.getFlag(carryFlag))
}

func (c *Cpu) tsx() uint8 {
	panic("not implemented")
}

func (c *Cpu) cpy() uint8 {
	panic("not implemented")
}

func (c *Cpu) cmp() uint8 {
	panic("not implemented")
}

func (c *Cpu) dec() uint8 {
	panic("not implemented")
}

func (c *Cpu) iny() uint8 {
	panic("not implemented")
}

func (c *Cpu) dex() uint8 {
	panic("not implemented")
}

func (c *Cpu) cpx() uint8 {
	panic("not implemented")
}

func (c *Cpu) inc() uint8 {
	panic("not implemented")
}

func (c *Cpu) inx() uint8 {
	panic("not implemented")
}

func (c *Cpu) beq() uint8 {
	return c.branchOn(c.getFlag(zeroFlag))
}

func (c *Cpu) sed() uint8 {
	c.setFlag(decimalModeFlag, true)
	return 0
}

func (c *Cpu) nop() uint8 {
	panic("not implemented")
}

func (c *Cpu) illegal() uint8 {
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
