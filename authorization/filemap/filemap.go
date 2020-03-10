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
	"errors"
	"io"
	"path"

	"github.com/yagoggame/grpc_server/authorization"
	"github.com/yagoggame/grpc_server/authorization/filemap/json"
)

var (
	// ErrNotImpl error occurs when file name extension not recognized
	ErrNotImpl = errors.New("Not Implemented")
)

// FileMaper wraps Load, Save methods
type FileMaper interface {
	Load(io.Reader) (map[string]*authorization.User, error)
	Save(map[string]*authorization.User, io.Writer) error
}

func choseMaper(fileName string) (FileMaper, error) {
	ext := path.Ext(fileName)
	switch ext {
	case ".json":
		return json.New(), nil
	}
	return nil, ErrNotImpl
}

// Authorizator implements interfaces.Authorizator interface
type Authorizator struct {
	fileName string
	maper    FileMaper
	users    map[string]*authorization.User
}

// New constructs new Authorizator
func New(fileName string) (*Authorizator, error) {
	maper, err := choseMaper(fileName)
	if err != nil {
		return nil, err
	}

	authorizator := &Authorizator{
		fileName: fileName,
		maper:    maper,
		users:    nil,
	}
	return authorizator, nil
}
