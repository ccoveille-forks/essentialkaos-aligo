package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2019 ESSENTIAL KAOS                         //
//        Essential Kaos Open Source License <https://essentialkaos.com/ekol>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"path"
	"strings"

	"pkg.re/essentialkaos/ek.v10/fmtc"
	"pkg.re/essentialkaos/ek.v10/fmtutil"
	"pkg.re/essentialkaos/ek.v10/mathutil"
	"pkg.re/essentialkaos/ek.v10/strutil"

	"github.com/essentialkaos/aligo/inspect"
	"github.com/essentialkaos/aligo/report"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// MAX_TYPE_SIZE maximum type size
const MAX_TYPE_SIZE = 32

// ////////////////////////////////////////////////////////////////////////////////// //

// PrintFull prints all report info
func PrintFull(r *report.Report) {
	if r.IsEmpty() {
		fmtc.Println("{y}Given package doesn't have any structs{!}")
		return
	}

	for _, pkg := range r.Packages {
		if !pkg.IsEmpty() {
			printPackageInfo(pkg)
		}
	}

	fmtutil.Separator(true)
	fmtc.NewLine()
}

// PrintStruct prints info about struct
func PrintStruct(r *report.Report, strName string, detailed, optimal bool) {
	if r.IsEmpty() {
		fmtc.Println("{y}Given package doesn't have any structs{!}")
		return
	}

	switch {
	case r.IsEmpty():
		fmtc.Println("{y}Given package doesn't have any structs{!}")
		return
	case strName == "":
		fmtc.Println("{y}You should define struct name{!}")
		return
	}

	pkg, str := findStruct(r, strName)

	if pkg == nil && str == nil {
		fmtc.Printf("{y}Can't find struct with name {*}%s{!}\n", strName)
		return
	}

	fmtutil.Separator(false, pkg.Path)

	printStructInfo(str, pkg.Path, detailed, optimal)

	fmtutil.Separator(true)
	fmtc.NewLine()
}

// Check checks report for problems
func Check(r *report.Report, detailed bool) bool {
	if r.IsEmpty() {
		fmtc.Println("{y}Nothing to check - given package doesn't have any structs{!}")
		return false
	}

	var hasProblems bool

	for _, pkg := range r.Packages {
		if pkg.IsEmpty() {
			continue
		}

		if !isPackageHasProblems(pkg) {
			continue
		}

		hasProblems = true

		printPackageProblems(pkg, detailed)
	}

	if !hasProblems {
		fmtc.Println("{g}All structs is well aligned{!}")
	} else {
		fmtutil.Separator(true)
		fmtc.NewLine()
	}

	return hasProblems
}

// ////////////////////////////////////////////////////////////////////////////////// //

// printPackageInfo prints package info
func printPackageInfo(pkg *report.Package) {
	fmtutil.Separator(false, pkg.Path)

	for _, str := range pkg.Structs {
		printStructInfo(str, pkg.Path, true, false)
	}
}

// printPackageProblems prints problems in package
func printPackageProblems(pkg *report.Package, detailed bool) {
	fmtutil.Separator(false, pkg.Path)

	for _, str := range pkg.Structs {
		if !isStructHasProblems(str) {
			continue
		}

		printStructInfo(str, pkg.Path, detailed, true)
	}
}

// printStructInfo prints struct info
func printStructInfo(str *report.Struct, pkgPath string, detailed, optimal bool) {
	switch optimal {
	case false:
		fmtc.Printf(
			"{s-}// %s:%d | Size: %d (Optimal: %d){!}\n",
			str.Position.File, str.Position.Line, str.Size, str.OptimalSize,
		)
	default:
		fmtc.Printf(
			"Struct {*}%s{!} {s-}(%s:%d){!} fields order can be optimized (%d → %d)\n\n",
			str.Name, str.Position.File, str.Position.Line, str.Size, str.OptimalSize,
		)
	}

	fmtc.Printf("type {*}%s{!} struct {s}{{!}\n", str.Name)

	var fields []*report.Field

	if optimal {
		fields = str.OptimalFields
	} else {
		fields = str.Fields
	}

	if detailed {
		printDetailedFieldsInfo(fields, pkgPath)
	} else {
		printSimpleFieldsInfo(fields, pkgPath)
	}

	fmtc.Printf("{s}}{!}\n\n")
}

// printDetailedFieldsInfo prints verbose fields info
func printDetailedFieldsInfo(fields []*report.Field, pkgPath string) {
	f := getFieldFormat(fields, pkgPath)

	counter := int64(0)
	maxAlign := inspect.GetMaxAlign()

	for index, field := range fields {
		fType := getPrettyFieldType(field.Type, pkgPath)
		fmtc.Printf(f, field.Name, strutil.Ellipsis(fType, MAX_TYPE_SIZE))

		if counter != 0 {
			fmtc.Printf(strings.Repeat("  ", int(counter)))
		}

		for i := int64(0); i < field.Size; i++ {
			fmtc.Printf("{g}■ {!}")

			counter++

			if counter == maxAlign {
				if i+1 != field.Size {
					fmtc.NewLine()
					fmtc.Printf(f, "", "")
				}
				counter = 0
			}
		}

		if index+1 < len(fields) && counter != 0 && fields[index+1].Size > maxAlign-counter {
			fmtc.Printf(strings.Repeat("{r}□ {!}", int(maxAlign-counter)))
			counter = 0
		} else if index+1 == len(fields) && counter != 0 {
			fmtc.Printf(strings.Repeat("{g}□ {!}", int(maxAlign-counter)))
		}

		fmtc.NewLine()
	}
}

// printSimpleFieldsInfo prints verbose fields info
func printSimpleFieldsInfo(fields []*report.Field, pkgPath string) {
	f := getFieldFormat(fields, pkgPath)

	for _, field := range fields {
		fType := getPrettyFieldType(field.Type, pkgPath)
		fmtc.Printf(f, field.Name, strutil.Ellipsis(fType, MAX_TYPE_SIZE))
		fmtc.NewLine()
	}
}

// getFieldFormat generate format string for field output
func getFieldFormat(fields []*report.Field, pkgPath string) string {
	var lName, lType int

	for _, field := range fields {
		fType := getPrettyFieldType(field.Type, pkgPath)
		lName = mathutil.Max(lName, len(field.Name))
		lType = mathutil.Max(lType, len(strutil.Ellipsis(fType, MAX_TYPE_SIZE)))
	}

	return fmt.Sprintf("  %%-%ds {*}%%-%ds{!}  ", lName, lType)
}

// getPrettyFieldType formats type name
func getPrettyFieldType(typ string, pkgPath string) string {
	if !strings.Contains(typ, "/") {
		return typ
	}

	if strings.Contains(typ, pkgPath+".") {
		return strutil.Exclude(typ, pkgPath+".")
	}

	for i := 0; i < 128; i++ {
		k := strutil.ReadField(typ, i, true, "[", "]", "*")

		if k == "" {
			break
		}

		if strings.Contains(k, "/") {
			typ = strings.Replace(typ, k, path.Base(k), -1)
		}
	}

	return typ
}

// findStruct finds struct with given name
func findStruct(r *report.Report, name string) (*report.Package, *report.Struct) {
	for _, pkg := range r.Packages {
		for _, str := range pkg.Structs {
			if str.Name == name {
				return pkg, str
			}
		}
	}

	return nil, nil
}

// isPackageHasProblems returns true if package has structs with
// unaligned fields
func isPackageHasProblems(pkg *report.Package) bool {
	for _, str := range pkg.Structs {
		if isStructHasProblems(str) {
			return true
		}
	}

	return false
}

// isStructHasProblems returns true if struct has unaligned fields
func isStructHasProblems(str *report.Struct) bool {
	return str.Size != str.OptimalSize
}