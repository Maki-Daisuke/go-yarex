package main

import (
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"log"
	"os"
	"regexp"

	"github.com/Maki-Daisuke/go-yarex"
)

var reDirective = regexp.MustCompile(`^yarexgen\s*$`)

func main() {
	if len(os.Args) < 2 {
		log.Println("Specify a file name.")
		os.Exit(1)
	}
	filename := os.Args[1]

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		log.Println("Error: ", err)
		os.Exit(1)
	}

	generator := yarex.NewGoGenerator(filename, file.Name.Name)
LOOP:
	for n, cg := range ast.NewCommentMap(fset, file, file.Comments) {
		for _, c := range cg {
			if reDirective.MatchString(c.Text()) {
				strs := findRegex(n)
				if strs != nil {
					generator.Add(strs...)
				} else {
					log.Printf("couldn't find regexp string in %s at line %d\n", filename, fset.File(c.Pos()).Line(c.Pos()))
				}
				continue LOOP
			}
		}
	}

	re := regexp.MustCompile(`(?i:(_test)?\.go$)`)
	out := re.ReplaceAllString(filename, "_yarex$0")
	outfile, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	_, err = generator.WriteTo(outfile)
	if err != nil {
		panic(err)
	}
}

// find string litetal
func findRegex(n ast.Node) (out []string) {
	ast.Inspect(n, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok {
			return true
		}
		if lit.Kind != token.STRING {
			return true
		}
		v := constant.MakeFromLiteral(lit.Value, lit.Kind, 0)
		out = append(out, constant.StringVal(v))
		return true
	})
	return out
}
