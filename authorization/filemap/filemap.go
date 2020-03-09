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

import "errors"

var (
	// ErrNotImpl error occurs when file name extension not recognized
	ErrNotImpl = errors.New("Not Implemented")
	// ErrFileType error occurs when maper gets filename with wrong extension
	ErrFileType = errors.New("wrong extension")
	// ErrCreate error occurs when maper fails to create file storage
	ErrCreate = errors.New("fail to create file")
	// ErrRemove error occurs when maper fails to remove file storage
	ErrRemove = errors.New("fail to remove file")
)

// User contains user attributes
type User struct {
	Password string
	ID       int
}

// FileMaper wraps Load, Save methods
type FileMaper interface {
	Load() (map[string]*User, error)
	Save(map[string]*User) error
}

// Authorizator implements interfaces.Authorizator interface
type Authorizator struct {
	mapper FileMaper
	users  map[string]*User
}

func createMapper(file string) {

}

// New constructs new Authorizator
func New(file string) (*Authorizator, error) {
	return nil, ErrNotImpl
}
