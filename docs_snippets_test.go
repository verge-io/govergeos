package vergeos

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// TestDocumentationSnippetsCompile type-checks every ```go fence in the README
// and docs. Those snippets are what users copy. A method or field that no
// longer matches the library fails this test.
func TestDocumentationSnippetsCompile(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Dir(file)

	snippets, err := loadDocSnippets(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(snippets) == 0 {
		t.Fatal("no Go snippets found in README.md or docs/")
	}

	analyzed := make([]analyzedSnippet, 0, len(snippets))
	for _, snip := range snippets {
		a, err := analyzeSnippet(snip)
		if err != nil {
			t.Errorf("%s:%d: %v\n%s", snip.file, snip.line, err, snip.code)
			continue
		}
		analyzed = append(analyzed, a)
	}
	if t.Failed() {
		return
	}

	overrides := make([]map[string]string, len(analyzed))
	for i := range overrides {
		overrides[i] = map[string]string{}
	}

	var lastErr string
	for attempt := 0; attempt < 8; attempt++ {
		src, ranges := renderSnippetProgram(analyzed, overrides)
		stderr, err := compileSnippetProgram(t, root, src)
		if err == nil {
			t.Logf("compiled %d documentation snippets", len(analyzed))
			return
		}
		if stderr == lastErr {
			t.Fatalf("documentation snippets failed to compile:\n%s", stderr)
		}
		lastErr = stderr
		if !applySnippetTypeOverrides(stderr, ranges, overrides) {
			t.Fatalf("documentation snippets failed to compile:\n%s", stderr)
		}
	}
	t.Fatalf("documentation snippets failed to compile:\n%s", lastErr)
}

type docSnippet struct {
	file string
	line int
	code string
}

type analyzedSnippet struct {
	file      string
	line      int
	code      string
	rewritten string
	free      []string
}

type snippetRange struct {
	index     int
	startLine int
	endLine   int
}

func loadDocSnippets(root string) ([]docSnippet, error) {
	paths := []string{filepath.Join(root, "README.md")}
	docs, err := filepath.Glob(filepath.Join(root, "docs", "*.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(docs)
	paths = append(paths, docs...)

	var snippets []docSnippet
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		snippets = append(snippets, extractGoSnippets(rel, string(data))...)
	}
	return snippets, nil
}

func extractGoSnippets(file, src string) []docSnippet {
	lines := strings.Split(src, "\n")
	var out []docSnippet
	for i := 0; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "```go" {
			continue
		}
		start := i + 1
		var b strings.Builder
		j := start
		for ; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "```" {
				break
			}
			b.WriteString(lines[j])
			b.WriteByte('\n')
		}
		code := stripImports(b.String())
		if strings.TrimSpace(stripGoComments(code)) != "" {
			out = append(out, docSnippet{file: filepath.ToSlash(file), line: start + 1, code: code})
		}
		i = j
	}
	return out
}

func stripImports(code string) string {
	lines := strings.Split(code, "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		trim := strings.TrimSpace(lines[i])
		if trim == "import (" {
			for i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != ")" {
				i++
			}
			if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == ")" {
				i++
			}
			continue
		}
		if strings.HasPrefix(trim, "import ") {
			continue
		}
		out = append(out, lines[i])
	}
	return strings.Join(out, "\n")
}

func stripGoComments(code string) string {
	var b strings.Builder
	for _, line := range strings.Split(code, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "//") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func analyzeSnippet(snip docSnippet) (analyzedSnippet, error) {
	body, fset, err := parseSnippetBody(snip.code)
	if err != nil {
		return analyzedSnippet{}, err
	}
	free := map[string]bool{}
	walkStmts(body.List, map[string]bool{}, free)
	initial := map[string]bool{}
	for name := range free {
		initial[name] = true
	}
	body.List = transformStmtList(body.List, initial)

	var printed bytes.Buffer
	cfg := &printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}
	for _, stmt := range body.List {
		if err := cfg.Fprint(&printed, fset, stmt); err != nil {
			return analyzedSnippet{}, err
		}
		printed.WriteByte('\n')
	}

	names := make([]string, 0, len(free))
	for name := range free {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return freeRank(names[i]) < freeRank(names[j]) || (freeRank(names[i]) == freeRank(names[j]) && names[i] < names[j])
	})
	return analyzedSnippet{
		file:      snip.file,
		line:      snip.line,
		code:      snip.code,
		rewritten: printed.String(),
		free:      names,
	}, nil
}

const snippetParsePrefix = "package p\nfunc _() {\n"

func parseSnippetBody(code string) (*ast.BlockStmt, *token.FileSet, error) {
	src := snippetParsePrefix + code + "\n}\n"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "snippet.go", src, parser.SkipObjectResolution|parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	if len(file.Decls) != 1 {
		return nil, nil, fmt.Errorf("expected one function, got %d declarations", len(file.Decls))
	}
	fn, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok || fn.Body == nil {
		return nil, nil, fmt.Errorf("snippet is not a function body")
	}
	return fn.Body, fset, nil
}

// transformStmtList makes a documentation snippet compile as one function.
// A repeated := shadows the earlier name so each example is checked on its
// own, and every new local is referenced because snippets often ignore results.
func transformStmtList(stmts []ast.Stmt, defined map[string]bool) []ast.Stmt {
	var out []ast.Stmt
	for _, stmt := range stmts {
		out = append(out, transformStmt(stmt, defined)...)
	}
	return out
}

func transformStmt(stmt ast.Stmt, defined map[string]bool) []ast.Stmt {
	switch st := stmt.(type) {
	case *ast.BadStmt, *ast.EmptyStmt, *ast.BranchStmt:
		return []ast.Stmt{st}
	case *ast.DeclStmt:
		return append([]ast.Stmt{st}, declSinks(st.Decl, defined)...)
	case *ast.LabeledStmt:
		stmts := transformStmt(st.Stmt, defined)
		st.Stmt = stmts[0]
		return append([]ast.Stmt{st}, stmts[1:]...)
	case *ast.ExprStmt:
		transformExpr(st.X, defined)
		return []ast.Stmt{st}
	case *ast.SendStmt:
		transformExpr(st.Chan, defined)
		transformExpr(st.Value, defined)
		return []ast.Stmt{st}
	case *ast.IncDecStmt:
		transformExpr(st.X, defined)
		return []ast.Stmt{st}
	case *ast.AssignStmt:
		return transformAssign(st, defined)
	case *ast.GoStmt:
		transformExpr(st.Call, defined)
		return []ast.Stmt{st}
	case *ast.DeferStmt:
		transformExpr(st.Call, defined)
		return []ast.Stmt{st}
	case *ast.ReturnStmt:
		for _, result := range st.Results {
			transformExpr(result, defined)
		}
		return []ast.Stmt{st}
	case *ast.BlockStmt:
		st.List = transformStmtList(st.List, cloneDefined(defined))
		return []ast.Stmt{st}
	case *ast.IfStmt:
		transformIf(st, defined)
		return []ast.Stmt{st}
	case *ast.SwitchStmt:
		transformSwitch(st, defined)
		return []ast.Stmt{st}
	case *ast.TypeSwitchStmt:
		transformTypeSwitch(st, defined)
		return []ast.Stmt{st}
	case *ast.SelectStmt:
		transformSelect(st, defined)
		return []ast.Stmt{st}
	case *ast.ForStmt:
		transformFor(st, defined)
		return []ast.Stmt{st}
	case *ast.RangeStmt:
		transformRange(st, defined)
		return []ast.Stmt{st}
	default:
		panic(fmt.Sprintf("unhandled statement %T", stmt))
	}
}

func firstStmt(stmts []ast.Stmt) ast.Stmt {
	if len(stmts) == 0 {
		return &ast.EmptyStmt{}
	}
	return stmts[0]
}

func cloneDefined(defined map[string]bool) map[string]bool {
	out := make(map[string]bool, len(defined))
	for k, v := range defined {
		out[k] = v
	}
	return out
}

func transformAssign(st *ast.AssignStmt, defined map[string]bool) []ast.Stmt {
	for _, rhs := range st.Rhs {
		transformExpr(rhs, defined)
	}
	return finishAssign(st, defined)
}

func finishAssign(st *ast.AssignStmt, defined map[string]bool) []ast.Stmt {
	if st.Tok != token.DEFINE {
		for _, lhs := range st.Lhs {
			transformExpr(lhs, defined)
		}
		return []ast.Stmt{st}
	}
	var newNames []string
	var existing []string
	for _, lhs := range st.Lhs {
		id, ok := lhs.(*ast.Ident)
		if !ok || id.Name == "_" {
			continue
		}
		if defined[id.Name] {
			existing = append(existing, id.Name)
			continue
		}
		newNames = append(newNames, id.Name)
	}
	// A later example often reuses a name with :=. Shadow it so the new call
	// is type-checked on its own and does not have to match the previous type.
	if len(newNames) == 0 && len(existing) > 0 {
		inner := cloneDefined(defined)
		for _, name := range existing {
			delete(inner, name)
		}
		return []ast.Stmt{&ast.BlockStmt{List: finishAssign(st, inner)}}
	}
	for _, name := range newNames {
		defined[name] = true
	}
	return append([]ast.Stmt{st}, sinkStmts(newNames)...)
}

func declSinks(decl ast.Decl, defined map[string]bool) []ast.Stmt {
	gen, ok := decl.(*ast.GenDecl)
	if !ok {
		return nil
	}
	var names []string
	for _, spec := range gen.Specs {
		switch sp := spec.(type) {
		case *ast.ValueSpec:
			if sp.Type != nil {
				transformExpr(sp.Type, defined)
			}
			for _, value := range sp.Values {
				transformExpr(value, defined)
			}
			for _, name := range sp.Names {
				if name.Name == "_" || defined[name.Name] {
					continue
				}
				defined[name.Name] = true
				names = append(names, name.Name)
			}
		case *ast.TypeSpec:
			if sp.Type != nil {
				transformExpr(sp.Type, defined)
			}
			defined[sp.Name.Name] = true
		}
	}
	return sinkStmts(names)
}

func sinkStmts(names []string) []ast.Stmt {
	var out []ast.Stmt
	for _, name := range names {
		out = append(out, &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("_")},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(name)},
		})
	}
	return out
}

func transformIf(st *ast.IfStmt, defined map[string]bool) {
	child := cloneDefined(defined)
	if st.Init != nil {
		stmts := transformStmt(st.Init, child)
		st.Init = stmts[0]
		if len(stmts) > 1 {
			st.Body.List = append(stmts[1:], st.Body.List...)
		}
	}
	afterInit := cloneDefined(child)
	if st.Cond != nil {
		transformExpr(st.Cond, child)
	}
	st.Body.List = transformStmtList(st.Body.List, child)
	if st.Else == nil {
		return
	}
	switch e := st.Else.(type) {
	case *ast.BlockStmt:
		e.List = transformStmtList(e.List, afterInit)
	case *ast.IfStmt:
		transformIf(e, afterInit)
	default:
		st.Else = firstStmt(transformStmt(e, afterInit))
	}
}

func transformSwitch(st *ast.SwitchStmt, defined map[string]bool) {
	child := cloneDefined(defined)
	if st.Init != nil {
		stmts := transformStmt(st.Init, child)
		st.Init = stmts[0]
		if extra := stmts[1:]; len(extra) > 0 {
			prependCase(st.Body, extra)
		}
	}
	if st.Tag != nil {
		transformExpr(st.Tag, child)
	}
	transformCases(st.Body, child)
}

func transformTypeSwitch(st *ast.TypeSwitchStmt, defined map[string]bool) {
	child := cloneDefined(defined)
	if st.Init != nil {
		stmts := transformStmt(st.Init, child)
		st.Init = stmts[0]
		if extra := stmts[1:]; len(extra) > 0 {
			prependCase(st.Body, extra)
		}
	}
	if st.Assign != nil {
		stmts := transformStmt(st.Assign, child)
		st.Assign = stmts[0]
		if extra := stmts[1:]; len(extra) > 0 {
			prependCase(st.Body, extra)
		}
	}
	transformCases(st.Body, child)
}

func prependCase(body *ast.BlockStmt, stmts []ast.Stmt) {
	if body == nil || len(stmts) == 0 || len(body.List) == 0 {
		return
	}
	clause, ok := body.List[0].(*ast.CaseClause)
	if !ok {
		return
	}
	clause.Body = append(append([]ast.Stmt{}, stmts...), clause.Body...)
}

func transformCases(body *ast.BlockStmt, parent map[string]bool) {
	if body == nil {
		return
	}
	for _, stmt := range body.List {
		clause, ok := stmt.(*ast.CaseClause)
		if !ok {
			continue
		}
		for _, expr := range clause.List {
			transformExpr(expr, parent)
		}
		clause.Body = transformStmtList(clause.Body, cloneDefined(parent))
	}
}

func transformSelect(st *ast.SelectStmt, defined map[string]bool) {
	if st.Body == nil {
		return
	}
	for _, stmt := range st.Body.List {
		clause, ok := stmt.(*ast.CommClause)
		if !ok || clause.Comm == nil {
			continue
		}
		child := cloneDefined(defined)
		stmts := transformStmt(clause.Comm, child)
		clause.Comm = stmts[0]
		body := append(append([]ast.Stmt{}, stmts[1:]...), clause.Body...)
		clause.Body = transformStmtList(body, child)
	}
}

func transformFor(st *ast.ForStmt, defined map[string]bool) {
	child := cloneDefined(defined)
	if st.Init != nil {
		stmts := transformStmt(st.Init, child)
		st.Init = stmts[0]
		if len(stmts) > 1 {
			st.Body.List = append(stmts[1:], st.Body.List...)
		}
	}
	if st.Cond != nil {
		transformExpr(st.Cond, child)
	}
	if st.Post != nil {
		stmts := transformStmt(st.Post, child)
		st.Post = stmts[0]
	}
	st.Body.List = transformStmtList(st.Body.List, cloneDefined(child))
}

func transformRange(st *ast.RangeStmt, defined map[string]bool) {
	transformExpr(st.X, defined)
	child := cloneDefined(defined)
	var names []string
	if st.Tok == token.DEFINE {
		names = append(names, bindRangeIdent(st.Key, child)...)
		names = append(names, bindRangeIdent(st.Value, child)...)
	} else {
		if st.Key != nil {
			transformExpr(st.Key, defined)
		}
		if st.Value != nil {
			transformExpr(st.Value, defined)
		}
	}
	st.Body.List = append(sinkStmts(names), st.Body.List...)
	st.Body.List = transformStmtList(st.Body.List, child)
}

func bindRangeIdent(expr ast.Expr, defined map[string]bool) []string {
	id, ok := expr.(*ast.Ident)
	if !ok || id.Name == "_" {
		if expr != nil && !ok {
			transformExpr(expr, defined)
		}
		return nil
	}
	defined[id.Name] = true
	return []string{id.Name}
}

func transformExpr(expr ast.Expr, defined map[string]bool) {
	switch ex := expr.(type) {
	case nil, *ast.BadExpr, *ast.BasicLit, *ast.Ident:
	case *ast.Ellipsis:
		if ex.Elt != nil {
			transformExpr(ex.Elt, defined)
		}
	case *ast.FuncLit:
		transformFuncLit(ex, defined)
	case *ast.CompositeLit:
		if ex.Type != nil {
			transformExpr(ex.Type, defined)
		}
		for _, elt := range ex.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				transformExpr(kv.Value, defined)
				continue
			}
			transformExpr(elt, defined)
		}
	case *ast.ParenExpr:
		transformExpr(ex.X, defined)
	case *ast.SelectorExpr:
		transformExpr(ex.X, defined)
	case *ast.IndexExpr:
		transformExpr(ex.X, defined)
		transformExpr(ex.Index, defined)
	case *ast.IndexListExpr:
		transformExpr(ex.X, defined)
		for _, idx := range ex.Indices {
			transformExpr(idx, defined)
		}
	case *ast.SliceExpr:
		transformExpr(ex.X, defined)
		transformExpr(ex.Low, defined)
		transformExpr(ex.High, defined)
		transformExpr(ex.Max, defined)
	case *ast.TypeAssertExpr:
		transformExpr(ex.X, defined)
		transformExpr(ex.Type, defined)
	case *ast.CallExpr:
		transformExpr(ex.Fun, defined)
		for _, arg := range ex.Args {
			transformExpr(arg, defined)
		}
	case *ast.StarExpr:
		transformExpr(ex.X, defined)
	case *ast.UnaryExpr:
		transformExpr(ex.X, defined)
	case *ast.BinaryExpr:
		transformExpr(ex.X, defined)
		transformExpr(ex.Y, defined)
	case *ast.KeyValueExpr:
		transformExpr(ex.Value, defined)
	default:
		panic(fmt.Sprintf("unhandled expression %T", expr))
	}
}

func transformFuncLit(fn *ast.FuncLit, defined map[string]bool) {
	child := cloneDefined(defined)
	var names []string
	if fn.Type != nil {
		names = append(names, bindFields(fn.Type.Params, child)...)
		bindFields(fn.Type.Results, child)
	}
	if fn.Body == nil {
		return
	}
	fn.Body.List = append(sinkStmts(names), fn.Body.List...)
	fn.Body.List = transformStmtList(fn.Body.List, child)
}

func bindFields(fields *ast.FieldList, defined map[string]bool) []string {
	if fields == nil {
		return nil
	}
	var names []string
	for _, field := range fields.List {
		if field.Type != nil {
			transformExpr(field.Type, defined)
		}
		for _, name := range field.Names {
			if name.Name == "_" {
				continue
			}
			defined[name.Name] = true
			names = append(names, name.Name)
		}
	}
	return names
}

func freeRank(name string) int {
	switch name {
	case "ctx":
		return 0
	case "client":
		return 1
	case "err":
		return 2
	default:
		return 3
	}
}

type scope struct {
	defined map[string]bool
	free    map[string]bool
}

func walkStmts(stmts []ast.Stmt, defined, free map[string]bool) {
	s := &scope{defined: defined, free: free}
	for _, stmt := range stmts {
		s.stmt(stmt)
	}
}

func (s *scope) child() *scope {
	defined := make(map[string]bool, len(s.defined))
	for k, v := range s.defined {
		defined[k] = v
	}
	return &scope{defined: defined, free: s.free}
}

func (s *scope) use(name string) {
	if name == "" || name == "_" || s.defined[name] || excludedIdent(name) {
		return
	}
	s.free[name] = true
}

func (s *scope) stmt(stmt ast.Stmt) {
	switch st := stmt.(type) {
	case *ast.BadStmt, *ast.EmptyStmt, *ast.BranchStmt:
	case *ast.DeclStmt:
		s.decl(st.Decl)
	case *ast.LabeledStmt:
		s.stmt(st.Stmt)
	case *ast.ExprStmt:
		s.expr(st.X)
	case *ast.SendStmt:
		s.expr(st.Chan)
		s.expr(st.Value)
	case *ast.IncDecStmt:
		s.expr(st.X)
	case *ast.AssignStmt:
		s.assign(st)
	case *ast.GoStmt:
		s.expr(st.Call)
	case *ast.DeferStmt:
		s.expr(st.Call)
	case *ast.ReturnStmt:
		for _, result := range st.Results {
			s.expr(result)
		}
	case *ast.BlockStmt:
		walkStmts(st.List, s.child().defined, s.free)
	case *ast.IfStmt:
		s.ifStmt(st)
	case *ast.SwitchStmt:
		s.switchStmt(st)
	case *ast.TypeSwitchStmt:
		s.typeSwitch(st)
	case *ast.SelectStmt:
		s.selectStmt(st)
	case *ast.ForStmt:
		s.forStmt(st)
	case *ast.RangeStmt:
		s.rangeStmt(st)
	default:
		panic(fmt.Sprintf("unhandled statement %T", stmt))
	}
}

func (s *scope) decl(decl ast.Decl) {
	gen, ok := decl.(*ast.GenDecl)
	if !ok {
		return
	}
	for _, spec := range gen.Specs {
		switch sp := spec.(type) {
		case *ast.ValueSpec:
			if sp.Type != nil {
				s.expr(sp.Type)
			}
			for _, value := range sp.Values {
				s.expr(value)
			}
			for _, name := range sp.Names {
				s.defined[name.Name] = true
			}
		case *ast.TypeSpec:
			if sp.Type != nil {
				s.expr(sp.Type)
			}
			s.defined[sp.Name.Name] = true
		}
	}
}

func (s *scope) assign(st *ast.AssignStmt) {
	for _, rhs := range st.Rhs {
		s.expr(rhs)
	}
	if st.Tok != token.DEFINE {
		for _, lhs := range st.Lhs {
			s.expr(lhs)
		}
		return
	}
	for _, lhs := range st.Lhs {
		id, ok := lhs.(*ast.Ident)
		if !ok {
			s.expr(lhs)
			continue
		}
		if id.Name == "_" {
			continue
		}
		s.defined[id.Name] = true
	}
}

func (s *scope) ifStmt(st *ast.IfStmt) {
	child := s.child()
	if st.Init != nil {
		child.stmt(st.Init)
	}
	child.expr(st.Cond)
	walkStmts(st.Body.List, child.child().defined, s.free)
	if st.Else == nil {
		return
	}
	elseScope := child.child()
	switch e := st.Else.(type) {
	case *ast.BlockStmt:
		walkStmts(e.List, elseScope.defined, s.free)
	default:
		elseScope.stmt(e)
	}
}

func (s *scope) switchStmt(st *ast.SwitchStmt) {
	child := s.child()
	if st.Init != nil {
		child.stmt(st.Init)
	}
	if st.Tag != nil {
		child.expr(st.Tag)
	}
	s.cases(st.Body, child)
}

func (s *scope) typeSwitch(st *ast.TypeSwitchStmt) {
	child := s.child()
	if st.Init != nil {
		child.stmt(st.Init)
	}
	if st.Assign != nil {
		child.stmt(st.Assign)
	}
	s.cases(st.Body, child)
}

func (s *scope) cases(body *ast.BlockStmt, parent *scope) {
	if body == nil {
		return
	}
	for _, stmt := range body.List {
		clause, ok := stmt.(*ast.CaseClause)
		if !ok {
			parent.stmt(stmt)
			continue
		}
		for _, expr := range clause.List {
			parent.expr(expr)
		}
		walkStmts(clause.Body, parent.child().defined, s.free)
	}
}

func (s *scope) selectStmt(st *ast.SelectStmt) {
	if st.Body == nil {
		return
	}
	for _, stmt := range st.Body.List {
		clause, ok := stmt.(*ast.CommClause)
		if !ok {
			s.stmt(stmt)
			continue
		}
		child := s.child()
		if clause.Comm != nil {
			child.stmt(clause.Comm)
		}
		walkStmts(clause.Body, child.defined, s.free)
	}
}

func (s *scope) forStmt(st *ast.ForStmt) {
	child := s.child()
	if st.Init != nil {
		child.stmt(st.Init)
	}
	if st.Cond != nil {
		child.expr(st.Cond)
	}
	if st.Post != nil {
		child.stmt(st.Post)
	}
	walkStmts(st.Body.List, child.child().defined, s.free)
}

func (s *scope) rangeStmt(st *ast.RangeStmt) {
	s.expr(st.X)
	child := s.child()
	if st.Tok == token.DEFINE {
		if id, ok := st.Key.(*ast.Ident); ok && id.Name != "_" {
			child.defined[id.Name] = true
		} else if st.Key != nil {
			s.expr(st.Key)
		}
		if id, ok := st.Value.(*ast.Ident); ok && id.Name != "_" {
			child.defined[id.Name] = true
		} else if st.Value != nil {
			s.expr(st.Value)
		}
	} else {
		if st.Key != nil {
			s.expr(st.Key)
		}
		if st.Value != nil {
			s.expr(st.Value)
		}
	}
	walkStmts(st.Body.List, child.defined, s.free)
}

func (s *scope) expr(expr ast.Expr) {
	switch ex := expr.(type) {
	case nil, *ast.BadExpr, *ast.BasicLit:
	case *ast.Ident:
		s.use(ex.Name)
	case *ast.Ellipsis:
		if ex.Elt != nil {
			s.expr(ex.Elt)
		}
	case *ast.FuncLit:
		s.funcLit(ex)
	case *ast.CompositeLit:
		if ex.Type != nil {
			s.expr(ex.Type)
		}
		for _, elt := range ex.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				s.expr(kv.Value)
				continue
			}
			s.expr(elt)
		}
	case *ast.ParenExpr:
		s.expr(ex.X)
	case *ast.SelectorExpr:
		s.expr(ex.X)
	case *ast.IndexExpr:
		s.expr(ex.X)
		s.expr(ex.Index)
	case *ast.IndexListExpr:
		s.expr(ex.X)
		for _, idx := range ex.Indices {
			s.expr(idx)
		}
	case *ast.SliceExpr:
		s.expr(ex.X)
		s.expr(ex.Low)
		s.expr(ex.High)
		s.expr(ex.Max)
	case *ast.TypeAssertExpr:
		s.expr(ex.X)
		s.expr(ex.Type)
	case *ast.CallExpr:
		s.expr(ex.Fun)
		for _, arg := range ex.Args {
			s.expr(arg)
		}
	case *ast.StarExpr:
		s.expr(ex.X)
	case *ast.UnaryExpr:
		s.expr(ex.X)
	case *ast.BinaryExpr:
		s.expr(ex.X)
		s.expr(ex.Y)
	case *ast.KeyValueExpr:
		s.expr(ex.Value)
	default:
		panic(fmt.Sprintf("unhandled expression %T", expr))
	}
}

func (s *scope) funcLit(fn *ast.FuncLit) {
	child := s.child()
	if fn.Type != nil {
		child.bindFields(fn.Type.Params)
		child.bindFields(fn.Type.Results)
	}
	if fn.Body != nil {
		walkStmts(fn.Body.List, child.defined, s.free)
	}
}

func (s *scope) bindFields(fields *ast.FieldList) {
	if fields == nil {
		return
	}
	for _, field := range fields.List {
		if field.Type != nil {
			s.expr(field.Type)
		}
		for _, name := range field.Names {
			s.defined[name.Name] = true
		}
	}
}

func excludedIdent(name string) bool {
	switch name {
	case "vergeos", "fmt", "time", "sync", "io", "os", "log", "context", "ptr",
		"true", "false", "nil", "iota",
		"append", "cap", "clear", "close", "complex", "copy", "delete", "imag",
		"len", "make", "max", "min", "new", "panic", "print", "println", "real", "recover",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"float32", "float64", "complex64", "complex128",
		"byte", "rune", "string", "bool", "error", "any", "comparable":
		return true
	default:
		return false
	}
}

func renderSnippetProgram(snippets []analyzedSnippet, overrides []map[string]string) (string, []snippetRange) {
	var b strings.Builder
	b.WriteString(`package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

func ptr[T any](v T) *T { return &v }

func main() {}

func keepImports() {
	fmt.Sprint(context.Background(), io.EOF, os.Stdout, time.Second, sync.Mutex{}, ptr(0))
	log.Print((*vergeos.Client)(nil))
}

`)
	ranges := make([]snippetRange, len(snippets))
	for i, snip := range snippets {
		start := strings.Count(b.String(), "\n") + 1
		fmt.Fprintf(&b, "// %s:%d\nfunc snippet%d() error {\n", snip.file, snip.line, i)
		for _, name := range snip.free {
			if name == "ctx" {
				b.WriteString("\tctx := context.Background()\n")
				continue
			}
			fmt.Fprintf(&b, "\tvar %s %s\n", name, dummyType(name, overrides[i]))
		}
		b.WriteString(indentSnippet(snip.rewritten))
		if !strings.HasSuffix(snip.rewritten, "\n") {
			b.WriteByte('\n')
		}
		b.WriteString("\treturn nil\n}\n\n")
		end := strings.Count(b.String(), "\n")
		ranges[i] = snippetRange{index: i, startLine: start, endLine: end}
	}
	return b.String(), ranges
}

func indentSnippet(code string) string {
	var b strings.Builder
	for _, line := range strings.Split(code, "\n") {
		if strings.TrimSpace(line) == "" {
			b.WriteByte('\n')
			continue
		}
		b.WriteByte('\t')
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func dummyType(name string, overrides map[string]string) string {
	if typ, ok := overrides[name]; ok {
		return typ
	}
	switch name {
	case "client":
		return "*vergeos.Client"
	case "err":
		return "error"
	case "vm":
		return "*vergeos.VM"
	case "vmIDs":
		return "[]int"
	case "reader":
		return "io.ReadCloser"
	case "destination":
		return "io.Writer"
	case "fileSize", "since":
		return "int64"
	case "publicKeyPEM", "privateKeyPEM", "chainPEM":
		return "string"
	default:
		return "int"
	}
}

func compileSnippetProgram(t *testing.T, root, src string) (string, error) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	goMod := fmt.Sprintf(`module snippetcheck

go 1.21

require github.com/verge-io/govergeos v0.0.0

replace github.com/verge-io/govergeos => %s
`, filepath.ToSlash(root))
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", "-mod=mod", "-o", filepath.Join(dir, "snippetcheck"), ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GO111MODULE=on", "GOSUMDB=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return "", nil
}

var (
	cannotUseVar = regexp.MustCompile(`cannot use (\w+) \(variable of type ([^)]+)\) as (\S+)`)
	cannotUsePtr = regexp.MustCompile(`cannot use ptr\((\w+)\) \(value of type \*[^)]+\) as \*(\S+)`)
	errorLine    = regexp.MustCompile(`(?m)^(?:\./)?main\.go:(\d+):`)
)

func applySnippetTypeOverrides(stderr string, ranges []snippetRange, overrides []map[string]string) bool {
	changed := false
	lines := strings.Split(stderr, "\n")
	for _, line := range lines {
		loc := errorLine.FindStringSubmatch(line)
		if loc == nil {
			continue
		}
		var n int
		fmt.Sscanf(loc[1], "%d", &n)
		idx := snippetIndexAt(ranges, n)
		if idx < 0 {
			continue
		}
		if m := cannotUsePtr.FindStringSubmatch(line); m != nil {
			if setOverride(overrides[idx], m[1], m[2]) {
				changed = true
			}
			continue
		}
		if m := cannotUseVar.FindStringSubmatch(line); m != nil {
			want := m[3]
			if strings.HasPrefix(want, "*") {
				continue
			}
			if setOverride(overrides[idx], m[1], want) {
				changed = true
			}
		}
	}
	return changed
}

func snippetIndexAt(ranges []snippetRange, line int) int {
	for _, r := range ranges {
		if line >= r.startLine && line <= r.endLine {
			return r.index
		}
	}
	return -1
}

func setOverride(overrides map[string]string, name, typ string) bool {
	if _, ok := overrides[name]; ok {
		return false
	}
	overrides[name] = typ
	return true
}
