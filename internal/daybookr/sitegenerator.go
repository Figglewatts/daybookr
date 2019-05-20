package daybookr

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/otiai10/copy"

	"github.com/smallfish/simpleyaml"
)

const layoutsDir = "layouts"
const postsDir = "posts"
const includesDir = "includes"
const pagesDir = "pages"

func Generate(baseURL string, inputFolder string, outputFolder string, configPath string) error {
	// check to see if the input folder exists
	inputFolderExists, err := exists(inputFolder)
	if err != nil || !inputFolderExists {
		return fmt.Errorf("input folder '%s' did not exist", inputFolder)
	}

	// check to see if the config file exists
	configPathExists, err := exists(configPath)
	if err != nil || !configPathExists {
		return fmt.Errorf("config file '%s' did not exist", configPath)
	}

	// remove the output folder and remake it to clear contents
	os.RemoveAll(outputFolder)
	os.MkdirAll(outputFolder, os.ModePerm)

	// load the config
	config, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("could not load config: %v", err)
	}

	// get data folders from config
	dataFolders, err := getDataFoldersFromConfig(config, inputFolder)
	if err != nil {
		return fmt.Errorf("could not load data folders: %v", err)
	}

	site, err := createSite(baseURL, config, inputFolder)
	if err != nil {
		return fmt.Errorf("could not create site: %v", err)
	}

	// get includes filenames
	includes, err := getFilesInDir(path.Join(inputFolder, includesDir), "*.html")
	if err != nil {
		return err
	}

	index, err := loadTemplate(path.Join(inputFolder, "index.html"), includes)
	if err != nil {
		return err
	}

	// render the index page
	renderedIndex, err := renderTemplate(index, site)
	if err != nil {
		return err
	}

	// write the index page
	err = writeHTML(renderedIndex, path.Join(outputFolder, "index"))
	if err != nil {
		return fmt.Errorf("could not write index: %v", err)
	}

	// now load site layouts
	layouts, err := loadAllTemplates(path.Join(inputFolder, layoutsDir), includes)

	// render all of the pages
	for _, page := range site.Pages {
		// check to see if the layout exists
		if layout, ok := layouts[page.Layout]; ok {
			// if it does, render the page and write it
			renderedPage, err := renderTemplate(layout, page)
			if err != nil {
				return fmt.Errorf("unable to render page '%s': %v", page.Name, err)
			}
			err = writeHTML(renderedPage, path.Join(outputFolder, page.Name))
			if err != nil {
				return fmt.Errorf("unable to write rendered page '%s': %v", page.Name, err)
			}
		} else {
			return fmt.Errorf("unable to generate page '%s' with unknown layout '%s'", page.Name, page.Layout)
		}
	}

	// render all of the posts
	for _, post := range site.Posts {
		// check to see if the layout exists
		if layout, ok := layouts[post.Layout]; ok {
			// if it does, render the page and write it
			renderedPost, err := renderTemplate(layout, post)
			if err != nil {
				return fmt.Errorf("unable to render post '%s': %v", post.Name, err)
			}
			err = writeHTML(renderedPost, path.Join(outputFolder, post.Name))
			if err != nil {
				return fmt.Errorf("unable to write rendered post '%s': %v", post.Name, err)
			}
		} else {
			return fmt.Errorf("unable to generate post '%s' with unknown layout '%s'", post.Name, post.Layout)
		}
	}

	// copy data folders to output
	err = copyDataFoldersToOutput(dataFolders, outputFolder)
	if err != nil {
		return err
	}

	return nil
}

func getDataFoldersFromConfig(config *simpleyaml.Yaml, inputPath string) ([]string, error) {
	dataFoldersYAML := config.Get(configDataFoldersField)
	dataFoldersArr, err := dataFoldersYAML.Array()
	if err != nil {
		return nil, fmt.Errorf("%s was not of array type", configDataFoldersField)
	}
	dataFolders := make([]string, len(dataFoldersArr))
	for i := range dataFoldersArr {
		folder, err := dataFoldersYAML.GetIndex(i).String()
		if err != nil {
			return nil, fmt.Errorf("folder %d was not of string type", i)
		}
		dataFolders[i] = path.Join(inputPath, folder)
	}
	return dataFolders, nil
}

func copyDataFoldersToOutput(dataFolders []string, outputFolder string) error {
	for _, folder := range dataFolders {
		folderName := path.Base(folder)
		outputFolder := path.Join(outputFolder, folderName)
		err := copy.Copy(folder, outputFolder)
		if err != nil {
			return fmt.Errorf("error copying data folder %s: %v", folderName, err)
		}
	}
	return nil
}

func writeHTML(html string, filename string) error {
	return ioutil.WriteFile(filename+".html", []byte(html), 0644)
}
