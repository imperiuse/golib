package dirhelper

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
)

// CopyFile -  копирует файл из src в dst
func CopyFile(src, dst string) (err error) {
	srcfd, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = srcfd.Close()
		}
		err = srcfd.Close()
	}()

	dstfd, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = dstfd.Close()
		}
		err = dstfd.Close()
	}()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}

	srcinfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcinfo.Mode())
}

// CopyDir - рекурсивно копирует директорию (рекурсивно - т.е. со всеми поддиректориями и файлами)
func CopyDir(src string, dst string) error {

	srcinfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	fds, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				return err
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				return err
			}
		}
	}

	return nil
}

// MakeDir - создает новую директорию (есть проверка на существование директории, если нет - создает)
func MakeDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

// CleanDir - удаляет содержимое директории (все папки и файлы)
func CleanDir(dir string) (err error) {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = d.Close()
			return
		}
		err = d.Close()
	}()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		if err = os.RemoveAll(filepath.Join(dir, name)); err != nil {
			return
		}
	}

	return
}

// ListSubeditors - получает список поддиректорий в текущей директории
func ListSubeditors(src string) ([]string, error) {
	fds, err := ioutil.ReadDir(src)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("can't ioutil.ReadDir for dir:%s", src))
	}

	var subDirs = make([]string, 0)
	for _, fd := range fds {
		if fd.IsDir() {
			subDirs = append(subDirs, fd.Name())
		}
	}

	return subDirs, nil
}
