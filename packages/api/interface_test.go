// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package api

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInterfaceRow(t *testing.T) {
	cases := []struct {
		url        string
		contract   string
		equalAttrs []string
	}{
		{"interface/page/", "NewPage", []string{"Name", "Value", "Menu", "Conditions"}},
		{"interface/menu/", "NewMenu", []string{"Name", "Value", "Title", "Conditions"}},
		{"interface/block/", "NewBlock", []string{"Name", "Value", "Conditions"}},
	}

	checkEqualAttrs := func(form url.Values, result map[string]interface{}, equalKeys []string) {
		for _, key := range equalKeys {
			v := result[strings.ToLower(key)]
			assert.EqualValues(t, form.Get(key), v)
		}
	}

	errUnauthorized := `401 {"error": "E_UNAUTHORIZED", "msg": "Unauthorized" }`
	for _, c := range cases {
		assert.EqualError(t, sendGet(c.url+"-", &url.Values{}, nil), errUnauthorized)
	}

	assert.NoError(t, keyLogin(1))

	for _, c := range cases {
		name := randName("component")
		form := url.Values{
			"Name": {name}, "Value": {"value"}, "Menu": {"default_menu"}, "Title": {"title"},
			"Conditions": {"true"},
		}
		assert.NoError(t, postTx(c.contract, &form))
		result := map[string]interface{}{}
		assert.NoError(t, sendGet(c.url+name, &url.Values{}, &result))
		checkEqualAttrs(form, result, c.equalAttrs)
	}
}

func TestNewMenuNoError(t *testing.T) {
	require.NoError(t, keyLogin(1))
	menuname := "myTestMenu"
	form := url.Values{"Name": {menuname}, "Value": {`first
		second
		third`}, "Title": {`My Test Menu`},
		"Conditions": {`true`}}
	assert.NoError(t, postTx(`NewMenu`, &form))

	err := postTx(`NewMenu`, &form)
	assert.Equal(t, fmt.Sprintf(`{"type":"warning","error":"Menu %s already exists"}`, menuname), cutErr(err))
}

func TestEditMenuNoError(t *testing.T) {
	require.NoError(t, keyLogin(1))
	form := url.Values{
		"Id": {"1"},
		"Value": {`first
		second
		third
		andmore`},
		"Title": {`My edited Test Menu`},
	}
	assert.NoError(t, postTx(`EditMenu`, &form))
}

func TestAppendMenuNoError(t *testing.T) {
	require.NoError(t, keyLogin(1))
	form := url.Values{
		"Id":    {"3"},
		"Value": {"appended item"},
	}

	assert.NoError(t, postTx("AppendMenu", &form))
}
