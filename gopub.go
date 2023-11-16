package gopub

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"path"
	"strings"
)

type epubReader struct {
	zipReader *zip.ReadCloser
	options   *epubReaderOptions
}

type epubReaderOptions struct {
	rootFilePath string
}

type Book struct {
	FilePath     string
	Title        string
	Author       string
	Authors      []string
	Description  string
	CoverImage   []byte
	ReadingOrder []*LocalTextContentFile
	Navigation   []*Navigation
	Content      *Content
}

func ReadBook(filePath string) (*Book, error) {
	zf, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer zf.Close()

	container, err := findRootFile(zf)
	if err != nil {
		return nil, err
	}

	reader := &epubReader{
		zipReader: zf,
		options: &epubReaderOptions{
			rootFilePath: container.FullPath,
		},
	}

	schema, err := readSchema(reader)
	if err != nil {
		return nil, err
	}

	coverImage, err := schema.pkg.Manifest.getCoverImage(zf)
	if err != nil {
		return nil, err
	}

	readingOrder, err := schema.getReadingOrder(reader)
	if err != nil {
		return nil, err
	}

	navigationPath, err := schema.pkg.Manifest.getNavigationFilePath()
	if err != nil {
		return nil, err
	}

	navDoc, err := readNavigation(zf, path.Join(schema.contentDirectoryPath, navigationPath))
	if err != nil {
		return nil, err
	}

	schema.navigationDocument = navDoc

	content, err := readContent(schema, reader)
	if err != nil {
		return nil, err
	}

	book := &Book{
		FilePath:     filePath,
		Title:        schema.pkg.Metadata.Titles[0].Value,
		Author:       schema.pkg.Metadata.Creators[0].Value,
		Authors:      schema.pkg.Metadata.getCreators(),
		Description:  schema.pkg.Metadata.getFirstOrDefaultDescription(),
		CoverImage:   coverImage,
		ReadingOrder: readingOrder,
		Navigation:   navDoc.Navigations,
		Content:      content,
	}

	return book, nil
}

func findRootFile(r *zip.ReadCloser) (*RootFile, error) {
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "META-INF/container.xml") {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			var container Container
			if err := xml.NewDecoder(rc).Decode(&container); err != nil {
				return nil, err
			}

			return &container.RootFile, nil
		}
	}
	return nil, fmt.Errorf("root file not found in META-INF/container.xml")
}
