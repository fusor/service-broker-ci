package ci

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// TODO: Consider renaming to getRepoAddr
func getScriptAddr(repoScriptAndArgs string, repo string, dir string) (string, string) {
	var script, args string

	// TODO: Split this function into pieces. It has mutiple uses:
	//    - check for local template
	//    - check for valid tempalte
	//    - render template url

	// Split rthallisey/service-broker-ci/wait-for-resource.sh create mediawiki
	//
	// addr="rthallisey/service-broker-ci/wait-for-resource.sh"
	// args="create mediawiki"
	//

	// Using a local file
	if repo == "" {
		items := strings.Split(repoScriptAndArgs, " ")
		script = items[0]
		if dir == "template" && len(items) > 1 {
			fmt.Println("Error: %s should only be the name of the apb.")
			return "", ""
		} else if dir == "template" {
			return fmt.Sprintf("templates/%s.yaml", script), args
		}

		if len(items) > 1 {
			args = strings.Join(items[1:len(items)], " ")
		}
		return script, args
	} else {
		// Using a Git Repo
		script, args = getScriptAndArgs(repo, repoScriptAndArgs, dir)
		if dir == "template" {
			return fmt.Sprintf("%s/%s/%s/templates/%s.yaml", BaseURL, repo, Branch, script), args
		} else if dir == "script" {
			return fmt.Sprintf("%s/%s/%s/%s", BaseURL, repo, Branch, script), args
		}
		return "", ""
	}
}

func resolveGitRepo(repo string) (string, error) {
	var validRepo string

	// Loop through each string in a git repo and combine them to test
	// for a valid git repo. If there is no valid repo found, look locally
	// for the file.
	//
	// `curl https://github.com/fake-git-user/fake-git-repo`  - FAIL
	// `curl https://github.com/fusor/service-broker-ci` - PASS
	//
	addr := strings.Split(repo, "/")
	if len(addr) >= 2 {
		// A git repo's address is always the first two items
		//     rthallisey/service-broker-ci/...
		baseRepo := addr[0:2]
		gitURL := []string{"https://github.com"}
		for count, _ := range addr {
			// Combine 0...N items to form the url for testing
			validRepo = strings.Join(addr[0:count], "/")
			if validRepo == "" {
				validRepo = strings.Join(baseRepo, "/")
			}

			// Combine: 'https://github.com' + '/' + 'rthallisey/service-broker-ci'
			gitURL = append(gitURL, validRepo)
			validRepo = strings.Join(gitURL, "/")

			req, _ := http.Get(validRepo)
			defer req.Body.Close()

			if req.StatusCode == http.StatusOK {
				validRepo = strings.Split(validRepo, "https://github.com/")[1]
				fmt.Printf("REPO: %s\n", validRepo)
				break
			}
		}
	}
	if strings.Contains(validRepo, "https://github.com") {
		return "", errors.New(fmt.Sprintf("Invalid git repo %s", validRepo))
	}
	return validRepo, nil
}

func getScriptAndArgs(repo string, repoScriptAndArgs string, dir string) (string, string) {
	var s []string

	if repo == "" || repoScriptAndArgs == "" {
		return repo, repoScriptAndArgs
	}
	// Split 'openshift/ansible-service-broker' and
	// '/scripts/broker-ci/wait-for-resource.sh create mediawiki'
	a := strings.Split(repoScriptAndArgs, repo)

	if len(a) <= 1 {
		panic(fmt.Sprintf("Repo: %s. ScriptAndArgs: %s. Config.yaml has: %s. Splitting repo from args failed.", repo, a, repoScriptAndArgs))
	}
	scriptAndArgs := a[1]
	if repoScriptAndArgs == repo || scriptAndArgs == " " {
		// Return the resource name
		r := strings.Split(repo, "/")
		return r[len(r)-1], ""
	}

	// Templates have no args
	if dir == "template" {
		return scriptAndArgs, ""
	}

	s = strings.Split(scriptAndArgs, " ")
	// Script with no args
	if len(s) == 1 {
		return s[0], ""
	}

	script := s[0]
	fmt.Printf("SCRIPT: %s\n", script)

	listArgs := s[1:len(s)]
	args := strings.Join(listArgs, " ")
	fmt.Printf("ARGS: %s\n", args)

	return script, args
}

func findBindTarget(repo string, provisioned []string) (string, []string, error) {
	var usedTargets []int
	foundTarget := false
	foundBind := false
	var bindTarget string

	// The config in imperative so order matters
	for count, r := range provisioned {
		// Remove the first Provisioned app that matches the Bind repo
		// and the first Provisioned app that doesn't.

		// The first Provisioned app that doesn't match the bindApp is
		// the bindTarget.
		if r != repo && !foundTarget {
			bindTarget = r
			foundTarget = true
			usedTargets = append(usedTargets, count)
		}

		// The first Provisioned app that matches the Bind repo is the
		// bind app.
		if r == repo && !foundBind {
			foundBind = true
			usedTargets = append(usedTargets, count)
		}

		if foundBind && foundTarget {
			cleanupUsedTargets(usedTargets, provisioned)
			return bindTarget, provisioned, nil
		}
	}

	return "", provisioned, errors.New("Failed to find a provisioned bind target and bind app")
}

func cleanupUsedTargets(usedTargets []int, provisioned []string) []string {
	for _, d := range usedTargets {
		// Cleanup the bindTarget and the bind app
		if len(provisioned) == 1 {
			provisioned = provisioned[:0]
		} else {
			provisioned = append(provisioned[:d], provisioned[d+1:]...)
		}
	}
	return provisioned
}
