package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
)

// Clones the repository into the given dir, just as a normal git clone does
func cloneRepository(repoUrl string, dir string) error {

	// Cloned the repository onto my desktop under Devops-Node folder
	_, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:      repoUrl,
		Progress: os.Stdout,
	})

	if err != nil {
		log.Fatal(err)
	}

	return err
}

// function to fetch the commit history
func getCommitHistory(dirName string) ([]string, error) {

	// creates a command to execute the Git command-line tool
	cmd := exec.Command("git", "-C", dirName, "rev-list", "HEAD")

	output, err := cmd.Output()
	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	return commits, err
}

// function to fetch all of the branches
func getAllBranches(dirName string) ([]string, error) {
	cmd := exec.Command("git", "-C", dirName, "branch", "-r", "--format", "%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	branches := strings.Split(strings.TrimSpace(string(output)), "\n")
	var newBranch []string
	for _, val := range branches {
		parts := strings.Split(val, "/")
		if len(parts) > 0 {
			newBranch = append(newBranch, parts[len(parts)-1])
		}
	}
	return newBranch, nil
}

func switchBranch(dirName, branchName string) error {
	cmd := exec.Command("git", "-C", dirName, "checkout", branchName)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error switching branch: %w", err)
	}
	return nil
}

// Gets the content of a file from a specific commit
func getFileContentFromCommit(dirName, commitHash, filePath string) (string, error) {
	cmd := exec.Command("git", "-C", dirName, "show", fmt.Sprintf("%s:%s", commitHash, filePath))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func scanFileWithRegex(fileContent string) ([]string, error) {
	r := regexp.MustCompile(`(?m)(?i)AKIA[0-9A-Z]{16}\s+\S{40}|AWS[0-9A-Z]{38}\s+?\S{40}`)

	matches := r.FindAllString(fileContent, -1)
	var matchArr []string
	for _, match := range matches {
		matchArr = regexp.MustCompile(`[^\S]+`).Split(match, 2)
	}
	return matchArr, nil
}

// Lists all files and directories in a given commit
func listFilesInCommit(dirName string, commitHash string) ([]string, error) {
	cmd := exec.Command("git", "-C", dirName, "ls-tree", "-r", commitHash)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) > 3 {
			files = append(files, parts[3]) // the file path
		}
	}

	return files, nil
}

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go [directory path]")
	}

	repoUrl := os.Args[1]

	dir, e := os.MkdirTemp("", "example")
	if e != nil {
		log.Fatal(e)
	}

	// invoke method to clone the repository locally
	err := cloneRepository(repoUrl, dir)
	if err != nil {
		fmt.Println("Error cloning the repository")
	}

	branches, err := getAllBranches(dir)

	if err != nil {
		fmt.Println("Error getting all the branches")
	}

	// Below loop prints the commits that are present in their respective branches
	for _, val := range branches {
		switchBranch(dir, val)
		commits, _ := getCommitHistory(dir)
		for _, commit := range commits {
			files, _ := listFilesInCommit(dir, commit)
			for _, file := range files {
				fileContent, err := getFileContentFromCommit(dir, commit, file)
				if err != nil {
					fmt.Println("Error getting file content for", file, "in commit", commit)
					continue
				}

				matches, err := scanFileWithRegex(fileContent)
				if err != nil {
					fmt.Println("Error scanning file", file, "in commit", commit)
					continue
				}

				if len(matches) > 0 {
					fmt.Printf("Matches in %s for commit %s on branch %s:\n", file, commit, val)
					fmt.Println("Access Key: ", matches[0])
					fmt.Println("Secret Token: ", matches[1])
				}
			}
		}

	}

	// clean up temp dir
	defer os.RemoveAll(dir)

}
