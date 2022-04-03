package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func selectOption(label string, repos []string) string {
	prompt := promptui.Select{
		Label: label,
		Items: repos,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return ""
	}
	fmt.Print("\033[H\033[2J")
	return result
}

func getUser() string {
	validate := func(input string) error {
		if len(input) <= 4 {
			return errors.New("invalid input")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    "Username",
		Validate: validate,
	}
	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
	}
	fmt.Print("\033[H\033[2J")
	return result
}

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func GET(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func main() {
	var user string
	var selectRepo string
	var names []string
	var assets []string
	var binaries []string
	var downloadUri string
	var Releases []map[string]interface{}
	var Repos []map[string]interface{}
	app := &cli.App{
		Name:  "gpm",
		Usage: "get github releases",
		Action: func(c *cli.Context) error {
			user = getUser()
			if user == "" {
				os.Exit(0)
			}
			getRepos, err := GET("https://api.github.com/users/" + user + "/repos")
			if err != nil {
				log.Fatalln(err)
			}
			json.Unmarshal(getRepos, &Repos)
			for _, repo := range Repos {
				names = append(names, repo["name"].(string))
			}
			selectRepo = selectOption("Select a repo", names)
			if selectRepo == "" {
				os.Exit(0)
			}
			getReleases, err := GET("https://api.github.com/repos/" + user + "/" + selectRepo + "/releases")
			if err != nil {
				log.Fatalln(err)
			}
			json.Unmarshal(getReleases, &Releases)
			if len(Releases) == 0 {
				fmt.Println("No releases found")
				os.Exit(0)
			}
			for _, repo := range Releases {
				binaries = append(binaries, repo["name"].(string))
			}
			release := selectOption("Select a Release", binaries)
			for _, repo := range Releases {
				if repo["name"].(string) == release {
					if len(repo["assets"].([]interface{})) > 0 {
						for _, asset := range repo["assets"].([]interface{}) {
							assets = append(assets, asset.(map[string]interface{})["name"].(string))
						}
					} else {
						fmt.Println("No assets found")
						os.Exit(0)
					}
				}
			}
			assetSelection := selectOption("Select an asset", assets)
			for _, repo := range Releases {
				if repo["name"].(string) == release {
					for _, asset := range repo["assets"].([]interface{}) {
						if asset.(map[string]interface{})["name"].(string) == assetSelection {
							downloadUri = asset.(map[string]interface{})["browser_download_url"].(string)
						}
					}
				}
			}
			dir, err := os.Getwd()
			if err != nil {
				log.Fatalln(err)
			}
			DownloadFile(dir+"/"+assetSelection, downloadUri)
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
