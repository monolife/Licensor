package prepender

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"go.uber.org/zap"
)

type Doc struct{
	Path string
	Contents []string
	Slogger *zap.SugaredLogger
	Type string
}

func NewDoc( docPath string, sLogger ...*zap.SugaredLogger  ) Doc {
	return Doc{
		Path: docPath,
		Contents: make([]string, 0),
		Slogger: sLogger[0],
		Type: "",
	}
}

func (doc *Doc) AlreadyPrepended( preText []string)( isAlready bool, err error){

	isAlready = true;

	f, err := os.OpenFile(doc.Path, os.O_RDONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for _, preLine := range preText {
		scanner.Scan();
		docLine := scanner.Text();
    if(preLine != docLine){
    	isAlready = false;
    	break;
    }
	}

	return;
}

func (doc *Doc) Prepend( preText []string)(err error){

	//Open document to be prepended (note: using backup)
	document, err := os.Open(doc.Path)
	if err != nil {
		log.Println("Error! Could not source file:", err);
		return;
	}
	defer document.Close()

	// Open document to write into
	out, err := os.OpenFile(doc.Path+".bak", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error! Could create prepended file:", err);
		return;
	}
	defer out.Close()

	// Add prepended text
	writer := bufio.NewWriter(out)
	for _, line := range preText {
		_, err := writer.WriteString(fmt.Sprintf("%s\n", line))
		if err != nil {
			log.Println("Error! Could prepend into file:", err)
			return err
		}
	}
	if err = writer.Flush(); err != nil {
		log.Println("Error! Could prepend into file:", err)
		return
	}

	_, err = io.Copy(out, document)
	if err != nil {
		log.Println("Error! Failed to append orignal file to output:", err);
		return;
	}
	// log.Printf("wrote %d bytes of %s to %s\n", n, doc.Path+".bak", doc.Path)

	if err = os.Rename(doc.Path+".bak", doc.Path); err != nil{
		log.Println("Error! Could not replace source file with prepended:", err);
		return;
	}

	return;
}

func (doc *Doc) ReadLines() error {
	if _, err := os.Stat(doc.Path); err != nil {
		return nil
	}

	f, err := os.OpenFile(doc.Path, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if tmp := scanner.Text(); len(tmp) != 0 {
			doc.Contents = append(doc.Contents, tmp)
		}
	}

	return nil
}