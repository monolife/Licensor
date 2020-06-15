package main

import(
	// "fmt"
	"path/filepath"
	"flag"
	// "log"
	"os"
	"go.uber.org/zap"
	"path"
	"strings"

	prep "dzynetech.com/licensor/prepender"
)

var slogger *zap.SugaredLogger;

func main(){

	// --- 0. Parse inputs
	fileDir := flag.String("s", "./testDir", "Directory of files to prepend with licenses (recursive)");
	licenseDir := flag.String("l", "./LicenseFiles", "Directory of license files where filename format is <extension>.license (ex. cpp.license)");
	debugMode := flag.Bool("d", false, "Flag for debug mode with verbose and prettier output");
	flag.Parse()

	logger, _ := zap.NewProduction();
	if(*debugMode){
		logger, _ = zap.NewDevelopment();
	}
	defer logger.Sync() // flushes buffer, if any
	slogger = logger.Sugar()

	// --- 1. Load license files
	licenses, err := LoadLicenses( licenseDir );
	if err != nil{
		slogger.Fatalw("Failed to load license files",
			"license directory", licenseDir,
			// "error", err,
		);
	}
	
	// --- 2. Process files (by kind)
	for _,lic := range licenses {
		fileNames, err := GetFilteredPaths( fileDir, lic.Type);
		if err != nil{
			slogger.Fatalw("Failed to load files to license",
				"File directory", fileDir,
				"License Type", lic.Type,
			);
		}
		for _,path := range fileNames{
			doc := prep.NewDoc(path, slogger);
			// --- 3. Prepend License block to file
			already, _ := doc.AlreadyPrepended(lic.Contents);
			if( !already ){
				doc.Prepend(lic.Contents);
			}
		}
	}

}

func LoadLicenses(licDirPath *string)( licenses []prep.Doc, err error){ 
	licenses = []prep.Doc{};
	licePaths, err := GetFilteredPaths(licDirPath, ".license"); 
	if err != nil{
		return licenses, err
	}

	for _, path := range licePaths{
		license := prep.NewDoc(path, slogger);
		if err = license.ReadLines(); err != nil{
			return;
		}else{
			license.Type = "."+FilenameWithoutExtension(filepath.Base(path));
			slogger.Infow("Adding license",
				"type", license.Type,
				"path", license.Path);
			licenses = append(licenses, license);
		}
	}

	return;
}

func GetFilePaths( dirPath *string) (fileNames []string, err error){
	return GetFilteredPaths(dirPath, "");
}

func GetFilteredPaths( dirPath *string, fileType string) (fileNames []string, err error){
	fileNames = []string{};

	err = filepath.Walk( *dirPath,
		func(path string, info os.FileInfo, e error) error {
			if e != nil {
				return e;
			}
			if info.IsDir() {
				return nil
			}
			if (fileType != "" && filepath.Ext(path) != fileType) {
				return nil
			}
			slogger.Debugw("Found",
				"path", path,
				"filter", fileType,
				"size", info.Size(),
			);
			// fmt.Println(path, info.Size())
			fileNames = append(fileNames, path);
			return nil;
		});

	if err != nil {
		// log.Println(err)
		slogger.Errorw("Failed to parse filepath",
			"path", *dirPath,
		);
		return;
	}
	return;
}

func FilenameWithoutExtension(fn string) string {
      return strings.TrimSuffix(fn, path.Ext(fn))
}