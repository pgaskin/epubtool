package epubtransform

import (
	"errors"
	"fmt"

	"github.com/beevik/etree"
)

// TransformTitle sets the epub opf dc:title.
func TransformTitle(title string) Transform {
	return TransformOPFMetadataElementContent(fmt.Sprintf("set title to %#v", title), "dc:title", title)
}

// TransformCreator sets the epub opf dc:creator.
func TransformCreator(creator string) Transform {
	return TransformOPFMetadataElementContent(fmt.Sprintf("set creator to %#v", creator), "dc:creator", creator)
}

// TransformDescription sets the epub opf dc:description.
func TransformDescription(description string) Transform {
	return TransformOPFMetadataElementContent(fmt.Sprintf("set description to %#v", description), "dc:description", description)
}

// TransformOPFMetadataElementContent sets the text content of the first instance of an element in package>metadata>element.
func TransformOPFMetadataElementContent(desc, tag, content string) Transform {
	return Transform{
		Desc: desc,
		OPFDoc: func(opf *etree.Document) error {
			me := opf.FindElement("//package/metadata")
			if me == nil {
				return errors.New("could not find package>metadata element")
			}
			tel := me.FindElement(tag)
			if tel == nil {
				tel = me.CreateElement(tag)
			}
			tel.SetText(content)
			return nil
		},
	}
}

// TransformOPFMetaElementContent sets the content attribute of the first instance of a package>metadata>meta[name][content].
func TransformOPFMetaElementContent(desc, name, content string) Transform {
	return Transform{
		Desc: desc,
		OPFDoc: func(opf *etree.Document) error {
			me := opf.FindElement("//package/metadata")
			if me == nil {
				return errors.New("could not find package>metadata element")
			}
			var mel *etree.Element
			for _, el := range me.FindElements("//meta") {
				if el.SelectAttrValue("name", "") == name {
					mel = el
					break
				}
			}
			if mel == nil {
				mel = me.CreateElement("meta")
				mel.CreateAttr("name", name)
			}
			mel.RemoveAttr("content")
			mel.CreateAttr("content", content)
			return nil
		},
	}
}

// TransformOPFBeautify indents the OPF file.
func TransformOPFBeautify(spaces int) Transform {
	return Transform{
		Desc: fmt.Sprintf("indent by %d spaces", spaces),
		OPFDoc: func(opf *etree.Document) error {
			opf.Indent(4)
			return nil
		},
	}
}
