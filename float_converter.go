package main

import (
	"math/big"
	"fmt"
	"strings"
	"os"
)


type float_plus struct{
	ExpWidth int32
	ExpOffset int32
	ManWidth int32
	AllowSubnorm bool
	Sign bool
	Exponent []bool
	Mantissa []bool
}


func validate_float(f *float_plus) bool {

	if int32(len(f.Exponent)) != f.ExpWidth {
		return false
	}

	if int32(len(f.Mantissa)) != f.ManWidth {
		return false
	}

	// TODO: check ExpOffset makes sense for ExpWidth

	return true
}


func float_from_uint32(v uint32) *float_plus {

	f := new(float_plus)

	f.ExpWidth = 8
	f.ExpOffset = 127
	f.ManWidth = 23
	f.AllowSubnorm = true

	f.Exponent = make([]bool, f.ExpWidth)
	f.Mantissa = make([]bool, f.ManWidth)

	for i := int32(0); i < f.ManWidth; i++ {
		if v % 2 == 1 {
			f.Mantissa[(f.ManWidth - 1) - i] = true
		}

		v /= 2
	}

	for i := int32(0); i < f.ExpWidth; i++ {
		if v % 2 == 1 {
			f.Exponent[(f.ExpWidth - 1) - i] = true
		}

		v /= 2
	}

	if v > 0 {
		f.Sign = true
	}

	return f
}


func float_from_uint64(v uint64) *float_plus {

	f := new(float_plus)

	f.ExpWidth = 11
	f.ExpOffset = 1023
	f.ManWidth = 52
	f.AllowSubnorm = true

	f.Exponent = make([]bool, f.ExpWidth)
	f.Mantissa = make([]bool, f.ManWidth)

	for i := int32(0); i < f.ManWidth; i++ {
		if v % 2 == 1 {
			f.Mantissa[(f.ManWidth - 1) - i] = true
		}

		v /= 2
	}

	for i := int32(0); i < f.ExpWidth; i++ {
		if v % 2 == 1 {
			f.Exponent[(f.ExpWidth - 1) - i] = true
		}

		v /= 2
	}

	if v > 0 {
		f.Sign = true
	}

	return f
}


func float_to_string(f *float_plus) (bool, string) {

	if !validate_float(f) {
		return false, ""
	}

	var exp int32
	var bdigit int32
	exp = 0
	bdigit= 1;

	fmt.Fprint(os.Stderr, "exp bits: ")
	for i := int32(0); i < f.ExpWidth; i++ {
		if f.Exponent[i] {
			fmt.Fprint(os.Stderr, "1")
		} else {
			fmt.Fprint(os.Stderr, "0")
		}
	}
	fmt.Fprint(os.Stderr, "\n")

	for i := f.ExpWidth - 1; i >= 0; i-- {
		if f.Exponent[i] {
			exp += bdigit
		}

		bdigit *= 2
	}

	exp -= f.ExpOffset

	fmt.Fprintf(os.Stderr, "exp val: %d\n", exp)

	lmant := make([]bool, f.ManWidth)
	copy(lmant, f.Mantissa)

	if !f.AllowSubnorm || (exp != 0 - f.ExpOffset) {
		lmant = append([]bool{true}, lmant...)
	} else {
		fmt.Fprint(os.Stderr, "subnormal number\n")
		//lmant = append([]bool{false}, lmant...)
	}

	fmt.Fprint(os.Stderr, "local mantissa bits: ")
	for i := 0; i < len(lmant); i++ {
		if lmant[i] {
			fmt.Fprint(os.Stderr, "1")
		} else {
			fmt.Fprint(os.Stderr, "0")
		}
	}
	fmt.Fprint(os.Stderr, "\n")

	// Check and do padding with zeros on the right if needed
	if exp > f.ManWidth {
		lmant = append(lmant, make([]bool, exp - f.ManWidth)...)
	}

	// Check and do padding with zeros on the left if needed
	if exp < -1 {
		lmant = append(make([]bool, -1 - exp), lmant...)
	}

	var ipart []bool
	var fpart []bool

	if exp >= f.ManWidth + 1 { // all int
		ipart = make([]bool, len(lmant))
		copy(ipart, lmant)
		fpart = make([]bool, 0)
	} else if exp >= 0 {        // some int, some frac
		ipart = make([]bool, exp + 1)
		fpart = make([]bool, len(lmant) - int(exp + 1))
		copy(ipart, lmant[:exp + 1])
		copy(fpart, lmant[exp + 1:])
	} else {                   // all frac
		ipart = make([]bool, 0)
		fpart = make([]bool, len(lmant))
		copy(fpart, lmant)
	}

	fmt.Fprint(os.Stderr, "ipart bits: ")
	for i := 0; i < len(ipart); i++ {
		if ipart[i] {
			fmt.Fprint(os.Stderr, "1")
		} else {
			fmt.Fprint(os.Stderr, "0")
		}
	}
	fmt.Fprint(os.Stderr, "\n")

	fmt.Fprint(os.Stderr, "fpart bits: ")
	for i := 0; i < len(fpart); i++ {
		if fpart[i] {
			fmt.Fprint(os.Stderr, "1")
		} else {
			fmt.Fprint(os.Stderr, "0")
		}
	}
	fmt.Fprint(os.Stderr, "\n")


	inum := big.NewInt(0)
	bival := big.NewInt(1)

	for i := len(ipart) - 1; i >= 0; i-- {
		if ipart[i] {
			inum = new(big.Int).Add(inum, bival)
		}

		bival = new(big.Int).Add(bival, bival) // bival *= 2
	}

	fnum := big.NewInt(0)
	bfval := big.NewInt(5)

	lastone := -1

	for i := 0; i < len(fpart); i++ {
		if fpart[i] {
			lastone = i
		}
	}

	for i := 0; i <= lastone; i++ {
		fnum = new(big.Int).Mul(fnum, big.NewInt(10))

		if fpart[i] {
			fnum = new(big.Int).Add(fnum, bfval)
		}

		bfval = new(big.Int).Mul(bfval, big.NewInt(5))
	}


	sign := ""

	if f.Sign {
		sign = "-"
	}

	istr := inum.Text(10)
	fstr := fnum.Text(10)

	zpad := ""

	if len(fstr) < lastone + 1 {
		zpad = strings.Repeat("0", (lastone + 1) - len(fstr))
	}

	return true, fmt.Sprintf("%s%s.%s%s", sign, istr, zpad, fstr)
}


func main() {

	// f := new(float_plus)

	// f.ExpWidth = 8
	// f.ExpOffset = 127
	// f.ManWidth = 23
	// f.AllowSubnorm = true

	// f.Exponent = make([]bool, f.ExpWidth)
	// f.Mantissa = make([]bool, f.ManWidth)

	// 1.75
	// f.Exponent = []bool{false, true, true, true, true, true, true, true}
	// f.Mantissa[0] = true
	// f.Mantissa[1] = true

	// 3.5
	// f.Exponent = []bool{true, false, false, false, false, false, false, false}
	// f.Mantissa[0] = true
	// f.Mantissa[1] = true

	// 2.0000002384185791015625
	//f.Exponent = []bool{true, false, false, false, false, false, false, false}
	//f.Mantissa[22] = true

	// 0.5
	//f.Exponent = []bool{false, true, true, true, true, true, true, false}

	// 0.1328125
	//f.Exponent = []bool{false, true, true, true, true, true, false, false}
	//f.Mantissa[3] = true

	// 1.248962747748680477216783E-38
	//f.Exponent = []bool{false, false, false, false, false, false, false, true}
	//f.Mantissa[3] = true

	// 237684506432258944259212705792
	// f.Exponent = []bool{true, true, true, false, false, false, false, false}
	// f.Mantissa[0] = true
	// f.Mantissa[22] = true

	// 6291456.5
	// f.Exponent = []bool{true, false, false, true, false, true, false, true}
	// f.Mantissa[0] = true
	// f.Mantissa[22] = true

	// 12582913
	// f.Exponent = []bool{true, false, false, true, false, true, true, false}
	// f.Mantissa[0] = true
	// f.Mantissa[22] = true

	// 25165826
	//f.Exponent = []bool{true, false, false, true, false, true, true, true}
	//f.Mantissa[0] = true
	//f.Mantissa[22] = true


	// f := float_from_uint32(0x40490fdb) // Pi -- 3.1415927410125732421875

	// ok, s := float_to_string(f)

	// if ok {
	// 	fmt.Printf("%s\n", s)
	// }


	f := float_from_uint32(0x00400000) // Subnormal 5.877471754111437539843683E-39

	ok, s := float_to_string(f)

	if ok {
		fmt.Printf("%s\n", s)
	}


	// d := float_from_uint64(0x400921FB54442D18) // Pi

	// ok, s = float_to_string(d)

	// if ok {
	// 	fmt.Printf("%s\n", s)
	// }

	d := float_from_uint64(0x000FFFFFFFFFFFFF) // Max subnormal 2.2250738585072009 * 10^-308

	ok, s = float_to_string(d)

	if ok {
		fmt.Printf("%s\n", s)
	}
}
