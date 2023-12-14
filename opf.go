package gopub

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"strings"
)

type Identifier struct {
	ID    string `xml:"id,attr"`
	Value string `xml:",chardata"`
}

type Title struct {
	ID    string `xml:"id,attr"`
	Value string `xml:",chardata"`
}

type Creator struct {
	ID    string `xml:"id,attr"`
	Value string `xml:",chardata"`
}

type Metadata struct {
	Identifiers  []Identifier `xml:"identifier"`
	Titles       []Title      `xml:"title"`
	Languages    []string     `xml:"language"`
	Contributers []string     `xml:contributor`
	Coverages    []string     `xml:coverage`
	Creators     []Creator    `xml:"creator"`
	Dates        []string     `xml:"date"`
	Descriptions []string     `xml:"description"`
	Formats      []string     `xml:"format"`
	Publishers   []string     `xml:"publisher"`
	Relations    []string     `xml:"relation"`
	Rights       []string     `xml:"right"`
	Sources      []string     `xml:"source"`
	Subjects     []string     `xml:"subject"`
	Types        []string     `xml:"types"`
}

func (m Metadata) getCreators() []string {
	var creators []string
	for _, creator := range m.Creators {
		creators = append(creators, creator.Value)
	}

	return creators
}

func (m Metadata) getFirstOrDefaultDescription() string {
	if len(m.Descriptions) > 0 {
		return m.Descriptions[0]
	} else {
		return ""
	}
}

type ManifestItem struct {
	ID                string `xml:"id,attr"`
	Href              string `xml:"href,attr"`
	MediaType         string `xml:"media-type,attr"`
	MediaOverlay      string `xml:"media-overlay,attr"`
	RequiredNamespace string `xml:"required-namespace,attr"`
	RequiredModules   string `xml:"required-modules,attr"`
	Fallback          string `xml:"fallback,attr"`
	FallbackStyle     string `xml:"fallback-style,attr"`
	Properties        string `xml:"properties,attr"`
}

type Manifest struct {
	Items []ManifestItem `xml:"item"`
}

func (m Manifest) getCoverImage(r *zip.ReadCloser) ([]byte, error) {
	for _, item := range m.Items {
		if strings.Contains(item.Properties, "cover-image") {
			rc, size, err := findFileInZip(r, path.Join("EPUB", item.Href))
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			image := make([]byte, size)
			_, err = io.ReadFull(rc, image)
			if err != nil {
				if err != io.EOF {
					return nil, err
				}
			}

			return image, nil
		}
	}

	return nil, fmt.Errorf("cover image not found")
}

func (m Manifest) getNavigationFilePath() (string, error) {
	for _, item := range m.Items {
		if strings.Contains(item.Properties, "nav") {
			return item.Href, nil
		}
	}

	return "", fmt.Errorf("navigation file not found")
}

type ItemRef struct {
	ID         string `xml:"id,attr"`
	IdRef      string `xml:"idref,attr"`
	Linear     string `xml:"linear,attr"`
	Properties string `xml:"properties"`
}

type Spine struct {
	ItemRefs []ItemRef `xml:"itemref"`
}

type Reference struct {
	Type  string `xml:"type,attr"`
	Title string `xml:"title,attr"`
	Href  string `xml:"href,attr"`
}

type Guide struct {
	References []Reference `xml:"reference"`
}

type Link struct {
	ID        string `xml:"id"`
	Href      string `xml:"href"`
	MediaType string `xml:"media-type"`
}

type Collection struct {
	ID          string       `xml:"id,attr"`
	Role        string       `xml:"role,attr"`
	Language    string       `xml:"language,attr"`
	Metadata    Metadata     `xml:"metadata"`
	Collections []Collection `xml:"collections"`
	Links       []Link       `xml:"link"`
}

type Package struct {
	XMLName          xml.Name     `xml:"package"`
	UniqueIdentifier string       `xml:"unique-identifier,attr"`
	Version          string       `xml:"version,attr"`
	Metadata         Metadata     `xml:"metadata"`
	Manifest         Manifest     `xml:"manifest"`
	Spine            Spine        `xml:"spine"`
	Guide            Guide        `xml:"guide"`
	Collections      []Collection `xml:"collection"`
}

func ReadPackage(r *zip.ReadCloser, rootfilePath string) (Package, error) {
	var pkg Package

	rc, _, err := findFileInZip(r, rootfilePath)
	if err != nil {
		return pkg, err
	}
	defer rc.Close()

	if err := xml.NewDecoder(rc).Decode(&pkg); err != nil {
		return pkg, err
	}

	return pkg, nil
}

type schema struct {
	pkg                  Package
	navigationDocument   NavigationDocument
	contentDirectoryPath string
}

func readSchema(er epubReader) (schema, error) {
	var schema schema

	pkg, err := ReadPackage(er.zipReader, er.options.rootFilePath)
	if err != nil {
		return schema, err
	}

	schema.pkg = pkg
	schema.contentDirectoryPath = path.Dir(er.options.rootFilePath)

	return schema, nil
}

func (s schema) getReadingOrder(er epubReader) ([]LocalTextContentFile, error) {
	var readingOrder []LocalTextContentFile

	for _, itemRef := range s.pkg.Spine.ItemRefs {
		for _, manifestItem := range s.pkg.Manifest.Items {
			if manifestItem.ID == itemRef.IdRef {
				contentFile := newContentFile(manifestItem, s.contentDirectoryPath, ContentLocationLocal, ContentFileTypeText)
				contentFilePath := path.Join(s.contentDirectoryPath, manifestItem.Href)

				localTextContentFile, err := newLocalTextContentFile(er, contentFile, contentFilePath)
				if err != nil {
					return nil, err
				}
				readingOrder = append(readingOrder, localTextContentFile)
			}
		}

	}

	return readingOrder, nil
}

func findFileInZip(r *zip.ReadCloser, filename string) (io.ReadCloser, uint64, error) {
	for _, f := range r.File {
		if f.Name == filename {
			rc, err := f.Open()
			if err != nil {
				return nil, 0, err
			}

			return rc, f.UncompressedSize64, nil
		}
	}
	return nil, 0, fmt.Errorf("file not found in EPUB: %s", filename)
}
