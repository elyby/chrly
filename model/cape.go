package model

import "os"

type Cape struct {
	File *os.File // TODO: нужно абстрагироваться в отдельный файл с инфой о скине
}
