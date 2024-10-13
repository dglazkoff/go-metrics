/*
Модуль staticlint включает в себя объединение различных анализаторов кода.

Чтобы собрать, необходимо перейти в папку с кодом и выполнить `go build .`

Чтобы запустить, необходимо перейти в корень вашего проекта и выполнить `{path-to-program}/multichecker ./...`

Используемые проверки:

- все проверки из пакета golang.org/x/tools/go/analysis/passes

- все анализаторы класса SA пакета staticcheck.io

- rules from staticcheck.io: "S1006", "S1008", "S1002", "S1011", "S1018", "S1028", "S1031"

- анализатор funlen для проверки размеров функций

- проверка вызова os.Exit в функции main пакета main
*/
package main

import (
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/ultraware/funlen"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"honnef.co/go/tools/staticcheck"
)

// так и не смог найти второй анализатор, который бы встраивался в multichecker и был более менее поддерживаемым

var FunLenAnalyzer = &analysis.Analyzer{
	Name: "funlen",
	Doc:  "check function length",
	Run:  runFunLen,
}

func runFunLen(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		fileName := filepath.Base(pass.Fset.Position(file.Pos()).Filename)
		if strings.Contains(fileName, "test") || strings.Contains(fileName, "templ") {
			continue
		}
		msgs := funlen.Run(file, pass.Fset, 60, 40, false)

		for _, msg := range msgs {
			pos := file.Pos()
			pass.Reportf(pos, msg.Message)
		}
	}
	return nil, nil
}

var ExitAnalyzer = &analysis.Analyzer{
	Name: "exit",
	Doc:  "check using os.Exit in main function",
	Run:  runExitCheck,
}

func runExitCheck(pass *analysis.Pass) (interface{}, error) {
	pkgName := pass.Pkg.Name()

	if pkgName != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		fileName := filepath.Base(pass.Fset.Position(file.Pos()).Filename)

		if fileName != "main.go" {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				if selExpr, ok := call.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "os" && selExpr.Sel.Name == "Exit" {
						pass.Reportf(call.Pos(), "Can't use os.Exit")
					}
				}
			}
			return true
		})

	}
	return nil, nil
}

func main() {
	checks := map[string]bool{
		"S1006": true,
		"S1008": true,
		"S1002": true,
		"S1011": true,
		"S1018": true,
		"S1028": true,
		"S1031": true,
	}

	mychecks := []*analysis.Analyzer{
		buildtag.Analyzer,
		asmdecl.Analyzer,
		appends.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		directive.Analyzer,
		defers.Analyzer,
		errorsas.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		httpmux.Analyzer,
		inspect.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		printf.Analyzer,
		pkgfact.Analyzer,
		reflectvaluecompare.Analyzer,
		shift.Analyzer,
		shadow.Analyzer,
		sigchanyzer.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		slog.Analyzer,
		sortslice.Analyzer,
		tests.Analyzer,
		testinggoroutine.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		FunLenAnalyzer,
		ExitAnalyzer,
	}
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			mychecks = append(mychecks, v.Analyzer)
		} else if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(
		mychecks...,
	)
}
