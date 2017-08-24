package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"unicode/utf8"
)

type ast []fmt.Stringer

type memberDef struct {
	Name    string
	Comment string
}

type funcDef struct {
	Name          string
	Params        string
	Body          string
	ProcessedBody string
	Comment       string
	IsOverride    bool
}

type classDef struct {
	Name              string
	Extends           string
	ConstructorParams string
	Members           []memberDef
	Funcs             []funcDef
	Comment           string
	ParentVarList     string
	Receiver          string
	Parent            string
}

var classes map[string]*classDef = make(map[string]*classDef)

func (a ast) String() string {
	var out string

	for _, s := range a {
		out += s.String()
	}

	return out
}

/**
Process a string that is gopp formatted code, and return the go code
*/
func parse(l *lexer) ast {
	var out ast
	var comment string

forloop:
	for {
		item := l.nextItem()

		//fmt.Printf("%v", item)

		switch item.typ {
		case itemText:
			out = append(out, stringer(comment+item.val))
			comment = ""
		case itemClass:
			out = append(out, parseClass(item.val, l, comment))
			comment = ""
		case itemError:
			out = append(out, stringer("Error: "+item.val))
			break forloop
		case itemComment:
			comment += item.val
		case itemLineComment:
			comment += item.val
		case itemEOF:
			if len(comment) > 0 {
				out = append(out, stringer(comment))
			}
			break forloop

		case itemPackage:
			out = append(out, stringer(item.val))
		default:
			out = append(out, stringer("Unexpected token: "+item.String()))
			break forloop
		}
	}

	return out
}

func parseClass(className string, l *lexer, comment string) fmt.Stringer {
	var curComment string
	var class classDef

	class.Comment = comment
	class.Name = className
	class.Receiver = strings.ToLower(string(class.Name[0])) + "_"

	item := l.nextItem()

	switch item.typ {
	case itemEOF:
		return stringer("Unexpected EOF")
	case itemExtends:
		class.Extends = item.val
		// keep going
	default:
		return stringer("Error: Extends keyword expected, got " + item.String())
	}

	item = l.nextItem()

	switch item.typ {
	case itemEOF:
		return stringer("Unexpected EOF")
	case itemLeftDelim:
	// keep going
	default:
		return stringer("Error: left delimiter expected, got " + item.String())
	}

	var isOverride bool

	// TODO: put comment after leftDelim into tree somehow
forloop:
	for {
		item = l.nextItem()
		switch item.typ {
		case itemEOF:
			return stringer("Unexpected EOF")
		case itemComment:
			curComment += item.val
		case itemLineComment:
			curComment += item.val
		case itemMember:
			class.Members = append(class.Members, memberDef{item.val, curComment})
			curComment = ""
		case itemOverride:
			isOverride = true
		case itemFunc:
			f, err := parseFunc(item.val, l, curComment)
			f.IsOverride = isOverride
			if err != "" {
				return stringer(err)
			}
			// Special constructor function
			if item.val == "Construct" {
				// The constructor
				class.ConstructorParams = strings.Trim(f.Params, "( ) ")
				f.IsOverride = true // constructor always overrides. This means base class MUST have a Construct function.
			}
			class.Funcs = append(class.Funcs, f)
			curComment = ""
			isOverride = false
		case itemRightDelim:
			break forloop
		}
	}
	class.Parent = parentName(class.Extends)

	classes[class.Name] = &class

	return &class
}

func parseFunc(name string, l *lexer, comment string) (f funcDef, e string) {
	params := l.nextItem()
	if params.typ != itemFuncParams {
		e = "Error: function parameters expected, got " + params.String()
		return
	}
	body := l.nextItem()
	if body.typ != itemFuncBody {
		e = "Error: function body expected, got " + body.String()
		return
	}

	f = funcDef{name, params.val, body.val, "", comment, false}
	return
}

// This struct will be sent in to the template to generate the go file
type TmplClass struct {
	Comments string
	Params   string
	Name     string
	VarList  string
	Receiver string
	Parent   string
	Members  []TmplMember
	Methods  []TmplMethod
}

type TmplMember struct {
	Comment string
	Name    string
}

type TmplMethod struct {
	Name   string
	Params string
	Body   string
}

/**
Output the class as a combination interface and struct.
*/
func (c *classDef) String() string {
	var vars []string

	//extends := c.Extends
	params := c.ConstructorParams
	//var parentClass *classDef
	/*
		for {
			if params == "" {
				parentClass = classes[extends]
				if parentClass == nil {
					params = "()"
					break
				}
				params = parentClass.ConstructorParams
				extends = parentClass.Extends
			} else {
				break
			}
		}*/

	// pull apart params to find variable names
	//sVarList := params[strings.Index(params, "(") + 1:strings.LastIndex(params, ")")]
	aVar := strings.Split(params, ",")
	for _, sVarDec := range aVar {
		fields := strings.Fields(sVarDec)
		if len(fields) > 0 {
			vars = append(vars, fields[0])
		}
	}
	c.ParentVarList = strings.Join(vars, ",")

	for i, f := range c.Funcs {
		c.Funcs[i].ProcessedBody = c.processFuncBody(f)
	}

	var tmpl = template.Must(template.New("Class").Parse(tmplString))

	var tpl bytes.Buffer

	if err := tmpl.Execute(&tpl, c); err != nil {
		panic(err)
	}

	return tpl.String()
}

type NewStruct struct {
	Params   string
	Name     string
	VarList  string
	Receiver string
}

// Write out the New function
func (c *classDef) outNew() string {
	var vars []string

	extends := c.Extends
	params := c.ConstructorParams
	var parentClass *classDef

	for {
		if params == "" {
			parentClass = classes[extends]
			if parentClass == nil {
				params = "()"
				break
			}
			params = parentClass.ConstructorParams
			extends = parentClass.Extends
		} else {
			break
		}
	}

	// pull apart params to find variable names
	sVarList := params[strings.Index(params, "(")+1 : strings.LastIndex(params, ")")]
	aVar := strings.Split(sVarList, ",")
	for _, sVarDec := range aVar {
		fields := strings.Fields(sVarDec)
		if len(fields) > 0 {
			vars = append(vars, fields[0])
		}
	}
	sVarList = strings.Join(vars, ",")

	tmplStruct := NewStruct{Params: params, Name: c.Name, VarList: sVarList, Receiver: strings.ToLower(string(c.Name[0]))}

	var tmpl = template.Must(template.New("New").Parse(tmplNew))

	var tpl bytes.Buffer

	tmpl.Execute(&tpl, tmplStruct)

	return tpl.String()
}

const tmplNew = `
// New {{.Name}} creates a new {{.Name}} object and returns its matching interface
func New{{.Name}} {{.Params}} {{.Name}}I {
	{{.Receiver}} := {{.Name}}{}
	{{.Receiver}}.Init(&{{.Receiver}})
	{{.Receiver}}.Construct({{.VarList}})
	return {{.Receiver}}.I().({{.Name}}I)
}
`

func (c *classDef) outFuncs() string {
	var out string

	for _, f := range c.Funcs {
		out += "func (this *" + c.Name + ") " + f.Name + f.Params
		out += c.processFuncBody(f)
		out += "\n\n"
	}

	return out
}

func (c *classDef) outReflect() string {
	parts := strings.Split(c.Extends, ".")
	extendsName := parts[len(parts)-1]

	out :=
		`
func (this *` + c.Name + `) IsA(className string) bool {
	if (className == "` + c.Name + `") {
		return true
	}
	return this.` + extendsName + `.IsA(className)
}

func (this *` + c.Name + `) Class() string {
	return "` + c.Name + `"
}
`
	return out
}

/**
Takes the raw body coming from the class definition and changes it to be go compatible. Some specific things it does:
- Strips all comments
- Converts method calls to be called against the interface, so that they are virtually called
- Converts template types to physical types
*/
func (c *classDef) processFuncBody(f funcDef) string {
	var out string
	var start, pos int

	in := f.Body
	length := len(in)

mainfor:
	for {
		r, width := utf8.DecodeRuneInString(in[pos:])
		pos += width
		if pos >= length {
			break
		}

		switch r {
		case '"', '\'':
			strLength := strings.IndexRune(in[pos:], r)
			if strLength < 0 {
				panic("String has open quote, but not close quote.")
			}
			pos += strLength + 1

			out += convertBody(in[start:pos], c)
			start = pos
		case '/':
			c2, width := utf8.DecodeRuneInString(in[pos:])
			pos += width
			if pos >= length {
				break mainfor
			}

			if c2 == '/' {
				out += convertBody(in[start:pos-2], c) // send out current stuff, and then skip the rest of the comment
				commentLength := strings.Index(in[pos:], "\n")
				start = pos + commentLength + 1
				pos = start
			} else if c2 == '*' {
				out += convertBody(in[start:pos-2], c) // send out current stuff, and then skip the rest of the comment
				commentLength := strings.Index(in[pos:], "*/")
				start = pos + commentLength + 2
				pos = start
			}
		}
		if pos >= length {
			break
		}

	}

	out += convertBody(in[start:pos], c)

	out = strings.Trim(out, "{} ")
	out = strings.TrimSpace(out)
	return out
}

func convertBody(in string, c *classDef) string {
	parts := strings.Split(c.Extends, ".")
	extendsName := parts[len(parts)-1] // get last item

	rFunc, _ := regexp.Compile("this\\.([a-zA-Z0-9_]+) ?\\(")
	out := rFunc.ReplaceAllString(in, c.Receiver+".I().("+c.Name+"I).$1(")
	out = strings.Replace(out, "parent::", c.Receiver+"."+extendsName+".", -1)
	out = strings.Replace(out, "this.", c.Receiver+".", -1)
	return out
}

// Simply strips off the package from the extends name
func parentName(extends string) string {
	a := strings.Split(extends, ".")
	return a[len(a)-1]
}

type stringer string

func (s stringer) String() string { return string(s) }

const tmplString = `
{{.Comment}}
type {{.Name}}I interface {
	{{.Extends}}I
{{range .Funcs}} {{if and (not (eq .Name "Construct")) (not .IsOverride)}}
	{{.Name}}{{.Params}}{{end}}{{end}}
}

type {{.Name}} struct {
	{{.Extends}}
{{range .Members}}{{if .Comment}}	{{.Comment}}
{{end}}	{{.Name}}{{end}}
}

// New {{.Name}} creates a new {{.Name}} object and returns its matching interface
func New{{.Name}} ({{.ConstructorParams}}) {{.Name}}I {
	{{.Receiver}} := {{.Name}}{}
	{{.Receiver}}.Init(&{{.Receiver}})
	{{.Receiver}}.Construct({{.ParentVarList}})
	return {{.Receiver}}.I().({{.Name}}I)
}


{{range .Funcs}}
func ({{$.Receiver}} *{{$.Name}}) {{.Name}} {{.Params}} {
{{.ProcessedBody}}
}
{{end}}
func ({{$.Receiver}} *{{$.Name}}) IsA(className string) bool {
	if className == "{{$.Name}}" {
		return true
	}
	return {{$.Receiver}}.{{$.Parent}}.IsA(className)
}

func ({{$.Receiver}} *{{$.Name}}) Class() string {
	return "{{$.Name}}"
}
`
