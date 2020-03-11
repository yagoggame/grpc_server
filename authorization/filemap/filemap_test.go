// Copyright ©2020 BlinnikovAA. All rights reserved.
// This file is part of yagogame.
//
// yagogame is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// yagogame is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with yagogame.  If not, see <https://www.gnu.org/licenses/>.

package filemap

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/yagoggame/grpc_server/interfaces"
)

type iderr struct {
	id  int
	err error
}

var (
	commonFileName = "tmp.json"
	wrongFileName  = "tmp.json.tar"
)

var jsonPrestoredContent = `[
	{
		"login": "Joe",
		"password": "aaa",
		"id": 2
	}
]
`

var newTests = []struct {
	name        string
	fileName    string
	fileContent string
	wantErr     error
	wantCount   int
}{
	{
		name:      "correct extension no file",
		fileName:  commonFileName,
		wantErr:   nil,
		wantCount: 2,
	},
	{
		name:        "correct extension prestored file content",
		fileName:    commonFileName,
		fileContent: jsonPrestoredContent,
		wantErr:     nil,
		wantCount:   1,
	},
	{
		name:      "wrong extension",
		fileName:  wrongFileName,
		wantErr:   ErrNotImpl,
		wantCount: 0,
	},
	{
		name:      "empty filename",
		fileName:  "",
		wantErr:   ErrNotImpl,
		wantCount: 0,
	},
}

var testsAuthorize = []struct {
	caseName   string
	requisites interfaces.Requisites
	want       iderr
}{
	{
		caseName: "unregistred login",
		requisites: interfaces.Requisites{
			Login:    "Piter",
			Password: "aaa",
		},
		want: iderr{id: 0, err: interfaces.ErrLogin},
	},
	{
		caseName: "registred login wrong password",
		requisites: interfaces.Requisites{
			Login:    "Joe",
			Password: "ababab",
		},
		want: iderr{id: 0, err: interfaces.ErrPassword},
	},
	{
		caseName: "registred login",
		requisites: interfaces.Requisites{
			Login:    "Joe",
			Password: "aaa",
		},
		want: iderr{id: 2, err: nil},
	},
	{
		caseName: "registred login",
		requisites: interfaces.Requisites{
			Login:    "Nick",
			Password: "bbb",
		},
		want: iderr{id: 3, err: nil},
	},
}

var testsRegister = []struct {
	caseName   string
	requisites interfaces.Requisites
	want       error
}{
	{
		caseName: "registred login",
		requisites: interfaces.Requisites{
			Login:    "Joe",
			Password: "aaa",
		},
		want: interfaces.ErrLoginOccupied},
	{
		caseName: "unregistred login first",
		requisites: interfaces.Requisites{
			Login:    "Piter",
			Password: "aaa",
		},
		want: nil},
	{
		caseName: "unregistred login second",
		requisites: interfaces.Requisites{
			Login:    "Mike",
			Password: "mmm",
		},
		want: nil},
}

var testsRemove = []struct {
	caseName   string
	requisites interfaces.Requisites
	want       error
}{
	{
		caseName: "unregistred login",
		requisites: interfaces.Requisites{
			Login:    "Piter",
			Password: "aaa",
		},
		want: interfaces.ErrLogin,
	},
	{
		caseName: "registred login wrong password",
		requisites: interfaces.Requisites{
			Login:    "Joe",
			Password: "ababab",
		},
		want: interfaces.ErrPassword,
	},
	{
		caseName: "registred login",
		requisites: interfaces.Requisites{
			Login:    "Joe",
			Password: "aaa",
		},
		want: nil,
	},
	{
		caseName: "registred login",
		requisites: interfaces.Requisites{
			Login:    "Nick",
			Password: "bbb",
		},
		want: nil,
	},
}

var testsChangeRequisites = []struct {
	caseName      string
	requisitesOld interfaces.Requisites
	requisitesNew interfaces.Requisites
	want          error
}{
	{
		caseName: "from unregistred login",
		requisitesOld: interfaces.Requisites{
			Login:    "Piter",
			Password: "aaa",
		},
		requisitesNew: interfaces.Requisites{
			Login:    "Teodor",
			Password: "ttt",
		},
		want: interfaces.ErrLogin,
	},
	{
		caseName: "from registred login wrong password",
		requisitesOld: interfaces.Requisites{
			Login:    "Joe",
			Password: "ababab",
		},
		requisitesNew: interfaces.Requisites{
			Login:    "Teodor",
			Password: "ttt",
		},
		want: interfaces.ErrPassword,
	},
	{
		caseName: "to registred",
		requisitesOld: interfaces.Requisites{
			Login:    "Joe",
			Password: "aaa",
		},
		requisitesNew: interfaces.Requisites{
			Login:    "Nick",
			Password: "fff",
		},
		want: interfaces.ErrLoginOccupied,
	},
	{
		caseName: "registred to unregistred",
		requisitesOld: interfaces.Requisites{
			Login:    "Joe",
			Password: "aaa",
		},
		requisitesNew: interfaces.Requisites{
			Login:    "Teodor",
			Password: "ttt",
		},
		want: nil,
	},
	{
		caseName: "only change password",
		requisitesOld: interfaces.Requisites{
			Login:    "Nick",
			Password: "bbb",
		},
		requisitesNew: interfaces.Requisites{
			Login:    "Nick",
			Password: "ccc",
		},
		want: nil,
	},
}

func TestNew(t *testing.T) {
	for _, test := range newTests {
		t.Run(test.name, func(t *testing.T) {
			if err := mkFileWithContent(test.fileName, test.fileContent); err != nil {
				t.Fatalf("Unexpected mkFile() err: %v", err)
			}

			authorizator, err := New(test.fileName)
			if !errors.Is(err, test.wantErr) {
				t.Errorf("Unexpected err:\nwant: %v,\ngot: %v.", test.wantErr, err)
			}
			if err == nil && authorizator.Len() != test.wantCount {
				t.Errorf("Unexpected users count:\nwant: %d,\ngot: %d.", test.wantCount, authorizator.Len())
			}

			errR := os.Remove(test.fileName)
			if err == nil && errR != nil {
				t.Errorf("Unexpected file %q remove err:\nwant: %v,\ngot: %v.", test.fileName, nil, errR)
			}
		})
	}
}

func TestAuthorize(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	authorizator, contentBefore := pretestActions(t, commonFileName)

	for _, test := range testsAuthorize {
		t.Run(test.caseName, func(t *testing.T) {
			id, err := authorizator.Authorize(&test.requisites)

			testIDErr(t, test.want, iderr{id: id, err: err})
		})
	}

	contentAfter := posttestActions(t, commonFileName)
	if bytes.Compare(contentBefore, contentAfter) != 0 {
		t.Errorf("unexpected file change")
	}
}

func TestRegister(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	authorizator, contentBefore := pretestActions(t, commonFileName)

	for _, test := range testsRegister {
		t.Run(test.caseName, func(t *testing.T) {
			err := authorizator.Register(&test.requisites)

			testErr(t, test.want, err)
		})
	}

	contentAfter := posttestActions(t, commonFileName)
	if bytes.Compare(contentBefore, contentAfter) == 0 {
		t.Errorf("unexpected file invariance")
	}
}

func TestRemove(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	authorizator, contentBefore := pretestActions(t, commonFileName)

	for _, test := range testsRemove {
		t.Run(test.caseName, func(t *testing.T) {
			usersLen := authorizator.Len()
			err := authorizator.Remove(&test.requisites)

			testErr(t, test.want, err)

			if err != nil && usersLen != authorizator.Len() {
				t.Errorf("Unexpected count of user delta:\nwant: 0\ngot: %d.", authorizator.Len()-usersLen)
			} else if err == nil && usersLen == authorizator.Len() {
				t.Errorf("Unexpected count of user delta:\nwant: -1\ngot: %d.", authorizator.Len()-usersLen)
			}

			_, authErr := authorizator.Authorize(&test.requisites)
			if err == nil && authErr != interfaces.ErrLogin {
				t.Errorf("Unexpected Authorize err:\nwant: %v\ngot: %v.", interfaces.ErrLogin, authErr)
			}

		})
	}

	contentAfter := posttestActions(t, commonFileName)
	if bytes.Compare(contentBefore, contentAfter) == 0 {
		t.Errorf("unexpected file invariance")
	}
}

func TestChangeRequisites(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	authorizator, contentBefore := pretestActions(t, commonFileName)

	for _, test := range testsChangeRequisites {
		t.Run(test.caseName, func(t *testing.T) {
			err := authorizator.ChangeRequisites(&test.requisitesOld, &test.requisitesNew)

			testErr(t, test.want, err)

			_, authErr := authorizator.Authorize(&test.requisitesNew)
			if err != nil && authErr == nil {
				t.Errorf("Unexpected Authorize err:\nwant: err!=nil\ngot: %v.", authErr)
			}

			if err == nil && authErr != nil {
				t.Errorf("Unexpected Authorize err:\nwant: %v\ngot: %v.", nil, authErr)
			}
		})
	}

	contentAfter := posttestActions(t, commonFileName)
	if bytes.Compare(contentBefore, contentAfter) == 0 {
		t.Errorf("unexpected file invariance")
	}
}

func mkFileWithContent(fileName, fileContent string) error {
	if len(fileContent) < 1 {
		return nil
	}

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(fileContent); err != nil {
		return err
	}

	return nil
}

func testIDErr(t *testing.T, want, got iderr) {
	if got.id != want.id {
		t.Errorf("Unexpected id:\nwant: %d,\ngot: %d.", want.id, got.id)
	}
	if got.err != want.err {
		t.Errorf("Unexpected err:\nwant: %v,\ngot: %v.", want.err, got.err)
	}
}

func testErr(t *testing.T, want, got error) {
	if got != want {
		t.Errorf("Unexpected err:\nwant: %v,\ngot: %v.", want, got)
	}
}

func pretestActions(t *testing.T, filename string) (*Authorizator, []byte) {
	authorizator, err := New(commonFileName)
	if err != nil {
		t.Fatalf("Unexpected New() err: %v", err)
	}
	contentBefore, err := ioutil.ReadFile(commonFileName)
	if err != nil {
		t.Fatalf("Unexpected ReadFile err: %v", err)
	}
	return authorizator, contentBefore
}

func posttestActions(t *testing.T, filename string) []byte {
	contentAfter, err := ioutil.ReadFile(commonFileName)
	if err != nil {
		t.Fatalf("Unexpected ReadFile err: %v", err)
	}
	if err := os.Remove(commonFileName); err != nil {
		t.Errorf("Unexpected os.Remove err: %v.", err)
	}
	return contentAfter
}
