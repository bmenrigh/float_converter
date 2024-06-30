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


func float_dump_string(f *float_plus) string {

	sign := "0"
	if f.Sign {
		sign = "1"
	}

	return fmt.Sprintf("float%d (1 / %d / %d): %s %s %s\n", 1 + f.ExpWidth + f.ManWidth, f.ExpWidth, f.ManWidth, sign, bits_to_string(f.Exponent), bits_to_string(f.Mantissa))
}


func bits_to_string(b []bool) string {

	bstr := make([]byte, 0, len(b))

	for _, v := range b {
		if v {
			bstr = append(bstr, byte('1'))
		} else {
			bstr = append(bstr, byte('0'))
		}
	}

	return string(bstr)
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

	fmt.Fprint(os.Stderr, "exp bits: %s\n", bits_to_string(f.Exponent))

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

	fmt.Fprint(os.Stderr, "local mantissa bits: %s\n", bits_to_string(lmant))

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

	fmt.Fprint(os.Stderr, "ipart bits: %s\n", bits_to_string(ipart))
	fmt.Fprint(os.Stderr, "fpart bits: %s\n", bits_to_string(fpart))


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

	fmt.Fprintf(os.Stderr, "is negative: %t\n", is_neg)

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

	seen_one := false
	if inum.Cmp(big.NewInt(0)) != 0 {
		seen_one = true
	}

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

	fmt.Fprintf(os.Stderr, "Got inum bits: %s\n", bits_to_string(ibits))

	mant_used := int32(0)
	if seen_one {
		mant_used = int32(len(ibits) - 1) // 1 less because because of the implicit 1 normalization
	}


	flimit := big.NewInt(1)
	if len(s_parts) > 1 {
		for i := 0; i < len(s_parts[1]); i++ {
			flimit = new(big.Int).Mul(flimit, big.NewInt(10))
		}
	}

	fbits := make([]bool, 0)
	for fnum.Cmp(big.NewInt(0)) > 0 && mant_used <= f.ManWidth {
		fnum = new(big.Int).Mul(fnum, big.NewInt(2))

		// This intentionally doesn't count the first 1 seen since
		// it will be implicit with normalization
		if seen_one {
			mant_used++
		}

		if fnum.Cmp(flimit) >= 0 {
			fbits = append(fbits, true)
			fnum = new(big.Int).Sub(fnum, flimit)

			if !seen_one {
				seen_one = true
			}
		} else {
			fbits = append(fbits, false)
		}
	}

	fmt.Fprintf(os.Stderr, "Got fnum bits: %s\n", bits_to_string(fbits))

	// Local mantissa (explicitly has normall implicit 1)
	lmant := append(ibits, fbits...)

	// find first one
	first_one := 0
	for i := 0; i < len(lmant); i++ {
		if lmant[i] {
			first_one = i
			break
		}
	}

	fmt.Fprintf(os.Stderr, "First 1 in local mantissa at %d\n", first_one)

	// reslice to remove leading zeros
	lmant = lmant[first_one:]

	fmt.Fprintf(os.Stderr, "Initial local mantissa bits: %s\n", bits_to_string(lmant))

	carry := false
	if int32(len(lmant)) <= f.ManWidth + 1 {
		// pad out with zeros
		for i := int32(len(lmant)); i <= f.ManWidth; i++ {
			lmant = append(lmant, false)
		}
	} else {
		// we got more bits than will fit so we need to round

		carry = lmant[f.ManWidth + 1]
		// truncate
		lmant = lmant[:f.ManWidth + 1]

		for i := f.ManWidth; i >= 0 && carry; i-- {
			if lmant[i] {
				lmant[i] = false
			} else {
				lmant[i] = true
				carry = false
			}
		}

		// Cary the one to the left into a new digit
		if carry {
			lmant = append([]bool{true}, lmant...)
		}
	}

	fmt.Fprintf(os.Stderr, "Pad/carry local mantissa bits: %s\n", bits_to_string(lmant))

	exp := int32((len(ibits) - 1) - first_one)
	if carry {
		exp++
	}

	fmt.Fprintf(os.Stderr, "Exponent: %d\n", exp)

	exp += f.ExpOffset

	if exp > 0 {
		ebits := make([]bool, 0, f.ExpWidth)

		for exp != 0 {
			if exp % 2 == 1 {
				ebits = append(ebits, true)
			} else {
				ebits = append(ebits, false)
			}

			exp /= 2
		}

		for int32(len(ebits)) < f.ExpWidth {
			ebits = append(ebits, false)
		}

		// put exponent in BE bit order
		slices.Reverse(ebits)

		if int32(len(ebits)) > f.ExpWidth {
			fmt.Fprintf(os.Stderr, "TODO: got infinity")
		} else {
			f.Sign = is_neg
			f.Exponent = ebits
			f.Mantissa = lmant[1:f.ManWidth + 1]

			return true
		}
	} else {
		// possibly denorm, or just zero
	}

	return false
}


func main() {

	f := float_from_uint32(0x40490fdb) // Pi -- 3.1415927410125732421875

	ok, s := float_to_string(f, true)

	if ok {
		fmt.Printf("%s\n", s)
	}

	//ok = float_from_string(f, "123.0354766845703125")
	//ok = float_from_string(f, "0.1")
	//ok = float_from_string(f, "0.01234567")
	//ok = float_from_string(f, "3.33333333333333333333333333333")
	//ok = float_from_string(f, "3.1415926535897932384626433832795028842")
	ok = float_from_string(f, "43252003274489856000")

	fmt.Fprintf(os.Stderr, "%s", float_dump_string(f))

	// d := float_from_uint64(0x000FFFFFFFFFFFFF) // Max subnormal 2.2250738585072009 * 10^-308

	// ok, s = float_to_string(d, true)

	// if ok {
	// 	fmt.Printf("%s\n", s)
	// }
}
