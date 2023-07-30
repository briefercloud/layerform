package main

import (
	"fmt"
	"os"
	"os/exec"
)

type PlanOutput struct {
    Path string
}

func (p *PlanOutput) Cleanup() {
    os.Remove(p.Path)
}

func main() {
    folderPath := os.Args[1]

	plan, err := plan(folderPath)
    if err != nil {
        fmt.Printf("Error running Terraform plan command: %v\n", err)
        os.Exit(1)
    }

	apply(folderPath, plan)
}

func plan(folderPath string) (*PlanOutput, error) {
    tmpFile, err := os.CreateTemp("", "terraform-plan-*.tfplan")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("terraform", "plan", "-out="+tmpFile.Name())
	cmd.Dir = folderPath

	err = cmd.Run()
	if err != nil {
		return nil, err
	}

    return &PlanOutput{Path: tmpFile.Name()}, nil
}

func apply(folderPath string, plan *PlanOutput) {
    // Use terraform show to ask for confirmation
    cmd := exec.Command("terraform", "show", plan.Path)
    cmd.Dir = folderPath
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Run()

	fmt.Println("Do you want to apply the changes? (yes/no)")

	var input string
	fmt.Scanln(&input)

	if input == "yes" {
		cmd := exec.Command("terraform", "apply", plan.Path)
		cmd.Dir = folderPath

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Env = append(os.Environ(), "TF_CLI_ARGS=-auto-approve")

		if err := cmd.Run(); err != nil {
			fmt.Printf("Error running Terraform apply command: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Aborted.")
	}
}
