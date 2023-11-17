package gopub

import (
	"io"
	"path"
	"strings"
)

type ContentType int

const (
	ContentTypeImageGif ContentType = iota + 1
	ContentTypeImageJpeg
	ContentTypeImagePng
	ContentTypeImageSvg
	ContentTypeImageWebp

	ContentTypeAudioMp3
	ContentTypeAudioMp4
	ContentTypeAudioOgg

	ContentTypeCss

	ContentTypeFontTruetype
	ContentTypeFontSfnt
	ContentTypeFontOpentype
	ContentTypeFontWoff
	ContentTypeFontWoff2

	ContentTypeXhtml
	ContentTypeXml
	ContentTypeScript
	ContentTypeDtb
	ContentTypeDtbNcx
	ContentTypeSmil

	ContentTypeOeb1Document
	ContentTypeOeb1Css

	ContentTypeOther
)

type ContentLocation int

const (
	ContentLocationLocal ContentLocation = iota + 1
	ContentLocationRemote
)

type ContentFileType int

const (
	ContentFileTypeText ContentFileType = iota + 1
	ContentFileTypeByteArray
)

type ContentFile struct {
	Key             string
	ContentType     ContentType
	ContentMimeType string
	ContentLocation ContentLocation
	ContentFileType ContentFileType
}

func newContentFile(item ManifestItem, contentDirectoryPath string, contentLocation ContentLocation, contentFileType ContentFileType) ContentFile {
	href := item.Href
	contentMimeType := item.MediaType
	contentType := getContentTypeByMimeType(contentMimeType)

	contentFile := ContentFile{
		Key:             href,
		ContentType:     contentType,
		ContentMimeType: contentMimeType,
		ContentLocation: contentLocation,
		ContentFileType: contentFileType,
	}

	return contentFile
}

type LocalContentFile struct {
	FilePath string
	ContentFile
}

type LocalByteContentFile struct {
	Content []byte
	ContentFile
}

func newLocalByteContentFile(er *epubReader, contentFile ContentFile, contentFilePath string) (LocalByteContentFile, error) {
	var localByteContentFile LocalByteContentFile

	rc, size, err := findFileInZip(er.zipReader, contentFilePath)
	if err != nil {
		return localByteContentFile, err
	}
	defer rc.Close()

	byteContent := make([]byte, size)

	_, err = rc.Read(byteContent)
	if err != nil {
		if err != io.EOF {
			return localByteContentFile, err
		}
	}

	localByteContentFile.Content = byteContent
	localByteContentFile.ContentFile = contentFile

	return localByteContentFile, nil
}

type LocalTextContentFile struct {
	Content string
	ContentFile
}

func newLocalTextContentFile(er *epubReader, contentFile ContentFile, contentFilePath string) (LocalTextContentFile, error) {
	var localTextContentFile LocalTextContentFile

	rc, size, err := findFileInZip(er.zipReader, contentFilePath)
	if err != nil {
		return localTextContentFile, err
	}
	defer rc.Close()

	buf := make([]byte, size)
	_, err = rc.Read(buf)
	if err != nil {
		if err != io.EOF {
			return localTextContentFile, err
		}
	}

	localTextContentFile.Content = string(buf)
	localTextContentFile.ContentFile = contentFile

	return localTextContentFile, nil
}

type Content struct {
	Cover              LocalByteContentFile
	NavigationHtmlFile LocalTextContentFile
	Html               []LocalTextContentFile
	Css                []LocalTextContentFile
	Images             []LocalByteContentFile
	Fonts              []LocalByteContentFile
	Audios             []LocalByteContentFile
	AllFiles           []LocalContentFile
}

func readContent(sc schema, er *epubReader) (Content, error) {
	var content Content

	var cover LocalByteContentFile
	var navigationHtmlFile LocalTextContentFile
	var htmlLocal []LocalTextContentFile
	var cssLocal []LocalTextContentFile
	var imagesLocal []LocalByteContentFile
	var fontsLocal []LocalByteContentFile
	var audiosLocal []LocalByteContentFile
	var allFilesLocal []LocalContentFile

	for _, item := range sc.pkg.Manifest.Items {
		href := item.Href
		contentMimeType := item.MediaType
		contentType := getContentTypeByMimeType(contentMimeType)
		contentDirectoryPath := sc.contentDirectoryPath
		contentFilePath := path.Join(contentDirectoryPath, href)

		contentFile := ContentFile{
			Key:             href,
			ContentType:     contentType,
			ContentMimeType: contentMimeType,
			ContentLocation: ContentLocationLocal,
			ContentFileType: ContentFileTypeText,
		}

		localContentFile := LocalContentFile{
			FilePath:    contentFilePath,
			ContentFile: contentFile,
		}

		switch contentType {
		case ContentTypeXhtml, ContentTypeCss, ContentTypeOeb1Document, ContentTypeOeb1Css, ContentTypeXml, ContentTypeDtb, ContentTypeDtbNcx, ContentTypeSmil, ContentTypeScript:
			localTextContentFile, err := newLocalTextContentFile(er, contentFile, contentFilePath)
			if err != nil {
				return content, err
			}

			if contentType == ContentTypeXhtml {
				htmlLocal = append(htmlLocal, localTextContentFile)

				if navigationHtmlFile.Content == "" && item.Properties != "" && strings.Contains(item.Properties, "nav") {
					navigationHtmlFile = localTextContentFile
				}
			} else if contentType == ContentTypeCss {
				cssLocal = append(cssLocal, localTextContentFile)
			}

			allFilesLocal = append(allFilesLocal, localContentFile)
		default:
			localByteContentFile, err := newLocalByteContentFile(er, contentFile, contentFilePath)
			if err != nil {
				return content, err
			}

			switch contentType {
			case ContentTypeImageGif, ContentTypeImageJpeg, ContentTypeImagePng, ContentTypeImageSvg, ContentTypeImageWebp:
				imagesLocal = append(imagesLocal, localByteContentFile)

				if strings.Contains(item.Properties, "cover-image") {
					cover = localByteContentFile
				}
			case ContentTypeFontTruetype, ContentTypeFontOpentype, ContentTypeFontSfnt, ContentTypeFontWoff, ContentTypeFontWoff2:
				fontsLocal = append(fontsLocal, localByteContentFile)
			case ContentTypeAudioMp3, ContentTypeAudioMp4, ContentTypeAudioOgg:
				audiosLocal = append(audiosLocal, localByteContentFile)
			}

			allFilesLocal = append(allFilesLocal, localContentFile)
		}

	}

	content.Cover = cover
	content.NavigationHtmlFile = navigationHtmlFile
	content.Html = htmlLocal
	content.Css = cssLocal
	content.Images = imagesLocal
	content.Fonts = fontsLocal
	content.Audios = audiosLocal
	content.AllFiles = allFilesLocal

	return content, nil
}

func getContentTypeByMimeType(contentMimeType string) ContentType {
	switch strings.ToLower(contentMimeType) {
	case "application/xhtml+xml":
		return ContentTypeXhtml
	case "application/x-dtbook+xml":
		return ContentTypeDtb
	case "application/x-dtbncx+xml":
		return ContentTypeDtbNcx
	case "text/x-oeb1-document":
		return ContentTypeOeb1Css
	case "application/xml":
		return ContentTypeXml
	case "text/css":
		return ContentTypeCss
	case "text/x-oeb1-css":
		return ContentTypeOeb1Css
	case "application/javascript", "application/ecmascript", "text/javascript":
		return ContentTypeScript
	case "image/gif":
		return ContentTypeImageGif
	case "image/jpeg":
		return ContentTypeImageJpeg
	case "image/png":
		return ContentTypeImagePng
	case "image/svg+xml":
		return ContentTypeImageSvg
	case "image/webp":
		return ContentTypeImageWebp
	case "font/truetype", "font/ttf", "application/x-font-truetype":
		return ContentTypeFontTruetype
	case "font/opentype", "font/otf", "application/vnd.ms-opentype":
		return ContentTypeFontOpentype
	case "font/sfnt", "application/font-sfnt":
		return ContentTypeFontSfnt
	case "font/woff", "application/font-woff":
		return ContentTypeFontWoff
	case "font/woff2":
		return ContentTypeFontWoff2
	case "application/smil+xml":
		return ContentTypeSmil
	case "audio/mpeg":
		return ContentTypeAudioMp3
	case "audio/mp4":
		return ContentTypeAudioMp4
	case "audio/ogg", "audio/ogg; codecs=opus":
		return ContentTypeAudioOgg
	default:
		return ContentTypeOther
	}
}
