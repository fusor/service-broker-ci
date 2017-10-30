package action

import (
	"fmt"
	"strings"
)

func Bind(repo string, cmd string, target string) error {
	template, err := getTemplate(repo, "template")
	if err != nil {
		return err
	}

	fmt.Printf("Running: %s create -f %s\n", cmd, template)
	args := fmt.Sprintf("create -f %s", template)
	output, err := RunCommand(cmd, args)

	fmt.Println(string(output))
	if err != nil {
		return err
	}

	// Wait for binding resource creation
	err = waitUntilReady(repo)
	if err != nil {
		return err
	}

	targetName := resourceName(target)
	instanceName, err := RunCommand("oc", fmt.Sprintf("get -f /tmp/%s.yaml -o jsonpath='{ .metadata.name }'", targetName))
	if err != nil {
		fmt.Println(string(instanceName))
		return err
	}
	fmt.Printf("Using Instance Name: %s\n", instanceName)

	err = waitUntilResourceReady("mediawiki-postgresql-binding", "secret")
	if err != nil {
		return err
	}

	// Get the name of the secret
	repoName := resourceName(repo)
	secretName, err := RunCommand("oc", fmt.Sprintf("get -f /tmp/%s -o jsonpath='{ .spec.secretName }'", repoName))
	if err != nil {
		fmt.Println(secretName)
		return err
	}

	// Gather bind data from secret
	bindData, err := RunCommand("oc", fmt.Sprintf("get secret %s -o jsonpath='{.data}'", secretName))
	if err != nil {
		fmt.Println(bindData)
		return err
	}
	data := strings.TrimPrefix(string(bindData), "map[")
	data = strings.TrimSuffix(data, "]")

	// "DB_PASSWORD=YWRtaW4= DB_PORT=NTQzMg== DB_NAME=YWRtaW4="
	data = strings.Replace(data, ":", "=", -1)

	fmt.Printf("Looking for a Deployment with the SAME name used in your ServiceInstance: %s\n", instanceName)
	// Inject bind data into the pod
	// oc env dc mediawiki123 DB_HOST=$DB_HOST DB_NAME=$DB_NAME
	output, err = RunCommand("oc", fmt.Sprintf("env dc %s %s", instanceName, data))
	fmt.Println(string(output))
	if err == nil {
		return nil
	}

	return nil
}
