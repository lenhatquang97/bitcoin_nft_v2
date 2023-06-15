package db

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
)

func applyMigrations(fs fs.FS, driver database.Driver, path,
	dbName string) error {

	migrateFileServer, err := httpfs.New(http.FS(fs), path)
	if err != nil {
		return err
	}

	sqlMigrate, err := migrate.NewWithInstance(
		"migrations", migrateFileServer, dbName, driver,
	)
	if err != nil {
		return err
	}
	err = sqlMigrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

type replacerFS struct {
	parentFS fs.FS
	replaces map[string]string
}

var _ fs.FS = (*replacerFS)(nil)

func newReplacerFS(parent fs.FS, replaces map[string]string) *replacerFS {
	return &replacerFS{
		parentFS: parent,
		replaces: replaces,
	}
}

func (t *replacerFS) Open(name string) (fs.File, error) {
	f, err := t.parentFS.Open(name)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return f, err
	}

	return newReplacerFile(f, t.replaces)
}

type replacerFile struct {
	parentFile fs.File
	buf        bytes.Buffer
}

var _ fs.File = (*replacerFile)(nil)

func newReplacerFile(parent fs.File, replaces map[string]string) (*replacerFile,
	error) {

	content, err := io.ReadAll(parent)
	if err != nil {
		return nil, err
	}

	contentStr := string(content)
	for from, to := range replaces {
		contentStr = strings.Replace(contentStr, from, to, -1)
	}

	var buf bytes.Buffer
	_, err = buf.WriteString(contentStr)
	if err != nil {
		return nil, err
	}

	return &replacerFile{
		parentFile: parent,
		buf:        buf,
	}, nil
}

func (t *replacerFile) Stat() (fs.FileInfo, error) {
	return t.parentFile.Stat()
}

func (t *replacerFile) Read(bytes []byte) (int, error) {
	return t.buf.Read(bytes)
}

func (t *replacerFile) Close() error {
	return nil
}
