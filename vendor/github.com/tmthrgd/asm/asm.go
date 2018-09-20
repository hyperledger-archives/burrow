// Copyright 2014 Benoît Amiaux. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package asm

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Opcodes struct {
	a *Asm
}

type Asm struct {
	Opcodes

	Seperator string

	w      *bufio.Writer
	errors []string

	// per function
	name  string
	args  int
	stack int
	split bool
}

func NewAsm(w io.Writer) *Asm {
	a := &Asm{
		Seperator: " ",

		w: bufio.NewWriter(w),
	}
	a.Opcodes.a = a

	a.write("\n#include \"textflag.h\"")
	return a
}

func (a *Asm) NewFunction(name string) {
	a.name = name
	a.args = 0
	a.stack = 0
	a.split = true
}

func (a *Asm) NoSplit() {
	a.split = false
}

func isZeroSlice(s []byte) bool {
	for _, b := range s {
		if b != 0 {
			return false
		}
	}

	return true
}

type Data string

func (d Data) String() string {
	return fmt.Sprintf("%v(SB)", string(d))
}

func (Data) Gas() string {
	panic("referencing GLOBL directives in unsupported opcodes is forbidden")
}

func (d Data) Offset(i int) Data {
	if i == 0 {
		return d
	}

	return Data(fmt.Sprintf("%v+0x%02x", string(d), i))
}

func (d Data) Address() Data {
	return Data(fmt.Sprintf("$%v", string(d)))
}

func (a *Asm) Data(name string, data []byte) Data {
	name = fmt.Sprintf("%v<>", name)

	a.write("")

	i := 0
	for ; i < len(data); i += 8 {
		if isZeroSlice(data[i : i+8]) {
			continue
		}

		a.write(fmt.Sprintf("DATA%v%v+0x%02x(SB)/8, $0x%016x", a.Seperator, name, i, data[i:i+8]))
	}

	for ; i < len(data); i++ {
		if data[i] == 0 {
			continue
		}

		a.write(fmt.Sprintf("DATA%v%v+0x%02x(SB)/1, $0x%02x", a.Seperator, name, i, data[i]))
	}

	a.write(fmt.Sprintf("GLOBL%v%v(SB),RODATA,$%v", a.Seperator, name, len(data)))
	return Data(name)
}

func (a *Asm) Data16(name string, data []uint16) Data {
	name = fmt.Sprintf("%v<>", name)

	a.write("")

	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			continue
		}

		a.write(fmt.Sprintf("DATA%v%v+0x%02x(SB)/2, $0x%04x", a.Seperator, name, 2*i, data[i]))
	}

	a.write(fmt.Sprintf("GLOBL%v%v(SB),RODATA,$%v", a.Seperator, name, 2*len(data)))
	return Data(name)
}

func (a *Asm) Data32(name string, data []uint32) Data {
	name = fmt.Sprintf("%v<>", name)

	a.write("")

	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			continue
		}

		a.write(fmt.Sprintf("DATA%v%v+0x%02x(SB)/4, $0x%08x", a.Seperator, name, 4*i, data[i]))
	}

	a.write(fmt.Sprintf("GLOBL%v%v(SB),RODATA,$%v", a.Seperator, name, 4*len(data)))
	return Data(name)
}

func (a *Asm) Data64(name string, data []uint64) Data {
	name = fmt.Sprintf("%v<>", name)

	a.write("")

	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			continue
		}

		a.write(fmt.Sprintf("DATA%v%v+0x%02x(SB)/8, $0x%016x", a.Seperator, name, 8*i, data[i]))
	}

	a.write(fmt.Sprintf("GLOBL%v%v(SB),RODATA,$%v", a.Seperator, name, 8*len(data)))
	return Data(name)
}

func (a *Asm) DataString(name string, data string) Data {
	name = fmt.Sprintf("%v<>", name)

	a.write("")

	i := 0
	for ; i < len(data); i += 8 {
		if isZeroSlice([]byte(data[i : i+8])) {
			continue
		}

		a.write(fmt.Sprintf("DATA%v%v+0x%02x(SB)/8, $%q", a.Seperator, name, i, data[i:i+8]))
	}

	for ; i < len(data); i++ {
		if data[i] == 0 {
			continue
		}

		a.write(fmt.Sprintf("DATA%v%v+0x%02x(SB)/1, $%q", a.Seperator, name, i, data[i]))
	}

	a.write(fmt.Sprintf("GLOBL%v%v(SB),RODATA,$%v", a.Seperator, name, len(data)))
	return Data(name)
}

type invalid int

func (invalid) String() string {
	panic("invalid operand")
}

func (invalid) Gas() string {
	panic("invalid operand")
}

var _ Operand = invalid(0)

const Invalid = invalid(0)

type argument struct {
	name   string
	offset int
}

func (s *argument) String() string {
	return fmt.Sprintf("%v+%v(FP)", s.name, s.offset)
}

func (*argument) Gas() string {
	panic("referencing arguments in unsupported opcodes is forbidden")
}

func (a *Asm) Argument(name string, size int) Operand {
	a.args += size
	return &argument{
		name:   name,
		offset: a.args - size,
	}
}

func (a *Asm) SliceArgument(name string) []Operand {
	var rpy []Operand

	for i := 0; i < 3; i++ {
		rpy = append(rpy, a.Argument(name, 8))
	}

	return rpy
}

type stackOperand struct {
	name   string
	offset int
}

func (s *stackOperand) String() string {
	return fmt.Sprintf("%v+-%v(SP)", s.name, s.offset)
}

func (*stackOperand) Gas() string {
	panic("referencing stack variables in unsupported opcodes is forbidden")
}

func (a *Asm) PushStack(name string, size int) Operand {
	a.stack += size
	return &stackOperand{
		name:   name,
		offset: a.stack,
	}
}

func (a *Asm) Start() {
	if a.split {
		a.write(fmt.Sprintf("\nTEXT ·%v(SB),0,$%v-%v", a.name, a.stack, a.args))
	} else {
		a.write(fmt.Sprintf("\nTEXT ·%v(SB),NOSPLIT,$%v", a.name, a.stack))
	}
}

func (a *Asm) Flush() error {
	err := a.w.Flush()
	a.save(err)
	return a.getErrors()
}

func (a *Asm) save(err error) {
	if err == nil {
		return
	}

	a.errors = append(a.errors, err.Error())
}

func (a *Asm) getErrors() error {
	if len(a.errors) == 0 {
		return nil
	}

	return fmt.Errorf("%s", strings.Join(a.errors, "\n"))
}

func (a *Asm) write(msg string) {
	_, err := a.w.WriteString(msg + "\n")
	a.save(err)
}

type Operand interface {
	String() string
	Gas() string
}

type constant string

func (cons constant) String() string {
	return string(cons)
}

func (cons constant) Gas() string {
	return string(cons)
}

func Constant(value interface{}) Operand {
	return constant(fmt.Sprintf("$%v", value))
}

type Scale uint

const (
	SX0 Scale = 0
	SX1 Scale = 1 << (iota - 1)
	SX2
	SX4
	SX8
)

type addressOperand struct {
	lit string
	gas string
}

func (a addressOperand) String() string {
	return a.lit
}

func (a addressOperand) Gas() string {
	return a.gas
}

func address(base Register) Operand {
	return addressOperand{
		fmt.Sprintf("(%v)", base.String()),
		fmt.Sprintf("(%v)", base.Gas()),
	}
}

func displaceaddress(base Register, index int) Operand {
	if index == 0 {
		return address(base)
	}

	return addressOperand{
		fmt.Sprintf("%v(%v)", index, base.String()),
		fmt.Sprintf("%v(%v)", index, base.Gas()),
	}
}

func scaledindex(index Register, scale Scale) string {
	if scale == SX0 {
		return ""
	}

	return fmt.Sprintf("(%v*%v)", index.String(), scale)
}

func indexaddress(base Register, index Register, scale Scale) Operand {
	return addressOperand{
		fmt.Sprintf("(%v)%v", base.String(), scaledindex(index, scale)),
		fmt.Sprintf("(%v, %v, %v)", base.Gas(), index.Gas(), scale),
	}
}

func fulladdress(base Register, index Register, scale Scale, displacement int) Operand {
	d := ""

	if displacement != 0 {
		d = fmt.Sprintf("%v", displacement)
	}

	return addressOperand{
		fmt.Sprintf("%v(%v)%v", d, base.String(), scaledindex(index, scale)),
		fmt.Sprintf("%v(%v, %v, %v)", d, base.Gas(), index.Gas(), scale),
	}
}

func Address(base Register, offsets ...interface{}) Operand {
	// happily panics if not given expected input
	switch len(offsets) {
	case 0:
		return address(base)
	case 1:
		switch t := offsets[0].(type) {
		case int:
			return displaceaddress(base, t)
		case uint:
			return displaceaddress(base, int(t))
		case Register:
			return indexaddress(base, t, SX1)
		case Scale:
			return addressOperand{
				scaledindex(base, t),
				fmt.Sprintf("(, %v, %v)", base.String(), t),
			}
		}
	case 2:
		index, ok := offsets[0].(Register)
		if !ok {
			break
		}

		switch t := offsets[1].(type) {
		case int:
			return fulladdress(base, index, SX1, t)
		case uint:
			return fulladdress(base, index, SX1, int(t))
		case Scale:
			return indexaddress(base, index, t)
		}
	case 3:
		index, ok := offsets[0].(Register)
		if !ok {
			break
		}

		scale, ok := offsets[1].(Scale)
		if !ok {
			break
		}

		switch t := offsets[2].(type) {
		case int:
			return fulladdress(base, index, scale, t)
		case uint:
			return fulladdress(base, index, scale, int(t))
		}
	}

	panic("unexpected input")
}

type Register string

func (r Register) String() string {
	return string(r)
}

func (r Register) Gas() string {
	var prefix string
	if r[0] != 'R' && (len(r) != 2 || (r[1] != 'H' && r[1] != 'L')) {
		prefix = "r"
	}

	return "%" + prefix + strings.ToLower(string(r))
}

type simdRegister string

func (r simdRegister) String() string {
	return string(r)
}

func (r simdRegister) Gas() string {
	return "%" + strings.ToLower(string(r[:1])) +
		"mm" + strings.ToLower(string(r[1:]))
}

var (
	_ Operand = Register("")
	_ Operand = simdRegister("")
)

const (
	SP  = Register("SP")
	AX  = Register("AX")
	AH  = Register("AH")
	AL  = Register("AL")
	BX  = Register("BX")
	BH  = Register("BH")
	BL  = Register("BL")
	CX  = Register("CX")
	CH  = Register("CH")
	CL  = Register("CL")
	DX  = Register("DX")
	DH  = Register("DH")
	DL  = Register("DL")
	BP  = Register("BP")
	DI  = Register("DI")
	SI  = Register("SI")
	R8  = Register("R8")
	R9  = Register("R9")
	R10 = Register("R10")
	R11 = Register("R11")
	R12 = Register("R12")
	R13 = Register("R13")
	R14 = Register("R14")
	R15 = Register("R15")
)

const (
	X0  = simdRegister("X0")
	X1  = simdRegister("X1")
	X2  = simdRegister("X2")
	X3  = simdRegister("X3")
	X4  = simdRegister("X4")
	X5  = simdRegister("X5")
	X6  = simdRegister("X6")
	X7  = simdRegister("X7")
	X8  = simdRegister("X8")
	X9  = simdRegister("X9")
	X10 = simdRegister("X10")
	X11 = simdRegister("X11")
	X12 = simdRegister("X12")
	X13 = simdRegister("X13")
	X14 = simdRegister("X14")
	X15 = simdRegister("X15")
)

const (
	Y0  = simdRegister("Y0")
	Y1  = simdRegister("Y1")
	Y2  = simdRegister("Y2")
	Y3  = simdRegister("Y3")
	Y4  = simdRegister("Y4")
	Y5  = simdRegister("Y5")
	Y6  = simdRegister("Y6")
	Y7  = simdRegister("Y7")
	Y8  = simdRegister("Y8")
	Y9  = simdRegister("Y9")
	Y10 = simdRegister("Y10")
	Y11 = simdRegister("Y11")
	Y12 = simdRegister("Y12")
	Y13 = simdRegister("Y13")
	Y14 = simdRegister("Y14")
	Y15 = simdRegister("Y15")
)

type Label struct{ name string }

func (a *Asm) NewLabel(name string) Label {
	return Label{name}
}

func (l Label) String() string {
	return l.name
}

func (Label) Gas() string {
	panic("referencing labels in unsupported opcodes is forbidden")
}

func (l Label) Suffix(suffix string) Label {
	return Label{fmt.Sprintf("%s_%s", l.name, suffix)}
}

type function string

func (f function) String() string {
	return string(f)
}

func (function) Gas() string {
	panic("referencing functions in unsupported opcodes is forbidden")
}

func Function(name string) Operand {
	return function(fmt.Sprintf("·%s(SB)", name))
}

func (a *Asm) op(instruction string, ops ...Operand) {
	if len(ops) == 0 {
		a.write("\t" + instruction)
		return
	}

	var sOps []string

	for i := len(ops) - 1; i >= 0; i-- {
		sOps = append(sOps, ops[i].String())
	}

	a.write(fmt.Sprintf("\t%v%v%v", instruction, a.Seperator, strings.Join(sOps, ", ")))
}

var objdumpRegex = regexp.MustCompile(`^\s+\d:\s+((?:[0-9a-fA-F]{2} )+)`)

func (a *Asm) unsupOp(instruction string, ops ...Operand) {
	tmp, err := ioutil.TempFile("", "asm-unsupOp-")
	if err != nil {
		panic(err)
	}
	defer tmp.Close()
	os.Remove(tmp.Name())

	var gOps []string

	for i := len(ops) - 1; i >= 0; i-- {
		gOps = append(gOps, ops[i].Gas())
	}

	cmd := exec.Command("as", "-o", "/dev/stdout", "-")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%v\t%s\n", instruction, strings.Join(gOps, ", ")))
	cmd.Stdout = tmp
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	cmd = exec.Command("objdump", "-d", "/dev/stdin")
	cmd.Stdin = tmp
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	if err = cmd.Start(); err != nil {
		panic(err)
	}

	gOps = gOps[:0]

	for i := len(ops) - 1; i >= 0; i-- {
		gOps = append(gOps, ops[i].String())
	}

	a.write(fmt.Sprintf("\t// %v%v%s", instruction, a.Seperator, strings.Join(gOps, ", ")))

	scan2 := bufio.NewScanner(stdout)

	for scan2.Scan() {
		m := objdumpRegex.FindStringSubmatch(scan2.Text())
		if m == nil {
			continue
		}

		a.write(fmt.Sprintf("\tBYTE $0x%s", strings.Join(strings.Split(strings.TrimSpace(m[1]), " "), "; BYTE $0x")))
	}

	if err = cmd.Wait(); err != nil {
		panic(err)
	}
}

func (a *Asm) Label(name Label) {
	a.write(name.String() + ":")
}

func Do(file, header string, run func(*Asm)) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := io.WriteString(f, header); err != nil {
		return err
	}

	a := NewAsm(f)
	run(a)
	return a.Flush()
}

//go:generate go run ./opcode_gen.go -i $GOROOT/src/cmd/internal/obj/x86/aenum.go -o opcode.go -p asm
//go:generate gofmt -w opcode.go

// general opcodes

func (o Opcodes) Nop(ops ...Operand)  { o.a.op("NOP", ops...) }
func (o Opcodes) NOP(ops ...Operand)  { o.a.op("NOP", ops...) }
func (o Opcodes) Ret(ops ...Operand)  { o.a.op("RET", ops...) }
func (o Opcodes) RET(ops ...Operand)  { o.a.op("RET", ops...) }
func (o Opcodes) Call(ops ...Operand) { o.a.op("CALL", ops...) }
func (o Opcodes) CALL(ops ...Operand) { o.a.op("CALL", ops...) }
func (o Opcodes) Jmp(ops ...Operand)  { o.a.op("JMP", ops...) }
func (o Opcodes) JMP(ops ...Operand)  { o.a.op("JMP", ops...) }

// other jumps

func (o Opcodes) Je(ops ...Operand)  { o.a.op("JE", ops...) }
func (o Opcodes) JE(ops ...Operand)  { o.a.op("JE", ops...) }
func (o Opcodes) Jb(ops ...Operand)  { o.a.op("JB", ops...) }
func (o Opcodes) JB(ops ...Operand)  { o.a.op("JB", ops...) }
func (o Opcodes) Jae(ops ...Operand) { o.a.op("JAE", ops...) }
func (o Opcodes) JAE(ops ...Operand) { o.a.op("JAE", ops...) }
func (o Opcodes) Jz(ops ...Operand)  { o.a.op("JZ", ops...) }
func (o Opcodes) JZ(ops ...Operand)  { o.a.op("JZ", ops...) }
func (o Opcodes) Jnz(ops ...Operand) { o.a.op("JNZ", ops...) }
func (o Opcodes) JNZ(ops ...Operand) { o.a.op("JNZ", ops...) }
func (o Opcodes) Jc(ops ...Operand)  { o.a.op("JC", ops...) }
func (o Opcodes) JC(ops ...Operand)  { o.a.op("JC", ops...) }
