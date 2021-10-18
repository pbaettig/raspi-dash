package photos

import (
	"fmt"
	"image"
	"image/jpeg"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/disintegration/imageorient"
	"github.com/nfnt/resize"
)

type Filetype int

type ThumbnailDestFunc func(string, uint, uint, int) (string, error)

type ImageWidth uint

type Photo struct {
	RelativePath   string
	LocalPath      string
	Filename       string
	ThumbnailPaths map[ImageWidth]string
	Info           fs.FileInfo
	Type           Filetype
}

type Photos struct {
	fs.FS
	root string
}

type ThumbnailConfig struct {
	Width, Height uint
	Quality       int
}

const (
	FiletypeUnknown = Filetype(0)
	FiletypeJPEG    = Filetype(1)
)

var (
	DefaultThumbnailDestFunc = LevelDestinationFunc(7, "thumbnails")
	PhotosBasePath           = "/Users/pascal/git/raspi-dash/photos/sample/photos"
	FS                       = NewPhotos(PhotosBasePath)

	ThumbnailBasePath = "/Users/pascal/git/raspi-dash/photos/sample/thumbnails"
	ThumbnailConfigs  = []ThumbnailConfig{
		{Width: 400, Height: 0, Quality: 25},
		{Width: 1920, Height: 0, Quality: 85},
	}
)

func (tc ThumbnailConfig) Path(p Photo) string {
	const photosBaseLevel = 7

	dirs := strings.Split(p.LocalPath, string(os.PathSeparator))
	dirs[photosBaseLevel] = "thumbnails"
	dirs = append(dirs[:photosBaseLevel+2], dirs[photosBaseLevel+1:]...)
	dirs[photosBaseLevel+1] = fmt.Sprintf("%dx%d-%d", tc.Width, tc.Height, tc.Quality)

	return strings.Join(dirs, string(os.PathSeparator))
}

// func (p Photo) ThumbnailPaths() []string {
// 	ps := make([]string, 0)
// 	for _, c := range ThumbnailConfigs {
// 		_, fname := path.Split(p.LocalPath)
// 		ps = append(ps, path.Join(ThumbnailBasePath, strconv.Itoa(int(c.Width)), fname))
// 	}

// 	return ps
// }

func (p Photo) GenerateThumbnails() (numOk int, numAlreadyExist int, numFailed int, err error) {
	var (
		dest  *os.File
		f     fs.File
		image image.Image
	)

	for _, tc := range ThumbnailConfigs {
		destPath := tc.Path(p)
		dest, err = os.OpenFile(destPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
		if err != nil {
			numAlreadyExist += 1
			continue
		}
		defer dest.Close()

		f, err = FS.Open(p.RelativePath)
		if err != nil {
			numFailed += 1
			continue
		}
		image, _, err = imageorient.Decode(f)
		if err != nil {
			numFailed += 1
			continue
		}
		// fmt.Println(image.Bounds().Size())

		newImage := resize.Resize(tc.Width, tc.Height, image, resize.Lanczos3)
		// fmt.Println(destPath)
		// fmt.Println(newImage.Bounds().Size())

		// Encode uses a Writer, use a Buffer if you need the raw []byte
		err = jpeg.Encode(dest, newImage, &jpeg.Options{Quality: tc.Quality})
		if err != nil {
			numFailed += 1
			continue
		}
		numOk += 1
	}
	return
}

func LevelDestinationFunc(level int, name string) ThumbnailDestFunc {
	fn := func(p string, w, h uint, q int) (string, error) {
		//    /Users/pascal/git/raspi-dash/photos/sample/photos/Nordkorea - Dezember 2020/20140315_092940925_iOS.jpg/20140315_092940925_iOS.jpg
		// -> /Users/pascal/git/raspi-dash/photos/sample/thumbnails/Nordkorea - Dezember 2020/20140315_092940925_iOS.jpg/20140315_092940925_iOS.jpg

		dirs := strings.Split(p, string(os.PathSeparator))
		dirs[level] = name
		dirs = append(dirs[:level+2], dirs[level+1:]...)
		dirs[level+1] = strconv.Itoa(int(w))
		fmt.Println(strings.Join(dirs, string(os.PathSeparator)))
		return strings.Join(dirs, string(os.PathSeparator)), nil
	}

	return fn
}

func NewPhotos(root string) Photos {
	return Photos{
		FS:   os.DirFS(root),
		root: root,
	}
}

func GetFiletype(p string) Filetype {
	suf := strings.TrimPrefix(path.Ext(p), ".")
	switch suf {
	case "jpg":
		fallthrough
	case "jpeg":
		return FiletypeJPEG
	default:
		return FiletypeUnknown
	}
}

func (ph Photos) ListAll() {
	fn := func(p string, e fs.DirEntry, err error) error {
		if e != nil && !e.IsDir() {
			if GetFiletype(p) == FiletypeJPEG {

				fmt.Println(path.Join(ph.root, p))
			}

		}
		return nil
	}

	fs.WalkDir(ph.FS, ".", fn)
}

func albumName(p string) (string, error) {
	ps := strings.Split(p, string(os.PathSeparator))
	if len(ps) == 0 {
		return "", fmt.Errorf("cannot split path")
	}
	return ps[0], nil
}

func (ph Photos) GetPhoto(relPath string, fi fs.FileInfo) Photo {
	fp := path.Join(ph.root, relPath)
	t := GetFiletype(relPath)
	_, fname := path.Split(relPath)
	p := Photo{
		RelativePath: relPath,
		Filename:     fname,
		LocalPath:    fp,
		Info:         fi,
		Type:         t,
	}

	tps := make(map[ImageWidth]string)
	for _, tc := range ThumbnailConfigs {
		tps[ImageWidth(tc.Width)] = tc.Path(p)
	}
	p.ThumbnailPaths = tps

	return p
}

func (ph Photos) GetAlbums() (map[string][]Photo, error) {
	albums := make(map[string][]Photo)
	fn := func(p string, e fs.DirEntry, err error) error {
		a, _ := albumName(p)
		t := GetFiletype(e.Name())
		if !e.IsDir() && t == FiletypeJPEG {
			// fmt.Println(p)
			fi, err := e.Info()
			if err != nil {
				return err
			}

			if _, ok := albums[a]; !ok {
				albums[a] = make([]Photo, 0)
			}
			albums[a] = append(albums[a], ph.GetPhoto(p, fi))
		}

		return nil
	}
	err := fs.WalkDir(ph.FS, ".", fn)
	if err != nil {
		return map[string][]Photo{}, err
	}

	return albums, nil
}

func (ph Photos) GetByAlbum(name string) ([]Photo, error) {
	found := make([]Photo, 0)

	fn := func(p string, e fs.DirEntry, err error) error {
		if a, _ := albumName(p); !e.IsDir() && a == name {
			fi, err := e.Info()
			if err != nil {
				return err
			}

			found = append(found, Photo{
				RelativePath: p,
				LocalPath:    path.Join(ph.root, p, e.Name()),
				Info:         fi,
			})
		}

		return nil
	}
	err := fs.WalkDir(ph.FS, ".", fn)
	if err != nil {
		return []Photo{}, err
	}

	if len(found) == 0 {
		return found, fmt.Errorf("not found")
	}

	return found, nil
}

func (ph Photos) ReadPhoto(p Photo) ([]byte, error) {
	f, err := ph.Open(p.RelativePath)
	if err != nil {
		return []byte{}, err
	}

	buf := make([]byte, p.Info.Size())
	_, err = f.Read(buf)
	if err != nil {
		return buf, err
	}

	return buf, nil
}

func (ph Photos) GenerateThumbnail(src Photo, width, height uint, quality int) error {

	if quality == 0 {
		quality = 85
	}
	destPath, err := DefaultThumbnailDestFunc(src.LocalPath, width, height, quality)
	if err != nil {
		return err
	}

	// return nil

	p, err := ph.Open(src.RelativePath)
	if err != nil {
		return err
	}
	image, _, err := imageorient.Decode(p)
	if err != nil {
		return err
	}
	// fmt.Println(image.Bounds().Size())

	newImage := resize.Resize(width, height, image, resize.Lanczos3)
	// fmt.Println(destPath)
	// fmt.Println(newImage.Bounds().Size())

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	// Encode uses a Writer, use a Buffer if you need the raw []byte
	err = jpeg.Encode(dest, newImage, &jpeg.Options{Quality: quality})
	if err != nil {
		return err
	}

	return nil
}

func (ph Photos) GenerateAllThumbnails() error {
	albs, err := ph.GetAlbums()
	if err != nil {
		return err
	}

	totalOK := 0
	totalExists := 0
	totalFailed := 0

	for album, photos := range albs {
		fmt.Printf("generating Thumbnails for \"%s\": ", album)
		for _, p := range photos {
			ok, exist, failed, _ := p.GenerateThumbnails()
			totalOK += ok
			totalExists += exist
			totalFailed += failed
			for i := 0; i < ok; i++ {
				fmt.Print(".")
			}
			for i := 0; i < exist; i++ {
				fmt.Print("A")
			}
			for i := 0; i < failed; i++ {
				fmt.Print("X")
			}
		}

		fmt.Println()
	}
	fmt.Printf("newly generated: %d\n", totalOK)
	fmt.Printf("already existing: %d\n", totalExists)
	fmt.Printf("failures: %d\n", totalFailed)

	return nil
}
