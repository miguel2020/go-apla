// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package templatev2

import (
	"encoding/json"
	//	"fmt"
	"html"
	"strings"
	//	"unicode/utf8"

	"github.com/shopspring/decimal"
)

const (
	tagText = `text`
)

type node struct {
	Tag      string                 `json:"tag"`
	Attr     map[string]interface{} `json:"attr,omitempty"`
	Text     string                 `json:"text,omitempty"`
	Children []*node                `json:"children,omitempty"`
	Tail     []*node                `json:"tail,omitempty"`
}

type parFunc struct {
	Owner *node
	Node  *node
	Vars  *map[string]string
	Pars  *map[string]string
	Tails *[]*[]string
}

type nodeFunc func(par parFunc) string

type tplFunc struct {
	Func   nodeFunc // process function
	Full   nodeFunc // full process function
	Tag    string   // HTML tag
	Params string   // names of parameters
}

type tailInfo struct {
	tplFunc
	Last bool
}

type forTails struct {
	Tails map[string]tailInfo
}

var (
	funcs = map[string]tplFunc{
		`Div`:       {defaultTag, defaultTag, `div`, `Class,Body`},
		`Em`:        {defaultTag, defaultTag, `em`, `Body,Class`},
		`Form`:      {defaultTag, defaultTag, `form`, `Class,Body`},
		`InputErr`:  {defaultTag, defaultTag, `inputerr`, `*`},
		`Label`:     {defaultTag, defaultTag, `label`, `Body,Class,For`},
		`MenuGroup`: {defaultTag, defaultTag, `menugroup`, `Title,Body,Icon`},
		`MenuItem`:  {defaultTag, defaultTag, `menuitem`, `Title,Page,PageParams,Icon`},
		`P`:         {defaultTag, defaultTag, `p`, `Body,Class`},
		`Span`:      {defaultTag, defaultTag, `span`, `Body,Class`},
		`Strong`:    {defaultTag, defaultTag, `strong`, `Body,Class`},
		`Style`:     {defaultTag, defaultTag, `style`, `Css`},
	}
	tails = map[string]forTails{
		`button`: {map[string]tailInfo{
			`Alert`: {tplFunc{alertTag, alertFull, `alert`, `Text,ConfirmButton,CancelButton,Icon`}, true},
		}},
		`if`: {map[string]tailInfo{
			`Else`:   {tplFunc{elseTag, elseFull, `else`, `Body`}, true},
			`ElseIf`: {tplFunc{elseifTag, elseifFull, `elseif`, `Condition,Body`}, false},
		}},
		`input`: {map[string]tailInfo{
			`Validate`: {tplFunc{validateTag, validateFull, `validate`, `*`}, true},
		}},
	}
	modes = [][]rune{{'(', ')'}, {'{', '}'}}
)

func init() {
	funcs[`Button`] = tplFunc{buttonTag, buttonTag, `button`, `Body,Page,Class,Contract,Params,PageParams`}
	funcs[`If`] = tplFunc{ifTag, ifFull, `if`, `Condition,Body`}
	funcs[`Input`] = tplFunc{inputTag, inputTag, `input`, `Name,Class,Placeholder,Type,Value`}
}

func setAttr(par parFunc, name string) {
	if len((*par.Pars)[name]) > 0 {
		par.Node.Attr[strings.ToLower(name)] = (*par.Pars)[name]
	}
}

func setAllAttr(par parFunc) {
	for key, v := range *par.Pars {
		if key != `Body` && len(v) > 0 {
			par.Node.Attr[strings.ToLower(key)] = v
		}
	}
}

func defaultTag(par parFunc) string {
	setAllAttr(par)
	par.Owner.Children = append(par.Owner.Children, par.Node)
	return ``
}

func alertTag(par parFunc) string {
	setAllAttr(par)
	par.Owner.Attr[`alert`] = par.Node.Attr
	return ``
}

func alertFull(par parFunc) string {
	setAllAttr(par)
	par.Owner.Tail = append(par.Owner.Tail, par.Node)
	return ``
}

func validateTag(par parFunc) string {
	setAllAttr(par)
	par.Owner.Attr[`validate`] = par.Node.Attr
	return ``
}

func validateFull(par parFunc) string {
	setAllAttr(par)
	par.Owner.Tail = append(par.Owner.Tail, par.Node)
	return ``
}

func defaultTail(par parFunc, tag string) {
	if par.Tails != nil {
		for _, v := range *par.Tails {
			name := (*v)[len(*v)-1]
			curFunc := tails[tag].Tails[name].tplFunc
			pars := (*v)[:len(*v)-1]
			callFunc(&curFunc, par.Node, par.Vars, &pars, nil)
		}
	}
}

func inputTag(par parFunc) string {
	defaultTag(par)
	defaultTail(par, `input`)
	return ``
}

func buttonTag(par parFunc) string {
	defaultTag(par)
	if len((*par.Pars)[`Params`]) > 0 {
		imap := make(map[string]string)
		for _, v := range strings.Split((*par.Pars)[`Params`], `,`) {
			v = strings.TrimSpace(v)
			if off := strings.IndexByte(v, '='); off == -1 {
				imap[v] = v
			} else {
				imap[strings.TrimSpace(v[:off])] = strings.TrimSpace(v[off+1:])
			}
		}
		if len(imap) > 0 {
			par.Node.Attr[`params`] = imap
		}
	}
	defaultTail(par, `button`)
	return ``
}

func ifValue(val string) bool {
	var sep string
	if strings.Index(val, `;base64`) < 0 {
		for _, item := range []string{`==`, `!=`, `<=`, `>=`, `<`, `>`} {
			if strings.Index(val, item) >= 0 {
				sep = item
				break
			}
		}
	}
	cond := []string{val}
	if len(sep) > 0 {
		cond = strings.SplitN(val, sep, 2)
		cond[0], cond[1] = strings.Trim(cond[0], `"`), strings.Trim(cond[1], `"`)
	}
	switch sep {
	case ``:
		return len(val) > 0 && val != `0` && val != `false`
	case `==`:
		return len(cond) == 2 && strings.TrimSpace(cond[0]) == strings.TrimSpace(cond[1])
	case `!=`:
		return len(cond) == 2 && strings.TrimSpace(cond[0]) != strings.TrimSpace(cond[1])
	case `>`, `<`, `<=`, `>=`:
		ret0, _ := decimal.NewFromString(strings.TrimSpace(cond[0]))
		ret1, _ := decimal.NewFromString(strings.TrimSpace(cond[1]))
		if len(cond) == 2 {
			var bin bool
			if sep == `>` || sep == `<=` {
				bin = ret0.Cmp(ret1) > 0
			} else {
				bin = ret0.Cmp(ret1) < 0
			}
			if sep == `<=` || sep == `>=` {
				bin = !bin
			}
			return bin
		}
	}
	return false
}

func ifTag(par parFunc) string {
	cond := ifValue((*par.Pars)[`Condition`])
	if cond {
		for _, item := range par.Node.Children {
			par.Owner.Children = append(par.Owner.Children, item)
		}
	}
	if !cond && par.Tails != nil {
		for _, v := range *par.Tails {
			name := (*v)[len(*v)-1]
			curFunc := tails[`if`].Tails[name].tplFunc
			pars := (*v)[:len(*v)-1]
			callFunc(&curFunc, par.Owner, par.Vars, &pars, nil)
			if (*par.Vars)[`_cond`] == `1` {
				(*par.Vars)[`_cond`] = `0`
				break
			}
		}
	}
	return ``
}

func ifFull(par parFunc) string {
	setAttr(par, `Condition`)
	par.Owner.Children = append(par.Owner.Children, par.Node)
	if par.Tails != nil {
		for _, v := range *par.Tails {
			name := (*v)[len(*v)-1]
			curFunc := tails[`if`].Tails[name].tplFunc
			pars := (*v)[:len(*v)-1]
			callFunc(&curFunc, par.Node, par.Vars, &pars, nil)
		}
	}
	return ``
}

func elseifTag(par parFunc) string {
	cond := ifValue((*par.Pars)[`Condition`])
	if cond {
		for _, item := range par.Node.Children {
			par.Owner.Children = append(par.Owner.Children, item)
		}
		(*par.Vars)[`_cond`] = `1`
	}
	return ``
}

func elseifFull(par parFunc) string {
	setAttr(par, `Condition`)
	par.Owner.Tail = append(par.Owner.Tail, par.Node)
	return ``
}

func elseTag(par parFunc) string {
	for _, item := range par.Node.Children {
		par.Owner.Children = append(par.Owner.Children, item)
	}
	return ``
}

func elseFull(par parFunc) string {
	par.Owner.Tail = append(par.Owner.Tail, par.Node)
	return ``
}

func appendText(owner *node, text string) {
	if len(strings.TrimSpace(text)) == 0 {
		return
	}
	if len(text) > 0 {
		owner.Children = append(owner.Children, &node{Tag: tagText, Text: html.EscapeString(text)})
	}
}

func callFunc(curFunc *tplFunc, owner *node, vars *map[string]string, params *[]string, tailpars *[]*[]string) {
	var (
		out     string
		curNode node
	)
	pars := make(map[string]string)
	parFunc := parFunc{
		Vars: vars,
	}
	if curFunc.Params == `*` {
		for _, v := range *params {
			val := strings.TrimSpace(v)
			off := strings.IndexByte(val, ':')
			if off != -1 {
				pars[val[:off]] = strings.TrimSpace(val[off+1:])
			}
		}
	} else {
		for i, v := range strings.Split(curFunc.Params, `,`) {
			if i < len(*params) {
				val := strings.TrimSpace((*params)[i])
				off := strings.IndexByte(val, ':')
				if off != -1 && strings.Contains(curFunc.Params, val[:off]) {
					pars[val[:off]] = strings.TrimSpace(val[off+1:])
				} else {
					pars[v] = val
				}
			} else if _, ok := pars[v]; !ok {
				pars[v] = ``
			}
		}
	}
	if len(curFunc.Tag) > 0 {
		curNode.Tag = curFunc.Tag
		curNode.Attr = make(map[string]interface{})
		if len(pars[`Body`]) > 0 {
			process(pars[`Body`], &curNode, vars)
		}
		parFunc.Owner = owner
		//		owner.Children = append(owner.Children, &curNode)
		parFunc.Node = &curNode
		parFunc.Tails = tailpars
	}
	parFunc.Pars = &pars
	if (*vars)[`_full`] == `1` {
		out = curFunc.Full(parFunc)
	} else {
		out = curFunc.Func(parFunc)
	}
	if len(out) > 0 {
		if len(owner.Children) > 0 && owner.Children[len(owner.Children)-1].Tag == tagText {
			owner.Children[len(owner.Children)-1].Text += out
		} else {
			appendText(owner, out)
		}
	}
}

func getFunc(input string, curFunc tplFunc) (*[]string, int, *[]*[]string) {
	var (
		curp, off, mode, lenParams int
		skip                       bool
		pair, ch                   rune
		tailpar                    *[]*[]string
	)
	params := make([]string, 1)
	if curFunc.Params == `*` {
		lenParams = 0xff
	} else {
		lenParams = len(strings.Split(curFunc.Params, `,`))
	}
	level := 1
	if input[0] == '{' {
		mode = 1
	}
	skip = true
main:
	for off, ch = range input {
		if skip {
			skip = false
			continue
		}
		if pair > 0 {
			if ch != pair {
				params[curp] += string(ch)
			} else {
				if off+1 == len(input) || rune(input[off+1]) != pair {
					pair = 0
				} else {
					params[curp] += string(ch)
					skip = true
				}
			}
			continue
		}
		if len(params[curp]) == 0 && ch != modes[mode][1] && ch != ',' {
			if ch >= '!' {
				if ch == '"' || ch == '`' {
					pair = ch
				} else {
					params[curp] += string(ch)
				}
			}
			continue
		}
		switch ch {
		case ',':
			if mode == 0 && level == 1 && len(params) < lenParams {
				params = append(params, ``)
				curp++
				continue
			}
		case modes[mode][0]:
			level++
		case modes[mode][1]:
			if level > 0 {
				level--
			}
			if level == 0 {
				if mode == 0 && off+1 < len(input) && rune(input[off+1]) == modes[1][0] &&
					strings.Contains(curFunc.Params, `Body`) {
					mode = 1
					params = append(params, `Body:`)
					curp++
					skip = true
					level = 1
					continue
				}
				for tail, ok := tails[curFunc.Tag]; ok && off+2 < len(input) && input[off+1] == '.'; {
					for key, tailFunc := range tail.Tails {
						if strings.HasPrefix(input[off+2:], key+`(`) || strings.HasPrefix(input[off+2:], key+`{`) {
							parTail, shift, _ := getFunc(input[off+len(key)+2:], tailFunc.tplFunc)
							off += shift + len(key) + 2
							if tailpar == nil {
								fortail := make([]*[]string, 0)
								tailpar = &fortail
							}
							*parTail = append(*parTail, key)
							*tailpar = append(*tailpar, parTail)
							if tailFunc.Last {
								break main
							}
						}
					}
				}
				break main
			}
		}
		params[curp] += string(ch)
		continue
	}
	return &params, off, tailpar
}

func process(input string, owner *node, vars *map[string]string) {
	var (
		nameOff, shift int
		curFunc        tplFunc
		isFunc         bool
		params         *[]string
		tailpars       *[]*[]string
	)
	//	fmt.Println(`Input`, input)
	name := make([]rune, 0, 128)
	//main:
	for off, ch := range input {
		if shift > 0 {
			shift--
			continue
		}
		if ch == '(' {
			if curFunc, isFunc = funcs[string(name[nameOff:])]; isFunc {
				appendText(owner, string(name[:nameOff]))
				name = name[:0]
				nameOff = 0
				params, shift, tailpars = getFunc(input[off:], curFunc)
				callFunc(&curFunc, owner, vars, params, tailpars)
				continue
			}
		}
		if (ch < 'A' || ch > 'Z') && (ch < 'a' || ch > 'z') {
			nameOff = len(name) + 1
		}
		name = append(name, ch)
	}
	appendText(owner, string(name))
}

// Template2JSON converts templates to JSON data
func Template2JSON(input string, full bool) []byte {
	vars := make(map[string]string)
	if full {
		vars[`_full`] = `1`
	} else {
		vars[`_full`] = `0`
	}
	root := node{}
	process(input, &root, &vars)
	if root.Children == nil {
		return []byte(`[]`)
	}
	out, err := json.Marshal(root.Children)
	if err != nil {
		return []byte(err.Error())
	}
	return out
}
