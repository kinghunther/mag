package main

import (
	"encoding/json"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
)

func newP(c *cli.Context) error {
	ngChan := make(chan error)
	if c.NArg() == 0 {
		return errNoArg
	}
	projectName := c.Args().Get(0)

	GOPATH := getGOPATH()
	GOPATH = filepath.ToSlash(GOPATH)
	// server := GOPATH + "/src/github.com/kinghunter58/server-templ"
	if err := os.Mkdir(projectName, os.ModePerm); err != nil {
		return err
	}
	go ngNew(projectName, ngChan)

	cmd := exec.Command("git", "clone", "http://github.com/kinghunter58/server-templ", projectName)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := os.RemoveAll(projectName + "/.git/"); err != nil {
		return err
	}

	//Create data for config.json
	u := getUsername()
	configData := config{
		Dir:     "./dist",
		Author:  u,
		Version: "0.0.1",
		CORS:    "http://localhost:4200",
		DBURL:   "localhost:27017",
		DB:      projectName,
	}

	j, err := json.MarshalIndent(configData, "", "\t")
	if err != nil {
		return errCreatingConfig
	}

	//Create the config.json
	_, err = createAndWrite(projectName+"/magconfig.json", j, os.ModePerm)
	if err != nil {
		return err
	}
	if e := <-ngChan; e != nil {
		return e
	}
	return nil
}

func ngNew(projectName string, c1 chan error) {
	cmd := exec.Command("ng", "new", projectName)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Dir = projectName
	fmt.Println("creating angular project")
	err := cmd.Run()
	if err != nil {
		c1 <- errNgNew
		return
	}
	err = os.Rename(projectName+"/"+projectName, projectName+"/angular")
	if err != nil {
		c1 <- errNgRename
		return
	}
	c1 <- nil
}

func copyFile(from, to string) error {
	data, err := ioutil.ReadFile(from)
	if err != nil {
		return errReadTemplate
	}
	_, err = createAndWrite(to, data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func createAndWrite(name string, data []byte, mode os.FileMode) (*os.File, error) {
	file, err := os.Create(name)
	if err != nil {
		fmt.Println(err)
		return nil, errCreatingFile(name)
	}
	err = ioutil.WriteFile(name, data, mode)
	if err != nil {
		return nil, errWritingFile(name)
	}
	return file, nil
}

func getUsername() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	if strings.Contains(u.Username, "\\") {
		s := strings.Split(u.Username, "\\")
		return s[len(s)-1]
	}
	return u.Username
}

func getGOPATH() string {
	gp := os.Getenv("GOPATH")
	if gp == "" {
		gp = build.Default.GOPATH
	}
	return gp
}
