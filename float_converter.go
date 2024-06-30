package main

import (
	"math/big"
	"fmt"
	"strings"
	"os"
	"slices"
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


func float_to_string(f *float_plus, scinot bool) (bool, string) {

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

	fnum_is_zero := fnum.Cmp(big.NewInt(0)) == 0

	if len(fstr) < lastone + 1 {
		zpad = strings.Repeat("0", (lastone + 1) - len(fstr))
	}

	if !scinot || fnum_is_zero || len(zpad) < 4 {
		var decimal_str string

		if fnum_is_zero {
			decimal_str = ""
		} else {
			decimal_str = fmt.Sprintf(".%s%s", zpad, fstr)
		}

		return true, fmt.Sprintf("%s%s%s", sign, istr, decimal_str)
	} else {
		// scientific notation

		return true, fmt.Sprintf("%s%s.%sE-%d", sign, fstr[0:1], fstr[1:], len(zpad) + 1)
	}
}


func float_from_string(f *float_plus, s string) bool {

	if len(s) < 1 {
		return false
	}

	is_neg := false
	if s[:1] == "-" {
		is_neg = true

		s = s[1:] // reslice to get rid of negative sign
	}

	fmt.Fprintf(os.Stderr, "is negative: %t\n", is_neg);

	s_parts := strings.Split(s, ".")

	if len(s_parts) > 2 {
		return false
	}

	inum, ok := new(big.Int).SetString(s_parts[0], 10)
	if !ok {
		return false
	}

	fnum := big.NewInt(0)
	if len(s_parts) > 1 {
		fnum, ok = new(big.Int).SetString(s_parts[1], 10)

		if !ok {
			return false
		}
	}

	fmt.Fprintf(os.Stderr, "got ipart: %s\n", inum.Text(10))
	fmt.Fprintf(os.Stderr, "got fpart: %s\n", fnum.Text(10))


	ibits := make([]bool, 0)

	for inum.Cmp(big.NewInt(0)) != 0 {

		if new(big.Int).Mod(inum, big.NewInt(2)).Cmp(big.NewInt(0)) != 0 {
			ibits = append(ibits, true)
		} else {
			ibits = append(ibits, false)
		}

		inum = new(big.Int).Div(inum, big.NewInt(2))
	}

	// Now put the bits in the correct "BE" order
	slices.Reverse(ibits)

	fmt.Fprintf(os.Stderr, "Got inum bits: ");
	for i := 0; i < len(ibits); i++ {
		if ibits[i] {
			fmt.Fprint(os.Stderr, "1")
		} else {
			fmt.Fprint(os.Stderr, "0")
		}
	}
	fmt.Fprint(os.Stderr, "\n")


	flimit := big.NewInt(1)

	for i := 0; i < len(s_parts[1]); i++ {
		flimit = new(big.Int).Mul(flimit, big.NewInt(10))
	}

	fbits := make([]bool, 0)
	for fnum.Cmp(big.NewInt(0)) > 0 {
		fnum = new(big.Int).Mul(fnum, big.NewInt(2))

		if fnum.Cmp(flimit) >= 0 {
			fbits = append(fbits, true)

			fnum = new(big.Int).Sub(fnum, flimit)
		} else {
			fbits = append(fbits, false)
		}
	}

	fmt.Fprintf(os.Stderr, "Got fnum bits: ");
	for i := 0; i < len(fbits); i++ {
		if fbits[i] {
			fmt.Fprint(os.Stderr, "1")
		} else {
			fmt.Fprint(os.Stderr, "0")
		}
	}
	fmt.Fprint(os.Stderr, "\n")

	return true
}


func main() {

	f := float_from_uint32(0x40490fdb) // Pi -- 3.1415927410125732421875

	ok, s := float_to_string(f, true)

	if ok {
		fmt.Printf("%s\n", s)
	}

	ok = float_from_string(f, "123.0354766845703125")


	// d := float_from_uint64(0x000FFFFFFFFFFFFF) // Max subnormal 2.2250738585072009 * 10^-308

	// ok, s = float_to_string(d, true)

	// if ok {
	// 	fmt.Printf("%s\n", s)
	// }
}
