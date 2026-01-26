// Package pptx provides PowerPoint (.pptx) file reading functionality.
package pptx

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Info contains metadata about a PowerPoint presentation.
type Info struct {
	Path       string `json:"path"`
	SlideCount int    `json:"slideCount"`
	Title      string `json:"title,omitempty"`
	Creator    string `json:"creator,omitempty"`
	Subject    string `json:"subject,omitempty"`
	Created    string `json:"created,omitempty"`
	Modified   string `json:"modified,omitempty"`
	Error      string `json:"error,omitempty"`
}

// Slide represents a single slide in the presentation.
type Slide struct {
	Number int    `json:"number"`
	Title  string `json:"title,omitempty"`
	Text   string `json:"text"`
}

// coreProperties represents docProps/core.xml
type coreProperties struct {
	Title    string `xml:"title"`
	Creator  string `xml:"creator"`
	Subject  string `xml:"subject"`
	Created  string `xml:"created"`
	Modified string `xml:"modified"`
}

// GetInfo returns metadata about a PowerPoint file.
func GetInfo(path string) (*Info, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = r.Close() }()

	info := &Info{
		Path: path,
	}

	// Count slides
	slidePattern := regexp.MustCompile(`^ppt/slides/slide\d+\.xml$`)
	for _, f := range r.File {
		if slidePattern.MatchString(f.Name) {
			info.SlideCount++
		}
	}

	// Read core properties
	for _, f := range r.File {
		if f.Name == "docProps/core.xml" {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			props, err := parseCoreProperties(rc)
			_ = rc.Close()
			if err == nil {
				info.Title = props.Title
				info.Creator = props.Creator
				info.Subject = props.Subject
				info.Created = props.Created
				info.Modified = props.Modified
			}
			break
		}
	}

	return info, nil
}

func parseCoreProperties(r io.Reader) (*coreProperties, error) {
	var props coreProperties
	decoder := xml.NewDecoder(r)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if se, ok := token.(xml.StartElement); ok {
			switch se.Name.Local {
			case "title":
				var val string
				if err := decoder.DecodeElement(&val, &se); err == nil {
					props.Title = val
				}
			case "creator":
				var val string
				if err := decoder.DecodeElement(&val, &se); err == nil {
					props.Creator = val
				}
			case "subject":
				var val string
				if err := decoder.DecodeElement(&val, &se); err == nil {
					props.Subject = val
				}
			case "created":
				var val string
				if err := decoder.DecodeElement(&val, &se); err == nil {
					props.Created = val
				}
			case "modified":
				var val string
				if err := decoder.DecodeElement(&val, &se); err == nil {
					props.Modified = val
				}
			}
		}
	}
	return &props, nil
}

// ExtractText extracts all text content from a PowerPoint file.
func ExtractText(path string) (string, error) {
	slides, err := GetSlides(path)
	if err != nil {
		return "", err
	}

	var parts []string
	for _, slide := range slides {
		if slide.Text != "" {
			header := fmt.Sprintf("--- Slide %d ---", slide.Number)
			if slide.Title != "" {
				header = fmt.Sprintf("--- Slide %d: %s ---", slide.Number, slide.Title)
			}
			parts = append(parts, header)
			parts = append(parts, slide.Text)
		}
	}

	return strings.Join(parts, "\n\n"), nil
}

// ExtractSlideText extracts text from a specific slide (1-indexed).
func ExtractSlideText(path string, slideNum int) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = r.Close() }()

	slidePath := fmt.Sprintf("ppt/slides/slide%d.xml", slideNum)
	for _, f := range r.File {
		if f.Name == slidePath {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open slide: %w", err)
			}
			text, err := extractTextFromXML(rc)
			_ = rc.Close()
			return text, err
		}
	}

	return "", fmt.Errorf("slide %d not found", slideNum)
}

// GetSlides returns all slides with their content.
func GetSlides(path string) ([]Slide, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = r.Close() }()

	// Find all slide files
	slidePattern := regexp.MustCompile(`^ppt/slides/slide(\d+)\.xml$`)
	var slideFiles []*zip.File
	slideNumbers := make(map[*zip.File]int)

	for _, f := range r.File {
		matches := slidePattern.FindStringSubmatch(f.Name)
		if matches != nil {
			var num int
			if _, err := fmt.Sscanf(matches[1], "%d", &num); err == nil {
				slideFiles = append(slideFiles, f)
				slideNumbers[f] = num
			}
		}
	}

	// Sort by slide number
	sort.Slice(slideFiles, func(i, j int) bool {
		return slideNumbers[slideFiles[i]] < slideNumbers[slideFiles[j]]
	})

	var slides []Slide
	for _, f := range slideFiles {
		rc, err := f.Open()
		if err != nil {
			continue
		}
		text, _ := extractTextFromXML(rc)
		_ = rc.Close()

		slide := Slide{
			Number: slideNumbers[f],
			Text:   text,
		}

		// Try to extract title (first text block is often the title)
		lines := strings.Split(text, "\n")
		if len(lines) > 0 && lines[0] != "" {
			slide.Title = lines[0]
		}

		slides = append(slides, slide)
	}

	return slides, nil
}

// extractTextFromXML extracts text from PowerPoint XML content.
// Text is contained in <a:t> tags.
func extractTextFromXML(r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var texts []string
	var currentParagraph []string

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch t := token.(type) {
		case xml.StartElement:
			// <a:t> contains text
			if t.Name.Local == "t" {
				var text string
				if err := decoder.DecodeElement(&text, &t); err == nil && text != "" {
					currentParagraph = append(currentParagraph, text)
				}
			}
		case xml.EndElement:
			// End of paragraph <a:p>
			if t.Name.Local == "p" && len(currentParagraph) > 0 {
				texts = append(texts, strings.Join(currentParagraph, ""))
				currentParagraph = nil
			}
		}
	}

	// Add any remaining text
	if len(currentParagraph) > 0 {
		texts = append(texts, strings.Join(currentParagraph, ""))
	}

	return strings.Join(texts, "\n"), nil
}

// ListSlides returns a summary of all slides.
func ListSlides(path string) ([]Slide, error) {
	return GetSlides(path)
}

// ValidatePPTX checks if a file is a valid PPTX file.
func ValidatePPTX(path string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("not a valid ZIP file: %w", err)
	}
	defer func() { _ = r.Close() }()

	// Check for required PPTX structure
	hasContentTypes := false
	hasPresentation := false

	for _, f := range r.File {
		switch f.Name {
		case "[Content_Types].xml":
			hasContentTypes = true
		case "ppt/presentation.xml":
			hasPresentation = true
		}
	}

	if !hasContentTypes {
		return fmt.Errorf("missing [Content_Types].xml - not a valid Office document")
	}
	if !hasPresentation {
		return fmt.Errorf("missing ppt/presentation.xml - not a valid PowerPoint file")
	}

	return nil
}

// IsPPTX checks if a file path appears to be a PowerPoint file.
func IsPPTX(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".pptx"
}

// CreateOptions configures PPTX creation.
type CreateOptions struct {
	Title   string
	Creator string
	Subject string
}

// SlideContent represents content for a slide to create.
type SlideContent struct {
	Title string
	Body  string
}

// Create creates a new PowerPoint file with the given slides.
func Create(outputPath string, slides []SlideContent, opts CreateOptions) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = f.Close() }()

	w := zip.NewWriter(f)
	defer func() { _ = w.Close() }()

	// Write required PPTX structure
	if err := writeContentTypes(w, len(slides)); err != nil {
		return err
	}
	if err := writeRels(w); err != nil {
		return err
	}
	if err := writeCoreProps(w, opts); err != nil {
		return err
	}
	if err := writeAppProps(w); err != nil {
		return err
	}
	if err := writePresentation(w, len(slides)); err != nil {
		return err
	}
	if err := writePresentationRels(w, len(slides)); err != nil {
		return err
	}

	// Write each slide
	for i, slide := range slides {
		if err := writeSlide(w, i+1, slide); err != nil {
			return err
		}
		if err := writeSlideRels(w, i+1); err != nil {
			return err
		}
	}

	// Write slide layouts and masters (minimal)
	if err := writeSlideLayout(w); err != nil {
		return err
	}
	if err := writeSlideMaster(w); err != nil {
		return err
	}
	if err := writeTheme(w); err != nil {
		return err
	}

	return w.Close()
}

// CreateFromText creates a simple presentation from text.
// Each paragraph becomes a slide with the first line as title.
func CreateFromText(outputPath string, text string, opts CreateOptions) error {
	paragraphs := strings.Split(strings.TrimSpace(text), "\n\n")
	var slides []SlideContent

	for _, para := range paragraphs {
		lines := strings.SplitN(strings.TrimSpace(para), "\n", 2)
		slide := SlideContent{
			Title: lines[0],
		}
		if len(lines) > 1 {
			slide.Body = lines[1]
		}
		slides = append(slides, slide)
	}

	return Create(outputPath, slides, opts)
}

func writeContentTypes(w *zip.Writer, slideCount int) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>
  <Override PartName="/ppt/slideMasters/slideMaster1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideMaster+xml"/>
  <Override PartName="/ppt/slideLayouts/slideLayout1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideLayout+xml"/>
  <Override PartName="/ppt/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>
  <Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>
  <Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>`

	for i := 1; i <= slideCount; i++ {
		content += fmt.Sprintf(`
  <Override PartName="/ppt/slides/slide%d.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>`, i)
	}

	content += `
</Types>`

	return writeFile(w, "[Content_Types].xml", content)
}

func writeRels(w *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>`
	return writeFile(w, "_rels/.rels", content)
}

func writeCoreProps(w *zip.Writer, opts CreateOptions) error {
	title := opts.Title
	if title == "" {
		title = "Presentation"
	}
	creator := opts.Creator
	if creator == "" {
		creator = "ps-pptx"
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>%s</dc:title>
  <dc:creator>%s</dc:creator>
  <dc:subject>%s</dc:subject>
</cp:coreProperties>`, escapeXML(title), escapeXML(creator), escapeXML(opts.Subject))
	return writeFile(w, "docProps/core.xml", content)
}

func writeAppProps(w *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties">
  <Application>ps-pptx</Application>
</Properties>`
	return writeFile(w, "docProps/app.xml", content)
}

func writePresentation(w *zip.Writer, slideCount int) error {
	slideList := ""
	for i := 1; i <= slideCount; i++ {
		slideList += fmt.Sprintf(`<p:sldId id="%d" r:id="rId%d"/>`, 255+i, i)
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:sldMasterIdLst>
    <p:sldMasterId id="2147483648" r:id="rId%d"/>
  </p:sldMasterIdLst>
  <p:sldIdLst>%s</p:sldIdLst>
  <p:sldSz cx="9144000" cy="6858000" type="screen4x3"/>
  <p:notesSz cx="6858000" cy="9144000"/>
</p:presentation>`, slideCount+1, slideList)
	return writeFile(w, "ppt/presentation.xml", content)
}

func writePresentationRels(w *zip.Writer, slideCount int) error {
	rels := ""
	for i := 1; i <= slideCount; i++ {
		rels += fmt.Sprintf(`
  <Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide%d.xml"/>`, i, i)
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">%s
  <Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="slideMasters/slideMaster1.xml"/>
  <Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="theme/theme1.xml"/>
</Relationships>`, rels, slideCount+1, slideCount+2)
	return writeFile(w, "ppt/_rels/presentation.xml.rels", content)
}

func writeSlide(w *zip.Writer, num int, slide SlideContent) error {
	title := escapeXML(slide.Title)
	body := escapeXML(slide.Body)

	// Build body paragraphs
	bodyContent := ""
	if body != "" {
		lines := strings.Split(body, "\n")
		for _, line := range lines {
			bodyContent += fmt.Sprintf(`
              <a:p>
                <a:r>
                  <a:rPr lang="en-US" sz="2400"/>
                  <a:t>%s</a:t>
                </a:r>
              </a:p>`, escapeXML(line))
		}
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:nvGrpSpPr>
        <p:cNvPr id="1" name=""/>
        <p:cNvGrpSpPr/>
        <p:nvPr/>
      </p:nvGrpSpPr>
      <p:grpSpPr/>
      <p:sp>
        <p:nvSpPr>
          <p:cNvPr id="2" name="Title"/>
          <p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr>
          <p:nvPr><p:ph type="title"/></p:nvPr>
        </p:nvSpPr>
        <p:spPr>
          <a:xfrm>
            <a:off x="457200" y="274638"/>
            <a:ext cx="8229600" cy="1143000"/>
          </a:xfrm>
        </p:spPr>
        <p:txBody>
          <a:bodyPr/>
          <a:lstStyle/>
          <a:p>
            <a:r>
              <a:rPr lang="en-US" sz="4400" b="1"/>
              <a:t>%s</a:t>
            </a:r>
          </a:p>
        </p:txBody>
      </p:sp>
      <p:sp>
        <p:nvSpPr>
          <p:cNvPr id="3" name="Content"/>
          <p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr>
          <p:nvPr><p:ph idx="1"/></p:nvPr>
        </p:nvSpPr>
        <p:spPr>
          <a:xfrm>
            <a:off x="457200" y="1600200"/>
            <a:ext cx="8229600" cy="4525963"/>
          </a:xfrm>
        </p:spPr>
        <p:txBody>
          <a:bodyPr/>
          <a:lstStyle/>%s
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
  <p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>
</p:sld>`, title, bodyContent)
	return writeFile(w, fmt.Sprintf("ppt/slides/slide%d.xml", num), content)
}

func writeSlideRels(w *zip.Writer, num int) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
</Relationships>`
	return writeFile(w, fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", num), content)
}

func writeSlideLayout(w *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldLayout xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" type="titleOnly">
  <p:cSld name="Title Only">
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr/>
    </p:spTree>
  </p:cSld>
  <p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>
</p:sldLayout>`

	// Write layout rels
	relsContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="../slideMasters/slideMaster1.xml"/>
</Relationships>`

	if err := writeFile(w, "ppt/slideLayouts/slideLayout1.xml", content); err != nil {
		return err
	}
	return writeFile(w, "ppt/slideLayouts/_rels/slideLayout1.xml.rels", relsContent)
}

func writeSlideMaster(w *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldMaster xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:bg>
      <p:bgRef idx="1001">
        <a:schemeClr val="bg1"/>
      </p:bgRef>
    </p:bg>
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr/>
    </p:spTree>
  </p:cSld>
  <p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/>
  <p:sldLayoutIdLst>
    <p:sldLayoutId id="2147483649" r:id="rId1"/>
  </p:sldLayoutIdLst>
</p:sldMaster>`

	relsContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="../theme/theme1.xml"/>
</Relationships>`

	if err := writeFile(w, "ppt/slideMasters/slideMaster1.xml", content); err != nil {
		return err
	}
	return writeFile(w, "ppt/slideMasters/_rels/slideMaster1.xml.rels", relsContent)
}

func writeTheme(w *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Office Theme">
  <a:themeElements>
    <a:clrScheme name="Office">
      <a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1>
      <a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1>
      <a:dk2><a:srgbClr val="44546A"/></a:dk2>
      <a:lt2><a:srgbClr val="E7E6E6"/></a:lt2>
      <a:accent1><a:srgbClr val="4472C4"/></a:accent1>
      <a:accent2><a:srgbClr val="ED7D31"/></a:accent2>
      <a:accent3><a:srgbClr val="A5A5A5"/></a:accent3>
      <a:accent4><a:srgbClr val="FFC000"/></a:accent4>
      <a:accent5><a:srgbClr val="5B9BD5"/></a:accent5>
      <a:accent6><a:srgbClr val="70AD47"/></a:accent6>
      <a:hlink><a:srgbClr val="0563C1"/></a:hlink>
      <a:folHlink><a:srgbClr val="954F72"/></a:folHlink>
    </a:clrScheme>
    <a:fontScheme name="Office">
      <a:majorFont><a:latin typeface="Calibri Light"/><a:ea typeface=""/><a:cs typeface=""/></a:majorFont>
      <a:minorFont><a:latin typeface="Calibri"/><a:ea typeface=""/><a:cs typeface=""/></a:minorFont>
    </a:fontScheme>
    <a:fmtScheme name="Office">
      <a:fillStyleLst>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
      </a:fillStyleLst>
      <a:lnStyleLst>
        <a:ln w="6350"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln>
        <a:ln w="12700"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln>
        <a:ln w="19050"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln>
      </a:lnStyleLst>
      <a:effectStyleLst>
        <a:effectStyle><a:effectLst/></a:effectStyle>
        <a:effectStyle><a:effectLst/></a:effectStyle>
        <a:effectStyle><a:effectLst/></a:effectStyle>
      </a:effectStyleLst>
      <a:bgFillStyleLst>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
      </a:bgFillStyleLst>
    </a:fmtScheme>
  </a:themeElements>
</a:theme>`
	return writeFile(w, "ppt/theme/theme1.xml", content)
}

func writeFile(w *zip.Writer, name, content string) error {
	f, err := w.Create(name)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(content))
	return err
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
