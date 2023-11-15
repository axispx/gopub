package gopub

import (
	"archive/zip"
	"encoding/xml"
)

type StructuralSemanticsProperty int

const (
	SspCover StructuralSemanticsProperty = iota + 1
	SspFrontMatter
	SspBodyMatter
	SspBackMatter
	SspVolume
	SspPart
	SspChapter
	SspSubChapter
	SspDivision
	SspAbstract
	SspForeword
	SspPreface
	SspPrologue
	SspIntroduction
	SspPreamble
	SspConclusion
	SspEpilogue
	SspAfterword
	SspEpigram
	SspToc
	SspTocBrief
	SspLandmarks
	SspLoa
	SspLoi
	SspLot
	SspLov
	SspAppendix
	SspColophon
	SspCredits
	SspKeywords
	SspIndex
	SspIndexHeadnotes
	SspIndexLegend
	SspIndexGroup
	SspIndexEntryList
	SspIndexEntry
	SspIndexTerm
	SspIndexEditorNote
	SspIndexLocator
	SspIndexLocatorList
	SspIndexLocatorName
	SspIndexXrefPreferred
	SspIndexXrefRelated
	SspIndexTermCategory
	SspIndexTermCategories
	SspGlossary
	SspGlossTerm
	SspGlossDef
	SspBibliography
	SspBibloEntry
	SspTitlePage
	SspHalfTitlePage
	SspCopyrightPage
	SspSeriesPage
	SspAcknowledgements
	SspImprint
	SspImprimatur
	SspContributors
	SspOtherCredits
	SspErrata
	SspDedication
	SspRevisionHistory
	SspCaseStudy
	SspHelp
	SspMarginalia
	SspNotice
	SspPullQuote
	SspSidebar
	SspTip
	SspWarning
	SspHalfTitle
	SspFullTitle
	SspCoverTitle
	SspTitle
	SspSubtitle
	SspLabel
	SspOrdinal
	SspBridgehead
	SspLearningObjective
	SspLearningObjectives
	SspLearningOutcome
	SspLearningOutcomes
	SspLearningResource
	SspLearningResources
	SspLearningStandard
	SspLearningStandards
	SspAnswer
	SspAnswers
	SspAssessment
	SspAssessments
	SspFeedback
	SspFillInTheBlankProblem
	SspGeneralProblem
	SspQna
	SspMatchProblem
	SspMultipleChoiceProblem
	SspPractice
	SspQuestion
	SspPractices
	SspTrueFalseProblem
	SspPanel
	SspPanelGroup
	SspBalloon
	SspTextArea
	SspSoundArea
	SspAnnotation
	SspNote
	SspFootnote
	SspEndnote
	SspRearnote
	SspFootnotes
	SspEndnotes
	SspRearnotes
	SspAnnoRef
	SspBiblioRef
	SspGlossRef
	SspNoteRef
	SspBacklink
	SspCredit
	SspKeyword
	SspTopicSentence
	SspConcludingSentence
	SspPagebreak
	SspPageList
	SspTable
	SspTableRow
	SspTableCell
	SspList
	SspListItem
	SspFigure
	SspAside
	SspUnknown
)

type NavAnchor struct {
	Href  string `xml:"href,attr"`
	Text  string `xml:",chardata"`
	Title string `xml:"title,attr"`
	Alt   string `xml:"alt,attr"`
	Type  string `xml:"type,attr"`
}

type NavSpan struct {
	Text  string `xml:"text"`
	Title string `xml:"title"`
	Alt   string `xml:"alt"`
}

type NavLi struct {
	Anchor  NavAnchor `xml:"a"`
	Span    NavSpan   `xml:"span"`
	ChildOl NavOl     `xml:"ol"`
}

type NavOl struct {
	IsHidden *string `xml:"hidden,attr"`
	Lis      []NavLi `xml:"li"`
}

type Navigation struct {
	Type     string  `xml:"type,attr"`
	IsHidden *string `xml:"hidden,attr"`
	H1       string  `xml:"h1"`
	H2       string  `xml:"h2"`
	H3       string  `xml:"h3"`
	H4       string  `xml:"h4"`
	H5       string  `xml:"h5"`
	H6       string  `xml:"h6"`
	Ol       NavOl   `xml:"ol"`
}

func (n *Navigation) getHeader() string {
	if n.H1 != "" {
		return n.H1
	} else if n.H2 != "" {
		return n.H2
	} else if n.H3 != "" {
		return n.H3
	} else if n.H4 != "" {
		return n.H4
	} else if n.H5 != "" {
		return n.H5
	} else {
		return n.H6
	}
}

type NavigationDocument struct {
	FilePath    string
	Title       string
	Navigations []*Navigation
}

func readNavigation(zf *zip.ReadCloser, filePath string) (*NavigationDocument, error) {
	rc, err := findFileInZip(zf, filePath)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var navs []*Navigation
	var title string

	decoder := xml.NewDecoder(rc)
	for {
		token, err := decoder.Token()
		if token == nil {
			break
		}

		if err != nil {
			return nil, err
		}

		switch se := token.(type) {
		case xml.StartElement:
			if se.Name.Local == "nav" {
				var nav Navigation
				err := decoder.DecodeElement(&nav, &se)
				if err != nil {
					return nil, err
				}

				navs = append(navs, &nav)
			} else if se.Name.Local == "title" {
				err := decoder.DecodeElement(&title, &se)
				if err != nil {
					return nil, err
				}
			}
		}

	}

	navDoc := &NavigationDocument{
		FilePath:    filePath,
		Title:       title,
		Navigations: navs,
	}

	return navDoc, nil
}
