package docs

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
)

type File struct {
	RelativePath string
	LocalPath    string
	Info         fs.FileInfo
}

func hidden(d fs.DirEntry) bool {
	return strings.HasPrefix(d.Name(), ".")
}

type Documents struct {
	fs.FS
	root string
}

func NewDocuments(root string) Documents {
	return Documents{
		FS:   os.DirFS(root),
		root: root,
	}
}

func (d Documents) fullPath(p string) string {
	return path.Join(d.root, p)
}

func (d Documents) FindById(id string) ([]File, error) {

	found := make([]File, 0)

	fn := func(path string, e fs.DirEntry, err error) error {
		if !e.IsDir() && strings.HasPrefix(e.Name(), fmt.Sprintf("%s_", id)) {
			fi, err := e.Info()
			if err != nil {
				return err
			}

			found = append(found, File{
				RelativePath: path,
				LocalPath:    d.fullPath(path),
				Info:         fi,
			})
		}
		return nil
	}
	err := fs.WalkDir(d.FS, ".", fn)
	if err != nil {
		return []File{}, err
	}
	if len(found) == 0 {
		return found, fmt.Errorf("not found")
	}

	return found, nil

}

func (d Documents) ReadFile(file File) ([]byte, error) {
	f, err := d.Open(file.RelativePath)
	if err != nil {
		return []byte{}, err
	}

	buf := make([]byte, file.Info.Size())
	_, err = f.Read(buf)
	if err != nil {
		return buf, err
	}

	return buf, nil
}

func (d Documents) FindByKeyword(kw string) ([]File, error) {
	found := make([]File, 0)

	fn := func(path string, e fs.DirEntry, err error) error {
		n := strings.ToLower(e.Name())
		if !hidden(e) && !e.IsDir() && strings.Contains(n, kw) {
			fi, err := e.Info()
			if err != nil {
				return err
			}
			found = append(found, File{
				RelativePath: path,
				LocalPath:    d.fullPath(path),
				Info:         fi,
			})
		}

		return nil
	}
	err := fs.WalkDir(d.FS, ".", fn)
	if err != nil {
		return found, err
	}

	if len(found) == 0 {
		return found, fmt.Errorf("not found")
	}

	return found, nil
}

func (d Documents) ZipFiles(files []File) ([]byte, error) {
	zbuf := new(bytes.Buffer)
	zw := zip.NewWriter(zbuf)

	for _, f := range files {
		zf, err := zw.Create(f.Info.Name())
		if err != nil {
			return []byte{}, err
		}

		fc, err := d.ReadFile(f)
		if err != nil {
			return []byte{}, err
		}
		_, err = zf.Write(fc)
		if err != nil {
			return []byte{}, err
		}
	}

	err := zw.Close()
	if err != nil {
		return []byte{}, err
	}

	return zbuf.Bytes(), nil
}
