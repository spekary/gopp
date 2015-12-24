package main
import (
	"fmt"
	"strings"
	"unicode/utf8"
	"regexp"
)

type ast []fmt.Stringer

type memberDef struct {
	name string
	comments string
}

type funcDef struct {
	name string
	params string
	body string
	comments string
}

type classDef struct {
	name string
	extends string
	constructorParams string
	members []memberDef
	funcs []funcDef
	comment string
}

var classes map[string]*classDef = make(map[string]*classDef)

func (a ast) String() string {
	var out string

	for _,s := range a {
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
			out = append(out, stringer(comment + item.val))
			comment = ""
		case itemClass:
			out = append(out, parseClass(item.val, l, comment))
			comment = ""
		case itemError:
			out = append(out, stringer("Error: " + item.val))
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
			out = append(out, stringer("Unexpected token: " + item.String()))
			break forloop
		}
	}

	return out
}


func parseClass(className string, l *lexer, comment string) fmt.Stringer {
	var curComment string
	var class classDef

	//_ = "breakpoint"

	class.comment = comment
	class.name = className
	item := l.nextItem()

	switch item.typ {
	case itemEOF:
		return stringer("Unexpected EOF")
	case itemExtends:
		class.extends = item.val
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
			class.members = append(class.members, memberDef{item.val, curComment})
			curComment = ""
		case itemFunc:
			f,err := parseFunc(item.val, l, curComment)
			if (err != "") {
				return stringer(err)
			}
			class.funcs = append(class.funcs, f)
			if (item.val == "Construct_") {
				// The constructor
				class.constructorParams = f.params
			}
			curComment = ""
		case itemRightDelim:
			break forloop
		}
	}

	classes[class.name] = &class

	return &class
}

func parseFunc(name string, l *lexer, comment string) (f funcDef, e string) {
	params := l.nextItem()
	if (params.typ != itemFuncParams) {
		e = "Error: function parameters expected, got " + params.String()
		return
	}
	body := l.nextItem()
	if (body.typ != itemFuncBody) {
		e = "Error: function body expected, got " + body.String()
		return
	}

	f = funcDef{name, params.val, body.val, comment}
	return
}

/**
Output the class as a combination interface and struct.
 */
func (c *classDef) String() string {
	var out string

	// output interface
	out = "type " + c.name + " interface {\n"
	out += "\t" + c.extends + "\n"
	for _,f := range c.funcs {
		out += "\t" + f.name + " " + f.params + "\n"
	}
	out += "}\n\n"

	// output the struct
	out += "type " + c.name + "_ struct {\n"
	out += "\t" + c.extends + "_\n"
	for _,m := range c.members {
		if (m.comments != "") {
			out += "\t" + m.comments
		}
		out += "\t" + m.name
	}

	out += "}\n\n"

	out += c.outNew()
	out += c.outFuncs()
	out += c.outReflect()

	return out
}

func (c *classDef) outNew() string {
	var out string
	var vars []string

	_ = "breakpoint"

	params := c.constructorParams
	var parentClass *classDef

	for {
		if (params == "") {
			parentClass = classes[c.extends]
			if (parentClass == nil) {
				params = "()"
				break
			}
			params = parentClass.constructorParams
		} else {
			break
		}
	}

	// pull apart params to find variable names
	sVarList := params[strings.Index(params, "(") + 1:strings.LastIndex(params, ")")]
	aVar := strings.Split(sVarList, ",")
	for _,sVarDec := range aVar {
		fields := strings.Fields(sVarDec)
		if len(fields) > 0 {
			vars = append(vars, fields[0])
		}
	}
	sVarList = strings.Join(vars, ",")

	out = "/**\n" +
		"New" + c.name + " creates a new " + c.name + " object.\n" +
		"*/\n"
	out += "func New" + c.name + params + " " + c.name + " {\n"
	out += "\tthis := " + c.name + "_{}\n"
	out += "\tthis.Init_(&this)\n"
	out += "\tthis.Construct_(" + sVarList + ")\n"
	out += "\treturn this.I_().(" + c.name + ")\n"
	out += "}\n\n"

	return out
}


func (c *classDef) outFuncs() string {
	var out string

	for _,f := range c.funcs {
		out += "func (this *" + c.name + "_) " + f.name + f.params
		out += c.processFuncBody(f)
		out += "\n\n"
	}

	return out
}

func (c *classDef) outReflect() string {
	parts := strings.Split(c.extends, ".")
	extendsName := parts[len(parts)-1]

	out :=
`
func (this *` + c.name + `_) IsA(className string) bool {
	if (className == "` + c.name + `") {
		return true
	}
	return this.` + extendsName + `_.IsA(className)
}

func (this *` + c.name + `_) Class() string {
	return "` + c.name + `"
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
	var  out string
	var start, pos int

	in := f.body
	len := len(in)

mainfor:
	for {
		r, width := utf8.DecodeRuneInString(in[pos:])
		pos += width
		if (pos >= len) {
			break
		}

		switch r {
		case '"', '\'':
			strLength := strings.IndexRune(in[pos:], r)
			if (strLength < 0) {
				panic ("String has open quote, but not close quote.")
			}
			pos += strLength + 1

			out += convertBody(in[start:pos], c)
			start = pos
		case '/':
			c2, width := utf8.DecodeRuneInString(in[pos:])
			pos += width
			if (pos >= len) {
				break mainfor
			}

			if (c2 == '/') {
				out += convertBody(in[start:pos - 2], c)	// send out current stuff, and then skip the rest of the comment
				commentLength := strings.Index(in[pos:], "\n")
				start = pos + commentLength + 1
				pos = start
			} else if (c2 == '*') {
				out += convertBody(in[start:pos - 2], c)	// send out current stuff, and then skip the rest of the comment
				commentLength := strings.Index(in[pos:], "*/")
				start = pos + commentLength + 2
				pos = start
			}
		}
		if pos >= len {
			break
		}

	}

	out += convertBody(in[start:pos], c)
	return out
}

func convertBody(in string, c *classDef) string {
	parts := strings.Split(c.extends, ".")
	extendsName := parts[len(parts)-1]	// get last item

	//_ = "breakpoint"
	rFunc, _ := regexp.Compile("this\\.([a-zA-Z0-9_]+) ?\\(")
	out := rFunc.ReplaceAllString(in, "this.I_().(" + c.name + ").$1(")
	out = strings.Replace(out, "parent::", "this." + extendsName + "_.", -1)
	return out
}


type stringer string

func (s stringer) String() string { return string(s) }

