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
	CanDenorm bool
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


func float_to_string(f *float_plus) (bool, string) {

	if !validate_float(f) {
		return false, ""
	}

	var exp int32
	var bdigit int32
	exp = 0
	bdigit= 1;

	for i := f.ExpWidth - 1; i >= 0; i-- {
		if f.Exponent[i] {
			exp += bdigit
		}

		bdigit *= 2
	}

	exp -= f.ExpOffset

	fmt.Fprintf(os.Stderr, "exp: %d\n", exp)

	lmant := append([]bool{true}, f.Mantissa...)

	// Check and do padding with zeros on the right if needed
	if exp > f.ManWidth + 1 {
		lmant = append(lmant, make([]bool, exp - (f.ManWidth + 1))...)
	}

	// Check and do padding with zeros on the left if needed
	if exp < 0 {
		lmant = append(make([]bool, 0 - exp), lmant...)
	}

	var ipart []bool
	var fpart []bool

	if exp >= f.ManWidth + 1 { // all int
		ipart = make([]bool, len(lmant))
		copy(ipart, lmant)
		fpart = make([]bool, 0)
	} else if exp >= 0 {        // some int, some frac
		ipart = make([]bool, exp + 1)
		fpart = make([]bool, (len(lmant) - 1) - int(exp + 1))
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

	f := new(float_plus)

	f.ExpWidth = 8
	f.ExpOffset = 127
	f.ManWidth = 23

	f.Exponent = make([]bool, f.ExpWidth)
	f.Mantissa = make([]bool, f.ManWidth)

	// 1.75
	f.Exponent = []bool{false, true, true, true, true, true, true, true}
	f.Mantissa[0] = true
	f.Mantissa[1] = true
	f.Mantissa[20] = true


	ok, s := float_to_string(f)

	if ok {
		fmt.Printf("%s\n", s)
	}
}
